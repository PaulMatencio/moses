package sindexd

import (
	"encoding/json"

	"fmt"
	"io/ioutil"
	// goLog "github.com/paulmatencio/moses/user/goLog"
	goLog "github.com/s3/gLog"
	"net/http"
)

type Response struct {
	Protocol      string                 `json:"protocol"`
	Status        int                    `json:"status"`
	Reason        string                 `json:"reason,omitempty"`
	Version       int                    `json:"version,omitempty"`
	Index_id      string                 `json:"index_id,omitempty"`
	Fork_id       string                 `json:"fork_id,omitempty"`
	Snapshot_id   string                 `json:"snapshot,omitempty"`
	Fetched       map[string]interface{} `json:"fetched,omitempty"`
	Not_found     []string               `json:"not_found,omitempty"`
	Common_prefix []string               `json:"common_prefix,omitempty"`
	Next_marker   string                 `json:"next_marker,omitempty"`
	Truncated     bool                   `json:"truncated,omitempty"`
}

func (r *Response) GetNMarker() string {
	return r.Next_marker
}

func (r *Response) GetStatus() int {
	return r.Status
}

func (r *Response) GetReason() string {
	return r.Reason
}

func GetResponse(resp *http.Response) (*Response, error) {

	var (
		response = new(Response)
		err      error
	)
	if resp != nil {
		body := GetBody(resp)
		if err = json.Unmarshal(body, &response); err != nil {
			goLog.Error.Println(err)
		}
	}
	return response, err
}

func GetBody(resp *http.Response) []byte {

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return body
}

func (r *Response) PrintFetched() string {

	for k, v := range r.Fetched {
		goLog.Info.Println("key:", k, "value:", v)
	}
	return r.Next_marker
}

func (r *Response) GetFetched() map[string]interface{} {
	return r.Fetched

}

func (r *Response) GetFetchedKeys() ([]string, string) {

	num := len(r.Fetched)
	keys := make([]string, num, num)
	i := 0
	for k, _ := range r.Fetched {
		keys[i] = k
		i++
	}
	return keys, r.Next_marker
}

func (r *Response) PrintNotFound() {
	fmt.Println("Key Not found:\n")
	for i := range r.Not_found {
		goLog.Warning.Println(r.Not_found[i])
	}
}

func (r *Response) PrintCommonPrefix() string {
	fmt.Println("Common Prefix:\n")
	for i := range r.Common_prefix {
		goLog.Info.Println(r.Common_prefix[i])
	}
	return r.Next_marker
}
