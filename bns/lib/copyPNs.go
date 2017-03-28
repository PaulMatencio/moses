package bns

// a DocId is composed of a TOC and pages

import (
	// "bytes"
	"encoding/json"
	"errors"
	"fmt"
	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	goLog "moses/user/goLog"
	"net/http"
	"os"
	"strconv"
	"time"
)

func AsyncCopyPns(pns []string, srcEnv string, targetEnv string) []*CopyResponse {
	var (
		duration      time.Duration
		media         string = "binary"
		treq          int    = 0
		ch                   = make(chan *CopyResponse)
		copyResponses        = []*CopyResponse{}
		pid                  = os.Getpid()
		hostname, _          = os.Hostname()
		start                = time.Now()
	)

	SetCPU("100%")

	if len(srcEnv) == 0 {
		srcEnv = sproxyd.Env
	}
	if len(targetEnv) == 0 {
		targetEnv = sproxyd.TargetEnv
	}

	//  launch concurrent requets
	for _, pn := range pns {

		var (
			srcPath    = srcEnv + "/" + pn
			dstPath    = targetEnv + "/" + pn
			srcUrl     = srcPath
			dstUrl     = dstPath
			bnsRequest = HttpRequest{
				Hspool: sproxyd.HP, // source sproxyd servers IP address and ports
				Client: &http.Client{
					Timeout:   sproxyd.WriteTimeout,
					Transport: sproxyd.Transport,
				},
				Media: media,
			}
		)
		go func(srcUrl string, dstUrl string) {
			treq++
			var (
				docmd                                         []byte
				encoded_docmd                                 string
				err                                           error
				statusCode                                    int
				num, num200, num412, num422, num404, numOther int = 0, 0, 0, 0, 0, 0
			)
			// Get the  Table of Content
			if encoded_docmd, err, statusCode = GetEncodedMetadata(&bnsRequest, srcUrl); err == nil {
				if len(encoded_docmd) > 0 {

					if docmd, err = base64.Decode64(encoded_docmd); err != nil {
						goLog.Error.Println(err) // Invalid meta data
						ch <- &CopyResponse{err, pn, num, num200}
						return
					}
				} else {
					if statusCode == 404 {
						err = errors.New("Document " + srcPath + " not found")
					} else {
						err = errors.New("Metadata is missing for " + srcPath)
					}
					goLog.Warning.Println(err)
					ch <- &CopyResponse{err, pn, num, num200}
					return
				}
			} else {
				goLog.Error.Println(err)
				ch <- &CopyResponse{err, pn, num, num200}
				return
			}
			// convert the PN  metadata into a go structure
			docmeta := DocumentMetadata{}
			// docmd := bytes.Replace(docmd1, []byte(`"\n  "`), []byte(`{}`), -1)

			if err := json.Unmarshal(docmd, &docmeta); err != nil {
				goLog.Error.Println("Document metadata is invalid ", srcUrl, err)
				goLog.Error.Println(string(docmd), docmeta)
				ch <- &CopyResponse{err, pn, num, num200}
				return
			} else {
				header := map[string]string{
					"Usermd": encoded_docmd,
				}
				buf0 := make([]byte, 0)
				bnsRequest.Hspool = sproxyd.TargetHP // Set Target sproxyd servers
				// Copy the document metadata to the destination buffer size = 0 byte
				// we could  update the meta data : TODO
				CopyBlob(&bnsRequest, dstUrl, buf0, header)
			}

			if num, err = docmeta.GetPageNumber(); err != nil {
				ch <- &CopyResponse{err, pn, num, num200}
			}

			var (
				urls      = make([]string, num, num)
				getHeader = map[string]string{
					"Content-Type": "application/binary",
				}
			)

			for i := 0; i < num; i++ {
				urls[i] = srcPath + "/p" + strconv.Itoa(i+1)
			}
			bnsRequest.Urls = urls
			bnsRequest.Hspool = sproxyd.HP // Set source sproxyd servers
			bnsRequest.Client = &http.Client{
				Timeout:   sproxyd.ReadTimeout,
				Transport: sproxyd.Transport,
			}
			// Get all the pages from the source Ring
			sproxydResponses := AsyncHttpGetBlobs(&bnsRequest, getHeader)
			// Build a response array of BnsResponse array to be used to update the pages  of  destination sproxyd servers
			bnsResponses := make([]BnsResponse, num, num)

			for i, sproxydResponse := range sproxydResponses {
				if err := sproxydResponse.Err; err == nil { //
					resp := sproxydResponse.Response                                        /* http response */ // http response
					body := *sproxydResponse.Body                                           // http response
					bnsResponse := BuildBnsResponse(resp, getHeader["Content-Type"], &body) // bnsResponse is a Go structure
					bnsResponses[i] = bnsResponse
					resp.Body.Close() // Close the connection after BuildBnsResponse()
				}
			}
			duration = time.Since(start)
			fmt.Println("Get elapsed time:", duration)
			goLog.Info.Println("Get elapsed time:", duration)

			// var sproxydResponses []*sproxyd.HttpResponse
			//   new &http.Client{}  and hosts pool are set to the target by the AsyncHttpCopyBlobs
			//  			sproxyd.TargetHP
			sproxydResponses = AsyncHttpPutBlobs(bnsResponses)

			if !sproxyd.Test {
				for _, v := range sproxydResponses {
					resp := v.Response
					url := v.Url
					switch resp.StatusCode {
					case 200:
						goLog.Trace.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Key"])
						num200++
					case 404:
						goLog.Trace.Println(hostname, pid, url, resp.Status)
						num404++
					case 412:
						goLog.Warning.Println(hostname, pid, url, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")
						num412++
					case 422:
						goLog.Error.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Status"])
						num422++
					default:
						goLog.Warning.Println(hostname, pid, url, resp.Status)
						numOther++
					}
					// close all the connection
					resp.Body.Close()
				}
				if num200 < num {
					goLog.Warning.Printf("\nHost name:%s,Pid:%d,Publication:%s,Ins:%d,Outs:%d,Notfound:%d,Existed:%d,Other:%d", hostname, pid, pn, num, num200, num404, num412, numOther)
					err = errors.New("Pages outs < Page ins")
				} else {
					goLog.Info.Printf("\nHost name:%s,Pid:%d,Publication:%s,Ins:%d,Outs:%d,Notfound:%d,Existed:%d,Other:%d", hostname, pid, pn, num, num200, num404, num412, numOther)
				}

			}
			ch <- &CopyResponse{err, pn, num, num200}

		}(srcUrl, dstUrl)
	}

	//  Loop wait for results
	for {
		select {
		case r := <-ch:
			copyResponses = append(copyResponses, r)
			if len(copyResponses) == treq {
				return copyResponses
			}
		case <-time.After(sproxyd.CopyTimeout * time.Millisecond):
			fmt.Printf("c")
		}
	}
	return copyResponses
}
