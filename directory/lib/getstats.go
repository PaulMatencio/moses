package directory

import (
	sindexd "github.com/paulmatencio/moses/sindexd/lib"
	"net/http"
)

func GetStats(client *http.Client, reset int, lowlevel int) (*http.Response, error) {

	s := &sindexd.Get_Stats{
		sindexd.Stats{
			Lowlevel:  lowlevel,
			Highlevel: 1,
			Cache:     1,
			Reset:     reset,
		},
	}

	return s.GetStats(client)

}
