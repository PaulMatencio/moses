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

//func UpdatePage(client *http.Client, path string, img *bytes.Buffer, usermd []byte) (error, time.Duration) {
func UpdatePage(hspool hostpool.HostPool, client *http.Client, path string, img *bytes.Buffer, putheader map[string]string) (error, time.Duration) {
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration
	// defer resp.Body.Close()
	if resp, err = sproxyd.UpdObject(hspool, client, path, img.Bytes(), putheader); err != nil {
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
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		default:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, elapse)
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err, elapse

}

func Updatepage(hspool hostpool.HostPool, client *http.Client, path string, img *bytes.Buffer, putheader map[string]string) (error, time.Duration) {
	/*
		putheader := map[string]string{
			"Usermd":       encoded_usermd := base64.Encode64(usermd),
			"Content-Type": "image/tiff",
		}
	*/
	// url:= base_url+string(pub)+docid+string(kc)+"_"+string(pagenum)
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration

	if resp, err = sproxyd.UpdObject(hspool, client, path, img.Bytes(), putheader); err != nil {
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
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		default:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, elapse)
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err, elapse

}

func AsyncHttpUpdates(hspool hostpool.HostPool, urls []string, bufa [][]byte, headera []map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0
	client := &http.Client{}
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
			resp, err = sproxyd.UpdObject(hspool, client, url, bufa[k], headera[k])
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
