package main

/*   ./DocumentGet  -action getPage -media pdf   -page p3  -pn /HR/P20020309/A2 -t 1   */
import (
	directory "directory/lib"
	"encoding/json"
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
	action, config, env, logPath, outDir, application, testname, hostname, pn, page, trace, media string
	Trace                                                                                         bool
	pid                                                                                           int
	timeout                                                                                       time.Duration
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

	if !file.Exist(outDir) {
		_ = os.MkdirAll(outDir, 0755)
	}
	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + ".md"
	goLog.Trace.Println("myfile:", myfile)
	err := ioutil.WriteFile(myfile, metadata, 0644)
	check(err)
}

func writeImage(outDir string, page string, media string, body []byte) {

	if !file.Exist(outDir) {
		_ = os.MkdirAll(outDir, 0755)
	}
	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + "." + strings.ToLower(media)
	goLog.Trace.Println("myfile:", myfile)
	err := ioutil.WriteFile(myfile, body, 0644)
	check(err)
}

func buildBnsResponse(resp *http.Response, contentType string, body []byte) (bsnImage bns.BnsImages) {

	bnsImage := bns.BnsImages{}
	if pagemd, err := base64.Decode64(resp.Header["X-Scal-Usermd"][0]); err == nil {
		bnsImage.Pagemd = string(pagemd)
		goLog.Trace.Println(bnsImage.Pagemd)
	}
	bnsImage.Image = body
	bnsImage.ContentType = contentType
	return bnsImage
}

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "action", "", "<getPageMeta> <getDocumentMeta> <getPage> <getDocument> <GetPagerange>")
	flag.StringVar(&config, "config", "storage", "Config file")
	flag.StringVar(&env, "env", "prod", "Environment")
	flag.StringVar(&trace, "t", "0", "Trace")    // Trace
	flag.StringVar(&testname, "T", "getDoc", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&page, "page", "", "page number")
	flag.StringVar(&media, "media", "tiff", "media type: tiff/png/pdf")
	flag.StringVar(&outDir, "outDir", "/home/paul/outPath", "output directory")
	Trace, _ = strconv.ParseBool(trace)
	flag.Parse()
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
			sproxyd.SetNewProxydHost(Config)
			sproxyd.Driver = Config.GetDriver()
			fmt.Println("INFO: Using config Hosts", sproxyd.Host, sproxyd.Driver, logPath)
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
		// the document metatadata is the metadata the <pathname>
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

	case "getDocument":
		// the document metatadata is the metadata the <pathname>
		docmd, err := bns.GetDocMetadata(client, pathname)
		docmeta := bns.DocumentMetadata{}

		if err == nil {
			// goLog.Info.Println(string(usermd))
			if err := json.Unmarshal(docmd, &docmeta); err != nil {
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
		sproxyResponses := bns.AsyncHttpGetPageType(urls, getHeader)
		bnsResponses := make([]bns.BnsImages, len, len)
		var pagemd []byte
		for i, v := range sproxyResponses {
			if err := v.Err; err == nil { //
				resp := v.Response
				body := v.Body
				bnsImage := buildBnsResponse(resp, getHeader["Content-Type"], body)
				bnsResponses[i] = bnsImage
				page = "p" + strconv.Itoa(i+1)
				writeImage(outDir, page, media, bnsImage.Image)
				if pagemd, err = base64.Decode64(resp.Header["X-Scal-Usermd"][0]); err == nil {
					writeMeta(outDir, page, pagemd)
				}
			}
		}

	case "getPage":
		pathname = pathname + "/" + page
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)
		var pagemd []byte
		if resp, err := bns.GetPageType(client, pathname, getHeader); err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			bnsImage := buildBnsResponse(resp, getHeader["Content-Type"], body)
			writeImage(outDir, page, media, bnsImage.Image)
			if pagemd, err = base64.Decode64(resp.Header["X-Scal-Usermd"][0]); err == nil {
				writeMeta(outDir, page, pagemd)
			}
		} else {
			goLog.Error.Println(action, pathname, err)
		}

	default:
		goLog.Info.Println("-action <action value> is missing")
	}

	goLog.Info.Println(time.Since(start))
}
