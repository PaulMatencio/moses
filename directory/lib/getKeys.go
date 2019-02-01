package directory

import (
	sindexd "github.com/moses/sindexd/lib"
	goLog "github.com/moses/user/goLog"
	"net/http"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

func GetKeys(client *http.Client, index *sindexd.Index_spec, aKey *[]string) (resp *http.Response, err error) {

	l := &sindexd.Load{
		Index_spec: sindexd.Index_spec{
			Index_id:  index.Index_id,
			Cos:       index.Cos,
			Vol_id:    index.Vol_id,
			Specific:  index.Specific,
			Read_only: index.Read_only,
		},
	}
	g := &sindexd.Get_Keys{
		Key: *aKey,
	}

	return g.GetKeys(client, l)
}

func Get(iIndex string, client *http.Client, hp hostpool.HostPool, l *sindexd.Load, aKey []string) {

	g := &sindexd.Get_Keys{
		Key: aKey,
	}
	start := time.Now()
	resp, err := g.GetKeys(client, l)
	if resp != nil {
		body := sindexd.GetBody(resp)
		time0 := time.Since(start)
		Print(iIndex, body)
		goLog.Info.Println("Total Elaspe:", time.Since(start), "Time to get:", time0)
	} else {
		goLog.Error.Println("Get Key >", err)
	}

}
