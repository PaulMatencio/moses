// bns project bns.go
package poc

import (
	"fmt"
	"io"
	/* "flag" */
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"user/base64"
	"user/ebc2asc"
	"user/files"
	"user/sproxyd"
)

type Configuration struct {
	//Input_directory string
	Storage_nodes []string
	//Output_tiff     string
	//Output_json     string
}

type bnsImages struct {
	Image  []byte
	Usermd map[string]string
	Index  int
}

type Date struct {
	Year       int16
	Month, Day byte
}

/*   Date structure */
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

func check(e error) {
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

/*  read buffer */
func ReadBuffer(InputFile string) (*bytes.Buffer, error) {
	fp, e := os.Open(InputFile)
	if e == nil {
		defer fp.Close()
		fi, e := fp.Stat()
		var n int64
		n = fi.Size() + bytes.MinRead
		buf := make([]byte, 0, n)
		buffer := bytes.NewBuffer(buf)
		_, e = buffer.ReadFrom(fp)
		return buffer, e
	} else {
		return nil, e
	}
}

/*
func ReadPage(Inputdir string) (*bytes.Buffer,map[string]interface{} ){

 	var img bytes.Buffer
  	var usermd map[string]interface{}
	return &img,usermd
}
*/

func PutPage(client *http.Client, url string, img *bytes.Buffer, usermd map[string]string) error {

	encoded_usermd, _ := base64.Encode64(usermd)
	putheader := map[string]string{
		"Usermd":       encoded_usermd,
		"Content-Type": "image/tiff",
	}
	// url:= base_url+string(pub)+docid+string(kc)+"_"+string(pagenum)
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	if resp, err = sproxyd.PutObject(client, url, img.Bytes(), putheader); err != nil {
		fmt.Println(err)
	} else {
		switch resp.StatusCode {
		case 412:
			fmt.Println(resp.Status, url, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")
		case 422:
			fmt.Println(resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			fmt.Println(url, resp.Status, time.Since(start))
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err
}

func GetPage(client *http.Client, url string) (*http.Response, error) {

	getHeader := map[string]string{}
	// no specific request , just give me an object
	resp, err := sproxyd.GetObject(client, url, getHeader)
	return resp, err
}

func UpdatePage(client *http.Client, url string, img *bytes.Buffer, usermd map[string]string) error {

	encoded_usermd, _ := base64.Encode64(usermd)
	putheader := map[string]string{
		"Usermd":       encoded_usermd,
		"Content-Type": "image/tiff",
	}
	// url:= base_url+string(pub)+docid+string(kc)+"_"+string(pagenum)
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	// defer resp.Body.Close()
	if resp, err = sproxyd.UpdObject(client, url, img.Bytes(), putheader); err != nil {
		fmt.Println(err)
	} else {

		switch resp.StatusCode {
		case 200:
			fmt.Println("OK", url, resp.Header["X-Scal-Ring-Key"], time.Since(start))
		case 404:
			fmt.Println(resp.Status, url, " not found")
		case 412:
			fmt.Println(resp.Status, url, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
		case 422:
			fmt.Println(resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			fmt.Println(url, resp.Status, time.Since(start))
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err

}

func DeletePage(client *http.Client, url string) error {

	deleteHeader := map[string]string{}
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	// defer resp.Body.Close()
	if resp, err = sproxyd.DeleteObject(client, url, deleteHeader); err != nil {
		fmt.Println(err)
	} else {

		switch resp.StatusCode {
		case 200:
			fmt.Println("OK", url, resp.Header["X-Scal-Ring-Key"], time.Since(start))
		case 404:
			fmt.Println(resp.StatusCode, url, " not found")
		case 412:
			fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
		case 422:
			fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
		default:
			fmt.Println(url, resp.Status, time.Since(start))
		}
		resp.Body.Close()
	}

	return err
}

func UpdMetadata(client *http.Client, url string, usermd map[string]string) error {
	encoded_usermd, _ := base64.Encode64(usermd)
	updheader := map[string]string{
		"Usermd":       encoded_usermd,
		"Content-Type": "image/tiff",
	}
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	// defer resp.Body.Close()
	if resp, err = sproxyd.UpdMetadata(client, url, updheader); err != nil {
		fmt.Println(err)
	} else {

		switch resp.StatusCode {
		case 200:
			fmt.Println("OK", url, resp.Header["X-Scal-Ring-Key"], time.Since(start))
		case 404:
			fmt.Println(resp.Status, url, " not found")
		case 412:
			fmt.Println(resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
		case 422:
			fmt.Println(resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			fmt.Println(url, resp.Status, time.Since(start))
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err

}

// Get former ST33 Header ++
func GetPageMetadata(url string) (map[string]interface{}, error) {
	return GetDocMetadata(url)
}

// Get  former BNS directory ++
func GetDocMetadata(url string) (map[string]interface{}, error) {
	return GetMetadata(url)
}

// Get  user metadata
func GetMetadata(url string) (map[string]interface{}, error) {
	client := &http.Client{}
	getHeader := map[string]string{}
	var usermd map[string]interface{}
	var resp *http.Response
	err := error(nil)
	if resp, err = sproxyd.GetMetadata(client, url, getHeader); err == nil {
		switch resp.StatusCode {
		case 200:
			encoded_usermd := resp.Header["X-Scal-Usermd"]
			usermd, err = base64.Decode64(encoded_usermd[0])
		case 404:
			fmt.Println(resp.Status, url)
		case 412:
			fmt.Println(resp.Status, url, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
		case 422:
			fmt.Println(resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			fmt.Println(url, resp.Status)
		}
	}
	/* the resp,Body is closed by sproxyd.getMetadata */
	return usermd, err
}

// Get total number of pages of a document
func GetPageNumber(usermd map[string]interface{}) (int, error) {

	if total_pages, ok := usermd["total_pages"]; ok {
		t_pages, err := strconv.Atoi(total_pages.(string))
		return t_pages, err
	} else {
		return 0, errors.New("Invalid user metadata")
	}

}
func GetPageNumber_1(usermd map[string]string) (int, error) {
	if total_pages, ok := usermd["total_pages"]; ok {
		t_pages, err := strconv.Atoi(total_pages)
		return t_pages, err
	} else {
		return 0, errors.New("Invalid user metadata")
	}
}

// Get  the publication date of a document
func GetPubDate(usermd map[string]interface{}) (Date, error) {
	date := Date{}
	err := error(nil)
	if pub_date, ok := usermd["Pub_date"]; ok {
		date, err = ParseDate(pub_date.(string))
	} else {
		err = errors.New("no Publication date")
	}
	return date, err
}

//  Get the Draw_up date of a document
func GetDrawUpDate(usermd map[string]interface{}) (Date, error) {
	date := Date{}
	err := error(nil)
	if drawup_date, ok := usermd["date_drawup"]; ok {
		date, err = ParseDate(drawup_date.(string))
	} else {
		err = errors.New("no Draw-up date")
	}
	return date, err
}

// get the layout of a Document  ( Bibl, Claim, Desc, etc ...)
func GetContent(usermd map[string]interface{}) (string, error) {
	var layout string
	err := error(nil)
	if content, ok := usermd["content"]; ok {
		layout = content.(string)
	} else {
		err = errors.New("no Content layout")
	}
	return layout, err
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

func ST33toTiff(action string, base_url string, inputFile string, outputusermdDir string, outputTiffDir string, outputContainerDir string, bns bool) error {

	// fmt.Println(action,base_url,bns)

	Little := binary.LittleEndian
	Big := binary.BigEndian
	enc := Big
	containermd := map[string]string{}

	var (
		file_containermd_o string
		//var container bytes.Buffer
		container = make([]string, 0, 200)
		url       string
		curl      string
		totaltime time.Duration
	)
	u := 0
	total := 0
	client := &http.Client{}
	//buf,err := ioutil.ReadFile(inputFile)
	abuf, err := ReadBuffer(inputFile)
	if err == nil {
		defer abuf.Reset()
		buf := abuf.Bytes()
		bufl := len(buf)
		usermd := map[string]string{}
		l := 0

		for l < bufl {
			k := l
			var (
				recs         uint16
				total_rec    uint16
				total_length uint32
				imgl         uint16
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
			usermd["pub_office"] = string(pub)
			kc := strings.Trim(string(st33[7:9]), " ")
			usermd["kc"] = kc
			docnum := st33[9:17]

			pagenum := st33[17:21]
			usermd["page_number"] = string(pagenum)
			// framenum := st33[21:25]
			//recs := byte2int(buf[k+25:k+27])

			//fmt.Println("recs",int(recs))
			pos9_ := st33[27:29]
			docid0 := strings.Trim(string(docnum)+string(pos9_), " ")
			docid := strings.TrimLeft(string(st33[33:45]), " ")

			if len(docid) == 0 {
				docid = docid0
			}

			usermd["doc_id"] = docid
			//usermd["doc_id"] = strings.TrimLeft(string(docid)," ")
			o_pub := st33[67:69]
			usermd["o_pub"] = string(o_pub)
			//date_drawup := st33[69:75]
			//usermd["date_drawup"]= string(date_drawup)
			//rec_stat :=  st33[75:76]
			total_pages := st33[76:80]
			usermd["total_pages"] = string(total_pages)
			/*
			   s_doc_h := st33[87:90]
			   s_doc_w := st33[90:93]
			*/
			f_date_drawup := st33[93:101]
			usermd["Date_drawup"] = string(f_date_drawup)
			f_pub_date := st33[101:109]
			usermd["Pub_date"] = string(f_pub_date)
			Biblio := string(st33[133:134])
			Claim := string(st33[134:135])
			Drawing := string(st33[135:136])
			Amendement := string(st33[136:137])
			Description := string(st33[137:138])
			Abstract := string(st33[138:139])
			Search_report := string(st33[139:140])

			var content bytes.Buffer

			if Biblio == "1" {
				content.WriteString(Bib)
			}
			if Abstract == "1" {
				content.WriteString(Abs)
			}
			if Claim == "1" {
				content.WriteString(Cla)
			}
			if Drawing == "1" {
				content.WriteString(Dra)
			}
			if Amendement == "1" {
				content.WriteString(Amd)
			}
			if Description == "1" {
				content.WriteString(Des)
			}
			if Search_report == "1" {
				content.WriteString(Srp)
			}
			container = append(container, content.String())
			/* +40*/
			usermd["content"] = content.String()
			data_type := string(st33[180:181])
			usermd["data_type"] = data_type
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
			var img = new(bytes.Buffer)

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

			if !bns {
				target := strings.Split(base_url, "/")
				src := target[len(target)-2]
				base_url1 := sproxyd.Host[u] + src + "/"
				u += 1
				if u == 3 {
					u = 0
				}
				url = base_url1 + string(pub) + docid + string(kc) + "/" + string(pagenum)
			} else {
				url = base_url + string(pub) + docid + string(kc) + "/" + string(pagenum)
			}

			switch action {
			case "Test":
				fmt.Println(url)
			case "PutObject":
				if err := PutPage(client, url, img, usermd); err != nil {
					fmt.Println(err)
				}
			case "CreateObject":
				if err := PutPage(client, url, img, usermd); err != nil {
					fmt.Println(err)
				}
			case "UpdateObject":
				if err := UpdatePage(client, url, img, usermd); err != nil {
					fmt.Println(err)
				}
			case "UpdateMetadata":
				if err := UpdMetadata(client, url, usermd); err != nil {
					fmt.Println(err)
				}
			case "DeleteObject":
				if err := DeletePage(client, url); err != nil {
					fmt.Println(err)
				}
			case "CreateFile":
				/* wrtite tiff to files  */
				//docid1        := strings.TrimLeft(string(docid)," ")
				fulldocid := string(pub) + docid + string(kc)
				udirname_o := outputusermdDir + fulldocid
				tdirname_o := outputTiffDir + fulldocid

				if !files.Exist(udirname_o) {
					if err := os.Mkdir(udirname_o, 0755); err != nil {
						fmt.Println(err)
					}
				}
				if !files.Exist(tdirname_o) {
					if err := os.Mkdir(tdirname_o, 0755); err != nil {
						fmt.Println(err)
					}

				}
				usermdB, _ := json.Marshal(usermd)

				file_usermd_o := udirname_o + string(os.PathSeparator) + string(pagenum) + ".json"
				file_tiff_o := tdirname_o + string(os.PathSeparator) + string(pagenum) + ".tiff"

				err = ioutil.WriteFile(file_tiff_o, img.Bytes(), 0644)
				check(err)
				err = ioutil.WriteFile(file_usermd_o, usermdB, 0644)
				check(err)

			default:
				fmt.Println("Wrong action! -a  Should be <CreateObject/DeleteObject/UpdateObject/UpdateMetadata/CreateFile/Test/GetUrl>")
				os.Exit(4)
			}
			total += 1

			/* update container md */
			for k, v := range usermd {
				containermd[k] = v
			}
			file_containermd_o = outputContainerDir + string(pub) + docid + string(kc) + ".json"
			curl = base_url + string(pub) + docid + string(kc)
			// reset img buffer for next image
			img.Reset()
			// get next images in the buffer
			l = k
		} // loop for
	} // end of inputfile

	if bns {
		// Create container metadata
		if action == "PutObject" || action == "CreateObject" || action == "UpdateObject" || action == "UpdateMetadata" || action == "CreateFile" {

			containermd["page_number"] = "0000"
			var container1 bytes.Buffer
			for _, v := range container {
				container1.WriteString(v)
				container1.WriteString(",")
			}
			containermd["content"] = container1.String()
			containermdB, _ := json.Marshal(containermd)

			// Always WriteFile the container metadata to a file
			err = ioutil.WriteFile(file_containermd_o, containermdB, 0644)
			check(err)
		}
		// PutObject  the container Metadata
		switch action {

		case "PutObject":

			img := new(bytes.Buffer) // EMPTY BUFFER
			if err := PutPage(client, curl, img, containermd); err != nil {
				fmt.Println(err)
			}
		case "CreateObject":

			img := new(bytes.Buffer) // EMPTY BUFFER
			if err := PutPage(client, curl, img, containermd); err != nil {
				fmt.Println(err)
			}
		case "UpdateObject":

			img := new(bytes.Buffer) // EMPTY BUFFER
			if err := UpdatePage(client, curl, img, containermd); err != nil {
				fmt.Println(err)
			}
		case "DeleteObject":
			// Delete the container metadata
			if err := DeletePage(client, curl); err != nil {
				fmt.Println(err)
			}

		default:
			{ // fmt.Println("CreateFile: Do nothing")
			}

		}
	} // end bns
	fmt.Printf("%s %d %5.2f|\n", "average ms=", total, float64(totaltime)/float64(total))
	return err
}
