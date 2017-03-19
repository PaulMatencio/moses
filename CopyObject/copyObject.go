package main

//  Copy  all the objects (pages + document metatdata) of a document  from one Ring  ( moses-dev) to another Ring (Moses-Prod)
//
//   GET he document metatada from the source  Ring
//   PUT  the document metadata  to the destination Ring
//     For every object ( header+ tiff+ png + pdf) of the document
//         GET The Object  from the source Ring
//         PUT the object  to the source Ring
//
//
//  Check the config file sproxyd/conf/<default config file> moses-dev for more detail before running this program
//  The <default config file>  can be changed via the -config parm
//

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
	"time"

	base64 "moses/user/base64j"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"

	// hostpool "github.com/bitly/go-hostpool"
)

var (
	action, config, env, targetEnv, logPath, outDir, application, testname, hostname, pn, page, trace, test, meta, image, media, doconly string
	Trace, Meta, Image, CopyObject, Test, Doconly                                                                                        bool
	pid                                                                                                                                  int
	timeout                                                                                                                              time.Duration
)

func usage() {
	default_config := "moses-dev"
	usage := "\n\nUsage:\n\nCopyObject  -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/storage]" +
		"\n -pn <Patent number>  \n -env <Source environment> \n -targEnv <Target environment> \n -t <trace 0/1>  \n -test <test mode 0/1>"

	what := "\nFunction:\n\n Copy  all the objects (pages + document metatdata) of a document  from one Ring  (Ex:moses-dev) to another Ring (Ex:moses-Prod)" +
		"\n GET he document metatada from the source  Ring " +
		"\n PUT  the document metadata  to the destination Ring" +
		"\n For every object ( header+ tiff+ png + pdf) of the document" +
		"\n      GET The Object  from the source Ring" +
		"\n      PUT the object  to the source Ring" +
		"\n\nCheck the config file $HOME/sproxyd/conf/<default config file name> moses-dev for more detail regarding source and destination Rings before running this program" +
		"\nThe <default config file name>:<" + default_config + "> can be changed via the -config parm  "

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
	defaultConfig := "moses-dev"
	flag.Usage = usage
	flag.StringVar(&config, "config", defaultConfig, "Config file")
	flag.StringVar(&env, "env", "", "Environment")
	flag.StringVar(&targetEnv, "targetEnv", "", "Target Environment")
	flag.StringVar(&trace, "t", "0", "Trace")     // Trace
	flag.StringVar(&testname, "T", "copyDoc", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&test, "test", "1", "Run copy in test mode")
	flag.StringVar(&doconly, "doconly", "0", "Only update the document meta")
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)
	Test, _ = strconv.ParseBool(test)
	sproxyd.Test = Test
	Doconly, _ = strconv.ParseBool(doconly)

	action = "CopyObject"
	application = "copyObject"
	if len(pn) == 0 {
		fmt.Println("Error:\n-pn <DocumentId> is missing, what Document objects do you want to copy ?")
		usage()
	}
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	usr, _ := user.Current()
	homeDir := usr.HomeDir
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
			if trace == "0" {
				goLog.Init(ioutil.Discard, io.Writer(inf), io.Writer(waf), io.Writer(erf))

			} else {
				goLog.Init(io.Writer(trf), io.Writer(inf), io.Writer(waf), io.Writer(erf))
				goLog.Trace.Println(hostname, pid, "Start", application, action)
			}
		}
	}
	var (
		err           error
		encoded_docmd string
		docmd         []byte
	)
	directory.SetCPU("100%")
	start := time.Now()
	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP, // source sproxyd servers IP address and ports
		Client: &http.Client{},
		Media:  media,
	}

	media = "binary"
	if len(env) == 0 {
		env = sproxyd.Env
	}
	if len(targetEnv) == 0 {
		targetEnv = sproxyd.TargetEnv
	}
	pathname := env + "/" + pn
	targetPath := targetEnv + "/" + pn
	url := pathname
	targetUrl := targetPath

	// Get the document metadata
	if encoded_docmd, err = bns.GetEncodedMetadata(&bnsRequest, url); err == nil {
		if docmd, err = base64.Decode64(encoded_docmd); err != nil {
			goLog.Error.Println(err)
			os.Exit(2)
		}
	} else {
		goLog.Error.Println(err)
		os.Exit(2)
	}

	// convert the json metadata into a go structure
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
		bnsRequest.Hspool = sproxyd.TargetHP // Set Target sproxyd servers
		// Write the document metadata to the destination with no buffer
		// we could only update the meta data : TODO
		bns.CopyBlob(&bnsRequest, targetUrl, buf0, header, Test)

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
			urls[i] = pathname + "/p" + strconv.Itoa(i+1)
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
			if err := sproxydResponse.Err; err == nil { //
				resp := sproxydResponse.Response                                            /* http response */ // http response
				body := *sproxydResponse.Body                                               // http response                                                          /* copy of the body */ // http body response
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
		sproxydResponses := bns.AsyncHttpCopyBlobs(bnsResponses)

		num200 := 0
		for _, sproxydResponse := range sproxydResponses {
			resp := sproxydResponse.Response
			resp.Body.Close()
			goLog.Trace.Println(sproxydResponse.Url, resp.StatusCode)
			if resp.StatusCode == 200 {
				num200++
			} else {
				goLog.Error.Println(sproxydResponse.Url, sproxydResponse.Err, resp.StatusCode)
			}
			resp.Body.Close()
		}

		if num200 < num {
			fmt.Println("Some pages of ", pn, " could not be copied, Check the error log for more details", num, num200)
		} else {
			fmt.Println("All the pages of ", pn, " are copied", num, num200)
		}
	}
	duration = time.Since(start)
	fmt.Println("Total copy elapsed time:", duration)
	goLog.Info.Println("Total copy elapsed time:", duration)
}
