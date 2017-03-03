package bns

import (
	"encoding/json"
	"os"
)

/*OLD DOCUMENT Meta Data */
type Docmeta struct {
	Date_drawup string `json:"date_drawup"`
	Pub_date    string `json:"pub_date"`
	Content     string `json:"content"`
	Data_type   string `json:"date_type"`
	Doc_id      string `json:"doc_id"`
	Kc          string `json:"kc"`
	O_pub       string `json:"o_pub"`
	Page_number string `json:"page_number"`
	Pub_office  string `json:"pub_office"`
	Total_pages string `json:"total_pages,omitempty"`
}

type Documentmeta struct {
	Abstract [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Abstract,omitempty"`
	Amendment [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Amendment,omitempty"`
	ApplicantCitations [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Applicant_citations,omitempty"`
	Bibliography [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Bibliography,omitempty"`
	Claims [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Claims"`
	Classification [1]string `json:"Classification"`
	Copyright      bool      `json:"Copyright"`
	DNASequence    []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"DNA_sequence,omitempty"`
	Description [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Description,omitempty"`
	DocumentID struct {
		CC string `json:"CC"`
		KC string `json:"KC"`
		PN string `json:"PN"`
	} `json:"Document_id"`
	DocumentType string `json:"Document_type"`
	Drawings     [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Drawings,omitempty"`
	FamilyID string `json:"Family_id"`
	LnkdocID [1]struct {
		CC string `json:"CC"`
		KC string `json:"KC"`
		PN string `json:"PN"`
	} `json:"Lnkdoc_id,omitempty"`
	LoadingDate string `json:"Loading_date,omitempty"`
	Multimedia  struct {
		PDF   bool `json:"PDF,omitempty"`
		PNG   bool `json:"PNG,omitempty"`
		TIFF  bool `json:"TIFF,omitempty"`
		VIDEO bool `json:"VIDEO,omitempty"`
	} `json:"Multimedia"`
	PageNumber        int    `json:"Page_number"`
	PublicationDate   string `json:"Publication_date"`
	PublicationID     string `json:"Publication_id"`
	PublicationOffice string `json:"Publication_office"`
	SearchReport      [1]struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"Search_report,omitempty"`
	TotalPages int `json:"Total_pages"`
}

type Range struct {
	Start int `json:"start"`
	end   int `json:"end"`
}

func (docmeta *Documentmeta) Encode(filename string) error {

	file, err := os.Create(filename)
	if err == nil {
		defer file.Close()
		encoder := json.NewEncoder(file)
		return encoder.Encode(&docmeta)
	} else {
		return err
	}
}

func (docmeta *Documentmeta) Decode(filename string) error {
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		return decoder.Decode(&docmeta)
	} else {
		return err
	}
}

/* OLD  PAGE META */

type Pagmeta struct {
	Date_drawup string `json:"date_drawup"`
	Pub_date    string `json:"pub_date"`
	Content     string `json:"content"`
	Data_type   string `json:"date_type"`
	Doc_id      string `json:"doc_id"`
	Kc          string `json:"kc"`
	O_pub       string `json:"o_pub"`
	Page_number string `json:"page_number"`
	Pub_office  string `json:"pub_office"`
	Page_size   string `json:"page_size"`
	Total_pages string `json:"total_pages,omitempty"`
}

type Pagemeta struct {
	DocumentID struct {
		CountryCode  string `json:"countryCode"`
		KindCode     string `json:"kindCode"`
		PatentNumber string `json:"patentNumber"`
	} `json:"documentId"`
	MultiMedia struct {
		Pdf   bool `json:"pdf"`
		Png   bool `json:"png"`
		Tiff  bool `json:"tiff"`
		Video bool `json:"video"`
	} `json:"multiMedia"`
	PageIndicator []string `json:"pageIndicator"`
	PageLength    int      `json:"pageLength"`
	PageNumber    int      `json:"pageNumber"`
	PdfOffset     struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pdfOffset,omitempty"`
	PngOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pngOffset,omitempty"`
	PublicationOffice string `json:"publicationOffice"`
	RotationCode      struct {
		Pdf  int `json:"pdf"`
		Png  int `json:"png"`
		Tiff int `json:"tiff"`
	} `json:"rotationCode"`
	TiffOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"tiffOffset,omitempty"`
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
	//Input_directory string
	Storage_nodes []string
	//Output_tiff     string
	//Output_json     string
}

type bnsImages struct {
	Pagemd Pagemeta
	Image  []byte
	Index  int
}

type Date struct {
	Year       int16
	Month, Day byte
}
