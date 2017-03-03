package sindexd

import (
	"net/http"
)

type Delete_Keys struct {
	Key      []string `json:"delete"`
	Prefetch bool     `json:"prefetch"`
}

func (p *Delete_Keys) DeleteKeys(client *http.Client, l *Load) (*http.Response, error) {
	/*
		l is a pointer to a Load (sindexd) structure
		keyObject is a map of "key" = obj  pair to be indexed
		[ { "hello":{ "protocol": "sindexd-1"} },
		{ "load":   {  "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x} },
		{ "delete": [ "k1", "k2", "kn"],"prefetch":false}]
	*/
	/*  p is a pointer to a Delete_leys structure, it can still be modified before sending a Post request to the sindexd server  */
	p.Prefetch = false
	return request(client, l, p)
}
