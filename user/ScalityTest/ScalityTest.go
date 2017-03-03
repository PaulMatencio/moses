package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
	"time"
	"user/base64"
	"user/files"
	"user/sproxyd"
)

func usage() {
	usage := "usage: ScalityTest -w <pdf/xpdf> -a [get/getMetadata]"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	action string
	what   string
)

func main() {

	const (
		sproxy string = "http://luo001t.internal.epo.org:81/proxy/chord/"
	)

	flag.Usage = usage
	flag.StringVar(&action, "a", "", "")
	flag.StringVar(&what, "w", "xpdf", "")
	flag.Parse()
	if len(action) == 0 {
		usage()
	}
	if action != "get" || action != "getMetaData" {

		fmt.Println("Deprecated for other actions, use CreatePdf instead\n")

		os.Exit(2)

	}
	client := &http.Client{}
	base_url := sproxy + what + "/"
	/* read   directory */
	cpath, _ := user.Current()
	//path	  := cpath.HomeDir+"/pdf/"
	path := path.Join(cpath.HomeDir, what)
	path_exist, _ := files.Exists(path)
	if path_exist == true {
		// read directory entries
		dirent, _ := ioutil.ReadDir(path)
		usermd := map[string]string{}
		var url string
		var buf []byte
		var filesize int64

		for _, file := range dirent {
			//os.fileInfo
			var err error
			filename := file.Name()
			url = base_url + filename
			doc := path + filename
			docid := strings.Split(filename, ".")[0]
			doctype := strings.Split(filename, ".")[1]
			kc := docid[len(docid)-2 : len(docid)]
			ip := docid[0:2]

			if action == "create" || action == "update" {
				filesize = file.Size()
				fmt.Println("reading file", doc, "size:", filesize)
				buf, err = ioutil.ReadFile(doc)
				if err != nil {
					fmt.Println(err)
				} else {

					usermd["Doc_id"] = docid
					usermd["KC"] = kc
					usermd["IP"] = ip
					usermd["Doc_type"] = doctype
					usermd["Total_page"] = "20"
					usermd["Bibl_SPN"] = "1"
					usermd["Bibl_PC"] = "2"
					usermd["Clai_SPN"] = "2"
					usermd["Clai_PC"] = "1"
					usermd["Desc_SPN"] = "5"
					usermd["Desc_PC"] = "5"
					usermd["Draw_SPN"] = "10"
					usermd["Draw_PC"] = "2"
					usermd["StRep_SPN"] = "18"
					usermd["StRep_PC"] = "2"

				}
			}

			switch action {
			case "create":
				time1 := time.Now()
				encoded_usermd, _ := base64.Encode64(usermd)
				//fmt.Println(encoded_usermd)
				putheader := map[string]string{
					"Usermd":       encoded_usermd,
					"Content-Type": "application/pdf",
				}
				resp, err := sproxyd.PutObject(client, url, buf, putheader)
				if err != nil {
					fmt.Println(err)
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

			case "update":
				encoded_usermd, _ := base64.Encode64(usermd)
				//fmt.Println(encoded_usermd)
				putheader := map[string]string{
					"Usermd":       encoded_usermd,
					"Content-Type": "application/pdf",
				}
				resp, err := sproxyd.UpdObject(client, url, buf, putheader)
				if err != nil {
					panic(err)
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

			case "delete":

				deleteHeader := map[string]string{}
				resp, err := sproxyd.DeleteObject(client, url, deleteHeader)
				if err != nil {
					panic(err)
				}

				switch resp.StatusCode {
				case 200:
					fmt.Println("OK", resp.Header["X-Scal-Ring-Key"])
				case 404:
					fmt.Println(resp.StatusCode, url, " not found")
				case 412:
					fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
				case 422:
					fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}

				fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Key"])

			case "get":

				time1 := time.Now()
				getHeader := map[string]string{}
				resp, err := sproxyd.GetObject(client, url, getHeader)
				if err != nil {
					panic(err)
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
					body, _ := ioutil.ReadAll(resp.Body)
					long := len(body)
					duration := time.Now().Sub(time1)
					fmt.Println("OK", "MB/sec=", 1000*float64(long)/float64(duration), "Duration=", duration)

				case 404:
					fmt.Println(resp.StatusCode, url, " not found")
				case 412:
					fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " Pre Condition failed")
				case 422:
					fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
				default:
					fmt.Println(resp.Status)
				}

			case "getMetadata":
				time1 := time.Now()
				getHeader := map[string]string{}
				resp, err := sproxyd.GetMetadata(client, url, getHeader)
				if err != nil {
					panic(err)
				}
				switch resp.StatusCode {
				case 200:
					encoded_usermd := resp.Header["X-Scal-Usermd"]
					// READ METADATA --> map[string]string
					usermd_1, _ := base64.Decode64(encoded_usermd[0])
					/* for k,v  := range usermd_1{
					   fmt.Println(k,"=",v) }
					*/
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

			case "updMetadata":
				//New user metadata
				usermd["Doc_id"] = docid
				usermd["KC"] = kc
				usermd["IP"] = ip
				usermd["Doc_type"] = doctype
				usermd["Total_page"] = "21"
				usermd["Bibl_SPN"] = "1"
				usermd["Bibl_PC"] = "2"
				usermd["Clai_SPN"] = "2"
				usermd["Clai_PC"] = "1"
				usermd["Desc_SPN"] = "5"
				usermd["Desc_PC"] = "5"
				usermd["Draw_SPN"] = "10"
				usermd["Draw_PC"] = "10"
				usermd["StRep_SPN"] = "19"
				usermd["StRep_PC"] = "2"
				encoded_usermd, _ := base64.Encode64(usermd)

				updHeader := map[string]string{
					"Usermd": encoded_usermd,
				}

				resp, err := sproxyd.UpdMetadata(client, url, updHeader)
				if err != nil {
					panic(err)
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

			default:
				fmt.Println("wrong Action")

			}
		}
	}
}
