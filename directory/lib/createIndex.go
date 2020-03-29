package directory

import (
	//hostpool "github.com/bitly/go-hostpool"
	sindexd "github.com/moses/sindexd/lib"
	"net/http"
)

// func Create(client *http.Client, hp hostpool.HostPool, index *sindexd.Index_spec) (*http.Response, error) {
func Create(client *http.Client, index *sindexd.Index_spec) (*http.Response, error) {
	// hp := sindexd.HP
	c := &sindexd.Create_Index{
		Index_spec: sindexd.Index_spec{
			Index_id: index.Index_id,
			Cos:      index.Cos,
			Vol_id:   index.Vol_id,
			Specific: index.Specific,
		},
	}
	// return c.CreateIndex(client, hp.Get().Host())
	return c.CreateIndex(client)
}
