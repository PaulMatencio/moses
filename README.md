# Moses
Multimedia Object Storage 

## Application functions overview
This application is to `store`, `update`, `delete`, `list` and retrieve scanned Patent Documents. A single patent document is stored in multiple objects, an object to describe the document itself ( Table of Content) and one object per page of the document. Each page contain a tiff and png imags and optionally a pdf image. Documents are objects stored in the Scality Ring Object storage using the sproxyd restful API ( low level Object storage) and are indexed with the sindexd key/value storage of Scality 


## Table of Content
The Table of Content is only meta data that describes the lyaout of the document :  `page`, `subpart`  

## Page
A pageof a patent document is an object with metadata and data contaning a tiff and png images, and optionally a pdf or video 
of this page. The page's metadata describe the content of the page.

## Subpart 
Subpart of a patent document is one or multiple range of pages that contains the Bibliographie, Description, Abstract, Draws, Claims, Citations, DNA sequences, etc ... of a patent.

## Access
Document, pages and subparts are accessed by path name using Scality Ring Sproxyd Driver ( low level object storage restful API)


## Indexing

Documents are indexed with the Scality Ring Sindexd, a key/value distributed storage

## Document Publication ID

CC/PN/KC 
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


## extenal librries



## Moses commands Line

Publication number: 

	TOC ( one per Publication Number )
	Pages  (a few to thousands pages)

`CopyObject`:   Copy a Publication Number (TOC + Pages) from one Ring to another Ring ( or the same Ring)
`CopyPNs` : Copy list of Publication Numbers (ROC + Pages) from one Ring to another Ring ( or the same Ring)
`DeleteObject`: Delete a Publication Number (TOC + Pages) from one Ring to another Ring ( or the same Ring)
`DeletePNs`: Delete list of Publication Numbers (ROC + Pages) from one Ring to another Ring ( or the same Ring)
`UpdateObject`: Update a Publication Number (TOC + Pages) from one Ring to another Ring ( or the same Ring)
`UpdatePNs`: Delete list of Publication Numbers (ROC + Pages) from one Ring to another Ring ( or the same Ring)

`GetDocument`: Retrieve document object data and metadata, document type (tiff/png/pdf),  pages, range of pages, subparts ( Claims, Biblio, Description, etc )

`Sindexd`  :
`GetPrefix`
`BuildIndexparm` 




## Moses RestFul API (coming)


## Moses micro services ( coming)


