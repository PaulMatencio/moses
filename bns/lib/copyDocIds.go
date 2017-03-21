package bns

// a DocId is composed of a TOC and pages

import (
	"encoding/json"
	"errors"
	"fmt"

	// "io"
	// "io/ioutil"

	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	// file "moses/user/files/lib"
	goLog "moses/user/goLog"
	"net/http"
	"os"
	// "os/user"
	// "path"
	"strconv"
	"time"

	// hostpool "github.com/bitly/go-hostpool"
)

func AsyncCopyDocIds(pns []string, srcEnv string, targetEnv string) []*CopyResponse {

	SetCPU("100%")
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	start := time.Now()
	var (
		err           error
		encoded_docmd string
		docmd         []byte
	)
	media := "binary"
	if len(srcEnv) == 0 {
		srcEnv = sproxyd.Env
	}
	if len(targetEnv) == 0 {
		targetEnv = sproxyd.TargetEnv
	}

	ch := make(chan *CopyResponse)
	copyResponses := []*CopyResponse{}
	treq := 0

	//  launch concurrent requets
	for _, pn := range pns {

		srcPath := srcEnv + "/" + pn
		dstPath := targetEnv + "/" + pn
		srcUrl := srcPath
		dstUrl := dstPath

		bnsRequest := HttpRequest{
			Hspool: sproxyd.HP, // source sproxyd servers IP address and ports
			Client: &http.Client{},
			Media:  media,
		}

		go func(srcUrl string, dstUrl string) {

			treq++
			num200 := 0
			num := 0
			// Get the PN metadata ( Table of Content)
			if encoded_docmd, err = GetEncodedMetadata(&bnsRequest, srcUrl); err == nil {
				if len(encoded_docmd) > 0 {
					if docmd, err = base64.Decode64(encoded_docmd); err != nil {
						goLog.Error.Println(err)
						ch <- &CopyResponse{err, pn, num, num200}
						return
					}
				} else {
					err = errors.New("Metadata is missing for " + srcPath)
					goLog.Error.Println(err)
					ch <- &CopyResponse{err, pn, num, num200}
					return
				}
			} else {
				goLog.Error.Println(err)
				ch <- &CopyResponse{err, pn, num, num200}
				return
			}

			// The PN meta data is valid
			// convert the PN  metadata into a go structure
			docmeta := DocumentMetadata{}
			if err := json.Unmarshal(docmd, &docmeta); err != nil {
				goLog.Error.Println(docmeta, err)
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
			var duration time.Duration

			num = docmeta.TotalPage
			urls := make([]string, num, num)
			// dstUrls := make([]string, num, num)
			getHeader := map[string]string{}
			getHeader["Content-Type"] = "application/binary"
			for i := 0; i < num; i++ {
				urls[i] = srcPath + "/p" + strconv.Itoa(i+1)
				// dstUrls[i] = dstPath + "/p" + strconv.Itoa(i+1)
			}
			bnsRequest.Urls = urls
			bnsRequest.Hspool = sproxyd.HP // Set source sproxyd servers
			bnsRequest.Client = &http.Client{}
			// Get all the pages from the source Ring
			sproxyResponses := AsyncHttpGetBlobs(&bnsRequest, getHeader)
			// Build a response array of BnsResponse array to be used to update the pages  of  destination sproxyd servers
			bnsResponses := make([]BnsResponse, num, num)

			for i, sproxydResponse := range sproxyResponses {
				if err := sproxydResponse.Err; err == nil { //
					resp := sproxydResponse.Response                                        /* http response */ // http response
					body := *sproxydResponse.Body                                           // http response                                                          /* copy of the body */ // http body response
					bnsResponse := BuildBnsResponse(resp, getHeader["Content-Type"], &body) // bnsResponse is a Go structure
					bnsResponses[i] = bnsResponse
					resp.Body.Close()
				}
			}
			duration = time.Since(start)
			fmt.Println("Get elapsed time:", duration)
			goLog.Info.Println("Get elapsed time:", duration)

			// var sproxydResponses []*sproxyd.HttpResponse
			//   new &http.Client{}  and hosts pool are set to the target by the AsyncHttpCopyBlobs
			//  			sproxyd.TargetHP
			sproxydResponses := AsyncHttpCopyBlobs(bnsResponses)

			if !sproxyd.Test {
				for _, v := range sproxydResponses {
					resp := v.Response
					url := v.Url
					switch resp.StatusCode {
					case 200:
						goLog.Trace.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Key"])
						num200++
					case 412:
						goLog.Warning.Println(hostname, pid, url, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")

					case 422:
						goLog.Error.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Status"])
					default:
						goLog.Warning.Println(hostname, pid, url, resp.Status)
					}
					// close all the connection
					resp.Body.Close()
				}

				if num200 < num {
					goLog.Warning.Println("\nPublication id:", pn, num, " Pages in;", num200, " Pages out")
					err = errors.New("Pages out < Pages in")
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
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf("w")
		}
	}
	return copyResponses
}
