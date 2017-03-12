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

	// hostpool "github.com/bitly/go-hostpool"
)

// new function
func GetPage(sproxydRequest *sproxyd.HttpRequest) (*http.Response, error) {
	sproxydRequest.ReqHeader = map[string]string{}
	return sproxyd.Getobject(sproxydRequest)
}

// new function
func AsyncHttpGetPage(bnsRequest *HttpRequest, getHeader map[string]string) []*sproxyd.HttpResponse {

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
		sproxydRequest.Path = url
		go func(url string) {
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

// new  function
func GetPageType(bnsRequest *HttpRequest) (*http.Response, error) {
	/*
		bnsRequest structure
			Hspool hostpool.HostPool
			Urls   []string
			Path		string
			Client *http.Client
			Media  string
	*/
	var (
		usermd []byte
		err    error
		resp   *http.Response
	)
	// sproxydRequest := &sproxyd.HttpRequest{}
	usermd, err = GetMetadata(bnsRequest)
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

		sproxydRequest := &sproxyd.HttpRequest{
			Path:      bnsRequest.Path,
			Hspool:    bnsRequest.Hspool,
			Client:    bnsRequest.Client,
			ReqHeader: map[string]string{},
		}

		sproxydRequest.ReqHeader["Content-Type"] = "image/" + strings.ToLower(bnsRequest.Media)

		if contentType, ok := sproxydRequest.ReqHeader["Content-Type"]; ok {

			switch contentType {
			case "image/tiff", "image/tif":
				start := strconv.Itoa(pagemeta.TiffOffset.Start)
				end := strconv.Itoa(pagemeta.TiffOffset.End)
				sproxydRequest.ReqHeader["Range"] = "bytes=" + start + "-" + end
				goLog.Trace.Println(sproxydRequest.ReqHeader)
				// resp, err = sproxyd.GetObject(hspool, client, path, getHeader)
				resp, err = sproxyd.Getobject(sproxydRequest)

			case "image/png":
				start := strconv.Itoa(pagemeta.PngOffset.Start)
				end := strconv.Itoa(pagemeta.PngOffset.End)
				sproxydRequest.ReqHeader["Range"] = "bytes=" + start + "-" + end
				goLog.Trace.Println(sproxydRequest.ReqHeader)
				// resp, err = sproxyd.GetObject(hspool, client, path, getHeader)
				resp, err = sproxyd.Getobject(sproxydRequest)

			case "image/pdf":
				if pagemeta.PdfOffset.Start > 0 {
					start := strconv.Itoa(pagemeta.PdfOffset.Start)
					end := strconv.Itoa(pagemeta.PdfOffset.End)
					sproxydRequest.ReqHeader["Range"] = "bytes=" + start + "-" + end
					goLog.Trace.Println(sproxydRequest.ReqHeader)
					resp, err = sproxyd.Getobject(sproxydRequest)
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

func AsyncHttpGetpageType(bnsRequest *HttpRequest) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0
	// fmt.Printf("\n")
	bnsRequest.Client = &http.Client{} // one http connection for all requests

	for _, url := range bnsRequest.Urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		bnsRequest.Path = url
		//sproxydRequest.Path = url
		go func(url string) {
			resp, err := GetPageType(bnsRequest)
			defer resp.Body.Close()
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			}
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
