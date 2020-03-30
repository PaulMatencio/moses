// sindexd project sindexd.go
package sindexd

import (
	"bytes"
	"errors"
	// goLog "github.com/moses/user/goLog"
	goLog "github.com/s3/gLog"
	"net/http"
	"strconv"

	hostpool "github.com/bitly/go-hostpool"
	//"time"
)

func PostRequest(client *http.Client, d []byte) (*http.Response, error) {

	var (
		resp *http.Response
		// response *Response
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
			goLog.Trace.Println(r, "REQUEST_HEADER:> ", req.Header, "REQUEST_URL:>", req.URL, "REQUEST_BODY:> ", req.Body)
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
			/*
				if response, err = GetResponse(resp); err == nil {
					if response.Status == 200 {
						hpool.Mark(nil)
						break
					} else {
						hpool.Mark(err)
					}
				}*/
			hpool.Mark(err)
			break
		}
	}
	return resp, err
}
