package bns

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	sproxyd "sproxyd/lib"
	"strconv"
	"strings"
	"time"
	base64 "user/base64j"
	"user/ebc2asc"
	files "user/files/lib"
	goLog "user/goLog"
	// imaging "github.com/desintegration/imaging"
)

func ST33toFiles(inputFile string, outputusermdDir string, outputTiffDir string, outputContainerDir string, combine bool) error {

	//   EXTRACT ST33 Files
	//
	//  FOR EACH DOCUMENT in ST33 {
	// 		CREATE  page metadata			 ( ST33 HEADER  ++)
	// 		CREATE  page data  (Tiff)      ( ST33 TIFF RECORDS )
	//
	// 		FOR EACH PAGE of a document {
	// 			 if combine  {
	//      		  COMBINE  page data (TIFF) and metadata => PAGE Struct
	//      		  WRITE PAGE struct
	//       	 }
	//  		else {
	//     		  	  WRITE data ( TIFF)
	//	     		  WRITE Metadata  ( user metadata)
	//       	}
	//      }
	//      CREATE DOCUMENT metadata
	//      WRITE DOCUMENT  metadata
	//  }
	//
	goLog.Info.Println("Combine output:", combine)

	Little := binary.LittleEndian
	Big := binary.BigEndian
	enc := Big
	documentmeta := &Documentmeta{}

	var (
		// container          = make([]string, 0, 500)
		abs                = make([]int, 0, 100)
		bib                = make([]int, 0, 100)
		cla                = make([]int, 0, 100)
		desc               = make([]int, 0, 300)
		draw               = make([]int, 0, 300)
		amd                = make([]int, 0, 300)
		srp                = make([]int, 0, 300)
		file_containermd_o string
		elapsetm           time.Duration
		total_rec          uint16
		total_length       uint32
		// documentmd         []byte
	)
	total := 0
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	//buf,err := ioutil.ReadFile(inputFile)
	abuf, err := files.ReadBuffer(inputFile)
	if err == nil {

		defer abuf.Reset()

		buf := abuf.Bytes()
		bufl := len(buf)
		pagmeta := &Pagmeta{}

		l := 0

		for l < bufl {
			k := l

			var (
				recs uint16
				//total_rec    uint16
				//total_length uint32
				imgl uint16
			)
			buf1 := bytes.NewReader(buf[k+25 : k+27])
			err = binary.Read(buf1, Big, &recs)
			buf1 = bytes.NewReader(buf[k+84 : k+86])
			err = binary.Read(buf1, Big, &total_rec)
			buf1 = bytes.NewReader(buf[k+214 : k+218])
			err = binary.Read(buf1, Big, &total_length)
			buf1 = bytes.NewReader(buf[k+250 : k+252])
			err = binary.Read(buf1, Big, &imgl)

			// convert  Buf ( EBCDIC) to ST33 (ASCII)
			st33 := ebc2asc.Ebc2asci(buf[l : l+214])

			long, _ := strconv.Atoi(string(st33[0:5]))

			pub := st33[5:7]
			pagmeta.Pub_office = string(pub)
			kc := strings.Trim(string(st33[7:9]), " ")
			pagmeta.Kc = kc
			docnum := st33[9:17]

			pagenum := st33[17:21]
			// goLog.Info.Printf("long=%d pub=%s kc=%s pagenum=%s k= %d l = %d bufl=%d buf:%s", long, pub, kc, pagenum, l, k, bufl, ebc2asc.Ebc2asci(buf[117450357:117450457]))
			pagmeta.Page_number = string(pagenum)
			// framenum := st33[21:25]
			//recs := byte2int(buf[k+25:k+27])

			pos9_ := st33[27:29]
			docid0 := strings.Trim(string(docnum)+string(pos9_), " ")
			docid := strings.TrimLeft(string(st33[33:45]), " ")

			if len(docid) == 0 {
				docid = docid0
			}

			pagmeta.Doc_id = docid
			//usermd["doc_id"] = strings.TrimLeft(string(docid)," ")
			o_pub := st33[67:69]
			pagmeta.O_pub = string(o_pub)
			//date_drawup := st33[69:75]
			//usermd["date_drawup"]= string(date_drawup)
			//rec_stat :=  st33[75:76]
			total_pages := st33[76:80]
			pagmeta.Total_pages = string(total_pages)
			/*
			   s_doc_h := st33[87:90]
			   s_doc_w := st33[90:93]
			*/
			f_date_drawup := st33[93:101]
			pagmeta.Date_drawup = string(f_date_drawup)
			f_pub_date := st33[101:109]
			pagmeta.Pub_date = string(f_pub_date)
			Biblio := string(st33[133:134])
			Claim := string(st33[134:135])
			Drawing := string(st33[135:136])
			Amendement := string(st33[136:137])
			Description := string(st33[137:138])
			Abstract := string(st33[138:139])
			Search_report := string(st33[139:140])

			// var content bytes.Buffer

			// pnum:= strconv.Atoi(string(pagenum))
			pn, _ := strconv.Atoi(string(pagenum))
			if Biblio == "1" {

				bib = append(bib, pn)
			}
			if Abstract == "1" {
				abs = append(abs, pn)
			}
			if Claim == "1" {
				cla = append(cla, pn)
			}
			if Drawing == "1" {
				draw = append(draw, pn)
			}
			if Amendement == "1" {
				amd = append(amd, pn)
			}
			if Description == "1" {
				desc = append(desc, pn)
			}
			if Search_report == "1" {
				srp = append(srp, pn)
			}

			// container = append(container, content.String())

			/* +40*/
			// pagmeta.Content = content.String()
			data_type := string(st33[180:181])
			pagmeta.Data_type = data_type
			// comp_meth := st33[181:183]
			// k_fac := st33[183:185]
			// Resolution := st33[185:187]
			s_fr_h := st33[187:190]
			s_fr_w := st33[190:193]
			nl_fr_h := st33[193:197]
			nl_fr_w := st33[197:201]
			rotation_code := st33[201:202]
			// fr_x := st33[202:206]
			// fr_y := st33[206:210]
			// fr_stat := st33[210:211]
			version := st33[211:214]

			// Coniunue with Buffer
			buf1 = bytes.NewReader(buf[k+214 : k+218])
			if string(version) == "V30" {

				// buf1 = bytes.NewReader(buf[k+214 : k+218])
				// some V30 total_length are encoded with big Endian byte order
				err = binary.Read(buf1, Little, &total_length)
				if int(total_length) > 16777215 {
					buf1 = bytes.NewReader(buf[k+214 : k+218])
					err = binary.Read(buf1, Big, &total_length)
				}

			} else {
				err = binary.Read(buf1, Big, &total_length)
			}

			/* write tiff images */
			var (
				Page = new(PAGE)
				img  = new(bytes.Buffer)
			)

			_, err = io.WriteString(img, beHeader)          // magic number
			err = binary.Write(img, enc, uint32(ifdOffset)) // IFD offset
			err = binary.Write(img, enc, uint16(ifdLen))    // number of IFD entries

			err = binary.Write(img, enc, uint16(tImageWidth)) //  image Width
			err = binary.Write(img, enc, uint16(dtLong))      //  long
			err = binary.Write(img, enc, uint32(1))           //  value
			err = binary.Write(img, enc, getuint32(nl_fr_w))

			//Imagewidth,_ := strconv.Atoi(string(nl_fr_w))
			//err = binary.Write(img, enc, uint32(Imagewidth))  //

			err = binary.Write(img, enc, uint16(tImageLength)) // Image length
			err = binary.Write(img, enc, uint16(dtLong))       // long
			err = binary.Write(img, enc, uint32(1))            // value
			err = binary.Write(img, enc, getuint32(nl_fr_h))
			// ImageLength,_ := strconv.Atoi(string(nl_fr_h))
			// err = binary.Write(img, enc, uint32(ImageLength))  //

			err = binary.Write(img, enc, uint16(tCompression)) //  Compression
			err = binary.Write(img, enc, uint16(dtShort))      //  short
			err = binary.Write(img, enc, uint32(1))            //  value
			err = binary.Write(img, enc, uint16(cG4))          //  CCITT Group 4
			err = binary.Write(img, enc, uint16(0))            //  CCITT Group 4

			err = binary.Write(img, enc, uint16(tPhotometricInterpretation)) //  Photometric
			err = binary.Write(img, enc, uint16(dtShort))                    //  short
			err = binary.Write(img, enc, uint32(1))                          //  value
			err = binary.Write(img, enc, uint32(0))                          //  white

			err = binary.Write(img, enc, uint16(tStripOffsets)) //  StripOffsets
			err = binary.Write(img, enc, uint16(dtLong))        //  long
			err = binary.Write(img, enc, uint32(1))             //  value
			err = binary.Write(img, enc, uint32(150))           //  0xA0

			err = binary.Write(img, enc, uint16(tOrientation)) // Orientation
			err = binary.Write(img, enc, uint16(dtShort))      //  short
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, getOrientation(rotation_code)) // rotation code
			err = binary.Write(img, enc, uint16(0))

			err = binary.Write(img, enc, uint16(tStripByteCounts)) //  StripbyteCounts
			err = binary.Write(img, enc, uint16(dtLong))           //  long
			err = binary.Write(img, enc, uint32(1))
			// fmt.Println( total_length)
			err = binary.Write(img, enc, uint32(total_length)) //  image size

			err = binary.Write(img, enc, uint16(tXResolution)) // Xresolution
			err = binary.Write(img, enc, uint16(dtRational))   // rational
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, uint32(xoffset)) //

			err = binary.Write(img, enc, uint16(tYResolution)) // Xresolution
			err = binary.Write(img, enc, uint16(dtRational))   //  Rational
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, uint32(yoffset)) //

			err = binary.Write(img, enc, uint16(tResolutionUnit)) // Xresolution
			err = binary.Write(img, enc, uint16(dtShort))         //  value
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, uint16(3)) //  cm
			err = binary.Write(img, enc, uint16(0))

			err = binary.Write(img, enc, uint32(0)) // next IFD = 0

			// Xresolution value
			err = binary.Write(img, enc, getuint32(nl_fr_w)*10)
			err = binary.Write(img, enc, getuint32(s_fr_w))

			// Yresoluton value
			err = binary.Write(img, enc, getuint32(nl_fr_h)*10)
			err = binary.Write(img, enc, getuint32(s_fr_h))

			//write images
			start := time.Now()

			var elapse time.Duration

			for r := 0; r < int(total_rec); r++ {
				l1 := ebc2asc.Ebc2asci(buf[k : k+5])
				long, _ = strconv.Atoi(string(l1))

				//buf1 = bytes.NewReader(buf[k+250:k+252])
				//err = binary.Read(buf1, Big, &imgl0)
				err = binary.Read(bytes.NewReader(buf[k+250:k+252]), Big, &imgl)
				//imgl := int(imgl0)
				img.Write(buf[k+252 : k+252+int(imgl)])
				k = k + long
				//fmt.Println(r,imgl, "img len=",img.Len())
			}

			// Structure to Json

			pagemeta := &Pagemeta{}
			pagemeta.DocumentID.CountryCode = pagmeta.Pub_office
			pagemeta.DocumentID.PatentNumber = pagmeta.Doc_id
			pagemeta.DocumentID.KindCode = pagmeta.Kc

			pagemeta.PublicationOffice = pagmeta.Pub_office
			pagemeta.PageNumber, _ = strconv.Atoi(pagmeta.Page_number)
			// pagemeta.PageIndicator = pagmeta.Content
			pagemeta.MultiMedia.Tiff = true
			pagemeta.TiffOffset.Start = 0
			pagemeta.TiffOffset.End = pagemeta.TiffOffset.Start + img.Len()
			/* wrtite tiff to files  */
			//docid1        := strings.TrimLeft(string(docid)," ")
			fulldocid := string(pub) + docid + string(kc)
			// goLog.Info.Println("FULL:", len(fulldocid), fulldocid, string(pagenum))
			if len(fulldocid) > 15 {
				var meta []byte
				err := json.Unmarshal(meta, &pagmeta)
				goLog.Warning.Println(k, meta)
				dump := "/tmp/dump_" + strconv.Itoa(pid)
				files.WriteFile(dump, st33, 0644)
				panic(err)
			}
			udirname_o := outputusermdDir + fulldocid
			tdirname_o := outputTiffDir + fulldocid
			if !combine {
				Check(files.MakeDir(udirname_o, 0755))
				Check(files.MakeDir(tdirname_o, 0755))
				/*  WRITE TIFF IMAGES USING WriteFile*/
				file_tiff_o := tdirname_o + string(os.PathSeparator) + string(pagenum) + ".tiff"
				Check(files.WriteFile(file_tiff_o, img.Bytes(), 0644))
				/* WRITE USERMD using Encode : From Structure to JSON File */
				file_usermd_o := udirname_o + string(os.PathSeparator) + string(pagenum) + ".json"
				if err := pagemeta.Encode(file_usermd_o); err != nil {
					goLog.Warning.Println(hostname, pid, err, "Encoding", *pagemeta)
				}
			} else {
				Check(files.MakeDir(tdirname_o, 0755))
				Page.Metadata = *pagemeta
				Page.Tiff.Size = img.Len()
				Page.Tiff.Image = img.Bytes()
				file_json_o := tdirname_o + string(os.PathSeparator) + string(pagenum) + ".json"

				// goLog.Info.Println(file_json_o, Page.Size)
				if err := Page.Encode(file_json_o); err != nil {
					goLog.Warning.Println(hostname, pid, err, "Encoding", *Page)
				}
			}
			elapse = time.Since(start)
			total += 1
			elapsetm += elapse

			/* UPDATE CONTAINER MD STRUCTURE*/
			// documentmeta.Date_drawup = pagmeta.Date_drawup
			documentmeta.DocumentID.CC = pagmeta.Pub_office
			documentmeta.DocumentID.PN = pagmeta.Doc_id
			documentmeta.DocumentID.KC = pagmeta.Kc
			documentmeta.PublicationDate = pagmeta.Pub_date
			documentmeta.PublicationOffice = pagmeta.Pub_office
			documentmeta.Multimedia.TIFF = true
			documentmeta.Multimedia.PNG = false

			file_containermd_o = outputContainerDir + string(pub) + docid + string(kc) + ".json"

			// reset img buffer for next image
			img.Reset()
			// get next images in the buffer
			l = k
			// goLog.Info.Printf("L=%d", l)
			// } // loop for *****************************************
			if pagmeta.Page_number == pagmeta.Total_pages {
				// Create the container metadata

				documentmeta.PageNumber = 0
				documentmeta.TotalPages, _ = strconv.Atoi(pagmeta.Total_pages)
				documentmeta.PublicationID = pagmeta.Doc_id
				if len(abs) > 0 {
					documentmeta.Abstract[0].Start = abs[0]
					documentmeta.Abstract[0].End = abs[len(abs)-1]
				}
				if len(bib) > 0 {
					documentmeta.Bibliography[0].Start = bib[0]
					documentmeta.Bibliography[0].End = bib[len(bib)-1]
				}
				if len(cla) > 0 {
					documentmeta.Claims[0].Start = cla[0]
					documentmeta.Claims[0].End = cla[len(cla)-1]
				}
				if len(desc) > 0 {
					documentmeta.Description[0].Start = desc[0]
					documentmeta.Description[0].End = desc[len(desc)-1]
				}
				if len(draw) > 0 {
					documentmeta.Drawings[0].Start = draw[0]
					documentmeta.Drawings[0].End = draw[len(draw)-1]
				}
				if len(srp) > 0 {
					documentmeta.SearchReport[0].Start = srp[0]
					documentmeta.SearchReport[0].End = srp[len(srp)-1]
				}
				if len(amd) > 0 {
					documentmeta.Amendment[0].Start = amd[0]
					documentmeta.Amendment[0].End = amd[len(amd)-1]
				}
				if err := documentmeta.Encode(file_containermd_o); err != nil {
					goLog.Warning.Println(hostname, pid, err, "Encoding", *pagemeta)
				}
				bib = bib[:0]
				abs = abs[:0]
				desc = desc[:0]
				cla = cla[:0]
				draw = draw[:0]
				amd = amd[:0]
				srp = srp[:0]
			} // end write metadata
		}
	} // end input file

	total1 := total * 1000000
	goLog.Info.Println("Elapsetm:", elapsetm, "Number of images:", total, "Average ms:", float64(elapsetm)/float64(total1))
	// goLog.Info.Printf("%s %d %5.2f|\n", "average ms=", total, float64(totaltime)/float64(total))
	return err
}

