package main

//
//  ATTENTION ====>    USE updatePNs instead
//
import (
	"bytes"
	directory "directory/lib"
	"encoding/json"
	"errors"
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
	action, config, env, targetEnv, logPath, outDir, application, testname, hostname, pn, page, trace, test, meta, image, media, doconly string
	Trace, Meta, Image, CopyObject, Test, Doconly                                                                                        bool
	pid                                                                                                                                  int
	timeout                                                                                                                              time.Duration
)

func usage() {

	default_config := "moses-dev"
	usage := "\n\nUsage:\n\nUpdateObject  -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/storage]" +
		"\n -pn <Patent number> \n -docOnly <0/1> \n -env <Source environment> \n -targEnv <Target environment> \n -t <trace 0/1>  \n -test <test mode 0/1>"

	what := "\nFunction:\n\n Copy  all the objects (pages + document metatdata) of a document  from one Ring  (Ex:moses-dev) to another Ring (Ex:moses-Prod)" +
		"\n GET he document metatada from the source  Ring " +
		"\n UPDATE  the document metadata  to the destination Ring" +
		"\n For every object ( header+ tiff+ png + pdf) of the document" +
		"\n      GET The Object  from the source Ring" +
		"\n      UPDATE the object on the source Ring" +
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

	flag.Usage = usage
	flag.StringVar(&config, "config", "moses-dev", "Config file")
	flag.StringVar(&env, "env", "", "Environment")
	flag.StringVar(&targetEnv, "targetEnv", "", "Target Environment")
	flag.StringVar(&trace, "t", "0", "Trace")       // Trace
	flag.StringVar(&testname, "T", "updateDoc", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&test, "test", "0", "Run copy in test mode")
	flag.StringVar(&doconly, "doconly", "0", "Only update the document meta")
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)
	sproxyd.Test, _ = strconv.ParseBool(test)

	Doconly, _ = strconv.ParseBool(doconly)
	action = "UpdateObject"
	application = "updateObject"
	if len(pn) == 0 {
		fmt.Println("-pn <DocumentId> is missing")
		usage()
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
			// sproxyd.HP = hostpool.NewEpsilonGreedy(sproxyd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
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
	// client := &http.Client{}
	start := time.Now()
	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP, // set the source  sproxyd servers
		Client: &http.Client{},
		Media:  media,
	}
	var (
		err           error
		encoded_docmd string
		docmd, docmd1 []byte
	)
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
		if len(encoded_docmd) > 0 {
			if docmd1, err = base64.Decode64(encoded_docmd); err != nil {
				goLog.Error.Println(err)
				os.Exit(2)
			}
		} else {
			goLog.Error.Println("Metadata is missing for ", pathname)
			os.Exit(2)
		}
	} else {
		goLog.Error.Println(err)
		os.Exit(2)
	}

	docmeta := bns.DocumentMetadata{}
	// docmd = bytes.Replace(docmd1, []byte("\n"), []byte(""), -1)
	docmd = bytes.Replace(docmd1, []byte(`\n`), []byte(``), -1)
	// convert the document metadata into go structure : docmeta
	if err := json.Unmarshal(docmd, &docmeta); err != nil {
		goLog.Error.Println(docmeta)
		goLog.Error.Println(err)
		os.Exit(2)
	} else {
		// Update the metadata of the  document with its source content
		header := map[string]string{
			"Usermd": encoded_docmd,
		}
		buf0 := make([]byte, 0)
		bnsRequest.Hspool = sproxyd.TargetHP // set the destination sproxyd servers
		// Update the Document metadat first  on the destination sproxyd servers

		bns.UpdateBlob(&bnsRequest, targetUrl, buf0, header)

	}
	var (
		duration time.Duration
		startw   time.Time
	)

	//  if not update only the document metadata
	if !Doconly {
		// update all the objects (pages) of the document
		// Get all the document pages 's content
		num := docmeta.TotalPage
		urls := make([]string, num, num)
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "application/binary"

		for i := 0; i < num; i++ {
			urls[i] = pathname + "/p" + strconv.Itoa(i+1)
		}
		bnsRequest.Urls = urls
		bnsRequest.Hspool = sproxyd.HP // set the source  sproxyd servers
		bnsRequest.Client = &http.Client{}
		sproxyResponses := bns.AsyncHttpGetBlobs(&bnsRequest, getHeader)
		duration = time.Since(start)
		fmt.Println("Time to Get:", duration)
		goLog.Info.Println("Time to Get:", duration)

		// Build a response array of BnsResponse array to be used to update the pages  of  destination sproxyd servers
		bnsResponses := make([]bns.BnsResponse, num, num)
		bnsRequest.Client = &http.Client{}

		for i, v := range sproxyResponses {
			if err := v.Err; err == nil {
				bnsResponses[i] = bns.BuildBnsResponse(v.Response, getHeader["Content-Type"], v.Body)
			}
		}

		startw = time.Now()

		// update the destination pages using the bnsResponses structure array
		// return an array of sproxyResponse structure
		//   new &http.Client{}  and hosts pool are set to the target by the AsyncHttpCopyBlobs
		//  			sproxyd.TargetHP

		/*  Check the result
		for k, _ := range bnsResponses {

			fmt.Println("...Response Status=>", bnsResponses[k].HttpStatusCode)
			fmt.Println("...Page Meta=>", string(bnsResponses[k].Pagemd))
			fmt.Println("...Image length=>", len(bnsResponses[k].Image))

		}
		*/
		// sproxydResponses := []sproxyd.HttpResponse{}

		sproxydResponses := bns.AsyncHttpUpdateBlobs(bnsResponses)
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
					goLog.Warning.Println(hostname, pid, url, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "does not exist")

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
	duration = time.Since(startw)
	fmt.Println("Time to Update", duration)
	goLog.Info.Println("Time to Update", duration)
	duration = time.Since(start)
	fmt.Println("Total update elapsed time:", duration)
	goLog.Info.Println("Total update elapsed time:", duration)
}
