// sproxyd project sproxyd.go
package sproxyd

import (
	"bytes"
	"net/http"
	"strconv"

	hostpool "github.com/bitly/go-hostpool"
)

func PutObject(hspool hostpool.HostPool, client *http.Client, path string, object []byte, putHeader map[string]string) (*http.Response, error) {

	url := DummyHost + path
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(object))
	if usermd, ok := putHeader["Usermd"]; ok {
		req.Header.Add("X-Scal-Usermd", usermd)
	}
	if contentType, ok := putHeader["Content-Type"]; ok {
		req.Header.Add("Content-Type", contentType)
	}
	if contentLength, ok := putHeader["Content-Length"]; ok {
		req.Header.Add("Content-Length", contentLength)
	} else {
		req.Header.Add("Content-Length", strconv.Itoa(len(object)))
	}
	if policy, ok := putHeader["X-Scal-Replica-Policy"]; ok {
		req.Header.Add("X-Scal-Replica-Policy", policy)
	}
	req.Header.Add("If-None-Match", "*")
	return DoRequest(hspool, client, req, object)

}

func PutObjectTest(hspool hostpool.HostPool, client *http.Client, path string, object []byte, putHeader map[string]string) (*http.Response, error) {

	url := DummyHost + path
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(object))
	if usermd, ok := putHeader["Usermd"]; ok {
		req.Header.Add("X-Scal-Usermd", usermd)
	}
	if contentType, ok := putHeader["Content-Type"]; ok {
		req.Header.Add("Content-Type", contentType)
	}
	if contentLength, ok := putHeader["Content-Length"]; ok {
		req.Header.Add("Content-Length", contentLength)
	} else {
		req.Header.Add("Content-Length", strconv.Itoa(len(object)))
	}
	if policy, ok := putHeader["X-Scal-Replica-Policy"]; ok {
		req.Header.Add("X-Scal-Replica-Policy", policy)
	}
	req.Header.Add("If-None-Match", "*")

	return DoRequestTest(hspool, client, req, object)

}