func ST33toFiles_p(inputFile string, outputusermdDir string, outputTiffDir string, outputContainerDir string, combine bool) error {

	//   EXTRACT ST33 Files
	//
	//  FOR EACH DOCUMENT in ST33 {
	// 		CREATE  page metadata			 ( ST33 HEADER  ++)
	// 		CREATE  page data  (Tiff)      ( ST33 TIFF RECORDS )
	//
	// 		FOR EACH PAGE of a document {
	// 			 if combine  {
	//      		  COMBINE  page data (TIFF) and metadata => PAGE Struct
	//      		  WRITE PAGE struct
	//       	 }
	//  		else {
	//     		  	  WRITE data ( TIFF)
	//	     		  WRITE Metadata  ( user metadata)
	//       	}
	//      }
	//      CREATE DOCUMENT metadata
	//      WRITE DOCUMENT  metadata
	//  }
	//
	goLog.Info.Println("Combine output:", combine)

	Little := binary.LittleEndian
	Big := binary.BigEndian
	enc := Big
	documentmeta := &Documentmeta{}

	var (
		// container          = make([]string, 0, 500)
		abs                = make([]int, 0, 100)
		bib                = make([]int, 0, 100)
		cla                = make([]int, 0, 100)
		desc               = make([]int, 0, 300)
		draw               = make([]int, 0, 300)
		amd                = make([]int, 0, 300)
		srp                = make([]int, 0, 300)
		file_containermd_o string
		elapsetm           time.Duration
		total_rec          uint16
		total_length       uint32
		// documentmd         []byte
	)
	total := 0
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	//buf,err := ioutil.ReadFile(inputFile)
	abuf, err := files.ReadBuffer(inputFile)
	if err == nil {

		defer abuf.Reset()

		buf := abuf.Bytes()
		bufl := len(buf)
		pagmeta := &Pagmeta{}

		l := 0

		for l < bufl {
			k := l

			var (
				recs uint16
				//total_rec    uint16
				//total_length uint32
				imgl uint16
			)
			buf1 := bytes.NewReader(buf[k+25 : k+27])
			err = binary.Read(buf1, Big, &recs)
			buf1 = bytes.NewReader(buf[k+84 : k+86])
			err = binary.Read(buf1, Big, &total_rec)
			buf1 = bytes.NewReader(buf[k+214 : k+218])
			err = binary.Read(buf1, Big, &total_length)
			buf1 = bytes.NewReader(buf[k+250 : k+252])
			err = binary.Read(buf1, Big, &imgl)

			// convert  Buf ( EBCDIC) to ST33 (ASCII)
			st33 := ebc2asc.Ebc2asci(buf[l : l+214])

			long, _ := strconv.Atoi(string(st33[0:5]))

			pub := st33[5:7]
			pagmeta.Pub_office = string(pub)
			kc := strings.Trim(string(st33[7:9]), " ")
			pagmeta.Kc = kc
			docnum := st33[9:17]

			pagenum := st33[17:21]

			pagmeta.Page_number = string(pagenum)

			pos9_ := st33[27:29]
			docid0 := strings.Trim(string(docnum)+string(pos9_), " ")
			docid := strings.TrimLeft(string(st33[33:45]), " ")

			if len(docid) == 0 {
				docid = docid0
			}

			pagmeta.Doc_id = docid

			o_pub := st33[67:69]
			pagmeta.O_pub = string(o_pub)

			total_pages := st33[76:80]
			pagmeta.Total_pages = string(total_pages)

			f_date_drawup := st33[93:101]
			pagmeta.Date_drawup = string(f_date_drawup)
			f_pub_date := st33[101:109]
			pagmeta.Pub_date = string(f_pub_date)
			Biblio := string(st33[133:134])
			Claim := string(st33[134:135])
			Drawing := string(st33[135:136])
			Amendement := string(st33[136:137])
			Description := string(st33[137:138])
			Abstract := string(st33[138:139])
			Search_report := string(st33[139:140])

			pn, _ := strconv.Atoi(string(pagenum))
			if Biblio == "1" {

				bib = append(bib, pn)
			}
			if Abstract == "1" {
				abs = append(abs, pn)
			}
			if Claim == "1" {
				cla = append(cla, pn)
			}
			if Drawing == "1" {
				draw = append(draw, pn)
			}
			if Amendement == "1" {
				amd = append(amd, pn)
			}
			if Description == "1" {
				desc = append(desc, pn)
			}
			if Search_report == "1" {
				srp = append(srp, pn)
			}

			// container = append(container, content.String())

			/* +40*/
			// pagmeta.Content = content.String()
			data_type := string(st33[180:181])
			pagmeta.Data_type = data_type
			// comp_meth := st33[181:183]
			// k_fac := st33[183:185]
			// Resolution := st33[185:187]
			s_fr_h := st33[187:190]
			s_fr_w := st33[190:193]
			nl_fr_h := st33[193:197]
			nl_fr_w := st33[197:201]
			rotation_code := st33[201:202]
			// fr_x := st33[202:206]
			// fr_y := st33[206:210]
			// fr_stat := st33[210:211]
			version := st33[211:214]

			// Coniunue with Buffer
			buf1 = bytes.NewReader(buf[k+214 : k+218])
			if string(version) == "V30" {

				// buf1 = bytes.NewReader(buf[k+214 : k+218])
				// some V30 total_length are encoded with big Endian byte order
				err = binary.Read(buf1, Little, &total_length)
				if int(total_length) > 16777215 {
					buf1 = bytes.NewReader(buf[k+214 : k+218])
					err = binary.Read(buf1, Big, &total_length)
				}

			} else {
				err = binary.Read(buf1, Big, &total_length)
			}

			/* write tiff images */
			var (
				Page = new(PAGE)
				img  = new(bytes.Buffer)
			)

			_, err = io.WriteString(img, beHeader)          // magic number
			err = binary.Write(img, enc, uint32(ifdOffset)) // IFD offset
			err = binary.Write(img, enc, uint16(ifdLen))    // number of IFD entries

			err = binary.Write(img, enc, uint16(tImageWidth)) //  image Width
			err = binary.Write(img, enc, uint16(dtLong))      //  long
			err = binary.Write(img, enc, uint32(1))           //  value
			err = binary.Write(img, enc, getuint32(nl_fr_w))

			//Imagewidth,_ := strconv.Atoi(string(nl_fr_w))
			//err = binary.Write(img, enc, uint32(Imagewidth))  //

			err = binary.Write(img, enc, uint16(tImageLength)) // Image length
			err = binary.Write(img, enc, uint16(dtLong))       // long
			err = binary.Write(img, enc, uint32(1))            // value
			err = binary.Write(img, enc, getuint32(nl_fr_h))
			// ImageLength,_ := strconv.Atoi(string(nl_fr_h))
			// err = binary.Write(img, enc, uint32(ImageLength))  //

			err = binary.Write(img, enc, uint16(tCompression)) //  Compression
			err = binary.Write(img, enc, uint16(dtShort))      //  short
			err = binary.Write(img, enc, uint32(1))            //  value
			err = binary.Write(img, enc, uint16(cG4))          //  CCITT Group 4
			err = binary.Write(img, enc, uint16(0))            //  CCITT Group 4

			err = binary.Write(img, enc, uint16(tPhotometricInterpretation)) //  Photometric
			err = binary.Write(img, enc, uint16(dtShort))                    //  short
			err = binary.Write(img, enc, uint32(1))                          //  value
			err = binary.Write(img, enc, uint32(0))                          //  white

			err = binary.Write(img, enc, uint16(tStripOffsets)) //  StripOffsets
			err = binary.Write(img, enc, uint16(dtLong))        //  long
			err = binary.Write(img, enc, uint32(1))             //  value
			err = binary.Write(img, enc, uint32(150))           //  0xA0

			err = binary.Write(img, enc, uint16(tOrientation)) // Orientation
			err = binary.Write(img, enc, uint16(dtShort))      //  short
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, getOrientation(rotation_code)) // rotation code
			err = binary.Write(img, enc, uint16(0))

			err = binary.Write(img, enc, uint16(tStripByteCounts)) //  StripbyteCounts
			err = binary.Write(img, enc, uint16(dtLong))           //  long
			err = binary.Write(img, enc, uint32(1))
			// fmt.Println( total_length)
			err = binary.Write(img, enc, uint32(total_length)) //  image size

			err = binary.Write(img, enc, uint16(tXResolution)) // Xresolution
			err = binary.Write(img, enc, uint16(dtRational))   // rational
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, uint32(xoffset)) //

			err = binary.Write(img, enc, uint16(tYResolution)) // Xresolution
			err = binary.Write(img, enc, uint16(dtRational))   //  Rational
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, uint32(yoffset)) //

			err = binary.Write(img, enc, uint16(tResolutionUnit)) // Xresolution
			err = binary.Write(img, enc, uint16(dtShort))         //  value
			err = binary.Write(img, enc, uint32(1))
			err = binary.Write(img, enc, uint16(3)) //  cm
			err = binary.Write(img, enc, uint16(0))

			err = binary.Write(img, enc, uint32(0)) // next IFD = 0

			// Xresolution value
			err = binary.Write(img, enc, getuint32(nl_fr_w)*10)
			err = binary.Write(img, enc, getuint32(s_fr_w))

			// Yresoluton value
			err = binary.Write(img, enc, getuint32(nl_fr_h)*10)
			err = binary.Write(img, enc, getuint32(s_fr_h))

			//write images
			start := time.Now()

			var elapse time.Duration

			for r := 0; r < int(total_rec); r++ {
				l1 := ebc2asc.Ebc2asci(buf[k : k+5])
				long, _ = strconv.Atoi(string(l1))

				//buf1 = bytes.NewReader(buf[k+250:k+252])
				//err = binary.Read(buf1, Big, &imgl0)
				err = binary.Read(bytes.NewReader(buf[k+250:k+252]), Big, &imgl)
				//imgl := int(imgl0)
				img.Write(buf[k+252 : k+252+int(imgl)])
				k = k + long
				//fmt.Println(r,imgl, "img len=",img.Len())
			}

			// Convert Structure to Json

			pagemeta := &Pagemeta{}
			pagemeta.DocumentID.CountryCode = pagmeta.Pub_office
			pagemeta.DocumentID.PatentNumber = pagmeta.Doc_id
			pagemeta.DocumentID.CountryCode = pagmeta.Kc

			pagemeta.PublicationOffice = pagmeta.Pub_office
			pagemeta.PageNumber, _ = strconv.Atoi(pagmeta.Page_number)
			// pagemeta.PageIndicator = pagmeta.Content
			pagemeta.MultiMedia.Tiff = true
			pagemeta.TiffOffset.Start = 0
			pagemeta.TiffOffset.End = pagemeta.TiffOffset.Start + img.Len()
			/* wrtite tiff to files  */
			//docid1        := strings.TrimLeft(string(docid)," ")
			fulldocid := string(pub) + docid + string(kc)
			// goLog.Info.Println("FULL:", len(fulldocid), fulldocid, string(pagenum))

			if len(fulldocid) > 15 {
				var meta []byte
				err := json.Unmarshal(meta, &pagmeta)
				goLog.Warning.Println(k, meta)
				dump := "/tmp/dump_" + strconv.Itoa(pid)
				files.WriteFile(dump, st33, 0644)
				panic(err)
			}

			udirname_o := outputusermdDir + fulldocid
			tdirname_o := outputTiffDir + fulldocid
			if !combine {
				Check(files.MakeDir(udirname_o, 0755))
				Check(files.MakeDir(tdirname_o, 0755))
				/*  WRITE TIFF IMAGES USING WriteFile*/
				file_tiff_o := tdirname_o + string(os.PathSeparator) + string(pagenum) + ".tiff"
				Check(files.WriteFile(file_tiff_o, img.Bytes(), 0644))
				/* WRITE USERMD using Encode : From Structure to JSON File */
				file_usermd_o := udirname_o + string(os.PathSeparator) + string(pagenum) + ".json"
				if err := pagemeta.Encode(file_usermd_o); err != nil {
					goLog.Warning.Println(hostname, pid, err, "Encoding", *pagemeta)
				}
			} else {
				Check(files.MakeDir(tdirname_o, 0755))
				Page.Metadata = *pagemeta
				Page.Tiff.Size = img.Len()
				Page.Tiff.Image = img.Bytes()
				file_json_o := tdirname_o + string(os.PathSeparator) + string(pagenum) + ".json"
				// goLog.Info.Println(file_json_o, Page.Size)
				if err := Page.Encode(file_json_o); err != nil {
					goLog.Warning.Println(hostname, pid, err, "Encoding", *Page)
				}
			}
			elapse = time.Since(start)
			total += 1
			elapsetm += elapse

			/* UPDATE CONTAINER MD STRUCTURE*/
			// documentmeta.Date_drawup = pagmeta.Date_drawup
			documentmeta.DocumentID.CC = pagmeta.Pub_office
			documentmeta.DocumentID.PN = pagmeta.Doc_id
			documentmeta.DocumentID.KC = pagmeta.Kc
			documentmeta.PublicationDate = pagmeta.Pub_date
			documentmeta.PublicationOffice = pagmeta.Pub_office
			documentmeta.Multimedia.TIFF = true
			documentmeta.Multimedia.PNG = false

			file_containermd_o = outputContainerDir + string(pub) + docid + string(kc) + ".json"

			// reset img buffer for next image
			img.Reset()
			// get next images in the buffer
			l = k
			// goLog.Info.Printf("L=%d", l)
			// } // loop for *****************************************
			if pagmeta.Page_number == pagmeta.Total_pages {
				// Create the container metadata

				documentmeta.PageNumber = 0
				documentmeta.TotalPages, _ = strconv.Atoi(pagmeta.Total_pages)
				documentmeta.PublicationID = pagmeta.Doc_id
				if len(abs) > 0 {
					documentmeta.Abstract[0].Start = abs[0]
					documentmeta.Abstract[0].End = abs[len(abs)-1]
				}
				if len(bib) > 0 {
					documentmeta.Bibliography[0].Start = bib[0]
					documentmeta.Bibliography[0].End = bib[len(bib)-1]
				}
				if len(cla) > 0 {
					documentmeta.Claims[0].Start = cla[0]
					documentmeta.Claims[0].End = cla[len(cla)-1]
				}
				if len(desc) > 0 {
					documentmeta.Description[0].Start = desc[0]
					documentmeta.Description[0].End = desc[len(desc)-1]
				}
				if len(draw) > 0 {
					documentmeta.Drawings[0].Start = draw[0]
					documentmeta.Drawings[0].End = draw[len(draw)-1]
				}
				if len(srp) > 0 {
					documentmeta.SearchReport[0].Start = srp[0]
					documentmeta.SearchReport[0].End = srp[len(srp)-1]
				}
				if len(amd) > 0 {
					documentmeta.Amendment[0].Start = amd[0]
					documentmeta.Amendment[0].End = amd[len(amd)-1]
				}
				if err := documentmeta.Encode(file_containermd_o); err != nil {
					goLog.Warning.Println(hostname, pid, err, "Encoding", *pagemeta)
				}
				bib = bib[:0]
				abs = abs[:0]
				desc = desc[:0]
				cla = cla[:0]
				draw = draw[:0]
				amd = amd[:0]
				srp = srp[:0]
			} // end write metadata
		}
	} // end input file

	total1 := total * 1000000
	goLog.Info.Println("Elapsetm:", elapsetm, "Number of images:", total, "Average ms:", float64(elapsetm)/float64(total1))
	// goLog.Info.Printf("%s %d %5.2f|\n", "average ms=", total, float64(totaltime)/float64(total))
	return err
}

