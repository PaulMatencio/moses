package bns

//  Asynchronously Update object

import (
	"fmt"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"time"

	// hostpool "github.com/bitly/go-hostpool"
)

func AsyncHttpUpdateBlob(bnsRequest *HttpRequest, url string, buf []byte, header map[string]string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	// create a sproxyd request response
	sproxydResponse := &sproxyd.HttpResponse{}

	// create a sproxyd request structure
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		Client:    bnsRequest.Client,
		Path:      url,
		ReqHeader: header,
	}

	// asynchronously write the object
	go func(url string) {
		var err error
		var resp *http.Response
		resp, err = sproxyd.Updobject(&sproxydRequest, buf)
		if resp != nil {
			resp.Body.Close()
		}
		// the caller bns must issue resp.Body.Close()
		//
		ch <- &sproxyd.HttpResponse{url, resp, nil, err}
	}(url)

	for {
		select {
		case sproxydResponse = <-ch:
			return sproxydResponse
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf("w")
		}
	}

	return sproxydResponse
}

// func AsyncHttpUpdateBlobs(hspool hostpool.HostPool, urls []string, bufa [][]byte, headera []map[string]string) []*sproxyd.HttpResponse {

func AsyncHttpUpdateBlobs(bnsResponses []BnsResponse) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponses := []*sproxyd.HttpResponse{}
	client := &http.Client{}
	treq := 0
	for k, _ := range bnsResponses {
		treq += 1
		url := sproxyd.TargetEnv + "/" + bnsResponses[k].BnsId + "/" + bnsResponses[k].PageNumber
		go func(url string) {
			var err error
			var resp *http.Response

			sproxydRequest := sproxyd.HttpRequest{}
			sproxydRequest.ReqHeader = map[string]string{
				"Usermd": bnsResponses[k].Usermd,
			}
			sproxydRequest.Hspool = sproxyd.TargetHP
			sproxydRequest.Client = client
			sproxydRequest.Path = url
			resp, err = sproxyd.Updobject(&sproxydRequest, bnsResponses[k].Image)
			if resp != nil {
				resp.Body.Close()
			}
			if !sproxyd.Test {
				defer resp.Body.Close()
			} else {
				time.Sleep(1 * time.Millisecond)
			}
			ch <- &sproxyd.HttpResponse{sproxydRequest.Path, resp, nil, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			sproxydResponses = append(sproxydResponses, r)
			if len(sproxydResponses) == treq {
				return sproxydResponses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf("w")
		}
	}
	return sproxydResponses
}

/*
func AsyncHttpUpdateBlobsTest(bnsResponses []BnsResponse) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponses := []*sproxyd.HttpResponse{}
	client := &http.Client{}
	treq := 0
	for _, bnsResponse := range bnsResponses {

		treq += 1
		go func(bnsResponse *BnsResponse) {
			var err error
			var resp *http.Response

			sproxydRequest := sproxyd.HttpRequest{}
			header := map[string]string{
				"Usermd": bnsResponse.Usermd,
			}
			sproxydRequest.Hspool = sproxyd.TargetHP
			sproxydRequest.Client = client
			sproxydRequest.Path = sproxyd.TargetEnv + "/" + bnsResponse.BnsId + "/" + bnsResponse.PageNumber
			sproxydRequest.ReqHeader = header
			resp, err = sproxyd.Updobject(&sproxydRequest, bnsResponse.Image,test)
			if resp != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{sproxydRequest.Path, resp, nil, err}
		}(&bnsResponse)
	}
	for {
		select {
		case r := <-ch:
			sproxydResponses = append(sproxydResponses, r)
			if len(sproxydResponses) == treq {
				return sproxydResponses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf("w")
		}
	}
	return sproxydResponses
}

*/
