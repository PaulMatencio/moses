package main

/*   ./DocumentGet  -action getPage -media pdf   -page p3  -pn /HR/P20020309/A2 -t 1   */
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
	"strings"
	"time"

	base64 "moses/user/base64j"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"

	hostpool "github.com/bitly/go-hostpool"
)

var (
	action, config, env, logPath, outDir, application, testname, hostname, pn, page, trace, meta, image, media string
	Trace, Meta, Image, CopyObject                                                                             bool
	pid                                                                                                        int
	timeout                                                                                                    time.Duration
)

func usage() {

	usage := "DocumentGet: \n -action <action> -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/storage]\n" +
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

func writeMeta(outDir string, page string, metadata []byte) {
	if !Meta {
		return
	}

	if err := checkOutdir(outDir); err != nil {
		goLog.Error.Println(err)
		return
	}

	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + ".md"
	goLog.Trace.Println("myfile:", myfile)
	err := ioutil.WriteFile(myfile, metadata, 0644)
	check(err)
}

func writeImage(outDir string, page string, media string, body *[]byte) {

	if !Image {
		return
	}

	if err := checkOutdir(outDir); err != nil {
		goLog.Error.Println(err)
		return
	}

	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + "." + strings.ToLower(media)
	goLog.Trace.Println("myfile:", myfile)
	err := ioutil.WriteFile(myfile, *body, 0644)
	check(err)
}

func buildBnsResponse(resp *http.Response, contentType string, body *[]byte) (bsnImage bns.BnsImages) {

	bnsImage := bns.BnsImages{}
	bnsImage.Usermd = resp.Header["X-Scal-Usermd"][0]
	if pagemd, err := base64.Decode64(bnsImage.Usermd); err == nil {
		bnsImage.Pagemd = string(pagemd)
		goLog.Trace.Println(bnsImage.Pagemd)
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

func copyObject(bnsResponses *[]bns.BnsImages, urls []string) {

	// do not forget to write the document metadata
	// urls must be changed to use the target hostpool
	dim := len(*bnsResponses)
	headera := make([]map[string]string, dim, dim)
	bufera := make([][]byte, dim, dim)

	for i, bnsImage := range *bnsResponses {
		usermd := bnsImage.Usermd
		headera[i] = map[string]string{
			"Usermd": usermd,
		}
		bufera[i] = bnsImage.Image
		fmt.Println(i, headera[i], usermd, len(bufera[i]))
	}

	results := bns.AsyncHttpPuts(sproxyd.TargetHP, urls, bufera, headera)

	ok := 0
	req := len(urls)

	for _, result := range results {

		if result.Err != nil {
			goLog.Trace.Printf("%s %d %s status: %s\n", hostname, pid, result.Url, result.Err)
			continue
		}

		resp := result.Response
		url := result.Url
		if resp != nil {
			goLog.Trace.Printf("%s %d %s status: %s\n", hostname, pid, url,
				result.Response.Status)
		} else {
			goLog.Error.Printf("%s %d %s %s %s", hostname, pid, url, action, "failed")
			continue
		}

		switch resp.StatusCode {
		case 200:
			goLog.Trace.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Key"])
			ok += 1
		case 412:
			goLog.Warning.Println(hostname, pid, url, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")

		case 422:
			goLog.Error.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			goLog.Warning.Println(hostname, pid, url, resp.Status)
		}
		resp.Body.Close()
	}
	if ok < req {
		goLog.Warning.Println(hostname, pid, ok, req, "#ok < #req => Check Warning or Error log")
	}
}

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "action", "", "<getPageMeta> <getDocumentMeta> <getPage> <getDocumentType> <getObjectt> <copyObject> <GetPagerange>")
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
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)

	if action == "copyObject" {
		action = "getObject"
		CopyObject = true
	}

	if CopyObject {
		Meta = false  // do not decode Metadata and write to files
		Image = false // do not write Object to files
	}

	if len(action) == 0 {
		usage()
	}
	if len(pn) == 0 {
		fmt.Println("-pn <DocumentId> is missing")
	}
	application = "DocumentGet"
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

	//goLog.Init0(logPath, testname, application, action, Trace)
	directory.SetCPU("100%")
	client := &http.Client{}
	start := time.Now()
	page = "p" + page
	pathname := env + "/" + pn

	if action == "copyObject" {
		action = "getObject"
		CopyObject = true
	}

	switch action {
	case "getPageMeta":
		pathname = pathname + "/" + page
		pagemd, err := bns.GetPageMetadata(client, pathname)
		if err == nil {
			// goLog.Info.Println(string(pagemd))
			writeMeta(outDir, page, pagemd)
		} else {
			goLog.Error.Println(err)
		}
	case "getDocumentMeta":
		// the document's  metatadata is the metadata the object given <pathname>
		docmd, err := bns.GetDocMetadata(client, pathname)
		docmeta := bns.DocumentMetadata{}

		if err == nil {
			goLog.Info.Println(string(docmd))
			if err := json.Unmarshal(docmd, &docmeta); err != nil {
				goLog.Error.Println(err)
			} else {
				writeMeta(outDir, "", docmd)
			}
		} else {
			goLog.Error.Println(err)
		}

	case "getDocumentType":
		// the document metatadata is the metadata the <pathname>
		// all the Document's pages will be retrieved concurrently
		docmd, err := bns.GetDocMetadata(client, pathname)
		docmeta := bns.DocumentMetadata{}

		if err == nil {
			// goLog.Info.Println(string(usermd))
			if err := json.Unmarshal(docmd, &docmeta); err != nil {
				goLog.Error.Println(docmeta)
				goLog.Error.Println(err)
				os.Exit(2)
			} else {
				writeMeta(outDir, "", docmd)
			}

		} else {
			goLog.Error.Println(err)
			os.Exit(2)
		}
		// build []urls of pages  of the document to be fecthed
		len := docmeta.TotalPage
		urls := make([]string, len, len)

		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)

		for i := 0; i < len; i++ {
			urls[i] = pathname + "/p" + strconv.Itoa(i+1)
		}

		// sproxyResponses := bns.AsyncHttpGetPageType(urls, getHeader)
		sproxyResponses := bns.AsyncHttpGetPageType(urls, media)

		// AsyncHttpGetPageType should already  close [defer resp.Body.Clsoe()] all open connections
		bnsResponses := make([]bns.BnsImages, len, len)
		var pagemd []byte
		for i, v := range sproxyResponses {
			if err := v.Err; err == nil { //
				resp := v.Response
				body := *v.Body
				bnsImage := buildBnsResponse(resp, getHeader["Content-Type"], &body)
				bnsResponses[i] = bnsImage
				page = "p" + strconv.Itoa(i+1)
				if Image {
					writeImage(outDir, page, media, &bnsImage.Image)
				}
				if Meta {
					if pagemd, err = base64.Decode64(resp.Header["X-Scal-Usermd"][0]); err == nil {
						writeMeta(outDir, page, pagemd)
					}
				}
			}
		}
	case "getObject":
		// the document metatadata is the metadata the <pathname>
		// all the Document's pages will be retrieved concurrently
		var (
			err           error
			encoded_docmd string
			docmd         []byte
		)

		if encoded_docmd, err = bns.GetEncodedMetadata(client, pathname); err == nil {
			docmd, err = base64.Decode64(encoded_docmd)
		}

		docmeta := bns.DocumentMetadata{}

		if err == nil {
			// goLog.Info.Println(string(usermd))
			if err := json.Unmarshal(docmd, &docmeta); err != nil {
				goLog.Error.Println(docmeta)
				goLog.Error.Println(err)
				os.Exit(2)
			} else {
				writeMeta(outDir, "", docmd)
			}

		} else {
			goLog.Error.Println(err)
			os.Exit(2)
		}
		// build []urls of pages  of the document to be fecthed
		len := docmeta.TotalPage
		urls := make([]string, len, len)

		getHeader := map[string]string{}
		ct := "application/binary"
		getHeader["Content-Type"] = ct

		for i := 0; i < len; i++ {
			urls[i] = pathname + "/p" + strconv.Itoa(i+1)
		}

		sproxyResponses := bns.AsyncHttpGetPage(urls, getHeader)

		// AsyncHttpGetPageType should already  close [defer resp.Body.Clsoe()] all open connections
		bnsResponses := make([]bns.BnsImages, len, len)
		var pagemd []byte
		for i, v := range sproxyResponses {
			if err := v.Err; err == nil { //
				resp := v.Response
				body := *v.Body
				usermd := resp.Header["X-Scal-Usermd"][0]
				bnsImage := buildBnsResponse(resp, ct, &body)

				page = "p" + strconv.Itoa(i+1)
				if Image {
					writeImage(outDir, page, ct, &bnsImage.Image)
				}
				if Meta {
					if pagemd, err = base64.Decode64(usermd); err == nil {
						writeMeta(outDir, page, pagemd)
						bnsImage.Pagemd = string(pagemd)
					}
				}
				bnsResponses[i] = bnsImage
			}
		}

		if CopyObject {
			copyObject(&bnsResponses, urls)
		}

	case "getPage":
		// get a specific page of a document
		// if -page is missing p1 is the default
		pathname = pathname + "/" + page

		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)

		var pagemd []byte
		// if resp, err := bns.GetPageType(client, pathname, getHeader); err == nil {
		if resp, err := bns.GetPageType(client, pathname, media); err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			bnsImage := buildBnsResponse(resp, getHeader["Content-Type"], &body)
			writeImage(outDir, page, media, &bnsImage.Image)
			if pagemd, err = base64.Decode64(resp.Header["X-Scal-Usermd"][0]); err == nil {
				writeMeta(outDir, page, pagemd)
			}
		} else {
			goLog.Error.Println(action, pathname, err)
		}

	default:
		goLog.Info.Println("-action <action value> is missing")
	}
	duration := time.Since(start)
	fmt.Println("total elapsed time:", duration)
	goLog.Info.Println(duration)
}
