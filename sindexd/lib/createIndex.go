package sindexd

import (
	"bytes"
	"encoding/json"
	"net/http"

)

type Create_Index struct {
	Index_spec `json:"create"`
}

func (c *Create_Index) CreateIndex(client *http.Client) (*http.Response, error) {
	/*
		c is a pointer to a Create_Index structure
		CreateIndex is a method of the Create_Index structure
		 { "create" : { "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x, "readonly": x, "admin":  } }
	*/

	if cj, err := json.Marshal(c); err == nil {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), cj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		return PostRequest(client, request)
	} else {
		return nil, err
	}

}
