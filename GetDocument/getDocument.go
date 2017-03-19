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
	"os/user"
	"path"
	"strconv"
	"strings"
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

func checkOutdir(outDir string) (err error) {

	if len(outDir) == 0 {
		err = errors.New("Please specify an output directory with -outDir argument")
	} else if !file.Exist(outDir) {
		err = os.MkdirAll(outDir, 0755)
	}
	return err
}

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "action", "", "<getPageMeta> <getPageType> <getDocumentMeta> <getDocumentType> <GetPageRange>")
	flag.StringVar(&config, "config", "moses-dev", "Config file")
	flag.StringVar(&env, "env", "", "Environment")
	flag.StringVar(&trace, "t", "0", "Trace")       // Trace
	flag.StringVar(&test, "test", "0", "Test mode") // Test mode
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
	Test, _ = strconv.ParseBool(test)
	sproxyd.Test = Test

	if len(action) == 0 {
		usage()
	}
	if len(pn) == 0 {
		fmt.Println("-pn <DocumentId> is missing")
	}
	application = "DocumentGet"
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	if testname != "" {
		testname += string(os.PathSeparator)
	}

	if len(config) != 0 {
		if Config, err := sproxyd.GetConfig(config); err == nil {
			logPath = path.Join(homeDir, Config.GetLogPath())
			if len(outDir) == 0 {
				outDir = path.Join(homeDir, Config.GetOutputDir())
			}

			sproxyd.SetNewProxydHost(Config)
			sproxyd.Driver = Config.GetDriver()
			sproxyd.Env = Config.GetEnv()
			sproxyd.SetNewTargetProxydHost(Config)
			sproxyd.TargetDriver = Config.GetTargetDriver()
			sproxyd.TargetEnv = Config.GetTargetEnv()

			fmt.Println("INFO: Using config Hosts", sproxyd.Host, sproxyd.Driver, sproxyd.Env, logPath)
			fmt.Println("INFO: Using config target Hosts", sproxyd.TargetHost, sproxyd.TargetDriver, sproxyd.TargetEnv, logPath)
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
	if len(env) == 0 {
		env = sproxyd.Env
	}
	pathname := env + "/" + pn

	if action == "copyObject" {
		action = "getObject"
		CopyObject = true
	}

	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP,
		Client: client,
		Media:  media,
	}

	switch action {
	case "getPageMeta":
		Meta = true
		pathname = pathname + "/" + page
		// bnsRequest.Path = pathname
		if pagemd, err := bns.GetPageMetadata(&bnsRequest, pathname); err == nil {
			writeMeta(outDir, page, pagemd)
		} else {
			goLog.Error.Println(err)
		}

	case "getDocumentMeta":
		// the document's  metatadata is the metadata the object given <pathname>
		// bnsRequest.Path = pathname
		Meta = true
		if docmd, err := bns.GetDocMetadata(&bnsRequest, pathname); err == nil {
			goLog.Info.Println("Document Metadata=>\n", string(docmd))
			if len(docmd) != 0 {
				docmeta := bns.DocumentMetadata{}
				if err := json.Unmarshal(docmd, &docmeta); err != nil {
					goLog.Error.Println(err)
				} else {
					writeMeta(outDir, "", docmd)
				}
			} else {
				goLog.Error.Println(pathname, "Document Metadata is missing")
			}
		} else {
			goLog.Error.Println(err)
		}

	case "getDocumentType":
		docmeta := bns.DocumentMetadata{}
		if docmd, err := bns.GetDocMetadata(&bnsRequest, pathname); err == nil {
			goLog.Trace.Println("Document Metadata=>", string(docmd))
			if len(docmd) != 0 {

				if err := json.Unmarshal(docmd, &docmeta); err != nil {
					goLog.Error.Println(docmeta)
					goLog.Error.Println(err)
					os.Exit(2)
				} else {
					writeMeta(outDir, "", docmd)
				}
			} else {
				goLog.Error.Println(pathname, "Document Metadata is missing")
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

		bnsRequest.Urls = urls

		sproxyResponses := bns.AsyncHttpGetpageType(&bnsRequest)

		// AsyncHttpGetPageType should already  close [defer resp.Body.Clsoe()] all open connections
		bnsResponses := make([]bns.BnsResponse, len, len)
		var pagemd []byte
		for i, v := range sproxyResponses {
			if err := v.Err; err == nil {
				resp := v.Response
				body := *v.Body
				bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body)
				bnsResponses[i] = bnsResponse
				page = "p" + strconv.Itoa(i+1)
				if Image {
					writeImage(outDir, page, media, &bnsResponse.Image)
				}
				if Meta {
					if pagemd, err = base64.Decode64(resp.Header["X-Scal-Usermd"][0]); err == nil {
						writeMeta(outDir, page, pagemd)
					}
				}
			}
		}

	case "getObject":
		var (
			err           error
			encoded_docmd string
			docmd         []byte
		)
		media = "binary"
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
			writeMeta(outDir, "", docmd)
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
		sproxyResponses := bns.AsyncHttpGetBlobs(&bnsRequest, getHeader)
		bnsResponses := make([]bns.BnsResponse, len, len)
		var pagemd []byte
		bnsRequest.Client = &http.Client{}
		for i, v := range sproxyResponses {
			if err := v.Err; err == nil { //
				resp := v.Response
				body := *v.Body
				usermd := resp.Header["X-Scal-Usermd"][0]
				bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body) // bnsImage is a Go structure
				page = "p" + strconv.Itoa(i+1)
				if Image {
					writeImage(outDir, page, media, &bnsResponse.Image)
				}
				if Meta {
					if pagemd, err = base64.Decode64(usermd); err == nil {
						writeMeta(outDir, page, pagemd)
						bnsResponse.Pagemd = pagemd
					}
				}
				bnsResponses[i] = bnsResponse
			}
		}

	case "getPageType":

		pathname = pathname + "/" + page
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)
		var pagemd []byte

		bnsRequest := bns.HttpRequest{
			Hspool: sproxyd.HP,
			Client: client,
		}

		bnsRequest.Media = media

		if resp, err := bns.GetPageType(&bnsRequest, pathname); err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body)
			writeImage(outDir, page, media, &bnsResponse.Image)
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
