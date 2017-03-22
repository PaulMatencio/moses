package main

//  Delete  all the objects (pages + document metatdata) of a document  from one Ring  ( moses-dev) to another Ring (Moses-Prod)
//
//   GET he document metatada from the source  Ring
//   DELETE  the document metadata  to the destination Ring
//     For every object ( header+ tiff+ png + pdf) of the document
//
//         Delete  the object  on  the source Ring
//
//
//  Check the config file sproxyd/conf/<default config file> moses-dev for more detail before running this program
//  The <default config file>  can be changed via the -config parm
//

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	bns "moses/bns/lib"
	sproxyd "moses/sproxyd/lib"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	// hostpool "github.com/bitly/go-hostpool"
)

var (
	action, config, srcEnv, targetEnv, logPath, outDir, application, testname, hostname, pns, page, trace, test, meta, image, media, doconly string
	Trace, Meta, Image, CopyObject, Test, Doconly                                                                                            bool
	pid                                                                                                                                      int
	timeout                                                                                                                                  time.Duration
)

func usage() {
	default_config := "moses-dev"
	usage := "\n\nUsage:\n\nCopyPNs -config  <config file>  default is  $HOME/sproxyd/config/moses-dev]" +
		"\n -pns <List of PN separated by a comma>  \n -srcEnv <Source environment> \n -targEnv <Target environment> \n -t <trace 0/1>  \n -test <test mode 0/1>"

	what := "\nFunction:\nDelete PN's (Publication Numbers)  from one Ring  (Ex:moses-dev) to another Ring (Ex:moses-Prod)" +
		"\n" +
		"\nFor every PN { " +
		"\n     GET he PN's metatada from the target  Ring " +
		"\n     For every blob ( header+ tiff+ png + pdf) of the PN {" +
		"\n      	Delete the blob  to thedestination Ring" +
		"\n		{ " +
		"\n		Delete the PN's metadata " +
		"\n{ " +
		"\n" +
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
	flag.StringVar(&srcEnv, "srcEnv", "", "Environment")
	flag.StringVar(&targetEnv, "targetEnv", "", "Target Environment")
	flag.StringVar(&trace, "t", "0", "Trace")       // Trace
	flag.StringVar(&testname, "T", "deletePns", "") // Test name
	flag.StringVar(&pns, "pns", "", "Publication numbers")
	flag.StringVar(&test, "test", "1", "Run copy in test mode")
	flag.StringVar(&doconly, "doconly", "0", "Only update the document meta")
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)
	sproxyd.Test, _ = strconv.ParseBool(test)
	// sproxyd.Test = Test
	Doconly, _ = strconv.ParseBool(doconly)

	action = "DeletePNs"
	application = "deletePN"
	if len(pns) == 0 {
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
	pna := strings.Split(pns, ",")
	start := time.Now()
	targetEnv = sproxyd.TargetEnv
	copyResponses := bns.AsyncDeletePns(pna, targetEnv)
	duration := time.Since(start)
	for _, copyResponse := range copyResponses {
		fmt.Println(copyResponse.Err, copyResponse.Num, copyResponse.Num200)
		goLog.Info.Println(copyResponse.Err, copyResponse.Num, copyResponse.Num200)
	}
	fmt.Println("Total copy elapsed time:", duration)
	goLog.Info.Println("Total copy elapsed time:", duration)
}
