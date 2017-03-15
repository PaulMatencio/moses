package sproxyd

import (
	"bytes"
	"net/http"
	"strconv"

	hostpool "github.com/bitly/go-hostpool"
)

func UpdObject(hspool hostpool.HostPool, client *http.Client, path string, object []byte, putHeader map[string]string) (*http.Response, error) {

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
	req.Header.Add("If-Match", "*")
	return DoRequest(hspool, client, req, object)
}

func Updobject(sproxydRequest *HttpRequest, object []byte) (*http.Response, error) {

	url := DummyHost + sproxydRequest.Path
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(object))
	if usermd, ok := sproxydRequest.ReqHeader["Usermd"]; ok {
		req.Header.Add("X-Scal-Usermd", usermd)
	}
	if contentType, ok := sproxydRequest.ReqHeader["Content-Type"]; ok {
		req.Header.Add("Content-Type", contentType)
	}
	if contentLength, ok := sproxydRequest.ReqHeader["Content-Length"]; ok {
		req.Header.Add("Content-Length", contentLength)
	} else {
		req.Header.Add("Content-Length", strconv.Itoa(len(object)))
	}
	if policy, ok := sproxydRequest.ReqHeader["X-Scal-Replica-Policy"]; ok {
		req.Header.Add("X-Scal-Replica-Policy", policy)
	}
	req.Header.Add("If-Match", "*")
	return DoRequest(sproxydRequest.Hspool, sproxydRequest.Client, req, object)
}

func UpdobjectTest(sproxydRequest *HttpRequest, object []byte) (*http.Response, error) {

	url := DummyHost + sproxydRequest.Path
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(object))
	if usermd, ok := sproxydRequest.ReqHeader["Usermd"]; ok {
		req.Header.Add("X-Scal-Usermd", usermd)
	}
	if contentType, ok := sproxydRequest.ReqHeader["Content-Type"]; ok {
		req.Header.Add("Content-Type", contentType)
	}
	if contentLength, ok := sproxydRequest.ReqHeader["Content-Length"]; ok {
		req.Header.Add("Content-Length", contentLength)
	} else {
		req.Header.Add("Content-Length", strconv.Itoa(len(object)))
	}
	if policy, ok := sproxydRequest.ReqHeader["X-Scal-Replica-Policy"]; ok {
		req.Header.Add("X-Scal-Replica-Policy", policy)
	}
	req.Header.Add("If-Match", "*")
	return DoRequestTest(sproxydRequest.Hspool, sproxydRequest.Client, req, object)
}
