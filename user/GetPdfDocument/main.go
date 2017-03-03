// RetrievPdf project main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"user/base64"
	"user/sproxyd"
)

func usage() {
	usage := "\nusage: GetPdfDocument  -w  -a -d document\n-w <xpdf/pdf> \n-a <get/getMetadata/getPdfDirectory/Test>\n\nDefault options<\n"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	action string
	what   string
	docid  string
)

func main() {

	const (
		sproxy string = "http://luo001t.internal.epo.org:81/proxy/chord/"
	)

	flag.Usage = usage
	flag.StringVar(&action, "a", "get", "")
	flag.StringVar(&docid, "d", "", "")
	flag.StringVar(&what, "w", "xpdf", "")
	flag.Parse()
	if len(docid) == 0 {
		usage()
	}

	base_url := sproxy + what + "/"
	url := base_url + docid + ".pdf"

	if action == "Test" {
		fmt.Println(url)
		os.Exit(0)
	}
	if action != "get" && action != "getMetadata" && action != "getPdfDirectory" {

		fmt.Println("Invalid Action\n")
		usage()
		 
	}

	client := &http.Client{}

	switch action {

	case "get":
		time1 := time.Now()
		getHeader := map[string]string{}
		resp, err := sproxyd.GetObject(client, url, getHeader)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		switch resp.StatusCode {
		case 200:
			/* READ BODY */
			body, _ := ioutil.ReadAll(resp.Body)
			long := len(body)
			duration := time.Since(time1)
			fmt.Println("OK,Data Length:", long, ",MB/sec=", 1000*float64(long)/float64(duration), ",Duration=", duration)

		case 404:
			fmt.Println(resp.StatusCode, url, " not found")
		case 412:
			fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " Pre Condition failed")
		case 422:
			fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
		default:
			fmt.Println(resp.StatusCode, url)
		}
		resp.Body.Close()

	case "getMetadata":
		time1 := time.Now()
		getHeader := map[string]string{}
		resp, err := sproxyd.GetMetadata(client, url, getHeader)
		if err != nil {
			panic(err)
		}
		switch resp.StatusCode {
		case 200:
			encoded_usermd := resp.Header["X-Scal-Usermd"] // READ METADATA --> map[string]string
			usermd_1, _ := base64.Decode64(encoded_usermd[0])
			metaB, _ := json.Marshal(usermd_1)
			duration := time.Since(time1)
			fmt.Println(string(metaB), "Duration=", duration)

		case 404:
			fmt.Println(resp.StatusCode, url, " not found")
		case 412:
			fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " Pre Condition failed")
		case 422:
			fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])
		default:
			fmt.Println(resp.StatusCode, url)
		}
		resp.Body.Close()

	case "getPdfDirectory":

		time1 := time.Now()
		getHeader := map[string]string{}
		getHeader["Range"] = "bytes=0-1000"
		fmt.Println(getHeader)
		resp, err := sproxyd.GetObject(client, url, getHeader)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		switch resp.StatusCode {
		case 200:
			/* READ BODY */
			body, _ := ioutil.ReadAll(resp.Body)
			long := len(body)
			duration := time.Since(time1)
			fmt.Println("OK,Data length:", long, ",MB/sec=", 1000*float64(long)/float64(duration), ",Duration=", duration)
		case 206: // partial content request succeed
			body, _ := ioutil.ReadAll(resp.Body)
			//long := len(body)
			pdfHeader := string(body[0:256])
			duration := time.Since(time1)
			fmt.Println("OK", pdfHeader, ",Duration=", duration)

		case 404:
			fmt.Println(resp.StatusCode, url, " not found")
		case 412:
			fmt.Println(resp.StatusCode, "key=", resp.Header["X-Scal-Ring-Key"], " Pre Condition failed")
		case 422:
			fmt.Println(resp.StatusCode, resp.Header["X-Scal-Ring-Status"])

		default:
			fmt.Println(resp.StatusCode, url)
		}
		resp.Body.Close()

	default:
		fmt.Println("Wrong request")
	}
}
