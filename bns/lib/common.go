package bns

import (
	"encoding/json"
	"errors"
	"fmt"
	goLog "moses/user/goLog"
	"net/http"
	"os"
	"strconv"

	hostpool "github.com/bitly/go-hostpool"
)

type DocumentMetadata struct {
	PubId struct {
		CountryCode string `json: "countryCode`
		PubNumber   string `json: "pubNumber"`
		KindCode    string `json: "kindCode"`
	} `json: "PubId,omitempty"`

	BnsId struct {
		CountryCode string `json: "countryCode`
		PubNumber   string `json: "pubNumber"`
		KindCode    string `json: "kindCode"`
	} `json: "bnsId,omitempty"`

	DocId             interface{} `json:"docId` // could be integer  or string
	PublicationOffice string      `json:"publicationOffice`
	FamilyId          interface{} `json:"familyId"` // could be integer  or string
	TotalPage         int         `json:totalPage"`
	DocType           string      `json:docType"`
	PubDate           string      `json:pubDate"`
	LoadDate          string      `json:loadDate"`
	Copyright         string      `json:"copyright,omitempty"`

	LinkPubId []struct {
		CountryCode string `json: "countryCode`
		PubNumber   string `json: "pubNumber"`
		KindCode    string `json: "kindCode"`
	} `json: "linkPubId,omitemty`

	MultiMedia struct {
		Tiff  bool `json:"tiff"`
		Png   bool `json:"png"`
		Pdf   bool `json:"pdf"`
		Video bool `json:"video"`
	} `json:"multiMedia"`

	AbsRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"absRangePageNumber,omitempty"`

	AmdRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"amdRangePageNumber,omitempty"`

	BibliRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"bibliRangePageNumber,omitempty"`

	ClaimsRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"claimsRangePageNumber,omitempty"`

	DescRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"descRangePageNumber,omitempty"`

	DrawRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"drawRangePageNumber,omitempty"`

	SearchRepRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"searchRepRangePageNumber,omitempty"`

	DnaSequenceRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"dnaSequenceRangePageNumber,omoitempty"`

	ApplicantCitationsRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"applicantCitationsRangePageNumber,omitempty"`

	Classification []string `json:"classification,omitempty"`
}

type Range struct {
	Start int `json:"start"`
	end   int `json:"end"`
}

func (docmeta *DocumentMetadata) Encode(filename string) error {
	file, err := os.Create(filename)
	if err == nil {
		defer file.Close()
		encoder := json.NewEncoder(file)
		return encoder.Encode(&docmeta)
	} else {
		return err
	}
}

/* Read the content of a file and convert it into a Document metadata structure   */
func (docmeta *DocumentMetadata) Decode(filename string) error {
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		return decoder.Decode(&docmeta)
	} else {
		return err
	}
}

// Get total number of pages of a document
func (usermd *DocumentMetadata) GetPageNumber() (int, error) {
	if page := usermd.TotalPage; page > 0 {
		return usermd.TotalPage, nil
	} else {
		return 0, errors.New("Page number invalid")
	}
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

func (docmeta *DocumentMetadata) GetPagesRanges(section string) string {
	var pagesranges string
	switch section {
	case "Abstract":
		for _, ranges := range docmeta.AbsRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}
	case "Amendement":
		for _, ranges := range docmeta.AmdRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}

	case "Biblio":
		for _, ranges := range docmeta.BibliRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}
	case "Claims":
		for _, ranges := range docmeta.ClaimsRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}
	case "Desc":
		for _, ranges := range docmeta.DescRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}
	case "Draw":
		for _, ranges := range docmeta.DrawRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}
	case "SearchRep":
		for _, ranges := range docmeta.SearchRepRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}
	case "DnaSequence":
		for _, ranges := range docmeta.DnaSequenceRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}
	case "Citation":
		for _, ranges := range docmeta.ApplicantCitationsRangePageNumber {
			pagesranges += fmt.Sprintf("%s:%s,", strconv.Itoa(ranges.Start), strconv.Itoa(ranges.End))
		}

	default:
	}
	return pagesranges[0 : len(pagesranges)-1]
}

func (docmeta *DocumentMetadata) GetMetadata(bnsRequest *HttpRequest, pathname string) error {
	var (
		err        error
		docmd      []byte
		statusCode int
	)
	if docmd, err, statusCode = GetDocMetadata(bnsRequest, pathname); err == nil {
		goLog.Trace.Println("Document Metadata=>", string(docmd))
		if len(docmd) != 0 {
			if err = json.Unmarshal(docmd, &docmeta); err != nil {
				goLog.Error.Println(docmd, docmeta, err)

			}
		} else if statusCode == 404 {
			goLog.Warning.Printf("Document %s is not found", pathname)
			err = errors.New("Document not found")
		} else {
			goLog.Warning.Printf("Document's %s metadata is missing", pathname)
			err = errors.New("Document metadata is missing")
		}
	} else {
		goLog.Error.Println(err)

	}
	return err

}

type Pagemeta struct {
	PubId struct {
		CountryCode string `json:"countryCode"`
		PubNumber   string `json:"pubNumber`
		KindCode    string `json:"kindCode"`
	} `json:"pubId"`
	BnsId struct {
		CountryCode string `json:"countryCode"`
		PubNumber   string `json:"pubNumber`
		KindCode    string `json:"kindCode"`
	} `json:"bnsId"`
	PublicationOffice string `json:"publicationOffice"`
	PageNumber        int    `json:"pageNumber"`
	RotationCode      struct {
		Pdf  int `json:"pdf"`
		Png  int `json:"png"`
		Tiff int `json:"tiff"`
	} `json:"rotationCode"`
	Pubdate    string `json:"pubDate`
	Copyright  string `json:"copyright`
	MultiMedia struct {
		Pdf   bool `json:"pdf"`
		Png   bool `json:"png"`
		Tiff  bool `json:"tiff"`
		Video bool `json:"video"`
	} `json:"multiMedia"`
	PageIndicator []string `json:"pageIndicator"`
	PageLength    int      `json:"pageLength"`
	TiffOffset    struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"tiffOffset,omitempty"`
	PngOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pngOffset,omitempty"`
	PdfOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pdfOffset,omitempty"`
}

