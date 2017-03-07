package directory

import (
	// hostpool "github.com/bitly/go-hostpool"
	sindexd "moses/sindexd/lib"
	"net/http"
)

func Drop(client *http.Client, index *sindexd.Index_spec, force bool, admin bool) (*http.Response, error) {

	f, a := 0, 0
	if force {
		f = 1
	}
	if admin {
		a = 1
	}
	c := &sindexd.Drop_Index{
		Index_spec: sindexd.Index_spec{
			Index_id: index.Index_id,
			Cos:      index.Cos,
			Vol_id:   index.Vol_id,
			Specific: index.Specific,
		},
		Force: f,
		Admin: a,
	}
	return c.DropIndex(client)

}

/*
func DropMult(client *http.Client, url string, index *sindexd.Index_spec, force bool, pniod map[string]string) {
	f, a := 0, 0
	if force {
		f = 1
		a = 1
	}
	for _, v := range pniod {
		c := &sindexd.Drop_Index{
			Index_spec: sindexd.Index_spec{
				Index_id: v,
				Cos:      index.Cos,
				Vol_id:   index.Vol_id,
				Specific: index.Specific,
			},
			Force: f,
			Admin: a,
		}
		if resp, err := c.DropIndex(client, url); err != nil {
			goLog.Error.Println(err)
		} else {
			response := sindexd.GetResponse(resp)
			goLog.Info.Println("sindexd.Response:", response.Status, response.Reason)
		}
	}
}
*/
