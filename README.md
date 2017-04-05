# moses
Multimedia object stored on Scality Ring using sproxyd and sindexd API

This application is to store, update, delete and retrieve scaned Patent Documents. A patent document is stored in multiple objets,
one object to describe the document ( Table of Content) and one object per page of the document. Each page contain a tiff, 
png and pdf images. Not every page contain pdf. The Table of Content is meta data of a document which describe teh content of
the document:  pages, paragraph, type contents.
Page: A page of a scanned patent document. A page is an object contaning a tiff and  png images, and optionally a pdf or video 
of this page
Paragraph : Pages are grouped in paragraphes as for instance Bibliographie, Description, Abstract, Draws, Claims, Citations, DNA sequences, etc ...

A document is  accessed by path name via Scality Ring Sproxyd Driver ( object storage)
Documents are indexed with the Scality Ring Sindexd, a key/value distributed storage

The application is composed of 4 main modules

sindexd : Wrap the Scality Ring sindexd. This module provides functions to store, update, delete and retrive Scality Ring object 
and their meta data
sproxyd : wrap the Scality Ring sproxyd
directory : use the sindexd module to index and retrieve documents by name or publication date
  Publication number
  Publication date
bns : use the sproxyd and directory modules to store, copy, update, delete and retrieve documents and their meta data.

user:
command: line commands that uses the 


