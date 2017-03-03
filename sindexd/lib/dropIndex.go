package sindexd

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Drop_Index struct {
	Index_spec `json:"drop"`
	Version    int `json:"version,omitempty"`
	Force      int `json:"force,omitempty"`
	Admin      int `json:"admin,omitempty"`
}

func (d *Drop_Index) DropIndex(client *http.Client) (*http.Response, error) {

	/*
		d is a pointer to a Drop_Index structure
		DropIndex is a method of the Drop_Index structure
		 { "drop" : { "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x } }
	*/
	dj, err := json.Marshal(d)
	if err == nil {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), dj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		return PostRequest(client, request)
	} else {
		return nil, err
	}
}
