package sindexd

import "net/http"

type Get_Keys struct {
	Key []string `json:"get"`
}

func (p *Get_Keys) GetKeys(client *http.Client, l *Load) (*http.Response, error) {

	/*
		l is a pointer to a Load (sindexd) structure
		keyObject is a map of "key" = obj  pair to be indexed
		[ { "hello":{ "protocol": "sindexd-1"} },
		{ "load":   {  "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x} },
		{ "get": [ "k1", "k2", "kn"]}]
	*/

	return request(client, l, p)
}
