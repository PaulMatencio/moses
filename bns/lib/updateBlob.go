package bns

//  Asynchronously Update object

import (
	"fmt"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"time"
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

func AsyncHttpUpdateBlobTest(bnsRequest *HttpRequest, url string, buf []byte, header map[string]string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponse := &sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		Client:    bnsRequest.Client,
		Path:      url,
		ReqHeader: header,
	}

	go func(url string) {
		resp, err := sproxyd.UpdobjectTest(&sproxydRequest, buf)
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
