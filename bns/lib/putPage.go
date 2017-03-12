package bns

import (
	"bytes"
	"fmt"
	sproxyd "moses/sproxyd/lib"
	goLog "moses/user/goLog"
	"net/http"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

//func PutPage(client *http.Client, path string, img *bytes.Buffer, usermd []byte) (error, time.Duration) {
func PutPage(hspool hostpool.HostPool, client *http.Client, path string, img *bytes.Buffer, putheader map[string]string) (error, time.Duration) {
	/*
		putheader := map[string]string{
			"Usermd":       base64.Encode64(usermd),
			"Content-Type": "image/tiff",
		}
	*/
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration
	if resp, err = sproxyd.PutObject(hspool, client, path, img.Bytes(), putheader); err != nil {
		goLog.Error.Println(err)
	} else {
		elapse = time.Since(start)
		switch resp.StatusCode {
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist", elapse)
		case 422:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		case 200:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, "Elapse:", elapse)
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status, "Elapse:", elapse)
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err, elapse
}

func AsyncHttpPuts(hspool hostpool.HostPool, urls []string, bufa [][]byte, headera []map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0
	client := &http.Client{} // one client

	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			// clientw := &http.Client{}
			resp, err = sproxyd.PutObject(hspool, client, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, nil, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpPut2s(hspool hostpool.HostPool, urls []string, bufa [][]byte, bufb [][]byte, headera []map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0

	client := &http.Client{} // one connection for all request

	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			// clientw := &http.Client{}
			resp, err = sproxyd.PutObject(hspool, client, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, nil, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpCopy(hspool hostpool.HostPool, client *http.Client, url string, buf []byte, header map[string]string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	response := &sproxyd.HttpResponse{}

	go func(url string) {
		var err error
		var resp *http.Response
		// clientw := &http.Client{}
		resp, err = sproxyd.PutObject(hspool, client, url, buf, header)
		if resp != nil {
			resp.Body.Close()
		}
		ch <- &sproxyd.HttpResponse{url, resp, nil, err}
	}(url)

	for {
		select {
		case r := <-ch:
			response = r
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}

	return response
}

func AsyncHttpCopyBlob(bnsRequest *HttpRequest, buf []byte, header map[string]string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	spoxydResponse := &sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		Client:    bnsRequest.Client,
		Path:      bnsRequest.Path,
		ReqHeader: header,
	}
	url := bnsRequest.Path
	go func(url string) {
		var err error
		var resp *http.Response
		// clientw := &http.Client{}
		resp, err = sproxyd.Putobject(&sproxydRequest, buf)
		if resp != nil {
			resp.Body.Close()
		}
		ch <- &sproxyd.HttpResponse{url, resp, nil, err}
	}(url)

	for {
		select {
		case r := <-ch:
			spoxydResponse = r
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}

	return spoxydResponse
}

func AsyncHttpCopyTest(hspool hostpool.HostPool, client *http.Client, url string, buf []byte, header map[string]string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	response := &sproxyd.HttpResponse{}

	go func(url string) {
		_, _ = sproxyd.PutObjectTest(hspool, client, url, buf, header)
		ch <- &sproxyd.HttpResponse{url, nil, nil, nil}
	}(url)

	for {
		select {
		case r := <-ch:
			response = r
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}

	return response
}

func AsyncHttpCopyBlobTest(bnsRequest *HttpRequest, buf []byte, header map[string]string) *sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	spoxydResponse := &sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		Path:      bnsRequest.Path,
		ReqHeader: header,
	}
	url := sproxydRequest.Path

	go func(url string) {
		// _, _ = sproxyd.PutObjectTest(hspool, client, url, buf, header)
		_, _ = sproxyd.PutobjectTest(&sproxydRequest, buf)
		ch <- &sproxyd.HttpResponse{url, nil, nil, nil}
	}(url)

	for {
		select {
		case r := <-ch:
			spoxydResponse = r
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}

	return spoxydResponse
}
