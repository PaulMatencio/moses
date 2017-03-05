package bns

import (
	"fmt"
	"io/ioutil"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"time"
)

// Used to put/update/get a same object multiple times
// Used ONLY by test.go to valide the performance of the  Ring performance
// Use utilities.go for asynchronous operations with different objects

func AsyncHttpGet(urls []string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	treq := 0
	fmt.Printf("\n")
	for _, url := range urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			// fmt.Printf("Fetching %s \n", url)
			client := &http.Client{}
			//start := time.Now()
			//var elapse time.Duration
			resp, err := GetPage(client, url)
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			} else {

				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, len(body), err}

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

func AsyncHttpUpdate(urls []string, buf []byte, header map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	treq := 0

	for _, url := range urls {
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			clientw := &http.Client{}
			resp, err = sproxyd.UpdObject(clientw, url, buf, header)
			if resp != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
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

func AsyncHttpPut(urls []string, buf []byte, header map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	treq := 0

	for _, url := range urls {
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			clientw := &http.Client{}
			resp, err = sproxyd.PutObject(clientw, url, buf, header)
			if resp != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
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

func AsyncHttpDelete(urls []string, deleteheader map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0

	for _, url := range urls {
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			clientw := &http.Client{}
			resp, err = sproxyd.DeleteObject(clientw, url)

			if resp != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
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
