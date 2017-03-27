// bns project bns.go
package bns

import (
	"encoding/json"

	"io/ioutil"
	goLog "moses/user/goLog"
	"strings"
)

func GetPageMetadata(bnsRequest *HttpRequest, url string) ([]byte, error, int) {
	return GetMetadata(bnsRequest, url)
}

func GetDocMetadata(bnsRequest *HttpRequest, url string) ([]byte, error, int) {
	return GetMetadata(bnsRequest, url)
}

func BuildSubtable(content string, index string) []int {
	page_tab := make([]int, 0, Max_page)
	dpage := strings.Split(content, ",")
	for k, v := range dpage {
		if strings.Contains(v, index) {
			page_tab = append(page_tab, k+1)
		}
	}
	return page_tab
}

func (docmeta *DocumentMetadata) SetDocmd(filename string) error {
	var (
		buf []byte
		err error
	)
	if buf, err = ioutil.ReadFile(filename); err != nil {
		goLog.Warning.Println(err, "Reading", filename)
		return err
	} else if err = json.Unmarshal(buf, &docmeta); err != nil {
		goLog.Warning.Println(err, "Unmarshalling", filename)
		return err
	}
	return err
}

func (pagemeta *Pagemeta) SetPagemd(filename string) error {
	//* USE Encode
	var (
		buf []byte
		err error
	)
	if buf, err = ioutil.ReadFile(filename); err != nil {
		goLog.Warning.Println(err, "Reading", filename)
		return err
	} else if err = json.Unmarshal(buf, &pagemeta); err != nil {
		goLog.Warning.Println(err, "Unmarshalling", filename)
		return err
	}
	return err
}
