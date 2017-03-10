package bns

import (
	"fmt"
	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	goLog "moses/user/goLog"
	"net/http"
	"time"
)

func GetMetadata(client *http.Client, path string) ([]byte, error) {
	// client := &http.Client{}
	getHeader := map[string]string{}
	var usermd []byte
	var resp *http.Response
	err := error(nil)

	if resp, err = sproxyd.GetMetadata(client, path, getHeader); err == nil {
		switch resp.StatusCode {
		case 200:
			encoded_usermd := resp.Header["X-Scal-Usermd"]
			usermd, err = base64.Decode64(encoded_usermd[0])
		case 404:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status)
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
		case 422:
			goLog.Error.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status)
		}
	}
	/* the resp,Body is closed by sproxyd.getMetadata */
	return usermd, err
}

func GetEncodedMetadata(client *http.Client, path string) (string, error) {
	// client := &http.Client{}
	getHeader := map[string]string{}
	var (
		// usermd         []byte
		encoded_usermd string
		resp           *http.Response
	)
	err := error(nil)

	if resp, err = sproxyd.GetMetadata(client, path, getHeader); err == nil {
		switch resp.StatusCode {
		case 200:
			encoded_usermd = resp.Header["X-Scal-Usermd"][0]
		case 404:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status)
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
		case 422:
			goLog.Error.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status)
		}
	}
	/* the resp,Body is closed by sproxyd.getMetadata */
	return encoded_usermd, err
}

func AsyncHttpGetMetadatas(urls []string, getHeader map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	treq := 0
	fmt.Printf("\n")
	client := &http.Client{}
	for _, url := range urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			// fmt.Printf("Fetching %s \n", url)

			// client := &http.Client{}

			//start := time.Now()
			//var elapse time.Duration
			resp, err := sproxyd.GetMetadata(client, url, getHeader)
			if err != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, nil, err}

		}(url)
	}
	// wait for http response  message
	for {
		select {
		case r := <-ch:
			// fmt.Printf("%s was fetched\n", r.url)
			responses = append(responses, r)
			if len(responses) == treq /*len(urls)*/ {
				return responses
			}
		case <-time.After(100 * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}
