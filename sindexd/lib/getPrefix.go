package sindexd

import (
	// "fmt"
	"net/http"
	// "user/goLog"
)

type Get_Prefix struct {
	Prefix `json:"get"`
}

type Prefix struct {
	Prefix    string `json:"prefix"`
	Marker    string `json:"marker,omitempty"`
	Delimiter string `json:"delimiter,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

func (p *Get_Prefix) GetPrefix(client *http.Client, l *Load) (*http.Response, error) {

	/*
		l is a pointer to a Load (sindexd) structure
		keyObject is a map of "key" = obj  pair to be indexed
		[ { "hello":{ "protocol": "sindexd-1"} },
		{ "load":   {  "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x} },
		{   "get": {  "prefix": "key", "delimiter": "/","limit": n }  }   ]
	*/
	// p is a pointer to a Get_prefix structure, it can still be modified before sending a Post request to the sindexd server
	// p.Limit = 10000
	// goLog.Info.Println("p:", p, "l:", l)
	return request(client, l, p)

}
