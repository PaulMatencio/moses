package sproxyd

import (
	"net/http"
)

func Getobject(sproxydRequest *HttpRequest) (*http.Response, error) {
	req, _ := http.NewRequest("GET", DummyHost+sproxydRequest.Path, nil)
	/*
		Replica-Policy: "immutable" skips various checks to speed up responses based on the requesting application's certainty concerning
		the objects' immutability.
		Use this value only if the object has never been and never will be rewritten.
		If the "immutable" value is used for objects that have been rewritten, an erroneous version or an error may be returned.
	*/
	if len(ReplicaPolicy) > 0 {
		req.Header.Add("X-Scal-Replica-Policy", ReplicaPolicy)
	}
	if Range, ok := sproxydRequest.ReqHeader["Range"]; ok {
		req.Header.Add("Range", Range)
	}
	if ifmod, ok := sproxydRequest.ReqHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := sproxydRequest.ReqHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	return DoRequest(sproxydRequest.Hspool, sproxydRequest.Client, req, nil)
}
