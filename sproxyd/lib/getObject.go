package sproxyd

import (
	"net/http"
)

/*
func GetObject(hspool hostpool.HostPool, client *http.Client, path string, getHeader map[string]string) (*http.Response, error) {

	url := DummyHost + path
	req, _ := http.NewRequest("GET", url, nil)
	if Range, ok := getHeader["Range"]; ok {
		req.Header.Add("Range", Range)
	}
	if ifmod, ok := getHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := getHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	// resp, err := client.Do(req)
	return DoRequest(HP, client, req, nil)
}
*/

func Getobject(sproxydRequest *HttpRequest) (*http.Response, error) {
	// hspool hostpool.HostPool, client *http.Client, path string, getHeader map[string]string
	//url := DummyHost + url
	req, _ := http.NewRequest("GET", DummyHost+sproxydRequest.Path, nil)
	if Range, ok := sproxydRequest.ReqHeader["Range"]; ok {
		req.Header.Add("Range", Range)
	}
	if ifmod, ok := sproxydRequest.ReqHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := sproxydRequest.ReqHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	// resp, err := client.Do(req)
	return DoRequest(sproxydRequest.Hspool, sproxydRequest.Client, req, nil)
}
