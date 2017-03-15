package bns

// put objects

import (
	//"bytes"
	"fmt"
	sproxyd "moses/sproxyd/lib"
	// goLog "moses/user/goLog"
	"net/http"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

/*
func PutPage(hspool hostpool.HostPool, client *http.Client, path string, img *bytes.Buffer, putheader map[string]string) (error, time.Duration) {

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

*/
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
			fmt.Printf("w")
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
			fmt.Printf("w")
		}
	}
	return responses
}
