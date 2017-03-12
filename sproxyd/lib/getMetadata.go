package sproxyd

import (
	"net/http"

	// hostpool "github.com/bitly/go-hostpool"
)

/*
func GetMetadata(hspool hostpool.HostPool, client *http.Client, path string, getHeader map[string]string) (*http.Response, error) {

	url := DummyHost + path
	req, _ := http.NewRequest("HEAD", url, nil)
	if ifmod, ok := getHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := getHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	return DoRequest(hspool, client, req, nil)
}
*/

func GetMetadata(sproxydRequest *HttpRequest) (*http.Response, error) {

	// url := DummyHost + sproxydRequest.Path
	req, _ := http.NewRequest("HEAD", sproxydRequest.Path, nil)
	if ifmod, ok := sproxydRequest.ReqHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := sproxydRequest.ReqHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	return DoRequest(sproxydRequest.Hspool, sproxydRequest.Client, req, nil)
}
