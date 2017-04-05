# Moses

Multimedia Object Storage 

## Application functions overview

This purpose of this application is to `store`, `update`, `delete`, `list` and `retrieve` scanned Patent Documents. A single patent document is stored in multiple objects, an object to describe the document itself ( a ind of Table of Content) and one object per page of the document. Each page contain subpage section:s tiff, png imags and optionally a pdf image. Documents are objects stored in the Scality Ring Object storage using the sproxyd restful API ( low level Object storage API) and are indexed with the sindexd key/value storage of Scality 


## Table of Content

The Table of Content is only meta data that describes the layout of a document :  `page`, `subpart`, etc .  Below is a go structure that define the structure of a TOC

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



## Page

A page of a patent document is an object with metadata and data contaning a tiff and png images, and optionally a pdf or video 
of this page. The page's metadata describe the content of the page.

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

## Subpage

Specific tiff, png , pdf image and the metatadat of a page. Example of subpage : Tiff, Png 

	TiffOffset    struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"tiffOffset,omitempty"`
	
	PngOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pngOffset,omitempty"`

## Subpart 

Subpart of a patent document is one or multiple range of pages that contains the Bibliographie, Description, Abstract, Draws, Claims, Citations, DNA sequences, etc .. pages  of a patent. Example of subparts : AbsRangePageNumber (abstract), AmdRangePageNumber, etc  (amendement)

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


	
## Access

Whole Document, pages and subparts are accessed by path name using Scality Ring Sproxyd Driver ( low level object storage restful API)
 
Example of pathname : 

	US/4000000/A2    US is the Country code, 4000000 is the Publication number and A2 is the kind code
	US/4000000/A2/p20   Page 20 of the PID  US/4000000/A2

## Indexing

Documents are indexed with the Scality Ring Sindexd, a key/value distributed storage. Currently all documents are indexed by Publication Id ( key = CC+PN+KC) and Publication dat ( key = CC+YYYYMMDD)

## Document Publication ID

CC/PN/KC  ( as for instance US/40000000/A2)
CC : Country code. It's 2 characters long as for instance  'US','JP','KR','DE','FR','CH','GB' 
PN : Publication Number. It's naximum 15 characters long
KC : Kind Code. It's maximum 2 characters long , 1 char + 1 digit. Digit may be ommited.


## Main libraries 

The application is composed of 4 main libraries

`sindexd` : Wrap the Scality Ring sindexd API. This libary provides low level  functions to index and list patent documents by publication id or publication date

`sproxyd` : wrap the Scality Ring sproxyd restful API. This library provides low level functions to store, update, delete and update document objects ( both metadata and data)

`directory` : use the sindexd libray to index, list, update and retrieve any moses document by  publication id  and by publication date
   
`bns` : use the sproxyd and directory libaries to store, copy, update, delete, list and retrieve any moses documents ( metadata, page, range of pages, subpart, etc ... )


## Other libraries

`user`: files, log, hexa, encoding , etc 

## extenal libraries

 "github.com/bitly/go-hostpool"

## Moses commands Line

Patent Publication number's Path name

	TOC (one per Publication Number ). Object path name : CC/PN/KC. This object contain only metadata. 
	Page (from a  few to thousands pages). Object path name : CC/PN/KC/px , x is the number of the page. Thsi object contain both metadata en data 

`CopyObject`:   Copy a Publication Number (TOC + Pages) from one Ring to another Ring ( or the same Ring)

`CopyPNs` : Copy list of Publication Numbers (TOC + Pages) from one Ring to another Ring ( or the same Ring)

`DeleteObject`: Delete a Publication Number (TOC + Pages) from one Ring to another Ring ( or the same Ring)

`DeletePNs`: Delete list of Publication Numbers (TOC + Pages) from one Ring to another Ring ( or the same Ring)

`UpdateObject`: Update a Publication Number (TOC + Pages) from one Ring to another Ring ( or the same Ring)

`UpdatePNs`: Delete list of Publication Numbers (TOC + Pages) from one Ring to another Ring ( or the same Ring)


`GetDocument`: Retrieve document object ( TOC + Pages) data and metadata, a specific document type (TOC + tiff/png/pdf sub page),  one or multiple pages, one a multiple ranges of pages, one or multiple subparts ( Claims, Biblio, Description, etc )

`Sindexd`  :

`GetPrefix`

`BuildIndexparm` 



## Moses RestFul API (coming)


## Moses micro services ( coming)


