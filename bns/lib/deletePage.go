package bns

import (
	"fmt"
	sproxyd "moses/sproxyd/lib"
	goLog "moses/user/goLog"
	"net/http"
	"time"
)

// new delete page function
func DeletePage(bnsRequest *HttpRequest, url string) (error, time.Duration) {

	// deleteHeader := map[string]string{}
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration
	// defer resp.Body.Close()
	sproxydRequest := sproxyd.HttpRequest{
		Hspool: bnsRequest.Hspool,
		Client: bnsRequest.Client,
		Path:   url,
	}

	// if resp, err = sproxyd.DeleteObject(hspool, client, path); err != nil {
	if resp, err = sproxyd.Deleteobject(&sproxydRequest); err != nil {
		goLog.Error.Println(err)
	} else {
		elapse = time.Since(start)
		switch resp.StatusCode {
		case 200:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Key"], elapse)
		case 404:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, " not found", elapse)
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist", elapse)
		case 422:
			goLog.Error.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status, elapse)
		}
		resp.Body.Close()
	}
	return err, elapse
}

func AsyncHttpDeletePages(bnsRequest *HttpRequest, url string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponse := &sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool: bnsRequest.Hspool,
		Client: &http.Client{},
		Path:   url,
	}

	if len(url) == 0 {
		return sproxydResponse
	}

	go func(url string) {
		var err error
		var resp *http.Response
		resp, err = sproxyd.Deleteobject(&sproxydRequest)
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
			fmt.Printf("d")
		}
	}
	return sproxydResponse
}

func AsyncHttpDeletePagesTest(bnsRequest *HttpRequest, url string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponse := &sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool: bnsRequest.Hspool,
		Client: &http.Client{},
		Path:   url,
	}

	if len(url) == 0 {
		return sproxydResponse
	}

	go func(url string) {
		var err error
		var resp *http.Response
		resp, err = sproxyd.DeleteobjectTest(&sproxydRequest)
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
			fmt.Printf("d")
		}
	}
	return sproxydResponse
}
