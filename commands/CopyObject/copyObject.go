package main

//  Copy  all the objects (pages + document metatdata) of a document  from one Ring  ( moses-dev) to another Ring (Moses-Prod)
//
//  ATTENTION ====>    USE copyPNs instead
//
//  Check the config file sproxyd/conf/<default config file> moses-dev for more detail before running this program
//  The <default config file>  can be changed via the -config parm
//

import (
	// directory "directory/lib"
	"bytes"
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
	"time"

	base64 "moses/user/base64j"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"

	// hostpool "github.com/bitly/go-hostpool"
)

var (
	config, srcEnv, targetEnv, logPath, outDir, testname,
	pn, page, media, doconly string
	Trace, Meta, Image, CopyObject, Test, Doconly bool
	timeout                                       time.Duration
	action                                        = "CopyObject"
	application                                   = "copyObject"
	defaultConfig                                 = "moses-dev"
	pid                                           = os.Getpid()
	hostname, _                                   = os.Hostname()
	usr, _                                        = user.Current()
	homeDir                                       = usr.HomeDir
)

func usage() {
	// default_config := "moses-dev"
	usage := "\n\nUsage:\n\nCopyObject  -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/storage]" +
		"\n -pn <Patent number>  \n -srcEnv <Source environment> \n -targEnv <Target environment> \n -t <trace 0/1>  \n -test <test mode 0/1>"

	what := "\nFunction:\n\n Copy  all the objects (pages + document metatdata) of a document  from one Ring  (Ex:moses-dev) to another Ring (Ex:moses-Prod)" +
		"\n GET he document metatada from the source  Ring " +
		"\n PUT  the document metadata  to the destination Ring" +
		"\n For every object ( header+ tiff+ png + pdf) of the document" +
		"\n      GET The Object  from the source Ring" +
		"\n      PUT the object  to the source Ring" +
		"\n\nCheck the config file $HOME/sproxyd/conf/<default config file name> moses-dev for more detail regarding source and destination Rings before running this program" +
		"\nThe <default config file name>:<" + defaultConfig + "> can be changed via the -config parm  "

	fmt.Println(what, usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
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
	flag.StringVar(&config, "config", defaultConfig, "Config file")
	flag.StringVar(&srcEnv, "srcEnv", "", "Environment")
	flag.StringVar(&targetEnv, "targetEnv", "", "Target Environment")
	flag.BoolVar(&Trace, "t", false, "Trace")     // Trace
	flag.StringVar(&testname, "T", "copyDoc", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.BoolVar(&Test, "test", false, "Run copy in test mode")
	flag.StringVar(&doconly, "doconly", "0", "Only update the document meta")
	flag.Parse()
	// Trace, _ = strconv.ParseBool(trace)
	/*
		Meta, _ = strconv.ParseBool(meta)
		Image, _ = strconv.ParseBool(image)
	*/
	// Test, _ = strconv.ParseBool(test)
	sproxyd.Test = Test
	Doconly, _ = strconv.ParseBool(doconly)

	if len(pn) == 0 {
		fmt.Println("Error:\n-pn <DocumentId> is missing, what Document objects do you want to copy ?")
		usage()
	}

	if testname != "" {
		testname += string(os.PathSeparator)
	}

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

		fmt.Println("INFO: Using config Hosts=>", sproxyd.Host, sproxyd.Driver, sproxyd.Env)
		fmt.Println("INFO: Using config target Hosts=>", sproxyd.TargetHost, sproxyd.TargetDriver, sproxyd.TargetEnv)
		fmt.Println("INFO: Logs Path=>", logPath)
	} else {
		// sproxyd.HP = hostpool.NewEpsilonGreedy(sproxyd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
		fmt.Println(err, "WARNING: Using defaults :", "\nHosts=>", sproxyd.Host, sproxyd.TargetHost, "\nEnv", sproxyd.Env, sproxyd.TargetEnv)
		fmt.Println("$HOME/sproxyd/config/" + config + " must exist and well formed")
		os.Exit(100)
	}

	// init logging

	if logPath == "" {
		fmt.Println("WARNING: Using default logging")
		goLog.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	} else {
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
			if !Trace {
				goLog.Init(ioutil.Discard, io.Writer(inf), io.Writer(waf), io.Writer(erf))

			} else {
				goLog.Init(io.Writer(trf), io.Writer(inf), io.Writer(waf), io.Writer(erf))
				goLog.Trace.Println(hostname, pid, "Start", application, action)
			}
		}
	}

	bns.SetCPU("100%")

	start := time.Now()
	var (
		err           error
		encoded_docmd string
		docmd, docmd1 []byte
	)

	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP, // source sproxyd servers IP address and ports
		Client: &http.Client{},
		Media:  media,
	}

	media = "binary"
	if len(srcEnv) == 0 {
		srcEnv = sproxyd.Env
	}
	if len(targetEnv) == 0 {
		targetEnv = sproxyd.TargetEnv
	}
	srcPath := srcEnv + "/" + pn
	targetPath := targetEnv + "/" + pn
	url := srcPath
	targetUrl := targetPath

	// Get the document metadata
	if encoded_docmd, err, _ = bns.GetEncodedMetadata(&bnsRequest, url); err == nil {
		if len(encoded_docmd) > 0 {
			if docmd1, err = base64.Decode64(encoded_docmd); err != nil {
				goLog.Error.Println(err)
				os.Exit(2)
			}
		} else {
			goLog.Error.Println("Metadata is missing for ", srcPath)
			os.Exit(2)
		}
	} else {
		goLog.Error.Println(err)
		os.Exit(2)
	}

	// convert the json metadata into a go structure
	docmeta := bns.DocumentMetadata{}

	//docmd = bytes.Replace(docmd1, []byte("\n"), []byte(""), -1)
	docmd = bytes.Replace(docmd1, []byte(`\n`), []byte(``), -1)
	if err := json.Unmarshal(docmd, &docmeta); err != nil {
		goLog.Error.Println(docmeta)
		goLog.Error.Println(err)
		os.Exit(2)
	} else {
		header := map[string]string{
			"Usermd": encoded_docmd,
		}
		buf0 := make([]byte, 0)
		bnsRequest.Hspool = sproxyd.TargetHP // Set Target sproxyd servers
		// Write the document metadata to the destination with no buffer
		// we could only update the meta data : TODO
		bns.CopyBlob(&bnsRequest, targetUrl, buf0, header)

	}
	var duration time.Duration

	// update all the pages if requested
	if !Doconly {

		num := docmeta.TotalPage
		urls := make([]string, num, num)
		targetUrls := make([]string, num, num)
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "application/binary"
		for i := 0; i < num; i++ {
			urls[i] = srcPath + "/p" + strconv.Itoa(i+1)
			targetUrls[i] = targetPath + "/p" + strconv.Itoa(i+1)
		}
		bnsRequest.Urls = urls
		bnsRequest.Hspool = sproxyd.HP // Set source sproxyd servers
		bnsRequest.Client = &http.Client{}
		// Get all the pages from the source Ring
		sproxyResponses := bns.AsyncHttpGetBlobs(&bnsRequest, getHeader)
		// Build a response array of BnsResponse array to be used to update the pages  of  destination sproxyd servers
		bnsResponses := make([]bns.BnsResponse, num, num)

		// bnsRequest.Client = &http.Client{}
		for i, sproxydResponse := range sproxyResponses {
			if err := sproxydResponse.Err; err == nil {
				resp := sproxydResponse.Response                                            // http response
				body := *sproxydResponse.Body                                               /* copy of the body */ // http body response
				bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body) // bnsResponse is a Go structure
				bnsResponses[i] = bnsResponse
				resp.Body.Close()
			}
		}
		duration = time.Since(start)
		fmt.Println("Get elapsed time:", duration)
		goLog.Info.Println("Get elapsed time:", duration)

		// var sproxydResponses []*sproxyd.HttpResponse
		//   new &http.Client{}  and hosts pool are set to the target by the AsyncHttpCopyBlobs
		//  			sproxyd.TargetHP
		sproxydResponses := bns.AsyncHttpPutBlobs(bnsResponses)

		num200 := 0
		if !sproxyd.Test {
			for _, v := range sproxydResponses {
				resp := v.Response
				url := v.Url
				switch resp.StatusCode {
				case 200:
					goLog.Trace.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Key"])
					num200++
				case 412:
					goLog.Warning.Println(hostname, pid, url, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist")

				case 422:
					goLog.Error.Println(hostname, pid, url, resp.Status, resp.Header["X-Scal-Ring-Status"])
				default:
					goLog.Warning.Println(hostname, pid, url, resp.Status)
				}
				// close all the connection
				resp.Body.Close()
			}

			fmt.Println("\nPublication id:", pn, num, " Pages in;", num200, " Pages out")
			if num200 < num {
				goLog.Warning.Println("\nPublication id:", pn, num, " Pages in;", num200, " Pages out")
			} else {
				goLog.Info.Println("\nPublication id:", pn, num, " Pages in;", num200, " Pages out")
			}
		}
	}
	duration = time.Since(start)
	fmt.Println("Total copy elapsed time:", duration)
	goLog.Info.Println("Total copy elapsed time:", duration)
}
