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

func AsyncHttpGetBlobs(bnsRequest *HttpRequest, getHeader map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	sproxydResponses := []*sproxyd.HttpResponse{}
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

		sproxydRequest.Client = &http.Client{}

		go func(url string) {
			sproxydRequest.Path = url
			resp, err := sproxyd.Getobject(&sproxydRequest)
			defer resp.Body.Close()
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			} else {
				resp.Body.Close()
			}
			// WARNING The caller must close the Body after it is consumed
			ch <- &sproxyd.HttpResponse{url, resp, &body, err}
		}(url)
	}
	// wait for http response  message
	for {
		select {
		case r := <-ch:
			// fmt.Printf("%s was fetched\n", r.url)
			sproxydResponses = append(sproxydResponses, r)
			if len(sproxydResponses) == treq /*len(urls)*/ {
				return sproxydResponses
			}
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("r")
		}
	}
	return sproxydResponses
}
