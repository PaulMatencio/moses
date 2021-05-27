package directory

import (
	"errors"
	sindexd "github.com/paulmatencio/moses/sindexd/lib"
	"net/http"
)

/*
type HttpResponse struct { // used for get prefix
	indexId  string
	pref     string
	response *http.Response
	size     int
	err      error
}
*/

func GetPrefix(client *http.Client, index *sindexd.Index_spec, prefix string, delimiter string, marker string, limit int) (resp *http.Response, err error) {

	if index == nil {
		return nil, errors.New("Pointer to the index is nil")
	}
	l := &sindexd.Load{
		Index_spec: sindexd.Index_spec{
			Index_id:  index.Index_id,
			Cos:       index.Cos,
			Vol_id:    index.Vol_id,
			Specific:  index.Specific,
			Read_only: index.Read_only,
		},
	}
	p := &sindexd.Get_Prefix{
		Prefix: sindexd.Prefix{
			Prefix: prefix,
			Limit:  limit,
		},
	}
	if len(marker) > 0 {
		p.Marker = marker
	}
	if len(delimiter) > 0 {
		p.Delimiter = delimiter
	}

	return p.GetPrefix(client, l)

}
