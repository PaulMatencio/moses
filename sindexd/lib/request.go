package sindexd

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

// used by createindex, dropindex, etc ..
func request(client *http.Client, l interface{}, p interface{}) (*http.Response, error) {

	lj, err := json.Marshal(l)
	pj, err := json.Marshal(p)
	// goLog.Trace.Println(err, p, string(pj))
	if err == nil {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), lj, []byte(V), pj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		if !Test {
			return PostRequest(client, request)
		} else {
			return nil, errors.New("Post cancelled due to -test flag true")
		}
	} else {
		return nil, err
	}

}
