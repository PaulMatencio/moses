// sindexd project sindexd.go
package sindexd

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	goLog "user/goLog"

	hostpool "github.com/bitly/go-hostpool"
	//"time"
)

func PostRequest(client *http.Client, d []byte) (*http.Response, error) {

	var (
		resp  *http.Response
		err   error
		r     int
		hpool hostpool.HostPoolResponse
		url   string
	)

	for r = 1; r <= 3; r++ {
		query := bytes.NewBuffer(d)
		hpool = HP.Get()
		url = hpool.Host()
		req, _ := http.NewRequest("POST", url, query)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(query.Len()))
		if Debug {
			goLog.Info.Println(r, "REQUEST_HEADER:> ", req.Header, "REQUEST_URL:>", req.URL, "REQUEST_BODY:> ", req.Body)
		}
		// execute the request
		resp, err = client.Do(req)
		if err != nil {
			hpool.Mark(err)
			goLog.Error.Println(err)
		} else if resp.StatusCode != 200 {
			hpool.Mark(errors.New(resp.Status))
			if Debug {
				query = bytes.NewBuffer(d)
				goLog.Error.Println(r, req.URL, query, resp.StatusCode)
			} else {
				goLog.Error.Println(r, req.URL, resp.StatusCode)
			}

		} else {
			hpool.Mark(err)
			break
		}
	}
	return resp, err
}
