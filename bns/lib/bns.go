// bns project bns.go
package bns

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	goLog "moses/user/goLog"
	// "net/http"
	// "strconv"
	"strings"

	// hostpool "github.com/bitly/go-hostpool"
)

func GetPageMetadata(bnsRequest *HttpRequest, url string) ([]byte, error) {
	return GetMetadata(bnsRequest, url)
}

func GetDocMetadata(bnsRequest *HttpRequest, url string) ([]byte, error) {
	return GetMetadata(bnsRequest, url)
}

// Get total number of pages of a document

func (usermd *DocumentMetadata) GetPageNumber() (int, error) {
	return usermd.TotalPage, nil

}

// Get  the publication date of a document
func (usermd *DocumentMetadata) GetPubDate() (Date, error) {
	date := Date{}
	err := error(nil)
	if usermd.PubDate != "" {
		date, err = ParseDate(usermd.PubDate)
	} else {
		err = errors.New("no Publication date")
	}
	return date, err
}

/*
func (usermd *Documentmeta) GetDrawUpDate() string {
	return usermd.
}
*/

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
