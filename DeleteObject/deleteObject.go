package main

// delete documents  on  the Target environment
// update can be used instead of delete + copy again

import (
	directory "directory/lib"
	"encoding/json"
	// "errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	bns "moses/bns/lib"
	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"
)

var (
	action, config, srcEnv, targetEnv, logPath, application, testname, hostname, pn, page, trace, test string
	Trace, Meta, Image, CopyObject, Test                                                               bool
	pid                                                                                                int
	timeout                                                                                            time.Duration
)

func usage() {

	usage := "DeleteObject: \n -action <action> -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/storage]\n" +
		"-pn <pn>"

	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func deleteBlob(bnsRequest *bns.HttpRequest, url string) {

	result := bns.AsyncHttpDeleteBlob(bnsRequest, url)
	// if Test mode return
	if sproxyd.Test {
		goLog.Trace.Printf("Deleting URL => %s \n", result.Url)
		return
	}
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
	flag.StringVar(&srcEnv, "srcEnv", "prod", "Environment")
	flag.StringVar(&targetEnv, "targetEnv", "moses-prod", "Environment")
	flag.StringVar(&trace, "t", "0", "Trace")       // Trace
	flag.StringVar(&testname, "T", "deleteDoc", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&test, "test", "0", "Run copy in test mode")
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	sproxyd.Test, _ = strconv.ParseBool(test)

	action = "DeleteObject"
	application = "deleteObject"
	if len(pn) == 0 {
		fmt.Println("-pn <DocumentId> is missing")
	}

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
			sproxyd.SetNewProxydHost(Config)
			sproxyd.Driver = Config.GetDriver()
			sproxyd.SetNewTargetProxydHost(Config)
			sproxyd.TargetDriver = Config.GetTargetDriver()
			fmt.Println("INFO: Using config Hosts", sproxyd.Host, sproxyd.Driver, logPath)
			fmt.Println("INFO: Using config target Hosts", sproxyd.TargetHost, sproxyd.TargetDriver, logPath)
		} else {
			fmt.Println(err, "WARNING: Using defaults :", "\nHosts=>", sproxyd.Host, sproxyd.TargetHost, "\nEnv", sproxyd.Env, sproxyd.TargetEnv)
			fmt.Println("$HOME/sproxyd/config/" + config + " must exist and well formed")
			os.Exit(100)
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
	start := time.Now()
	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP,
		Client: &http.Client{},
	}
	var (
		err           error
		encoded_docmd string
		docmd         []byte
	)
	// READ THE DOCUMENT FROM THE SOURCE ENV to GET ITS METADATA
	// SOURCE AND TARGET COULD BE THE SAME . CHECK the config file

	targetPath := targetEnv + "/" + pn

	if encoded_docmd, err = bns.GetEncodedMetadata(&bnsRequest, targetPath); err == nil {
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
	}

	//  DELETE THE DOCUMENT ON THE TARGET ENVIRONMENT
	num := docmeta.TotalPage
	fmt.Println("len => ", num)
	bnsRequest.Urls = make([]string, num, num)

	targetPath = targetEnv + "/" + pn

	docPath := targetPath

	bnsRequest.Hspool = sproxyd.TargetHP // set target sproxyd servers
	bnsRequest.Client = &http.Client{}
	//  DELETE ALL THE PAGES FIRST
	for i := 0; i < num; i++ {
		bnsRequest.Urls[i] = targetPath + "/p" + strconv.Itoa(i+1)
		url := bnsRequest.Urls[i]
		deleteBlob(&bnsRequest, url)
	}

	// DELETE THE DOC METADATA AFTER DELING ALL THE PAGES

	deleteBlob(&bnsRequest, docPath)

	duration := time.Since(start)
	fmt.Println("total detelete elapsed time:", duration)
	goLog.Info.Println("total detelete elapsed time:", duration)
}
