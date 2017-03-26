package bns

import (
	"fmt"
	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	goLog "moses/user/goLog"
	"net/http"
	"time"
)

// new function
func GetMetadata(bnsRequest *HttpRequest, url string) ([]byte, error, int) {
	// client := &http.Client{}
	// getHeader := map[string]string{}

	var (
		usermd []byte
		resp   *http.Response
		err    error = error(nil)
	)

	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		Client:    bnsRequest.Client,
		Path:      url,
		ReqHeader: map[string]string{},
	}

	if resp, err = sproxyd.GetMetadata(&sproxydRequest); err == nil {
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
	return usermd, err, resp.StatusCode
}

//new  function
func GetEncodedMetadata(bnsRequest *HttpRequest, url string) (string, error, int) {

	getHeader := map[string]string{}
	var (
		encoded_usermd string
		resp           *http.Response
	)
	err := error(nil)
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		Client:    bnsRequest.Client,
		Path:      url,
		ReqHeader: getHeader,
	}

	if resp, err = sproxyd.GetMetadata(&sproxydRequest); err == nil {
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
	return encoded_usermd, err, resp.StatusCode
}

func AsyncHttpGetMetadatas(bnsRequest *HttpRequest, getHeader map[string]string) []*sproxyd.HttpResponse {
	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponses := []*sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		ReqHeader: getHeader,
	}
	treq := 0
	fmt.Printf("\n")
	sproxydRequest.Client = &http.Client{}
	for _, url := range bnsRequest.Urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}

		go func(url string) {
			sproxydRequest.Path = url
			resp, err := sproxyd.GetMetadata(&sproxydRequest)
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
			sproxydResponses = append(sproxydResponses, r)
			if len(sproxydResponses) == treq {
				return sproxydResponses
			}
		case <-time.After(100 * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return sproxydResponses
}
