package main

//  Copy  all the objects (pages + document metatdata) of a document  from one Ring  ( moses-dev) to another Ring (Moses-Prod)
//
//   GET he document metatada from the source  Ring
//   PUT  the document metadata  to the destination Ring
//     For every object ( header+ tiff+ png + pdf) of the document
//         GET The Object  from the source Ring
//         PUT the object to the destination  Ring
//
//
//  Check the config file sproxyd/conf/<default config file> moses-dev for more detail before running this program
//  The <default config file>  can be changed via the -config parm
//

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	/*
		"io"
		"io/ioutil"
	*/
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
	config, srcEnv, targetEnv, logPath, outDir, runname,
	hostname, pns, cpn, pnfile, trace, test, meta, image,
	media, doconly string
	Trace, Meta, Image, CopyObject, Test, Doconly bool
	pid, Cpn                                      int
	timeout, duration                             time.Duration
	scanner                                       *bufio.Scanner
	action, application                           string = "CopyPNs", "Moses"
	numloop, Numpns, NumpnsDone                   int    = 0, 0, 0
	Config                                        sproxyd.Configuration
	err                                           error
	defaultConfig                                 string = "moses-dev"
	start, start0                                 time.Time
)

func usage() {
	default_config := "moses-dev"
	usage := "\n\nUsage:\n\nCopyPNs -config  <config file>  default is  $HOME/sproxyd/config/moses-dev]" +
		"\n -pns <List of PN separated by a comma> " +
		"\n -pnfile <filename> " +
		"\n -srcEnv <Source environment> \n -targEnv <Target environment> \n -t <trace 0/1>  \n -test <test mode 0/1>"

	what := "\nFunction:\n\nCopy PN's (Publication Numbers)  from one Ring  (Ex:moses-dev) to another Ring (Ex:moses-Prod)" +
		"\n" +
		"\nFor every PN { " +
		"\n     GET he PN's metatada from the source  Ring " +
		"\n     PUT the PN's metadata  to the destination Ring" +
		"\n     For every blob ( header+ tiff+ png + pdf) of the PN {" +
		"\n      	GET The Object  from the source Ring" +
		"\n      	PUT the object  to thedestination Ring" +
		"\n		{ " +
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

	flag.Usage = usage
	flag.StringVar(&config, "config", defaultConfig, "Config file")
	flag.StringVar(&srcEnv, "srcEnv", "", "Environment")
	flag.StringVar(&targetEnv, "targetEnv", "", "Target Environment")
	flag.StringVar(&trace, "trace", "0", "Trace") // Trace
	flag.StringVar(&runname, "runname", "", "")   // Test name
	flag.StringVar(&pns, "pns", "", "Publication numbers -pns PN1,PN2,PN3,PN4")
	flag.StringVar(&pnfile, "pnfile", "", "File of publication numbers, one PN per line  -pnfile filename")
	flag.StringVar(&cpn, "cpn", "10", "Concurrent number of PN's reading from -pnfile")
	flag.StringVar(&test, "test", "0", "Run copy in test mode")
	flag.StringVar(&doconly, "doconly", "0", "Copy  only the document meta")
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)
	sproxyd.Test, _ = strconv.ParseBool(test)
	Doconly, _ = strconv.ParseBool(doconly)
	Cpn, _ = strconv.Atoi(cpn)
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	// Check input parameters
	if runname == "" {
		runname += time.Now().Format("2006-01-02:15:04:05.00")
	}
	runname += string(os.PathSeparator)
	if len(pnfile) > 0 {
		pnfile = path.Join(homeDir, pnfile)
		if scanner, err = file.Scanner(pnfile); err != nil {
			os.Exit(10)
		}
	} else if len(pns) == 0 {
		fmt.Println("Error:\n-pn <DocumentId list separated by comma>  or -pnfile <file name> is missing ?")
		usage()
	}

	/* INIT CONFIG */
	if Config, err = sproxyd.InitConfig(config); err != nil {
		os.Exit(12)
	}

	fmt.Printf("INFO: Logs Path=>%s", logPath)

	if len(outDir) == 0 {
		outDir = path.Join(homeDir, Config.GetOutputDir())
	}
	logPath = path.Join(homeDir, Config.GetLogPath())

	// init logging

	if defaut, trf, inf, waf, erf := goLog.InitLog(logPath, runname, application, action, Trace); !defaut {
		defer trf.Close()
		defer inf.Close()
		defer waf.Close()
		defer erf.Close()
	}

	pna := strings.Split(pns, ",")
	start0 = time.Now()
	stop := false

	if len(pns) == 0 {
		//  Take  the PNs from a file
		//  Cpn is the number of concuurent PN's to be processed
		for !stop {
			if linea, _ := file.ScanLines(scanner, Cpn); len(linea) > 0 {
				start = time.Now()
				copyResponses := bns.AsyncCopyPns(linea, srcEnv, targetEnv)
				duration = time.Since(start)
				for _, copyResponse := range copyResponses {
					fmt.Printf("Source Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
					goLog.Info.Printf("Source Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
					if copyResponse.Num > 0 && copyResponse.Num200 == copyResponse.Num {
						NumpnsDone++
					}

				}
				numloop++
				Numpns = Numpns + len(linea)
			} else {
				stop = true
			}
		}
	} else {
		// take the PN's from the pna ( -pns PN1,PN2,PN3,PN4 )
		start = time.Now()
		copyResponses := bns.AsyncCopyPns(pna, srcEnv, targetEnv)
		Numpns = len(pna)
		duration = time.Since(start)
		for _, copyResponse := range copyResponses {
			fmt.Printf("Source Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
			goLog.Info.Printf("Source Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
			if copyResponse.Num > 0 && copyResponse.Num200 == copyResponse.Num {
				NumpnsDone++
			}
		}
	}

	fmt.Printf("Total Elapsed Time %v \nNumber of PN's completed %d / Number of PN's", time.Since(start0), NumpnsDone, Numpns)
	goLog.Info.Printf("Total Elapsed Time %v \nNumber of PN's completed %d / Number of PN's", time.Since(start0), NumpnsDone, Numpns)
}
