/* CreatePdf.go project main.go */
package main

import (
	"container/ring"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
	/*"user/bns" */
	"user/base64"
	"user/files"
	"user/sproxyd"
)

func usage() {
	usage := "Usage: CreatePdfDocument -c config -w pdf/xpdf -a [create/update/delete/get/getMetadata/updMetadata/getUrl/CreateLinearPdf/Test] -i inputdir\nDefault options:"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	action   string
	inputDir string
	config   string
	what     string
)

func main() {

	const (
		// base_url string = "http://luo001t.internal.epo.org:81/proxy/chord/bns/"
		LAYOUT    = "Jan 2, 2006 at 3:04pm (MST)"
		PDF       = "Pdf"
		XPDF      = "xpdf"
		CONTAINER = "Container"
	)

	flag.Usage = usage
	flag.StringVar(&action, "a", "", "")
	flag.StringVar(&inputDir, "i", "", "")
	flag.StringVar(&config, "c", "", "")
	flag.StringVar(&what, "w", "xpdf", "")

	flag.Parse()
	if len(action) == 0 || action == "?"  || what == "?" {
		usage()
	}

	client := &http.Client{}
	sep := string(os.PathSeparator)
	cpath, _ := user.Current()
	input_dir := path.Join(cpath.HomeDir, inputDir)
	pdf_path := path.Join(input_dir, PDF)
	usermd_path := path.Join(input_dir, CONTAINER)
	xpdf_path := path.Join(input_dir, XPDF)
	pdf_exist, _ := files.Exists(pdf_path)
	usermd_exist, _ := files.Exists(usermd_path)
	xpdf_exist, _ := files.Exists(xpdf_path)
	if action == "CreateLinearPdf" && !xpdf_exist {
		_ = os.MkdirAll(xpdf_path, 0755)
	}
	if action == XPDF && !xpdf_exist {
		fmt.Println(xpdf_path, "Does not exist")
		os.Exit(2)
	}
	if action == PDF && !pdf_exist {
		fmt.Println(pdf_path, "Does not exist")
		os.Exit(2)
	}
	var url string
	var buf []byte
	var filesize int64
	var usermd map[string]string

	if len(config) != 0 {
		err := sproxyd.SetProxydHost(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}
	/* Create the ring of sproxyd servers */

	r := ring.New(len(sproxyd.Host)) // r  is a pointer to  the ring
	for i := 0; i < r.Len(); i++ {
		r.Value = sproxyd.Host[i]
		// fmt.Println(r.Value)
		r = r.Next()
	}

	/* read   directory */
	if what == XPDF {
		pdf_path = xpdf_path
          
	}
	if usermd_exist {
		dirent, _ := ioutil.ReadDir(pdf_path)
		for _, pdf_file := range dirent {

			base_url := r.Value.(string)
			base_url = base_url + what + "/"
			r = r.Next()

			//os.fileInfo
			var err error
			pdf_fn := pdf_file.Name()

			url = base_url + pdf_fn
			//doc_fn := pdf_path + sep + pdf_fn
			doc_fn := path.Join(pdf_path, pdf_fn)

			if action == "create" || action == "update" || action == "updMetadata" {
				filesize = pdf_file.Size()
				fmt.Println("reading file", doc_fn, "size:", filesize)
				buf, err = ioutil.ReadFile(doc_fn)
				if err != nil {
					fmt.Println(err)
					goto Next
				} else { // get metadata
					//usermd_fn := usermd_path + sep + strings.Split(pdf_fn, ".")[0] + ".json"  
					usermd_fn := path.Join(usermd_path, strings.Split(pdf_fn, ".")[0]+".json")
					usermd_buff, err := ioutil.ReadFile(usermd_fn)
					if err != nil {
						fmt.Println(err)
						goto Next
					} else {
						//usermd_string = string(usermd_buff)
						//var usermd_1 map[string]string
						err := json.Unmarshal(usermd_buff, &usermd)
						if err != nil {
							fmt.Println(err)
							goto Next
						}
					}
				}
			}

			switch action {

			case "Test":
				/*usermd["DocSize"] = strconv.Itoa(int(filesize))
				  encoded_usermd, _ := base64.Encode64(usermd)
				  usermd_1, _ := base64.Decode64(encoded_usermd)
				  metaB, _    := json.Marshal(usermd_1) */
				// fmt.Println(usermd, encoded_usermd, usermd_1, string(metaB))
				fmt.Println(url)
			case "getUrl":
				// generate url from the input file
				fmt.Println(url)
			case "CreateLinearPdf":
				// generate creation of linear pdf from pdf files using the qpdf command line
				cmd := "/usr/bin/qpdf --linearize "
				fmt.Println(cmd + doc_fn + " " + xpdf_path + sep + pdf_fn)

			case "create":
				time1 := time.Now()
				usermd["DocSize"] = strconv.Itoa(int(filesize))
				encoded_usermd, _ := base64.Encode64(usermd)
				//fmt.Println(encoded_usermd)
				putheader := map[string]string{
					"Usermd":       encoded_usermd,
					"Content-Type": "application/pdf",
				}
				var resp *http.Response
				if resp, err = sproxyd.PutObject(client, url, buf, putheader); err != nil {
					fmt.Println(err)
					resp.Body.Close()
					break
				}

				switch resp.StatusCode {
				case 200:
					duration := time.Now().Sub(time1)
					fmt.Println("OK", resp.Header["X-Scal-Ring-Key"], "MB/sec=", 1000*float64(filesize)/float64(duration), "Duration=", duration)
				case 412:
					fmt.Println(resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")
				case 422:
					fmt.Println(resp.Status, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}
				resp.Body.Close()

			case "update":
				time1 := time.Now()
				usermd["DocSize"] = strconv.Itoa(int(filesize))
				usermd["update"] = "Data"
				usermd["Time_of_update"] = time1.UTC().Format(LAYOUT)
				encoded_usermd, _ := base64.Encode64(usermd)
				putheader := map[string]string{
					"Usermd":       encoded_usermd,
					"Content-Type": "application/pdf",
				}
				var resp *http.Response
				if resp, err = sproxyd.UpdObject(client, url, buf, putheader); err != nil {
					fmt.Println(err)
					resp.Body.Close()
					break
				}

				switch resp.StatusCode {
				case 200:
					duration := time.Now().Sub(time1)
					fmt.Println("OK", resp.Header["X-Scal-Ring-Key"], "MB/sec=", 1000*float64(filesize)/float64(duration), "Duration=", duration)
				case 404:
					fmt.Println(resp.Status, url, " not found")
				case 412:
					fmt.Println(resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
				case 422:
					fmt.Println(resp.Status, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}
				resp.Body.Close()

			case "delete":
				time1 := time.Now()
				deleteHeader := map[string]string{}

				var resp *http.Response
				if resp, err = sproxyd.DeleteObject(client, url, deleteHeader); err != nil {
					fmt.Println(err)
					resp.Body.Close()
					break
				}

				switch resp.StatusCode {
				case 200:
					duration := time.Now().Sub(time1)
					fmt.Println(url, "OK", resp.Header["X-Scal-Ring-Key"], "Duration=", duration)
				case 404:
					fmt.Println(resp.StatusCode, url, " not found")
				case 412:
					fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
				case 422:
					fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}

				//fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Key"])
				resp.Body.Close()

			case "get":
				//client := &http.Client{}
				time1 := time.Now()
				getHeader := map[string]string{}

				var resp *http.Response
				if resp, err = sproxyd.GetObject(client, url, getHeader); err != nil {
					fmt.Println(err)
					resp.Body.Close()
					break
				}

				switch resp.StatusCode {
				case 200:
					/* READ METADATA */
					/* encoded_usermd := resp.Header["X-Scal-Usermd"]
					   	       fmt.Println(resp.StatusCode,encoded_usermd)
					    	       usermd_1,_ := base64.Decode64(encoded_usermd[0])

					   	       for k,v  := range usermd_1{
					    	          fmt.Println(k,"=",v)
					   	        }
					*/
					/* READ BODY */
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Println(err)
					}
					long := len(body)

					duration := time.Now().Sub(time1)
					fmt.Println(url, "Ok", "MB/sec=", 1000*float64(long)/float64(duration), "Duration=", duration)

				case 404:
					fmt.Println(resp.StatusCode, url, " not found")
				case 412:
					fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " Pre Condition failed")
				case 422:
					fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}
				resp.Body.Close()

			case "getMetadata":
				time1 := time.Now()
				getHeader := map[string]string{}

				var resp *http.Response
				if resp, err = sproxyd.GetMetadata(client, url, getHeader); err != nil {
					fmt.Println(err)
					resp.Body.Close()
					break
				}
				switch resp.StatusCode {
				case 200:
					encoded_usermd := resp.Header["X-Scal-Usermd"]
					// READ METADATA --> map[string]string
					usermd_1, _ := base64.Decode64(encoded_usermd[0])

					/* convert map[string]string to json Byte */
					metaB, _ := json.Marshal(usermd_1)
					duration := time.Now().Sub(time1)
					fmt.Println(string(metaB), "Duration=", duration)

				case 404:
					fmt.Println(resp.StatusCode, url, " not found")
				case 412:
					fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " Pre Condition failed")
				case 422:
					fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}
				resp.Body.Close()

			case "updMetadata":
				//New user metadata
				time1 := time.Now()
				usermd["update"] = "Metadata"
				usermd["Time_of_update"] = time1.UTC().Format(LAYOUT)
				encoded_usermd, _ := base64.Encode64(usermd)
				updHeader := map[string]string{
					"Usermd":       encoded_usermd,
					"Content-Type": "application/pdf",
				}

				resp, err := sproxyd.UpdMetadata(client, url, updHeader)
				if err != nil {
					fmt.Println(err)
					break
				}

				switch resp.StatusCode {
				case 200:
					fmt.Println("OK", resp.Header["X-Scal-Ring-Key"])
				case 404:
					fmt.Println(resp.Status, url, " not found")
				case 412:
					fmt.Println(resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
				case 422:
					fmt.Println(resp.Status, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}
				resp.Body.Close()

			default:
				fmt.Println("wrong Action")

			} // switch

		}
	Next: // dir entry
	} else {
		fmt.Println("Input File or Directrory", pdf_path, "Does not exist")
	}

	//}

}

