package main

import (
	directory "directory/lib"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	bns "moses/bns/lib"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"os"
	"strconv"
	"time"

	base64 "moses/user/base64j"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"

	hostpool "github.com/bitly/go-hostpool"
)

var (
	action, config, env, logPath, outDir, application, testname, hostname, pn, page, trace, test, meta, image, media string
	Trace, Meta, Image, CopyObject, Test                                                                             bool
	pid                                                                                                              int
	timeout                                                                                                          time.Duration
)

func usage() {

	usage := "CopyObject: \n  -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/storage]\n" +
		"-pn pn -page page"

	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func buildBnsResponse(resp *http.Response, contentType string, body *[]byte) (bsnImage bns.BnsImages) {

	bnsImage := bns.BnsImages{}

	if _, ok := resp.Header["X-Scal-Usermd"]; ok {
		bnsImage.Usermd = resp.Header["X-Scal-Usermd"][0]
		if pagemd, err := base64.Decode64(bnsImage.Usermd); err == nil {
			bnsImage.Pagemd = string(pagemd)
			goLog.Trace.Println(bnsImage.Pagemd)
		}
	} else {
		goLog.Warning.Println("X-Scal-Usermd is missing the resp header", resp.Status, resp.Header)
	}

	bnsImage.Image = *body
	bnsImage.ContentType = contentType
	return bnsImage
}

func checkOutdir(outDir string) (err error) {

	if len(outDir) == 0 {
		err = errors.New("Please specify an output directory with -outDir argument")
	} else if !file.Exist(outDir) {
		err = os.MkdirAll(outDir, 0755)
	}
	return err
}

func copyBlob(bnsRequest *bns.HttpRequest, url string, buf []byte, header map[string]string) {

	result := bns.AsyncHttpPutBlob(bnsRequest, url, buf, header)

	if result.Err != nil {
		goLog.Trace.Printf("%s %d %s status: %s\n", hostname, pid, result.Url, result.Err)
		return
	}

	resp := result.Response

	if resp != nil {
		goLog.Trace.Printf("%s %d %s status: %s\n", hostname, pid, url,
			result.Response.Status)
	} else {
		goLog.Error.Printf("%s %d %s %s %s", hostname, pid, url, action, "failed")
	}

	switch resp.StatusCode {
	case 200:
		goLog.Trace.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Key"])

	case 412:
		goLog.Warning.Println(hostname, pid, url, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")

	case 422:
		goLog.Error.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Status"])
	default:
		goLog.Warning.Println(hostname, pid, url, resp.Status)
	}
	resp.Body.Close()
}

func copyBlobTest(bnsRequest *bns.HttpRequest, url string, buf []byte, header map[string]string) {
	result := bns.AsyncHttpPutBlobTest(bnsRequest, url, buf, header)
	goLog.Trace.Printf("URL => %s \n", result.Url)
}

func deleteBlobTest(bnsRequest *bns.HttpRequest, url string) {
	result := bns.AsyncHttpDeleteBlobTest(bnsRequest, url)
	goLog.Trace.Printf("URL => %s \n", result.Url)
}

func deleteBlob(bnsRequest *bns.HttpRequest, url string) {
	result := bns.AsyncHttpDeleteBlob(bnsRequest, url)
	if result.Err != nil {
		goLog.Trace.Printf("%s %d %s status: %s\n", hostname, pid, result.Url, result.Err)
		return
	}
	resp := result.Response
	if resp != nil {
		goLog.Trace.Printf("%s %d %s status: %s\n", hostname, pid, url,
			result.Response.Status)
	} else {
		goLog.Error.Printf("%s %d %s %s %s", hostname, pid, url, action, "failed")
	}

	switch resp.StatusCode {
	case 200:
		goLog.Trace.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Key"])

	case 412:
		goLog.Warning.Println(hostname, pid, url, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")

	case 422:
		goLog.Error.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Status"])
	default:
		goLog.Warning.Println(hostname, pid, url, resp.Status)
	}
	resp.Body.Close()

}

func main() {

	flag.Usage = usage
	flag.StringVar(&config, "config", "storage", "Config file")
	flag.StringVar(&env, "env", "prod", "Environment")
	flag.StringVar(&trace, "t", "0", "Trace") // Trace
	flag.StringVar(&meta, "meta", "0", "Save object meta in output Directory")
	flag.StringVar(&image, "image", "0", "Save object image  type in output Directory")
	flag.StringVar(&testname, "T", "getDoc", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&page, "page", "1", "page number")
	flag.StringVar(&media, "media", "tiff", "media type: tiff/png/pdf")
	flag.StringVar(&outDir, "outDir", "", "output directory")
	flag.StringVar(&test, "test", "1", "Run copy in test mode")
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)
	Test, _ = strconv.ParseBool(test)

	action = "CopyObject"
	if len(pn) == 0 {
		fmt.Println("-pn <DocumentId> is missing")
	}
	// application = "DocumentGet"
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	if testname != "" {
		testname += string(os.PathSeparator)
	}

	if len(config) != 0 {

		if Config, err := sproxyd.GetConfig(config); err == nil {

			logPath = Config.GetLogPath()
			if len(outDir) == 0 {
				outDir = Config.GetOutputDir()
			}
			sproxyd.SetNewProxydHost(Config)
			sproxyd.Driver = Config.GetDriver()
			sproxyd.SetNewTargetProxydHost(Config)
			sproxyd.TargetDriver = Config.GetTargetDriver()
			fmt.Println("INFO: Using config Hosts", sproxyd.Host, sproxyd.Driver, logPath)
			fmt.Println("INFO: Using config target Hosts", sproxyd.TargetHost, sproxyd.TargetDriver, logPath)
		} else {
			sproxyd.HP = hostpool.NewEpsilonGreedy(sproxyd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
			fmt.Println(err, "WARNING: Using default Hosts:", sproxyd.Host)
		}
	}
	// init logging

	if logPath == "" {
		fmt.Println("WARNING: Using default logging")
		goLog.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	} else {

		// mkAll dir
		logPath = logPath + string(os.PathSeparator) + testname
		if !file.Exist(logPath) {
			_ = os.MkdirAll(logPath, 0755)
		}
		traceLog := logPath + application + "_trace.log"
		infoLog := logPath + application + "_info.log"
		warnLog := logPath + application + "_warning.log"
		errLog := logPath + application + "_error.log"

		trf, err1 := os.OpenFile(traceLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
		inf, err2 := os.OpenFile(infoLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
		waf, err3 := os.OpenFile(warnLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
		erf, err4 := os.OpenFile(errLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)

		defer trf.Close()
		defer inf.Close()
		defer waf.Close()
		defer erf.Close()

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			goLog.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
			goLog.Warning.Println(err1, err2, err3, err3)
			goLog.Warning.Println(hostname, pid, "Using default logging")
		} else {
			if trace == "0" {
				goLog.Init(ioutil.Discard, io.Writer(inf), io.Writer(waf), io.Writer(erf))

			} else {
				goLog.Init(io.Writer(trf), io.Writer(inf), io.Writer(waf), io.Writer(erf))
				goLog.Trace.Println(hostname, pid, "Start", application, action)
			}
		}
	}
	directory.SetCPU("100%")
	// client := &http.Client{}
	start := time.Now()
	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP,
		Client: &http.Client{},
		Media:  media,
	}
	var (
		err           error
		encoded_docmd string
		docmd         []byte
	)
	media = "binary"
	page = "p" + page
	pathname := env + "/" + pn
	url := pathname
	if encoded_docmd, err = bns.GetEncodedMetadata(&bnsRequest, url); err == nil {
		if docmd, err = base64.Decode64(encoded_docmd); err != nil {
			goLog.Error.Println(err)
			os.Exit(2)
		}
	} else {
		goLog.Error.Println(err)
		os.Exit(2)
	}
	docmeta := bns.DocumentMetadata{}
	if err := json.Unmarshal(docmd, &docmeta); err != nil {
		goLog.Error.Println(docmeta)
		goLog.Error.Println(err)
		os.Exit(2)
	} else {
		header := map[string]string{
			"Usermd": encoded_docmd,
		}
		buf0 := make([]byte, 0)
		if !Test {
			// copyObject(sproxyd.TargetHP, client, pathname, buf0, header)
			copyBlob(&bnsRequest, url, buf0, header)
		} else {
			copyBlobTest(&bnsRequest, url, buf0, header)
		}
	}
	len := docmeta.TotalPage
	urls := make([]string, len, len)
	getHeader := map[string]string{}
	getHeader["Content-Type"] = "application/binary"
	for i := 0; i < len; i++ {
		urls[i] = pathname + "/p" + strconv.Itoa(i+1)
	}
	bnsRequest.Urls = urls
	bnsRequest.Hspool = sproxyd.HP
	sproxyResponses := bns.AsyncHttpGetBlob(&bnsRequest, getHeader)
	bnsResponses := make([]bns.BnsImages, len, len)
	bnsRequest.Client = &http.Client{}
	for i, v := range sproxyResponses {
		if err := v.Err; err == nil { //
			resp := v.Response
			body := *v.Body
			usermd := resp.Header["X-Scal-Usermd"][0]
			bnsImage := buildBnsResponse(resp, getHeader["Content-Type"], &body) // bnsImage is a Go structure
			page = "p" + strconv.Itoa(i+1)
			bnsResponses[i] = bnsImage
			header := map[string]string{
				"Usermd": usermd,
			}
			url = urls[i]
			bnsRequest.Hspool = sproxyd.TargetHP
			if !Test {
				// copyObject(sproxyd.TargetHP, clientc, urls[i], bnsImage.Image, header)
				copyBlob(&bnsRequest, url, bnsImage.Image, header)
			} else {
				copyBlobTest(&bnsRequest, url, bnsImage.Image, header)
			}
		}
	}
	duration := time.Since(start)
	fmt.Println("total elapsed time:", duration)
	goLog.Info.Println(duration)
}
