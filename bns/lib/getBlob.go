package bns

//  Get full object
//  Get range of byte ( subpart of an object)

import (
	"fmt"
	"io/ioutil"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"time"

	// hostpool "github.com/bitly/go-hostpool"
)

// new function

func GetBlob(sproxydRequest *sproxyd.HttpRequest) (*http.Response, error) {
	sproxydRequest.ReqHeader = map[string]string{}
	return sproxyd.Getobject(sproxydRequest)
}

// new function

func AsyncHttpGetBlob(bnsRequest *HttpRequest, getHeader map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	sproxydRequest := sproxyd.HttpRequest{
		Hspool:    bnsRequest.Hspool,
		ReqHeader: getHeader,
	}

	treq := 0
	fmt.Printf("\n")
	for _, url := range bnsRequest.Urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}

		// client := &http.Client{} // one connection for all requests
		sproxydRequest.Client = &http.Client{}
		// sproxydRequest.Path = url
		go func(url string) {
			sproxydRequest.Path = url
			resp, err := sproxyd.Getobject(&sproxydRequest)
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			} else {

				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, &body, err}

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
