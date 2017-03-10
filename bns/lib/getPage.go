package bns

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	sproxyd "moses/sproxyd/lib"
	goLog "moses/user/goLog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GetPage  will be used by  AsyncHttpGetPage for concurrenr getPage
func GetPage(client *http.Client, path string) (*http.Response, error) {
	header := map[string]string{}
	return sproxyd.GetObject(client, path, header)

}

// GetPageType will be used by  AsyncHttpGetPageType for concurrent Getpagetype
// func GetPageType(client *http.Client, path string, getHeader map[string]string) (*http.Response, error) {
func GetPageType(client *http.Client, path string, media string) (*http.Response, error) {
	//  getrHeader must contain Content-type
	var (
		usermd []byte
		err    error
		resp   *http.Response
	)
	usermd, err = GetMetadata(client, path)
	if err != nil {
		return nil, errors.New("Page metadata is missing or invalid")
	} else {
		// c, _ := base64.Decode64(string(usermd))
		//goLog.Trace.Println("Usermd=", string(usermd))
		if len(usermd) == 0 {
			return nil, errors.New("Page metadata is missing. Please check the warning log for the reason")
		}
		var pagemeta Pagemeta
		if err := json.Unmarshal(usermd, &pagemeta); err != nil {
			return nil, err
		}

		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)

		if contentType, ok := getHeader["Content-Type"]; ok {

			switch contentType {
			case "image/tiff", "image/tif":
				start := strconv.Itoa(pagemeta.TiffOffset.Start)
				end := strconv.Itoa(pagemeta.TiffOffset.End)
				getHeader["Range"] = "bytes=" + start + "-" + end
				goLog.Trace.Println(getHeader)
				resp, err = sproxyd.GetObject(client, path, getHeader)

			case "image/png":
				start := strconv.Itoa(pagemeta.PngOffset.Start)
				end := strconv.Itoa(pagemeta.PngOffset.End)
				getHeader["Range"] = "bytes=" + start + "-" + end
				resp, err = sproxyd.GetObject(client, path, getHeader)

			case "image/pdf":
				if pagemeta.PdfOffset.Start > 0 {
					start := strconv.Itoa(pagemeta.PdfOffset.Start)
					end := strconv.Itoa(pagemeta.PdfOffset.End)
					getHeader["Range"] = "bytes=" + start + "-" + end
					resp, err = sproxyd.GetObject(client, path, getHeader)
				} else {
					resp = nil
					err = errors.New("Content-type " + contentType + " does not exist")
				}
			default:
				err = errors.New("Content-type is missing or invalid")
			}
		} else {
			err = errors.New("Content-type is missing or invalid")
		}
	}
	return resp, err
}

func AsyncHttpGetPage(urls []string, getHeader map[string]string) []*sproxyd.HttpResponse {

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

		client := &http.Client{} // one connection for all requests

		go func(url string) {
			// fmt.Printf("Fetching %s \n", url)
			// client := &http.Client{}
			resp, err := sproxyd.GetObject(client, url, getHeader)
			// resp, err := GetPage(client, url, getHeader)
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

// func AsyncHttpGetPageType(urls []string, getHeader map[string]string) []*sproxyd.HttpResponse {
func AsyncHttpGetPageType(urls []string, media string) []*sproxyd.HttpResponse {
	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	// fmt.Println(urls)
	treq := 0
	fmt.Printf("\n")

	client := &http.Client{} // one http connection for all requests

	for _, url := range urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}

		go func(url string) {
			// client := &http.Client{}
			// fmt.Printf("fetching %s\n", url)

			// resp, err := GetPageType(client, url, getHeader)
			resp, err := GetPageType(client, url, media)
			defer resp.Body.Close()
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			} /*else {
				if resp != nil { // resp == nil when there is no media type in the metadata
					resp.Body.Close()
				}
			} */
			ch <- &sproxyd.HttpResponse{url, resp, &body, err}

		}(url)
	}
	// wait for http response  message
	for {
		select {
		case r := <-ch:
			// fmt.Printf("%s was fetched\n", r.Url)
			responses = append(responses, r)
			if len(responses) == treq /*len(urls)*/ {
				// fmt.Println(responses)
				return responses
			}
		case <-time.After(100 * time.Millisecond):
			fmt.Printf(".")
		}
	}

	return responses
}
