package sproxyd

import (
	"net/http"
)

func GetMetadata(client *http.Client, path string, getHeader map[string]string) (*http.Response, error) {

	url := DummyHost + path
	req, _ := http.NewRequest("HEAD", url, nil)
	if ifmod, ok := getHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := getHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	return DoRequest(HP, client, req, nil)
}
