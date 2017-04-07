# Moses

Multimedia Object Storage 

## Overview

This purpose of this application is to `store`, `update`, `delete`, `list` and `retrieve` scanned patent documents. A patent document is stored in multiple objects, one object to describe the document itself (a Table of Content) and one object per page of the document. Each page contain metadata and sections such as tiff, png images and optionally a pdf image. Documents are stored as objects in the Scality Ring's' Object Storage using the sproxyd restful API ( low level Object storage API), then   indexed with the sindexd key/value store of the Scality Ring's Software Defined Storage  


## Table of Content

The Table of Content (ToC) conrains only meta data which describes the layout of a document :  `page`, `section`, etc .  Below is a go structure that defines the TOC of a document. 

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

## Section 

A document has multiple sections. Ecah section is one to multiple range of pages of a document. Following is the list of sections: Bibliographie, Description, Abstract, Draws, Claims, Citations, DNA sequences, etc .. pages  of a patent. 

Example of sections of a document : AbsRangePageNumber (abstract), AmdRangePageNumber, etc  (amendement)

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


## Page

A page of a patent document is an object with metadata which describe ist contents and data contaning subpages. Subpages are  tiff and png images, and optionally a pdf or video. 

Below is a go structure which describes the content of a page 

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

Subpage is a range of bytes of the object taht contains images such tiff,png and optionally pdf and video. Example of subpage : Tiff, Png 

	TiffOffset    struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"tiffOffset,omitempty"`
	
	PngOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pngOffset,omitempty"`



	
## Access

The document, pages and subparts are accessed by `pathname` using the Scality Ring Sproxyd Driver (a low level object storage restful API)
 
Example of `pathname`: 

	US/4000000/A2    US is the Country code, 4000000 is the Publication number and A2 is the kind code
	US/4000000/A2/p20   Page 20 of the PID  US/4000000/A2

## Indexing

The documents ( Table of Content ) object are indexed with the Scality Ring Sindexd, a key/value distributed store. Currently documents are indexed by Publication Id (key = CC+PN+KC) and Publication dat ( key = CC+YYYYMMDD). The value contain information such publication date, unique document id , etc ... Moreover, there are one index table per main country. Small countries are grouped in a separate index table


## Index tables

`Example of index tables`

	{"country":"CN","index_id":"148B0FD1AE4E918F84792B45C2FA0E0300002A20","cos":2,"volid":1170405902,"specific":42}
	{"country":"CA","index_id":"7B43BD0E90E3F64EF3B3D57883AC120300002A20","cos":2,"volid":2021895186,"specific":42}
	{"country":"DE","index_id":"58FE9CABF4E51CCFB7E33B3DB61C380300002A20","cos":2,"volid":1035344952,"specific":42}
	{"country":"EP","index_id":"4E598DFF1578EBE87CCDD066546F5F0300002A20","cos":2,"volid":1716809567,"specific":42}
	{"country":"FR","index_id":"A0458B6E535B0084ECF6C2EA189B640300002A20","cos":2,"volid":3927481188,"specific":42}
	{"country":"GB","index_id":"BEDE61296BEC137879AE67B7BF2E710300002A20","cos":2,"volid":3082759793,"specific":42}

### INDEX_SPEC:  Compound argument composed of the index_id, cos, volid, and specific arguments.

	index_id: Part of the INDEX_SPEC argument that designates the RING ID for the index. A 40-character hexadecimal number.
	cos Part: of the INDEX_SPEC argument that designates the class of service for index chunks in the RING.  
	volid: Part of the INDEX_SPEC argument that designates the volume ID for index chunks in the RING.  
	specific: Part of the INDEX_SPEC argument that designates the specific ID for index chunks in the RING

## Document Publication ID

Foramt of a piblication ID : CC/PN/KC  ( as for instance US/40000000/A2)

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

`Sindexd`  : Tool to 

	Create and Drop indexes tables. 
	Add and Delete indexes entries 
	Retrieve specific indexes entries
	List indexes entries ( with prefixing and delimiters)
	Retrieve the configuration of the key/value store (sindexd)
	Retrieve the statistics regarding the key/value store (sindexd)
	

`ScanPrefix`

	scan all entries of one or multiple index tables  

`BuildIndexparm` 

	Build the Index specification json file . This json file contains the index speficication ( index table definition) per main conutry. Small countries are grouped into a specific Index table  
	
	{"country":"CN","index_id":"148B0FD1AE4E918F84792B45C2FA0E0300002A20","cos":2,"volid":1170405902,"specific":42}
	{"country":"CA","index_id":"7B43BD0E90E3F64EF3B3D57883AC120300002A20","cos":2,"volid":2021895186,"specific":42} 



## Moses RestFul API (coming)


## Moses micro services ( coming)