func AsyncHttpGets(urls []string, getHeader map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	treq := 0
	fmt.Printf("\n")
	for _, url := range urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			// fmt.Printf("Fetching %s \n", url)
			client := &http.Client{}
			//start := time.Now()
			//var elapse time.Duration
			resp, err := sproxyd.GetObject(client, url, getHeader)
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			} else {

				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, len(body), err}

		}(url)
	}
	// wait for http response  message
	for {
		select {
		case r := <-ch:
			// fmt.Printf("%s was fetched\n", r.url)
			responses = append(responses, r)
			if len(responses) == treq /*len(urls)*/ {
				return responses
			}
		case <-time.After(100 * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpGetMetadatas(urls []string, getHeader map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	treq := 0
	fmt.Printf("\n")
	for _, url := range urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			// fmt.Printf("Fetching %s \n", url)
			client := &http.Client{}
			//start := time.Now()
			//var elapse time.Duration
			resp, err := sproxyd.GetMetadata(client, url, getHeader)
			if err != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, 0, err}

		}(url)
	}
	// wait for http response  message
	for {
		select {
		case r := <-ch:
			// fmt.Printf("%s was fetched\n", r.url)
			responses = append(responses, r)
			if len(responses) == treq /*len(urls)*/ {
				return responses
			}
		case <-time.After(100 * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpPuts(urls []string, bufa [][]byte, headera []map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0

	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			clientw := &http.Client{}
			resp, err = sproxyd.PutObject(clientw, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpPut2s(urls []string, bufa [][]byte, bufb [][]byte, headera []map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0

	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			clientw := &http.Client{}
			resp, err = sproxyd.PutObject(clientw, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpUpdates(urls []string, bufa [][]byte, headera []map[string]string) []*sproxyd.HttpResponse {

	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0

	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			clientw := &http.Client{}
			resp, err = sproxyd.UpdObject(clientw, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpUpdMetadatas(meta string, urls []string, headera []map[string]string) []*sproxyd.HttpResponse {
	// if meta == "Page"
	// Update meta data read from a File
	// TODO Update meta data reda from the Ring
	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0

	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var (
				pagmeta Pagemeta // OLD METADATA
				// usermd  []byte
				err  error
				resp *http.Response
			)
			clientw := &http.Client{}
			um, _ := base64.Decode64(headera[k]["Usermd"])
			if err = json.Unmarshal(um, &pagmeta); err == nil {
				// SET NEW METATA HERE
				// pmd := pagmeta.ToPagemeta()
				//	if usermd, err = json.Marshal(&pmd); err == nil {
				//	headera[k]["Usermd"] = base64.Encode64(usermd)
				//	}
			}
			resp, err = sproxyd.UpdMetadata(clientw, url, headera[k])

			if resp != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func AsyncHttpDeletes(urls []string) []*sproxyd.HttpResponse {
	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}
	treq := 0

	for _, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			clientw := &http.Client{}
			resp, err = sproxyd.DeleteObject(clientw, url)
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, 0, err}
		}(url)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(sproxyd.Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses

}

func ParseDate(str string) (dd Date, err error) {
	str = strings.TrimSpace(str)
	var (
		y, m, d int
	)
	if len(str) != 8 {
		goto invalid
	}
	if y, err = strconv.Atoi(str[0:4]); err != nil {
		return
	}
	if m, err = strconv.Atoi(str[4:6]); err != nil {
		return
	}
	if m < 1 || m > 12 {
		goto invalid
	}
	if d, err = strconv.Atoi(str[6:8]); err != nil {
		return
	}
	if d < 1 || d > 31 {
		goto invalid
	}
	dd.Year = int16(y)
	dd.Month = byte(m)
	dd.Day = byte(d)
	return
invalid:
	err = errors.New("Invalid metadata Date string: " + str)
	return
}

func (dd Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", dd.Year, dd.Month, dd.Day)
}

func noDate() Date {
	//return Date{int16(0),byte(0),byte(0)}
	return Date{}
}

func getuint16(in []byte) uint16 {
	out, _ := strconv.Atoi(string(in))
	return uint16(out)

}

func getuint32(in []byte) uint32 {
	out, _ := strconv.Atoi(string(in))
	return uint32(out)

}

func getConfig(configfile string) (Configuration, error) {
	cfile, err := os.Open(configfile)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(cfile)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	_ = cfile.Close()
	return configuration, err
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

/* image orientation */

func getOrientation(rotation_code []byte) uint16 {
	orientation, _ := strconv.Atoi(string(rotation_code))
	switch orientation {
	case 1:
		return uint16(1)
	case 2:
		return uint16(6)
	case 3:
		return uint16(3)
	case 4:
		return uint16(8)
	default:
		return uint16(1)
	}
}

func Tiff2Png(tiffile, pngfile string) error {
	// cmd := exec.Command("convert", "-resize", "950x", tiffile, pngfile)
	cmd := exec.Command("convert", tiffile, pngfile)
	return cmd.Run()

}

func RemoveSlash(input string) string {
	output := ""
	ar := strings.Split(input, "/")
	for _, word := range ar {
		output = output + word
	}
	return output
}
