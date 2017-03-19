package bns

// Asynchronouly PUT Object

import (
	"fmt"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"time"
)

func AsyncHttpPutBlob(bnsRequest *HttpRequest, url string, buf []byte, header map[string]string) *sproxyd.HttpResponse {

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
		// in test test mode , resp and err are nil
		resp, err = sproxyd.Putobject(&sproxydRequest, buf)
		if resp != nil {
			resp.Body.Close()
		}
		// the caller must close resp.Body.Close()
		// bns should close it ( in buildBnsResponse)
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

/*
func AsyncHttpPutBlobTest(bnsRequest *HttpRequest, url string, buf []byte, header map[string]string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponse := &sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		Client:    bnsRequest.Client,
		Path:      url,
		ReqHeader: header,
	}

	go func(url string) {
		resp, err := sproxyd.Putobject(&sproxydRequest, buf)

		time.Sleep(1 * time.Millisecond)
		ch <- &sproxyd.HttpResponse{url, resp, nil, err}

	}(url)

	for {
		select {
		case sproxydResponse = <-ch:
			return sproxydResponse
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}

	return sproxydResponse
}

*/
func AsyncHttpCopyBlobs(bnsResponses []BnsResponse) []*sproxyd.HttpResponse {
	// Put objects
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
			resp, err = sproxyd.Putobject(&sproxydRequest, bnsResponse.Image)

			if resp != nil {
				resp.Body.Close()
			}
			// the caller bns must close the Body after having consumed it
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

/*
func AsyncHttpCopyBlobsTest(bnsResponses []BnsResponse) []*sproxyd.HttpResponse {

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
			resp, err = sproxyd.PutobjectTest(&sproxydRequest, bnsResponse.Image)
			time.Sleep(1 * time.Millisecond)
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