func (pagemeta *Pagemeta) Encode(filename string) error {

	file, err := os.Create(filename)
	if err == nil {
		defer file.Close()
		encoder := json.NewEncoder(file)
		return encoder.Encode(&pagemeta)
	} else {
		return err
	}
}
func (pagemeta *Pagemeta) Decode(filename string) error {
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		return decoder.Decode(&pagemeta)
	} else {
		return err
	}
}

type PAGE struct {
	Metadata Pagemeta `json:"Metadata"`
	Tiff     struct {
		Size  int    `json:"size"`
		Image []byte `json:"image"`
	} `json:"tiff,omitempty"`
	Png struct {
		Size  int    `json:"size"`
		Image []byte `json:"image"`
	} `json:"Png,omitempty"`
}

func (page *PAGE) Encode(filename string) error {
	file, err := os.Create(filename)
	if err == nil {
		defer file.Close()
		encoder := json.NewEncoder(file)
		return encoder.Encode(&page)
	} else {
		return err
	}
}

type Configuration struct {
	Storage_nodes []string
}

type BnsResponse struct {
	HttpStatusCode int    `json: "httpstatuscode,omitempty"`
	Pagemd         []byte `json: "pagemeta,omitempty"` // decoded user meta data
	Usermd         string `json: "usermd,omitempty"`   // encoded user meta data
	ContentType    string `json: "content-type,omitempty"`
	Image          []byte `json: "images"`
	BnsId          string `json: "bnsId,omitempty"`
	PageNumber     string `json: "pageNunber,omitempty"`
	Page           int    `json: "page,omitempty`
	Err            error  `json: "errorCode"`
}

type BnsResponseLi struct {
	Page        int     `json: "page,omitempty`
	Pagemd      []byte  `json: "pagemeta,omitempty"` // decoded user meta data
	ContentType string  `json: "content-type,omitempty"`
	Image       *[]byte `json: "images"` // address of the image
	BnsId       string  `json: "bnsId,omitempty"`
}

type Date struct {
	Year       int16
	Month, Day byte
}

// bns Http request structure
type HttpRequest struct {
	Hspool hostpool.HostPool
	Urls   []string
	// Path   string
	Client *http.Client
	Media  string
}

type CopyResponse struct {
	Err    error
	SrcUrl string
	Num    int
	Num200 int
}

type MetaResponse struct {
	Err     error
	SrcUrl  string
	Encoded string
	Decoded []byte
}
