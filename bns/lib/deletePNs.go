package bns

//  DELETE <targetENV>  PNs on the Target RING
//

import (
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

func AsyncDeletePns(pns []string, targetEnv string) []*CopyResponse {
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	SetCPU("100%")
	var (
		err           error
		encoded_docmd string
		docmd         []byte
		pnPath        string
	)
	media := "binary"

	if len(targetEnv) == 0 {
		targetEnv = sproxyd.TargetEnv
	}

	ch := make(chan *CopyResponse)
	copyResponses := []*CopyResponse{}
	treq := 0

	//  launch concurrent requets
	for _, pn := range pns {
		targetPath := targetEnv + "/" + pn
		targetUrl := targetPath
		//
		//  Read the PN 's metadata  from the source RING
		//  The SOURCE RING may be the same as the DESTINATION RING
		//  Check the config file sproxyd.HP and sproxyd.TargetHP
		//
		bnsRequest := HttpRequest{
			Hspool: sproxyd.TargetHP, // <<<<<<  sproxyd.TargetHP is the Destination RING
			Client: &http.Client{},
			Media:  media,
		}

		goLog.Info.Println("Deleting", pnPath)

		go func(targetUrl string) {

			treq++
			num200 := 0
			num := 0
			// Get the PN metadata ( Table of Content)
			if encoded_docmd, err = GetEncodedMetadata(&bnsRequest, targetUrl); err == nil {
				if len(encoded_docmd) > 0 {
					if docmd, err = base64.Decode64(encoded_docmd); err != nil {
						goLog.Error.Println(err)
						ch <- &CopyResponse{err, pn, num, num200}
						return
					}
				} else {
					err = errors.New("Metadata is missing for " + targetPath)
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
			}

			num = docmeta.TotalPage
			// urls := make([]string, num, num)

			//  DELETE THE DOCUMENT ON THE TARGET ENVIRONMENT
			fmt.Println("len => ", num)
			bnsRequest.Hspool = sproxyd.TargetHP // set target sproxyd servers ( Destination RING)
			bnsRequest.Urls = make([]string, num, num)
			targetPath := targetEnv + "/" + pn
			pnPath = targetPath
			bnsRequest.Client = &http.Client{}
			//  DELETE ALL THE PAGES FIRST
			for i := 0; i < num; i++ {
				bnsRequest.Urls[i] = targetPath + "/p" + strconv.Itoa(i+1)
			}

			sproxydResponses := AsyncHttpDeleteBlobs(&bnsRequest)
			bnsResponses := make([]BnsResponse, num, num)

			for i, sproxydResponse := range sproxydResponses {
				if err := sproxydResponse.Err; err == nil { //
					resp := sproxydResponse.Response               /* http response */ // http response                                                      /* copy of the body */ // http body response
					bnsResponse := BuildBnsResponse(resp, "", nil) // bnsResponse is a Go structure
					bnsResponses[i] = bnsResponse
					resp.Body.Close()
				}
			}
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
					err = errors.New("Not all pages were deleted, Check the Warning/Errors logs for details")
				}
			}
			// Delete the PN metadata when all pages have been deleted
			if num == num200 {
				bnsRequest := HttpRequest{
					Hspool: sproxyd.TargetHP,
					Client: &http.Client{},
					Media:  media,
				}
				if err, statusCode := DeleteBlob(&bnsRequest, pnPath); err != nil {
					goLog.Error.Println("Error deleting PN", pnPath, " Error:", err, "Status Code:", statusCode)
				} else {
					goLog.Info.Println(pnPath, " is deleted")
				}
			}

			ch <- &CopyResponse{err, pn, num, num200}

		}(targetUrl)

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

	/* DELETE THE PN METADATA ONLY WHEN ALL THE PN'S PAGES ARE DELETED
	if *copyResponse.Num == *copyResponse.Num200 {
	bnsRequest := HttpRequest{
		Hspool: sproxyd.TargetHP,
		Client: &http.Client{},
		Media:  media,
	}

	err, _ = DeleteBlob(&bnsRequest, pnPath)
	}
	*/

	return copyResponses
}
