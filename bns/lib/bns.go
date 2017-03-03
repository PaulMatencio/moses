// bns project bns.go
package bns

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"poc/goLog"
	"strconv"
	"strings"
)

// OBSOLETE
func (usermd *Docmeta) GetPageNumber() (int, error) {
	if usermd.Total_pages != "" {
		t_pages, err := strconv.Atoi(usermd.Total_pages)
		return t_pages, err
	} else {
		return 0, errors.New("Invalid user metadata")
	}
}

func GetPageMetadata(client *http.Client, path string) ([]byte, error) {
	return GetMetadata(client, path)
}

func GetDocMetadata(client *http.Client, path string) ([]byte, error) {
	return GetMetadata(client, path)
}

// Get total number of pages of a document

func (usermd *DocumentMetadata) GetPageNumber() (int, error) {
	return usermd.TotalPages, nil

}

// Get  the publication date of a document
func (usermd *DocumentMetadata) GetPubDate() (Date, error) {
	date := Date{}
	err := error(nil)
	if usermd.PublicationDate != "" {
		date, err = ParseDate(usermd.PublicationDate)
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

/*
func (pagmeta *Pagmeta) ToPagemeta() *Pagemeta {
	// USE FOR META DATA CONVERSION => DO NOT USE IT FOR OTHER PURPOSE
	var err error
	pagemeta := &Pagemeta{}
	pagemeta.DocumentID.CountryCode = pagmeta.O_pub
	pagemeta.DocumentID.PatentNumber = pagmeta.Doc_id
	pagemeta.DocumentID.KindCode = pagmeta.Kc
	pagemeta.MultiMedia.Tiff = true
	pagemeta.MultiMedia.Png = false
	pagemeta.MultiMedia.Pdf = false

	if pagemeta.PageNumber, err = strconv.Atoi(pagmeta.Page_number); err != nil {
		pagemeta.PageNumber = -1
		goLog.Warning.Println(err)
	}
	pagemeta.PublicationOffice = pagmeta.Pub_office

	if pagemeta.PageLength, err = strconv.Atoi(pagmeta.Page_size); err != nil {
		pagemeta.PageLength = -1
		goLog.Warning.Println(err)
	}

	pagemeta.Tiff.Start = 0
	pagemeta.Tiff.End = pagemeta.PageLength
	pagemeta.Png.Start = 0
	pagemeta.Png.End = -1
	pagemeta.Pdf.Start = 0
	pagemeta.Pdf.End = -1

	return pagemeta
}
*/
/*
func ReadPage(Inputdir string) (*bytes.Buffer,Usermd ){

 	var img bytes.Buffer
  	var usermd Usermd
	return &img,usermd
}
*/
