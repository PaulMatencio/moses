package bns

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

	//   REMOVED  =>  Check old bns for how to do
	error := errors.New("Function has been removed")
	return error
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

	//   REMOVED  =>  Check old bns for how to do
	error := errors.New("Function has been removed")
	return error
}

func AsyncHttpGetPage(urls []string, getHeader map[string]string) []*sproxyd.HttpResponse {

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
			resp, err := sproxyd.GetObject(client, url, getHeader)
			// resp, err := GetPage(client, url, getHeader)
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			} else {

				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, &body, err}

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

// func AsyncHttpGetPageType(urls []string, getHeader map[string]string) []*sproxyd.HttpResponse {
func AsyncHttpGetPageType(urls []string, media string) []*sproxyd.HttpResponse {
	ch := make(chan *sproxyd.HttpResponse)
	responses := []*sproxyd.HttpResponse{}

	// fmt.Println(urls)
	treq := 0
	fmt.Printf("\n")
	client := &http.Client{}
	for _, url := range urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}

		go func(url string) {
			// client := &http.Client{}
			// fmt.Printf("fetching %s\n", url)

			// resp, err := GetPageType(client, url, getHeader)
			resp, err := GetPageType(client, url, media)
			defer resp.Body.Close()
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
			} /*else {
				if resp != nil { // resp == nil when there is no media type in the metadata
					resp.Body.Close()
				}
			} */
			ch <- &sproxyd.HttpResponse{url, resp, &body, err}

		}(url)
	}
	// wait for http response  message
	for {
		select {
		case r := <-ch:
			// fmt.Printf("%s was fetched\n", r.Url)
			responses = append(responses, r)
			if len(responses) == treq /*len(urls)*/ {
				// fmt.Println(responses)
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
	client := &http.Client{}
	for _, url := range urls {
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			// fmt.Printf("Fetching %s \n", url)

			// client := &http.Client{}

			//start := time.Now()
			//var elapse time.Duration
			resp, err := sproxyd.GetMetadata(client, url, getHeader)
			if err != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, nil, err}

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
	clientw := &http.Client{}
	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			// clientw := &http.Client{}
			resp, err = sproxyd.PutObject(clientw, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}
			ch <- &sproxyd.HttpResponse{url, resp, nil, err}
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
	clientw := &http.Client{}
	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			// clientw := &http.Client{}
			resp, err = sproxyd.PutObject(clientw, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, nil, err}
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
	clientw := &http.Client{}
	for k, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			// clientw := &http.Client{}
			resp, err = sproxyd.UpdObject(clientw, url, bufa[k], headera[k])
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, nil, err}
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
	clientw := &http.Client{}
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
			// clientw := &http.Client{}
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
			ch <- &sproxyd.HttpResponse{url, resp, nil, err}
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
	clientw := &http.Client{}
	for _, url := range urls {

		if len(url) == 0 {
			break
		} else {
			treq += 1
		}
		go func(url string) {
			var err error
			var resp *http.Response
			// clientw := &http.Client{}
			resp, err = sproxyd.DeleteObject(clientw, url)
			if resp != nil {
				resp.Body.Close()
			}

			ch <- &sproxyd.HttpResponse{url, resp, nil, err}
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
