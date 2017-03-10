package sproxyd

import (
	"net/http"
)

func GetObject(client *http.Client, path string, getHeader map[string]string) (*http.Response, error) {

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
