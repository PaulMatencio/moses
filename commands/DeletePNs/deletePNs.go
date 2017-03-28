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
	"bufio"
	"errors"
	"flag"
	"fmt"
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
	config, srcEnv, targetEnv, logPath, outDir,
	runname, hostname, pns, cpn, pnfile, page, trace, test,
	meta, image, media, doconly string
	Trace, Meta, Image, CopyObject, Test, Doconly bool
	pid, Cpn                                      int
	timeout, duration                             time.Duration
	scanner                                       *bufio.Scanner
	fp                                            *os.File
	Config                                        sproxyd.Configuration
	action, application                           string = "DeletePNs", "Moses"
	err                                           error
	defaultConfig                                 string = "moses-dev"
	start, start0                                 time.Time
	numloop, Numpns, NumpnsDone                   int = 0, 0, 0
	usr, _                                            = user.Current()
	homeDir                                           = usr.HomeDir
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
		"\n      	Delete the blob at the destination Ring" +
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
	flag.Usage = usage
	flag.StringVar(&config, "config", defaultConfig, "Config file")
	flag.StringVar(&srcEnv, "srcEnv", "", "Environment")
	flag.StringVar(&targetEnv, "targetEnv", "", "Target Environment")
	flag.StringVar(&trace, "trace", "0", "Trace") // Trace
	flag.StringVar(&runname, "runname", "", "")   // Test name
	flag.StringVar(&pns, "pns", "", "Publication numbers")
	flag.StringVar(&pnfile, "pnfile", "", "Publication numbers")
	flag.StringVar(&cpn, "cpn", "10", "Concurrent PN number")
	flag.StringVar(&test, "test", "0", "Run copy in test mode")
	flag.StringVar(&doconly, "doconly", "0", "Only update the document meta")
	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)
	sproxyd.Test, _ = strconv.ParseBool(test)
	Doconly, _ = strconv.ParseBool(doconly)
	Cpn, _ = strconv.Atoi(cpn)

	// Check input parameters
	if runname == "" {
		runname = action + "_"
		runname += time.Now().Format("2006-01-02:15:04:05.00")
	}
	runname += string(os.PathSeparator)
	if len(pnfile) > 0 {
		pnfile = path.Join(homeDir, pnfile)
		if scanner, err = file.Scanner(pnfile); err != nil {
			fmt.Println(err)
			os.Exit(10)
		}
	} else if len(pns) == 0 {
		fmt.Println("Error:\nOr -pn <DocumentId list separated by comma>  Either -pnfile <file name> is missing ?")
		usage()
	}

	/* INIT CONFIG */
	if Config, err = sproxyd.InitConfig(config); err != nil {
		os.Exit(12)
	}
	logPath = path.Join(homeDir, Config.GetLogPath())
	fmt.Printf("INFO: Logs Path=>%s", logPath)

	// init logging
	if defaut, trf, inf, waf, erf := goLog.InitLog(logPath, runname, application, action, Trace); !defaut {
		defer trf.Close()
		defer inf.Close()
		defer waf.Close()
		defer erf.Close()
	}

	var (
		pna    = strings.Split(pns, ",")
		start0 = time.Now()
		start  = start0
		stop   = false
	)

	if len(pns) == 0 {
		for !stop {
			if linea, err := file.ScanLines(scanner, Cpn); len(linea) > 0 && err == nil {
				start = time.Now()

				copyResponses := bns.AsyncDeletePns(linea, targetEnv)

				duration = time.Since(start)
				for _, copyResponse := range copyResponses {
					fmt.Printf("\nSource Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
					goLog.Info.Printf("\nSource Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
					if copyResponse.Num > 0 && copyResponse.Num == copyResponse.Num200 {
						NumpnsDone++
					}
				}
				numloop++
				Numpns = Numpns + len(linea)

			} else {
				if err != nil {
					goLog.Error.Println(err)
				}
				stop = true
			}
		}
	} else {
		copyResponses := bns.AsyncDeletePns(pna, targetEnv)
		Numpns = len(pna)
		duration = time.Since(start)
		for _, copyResponse := range copyResponses {
			fmt.Printf("\nSource Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
			goLog.Info.Printf("\nSource Url=%s,Error=%v,#Input=%d, #Ouput=%d, Duration %v", copyResponse.SrcUrl, copyResponse.Err, copyResponse.Num, copyResponse.Num200, duration)
			if copyResponse.Num > 0 && copyResponse.Num == copyResponse.Num200 {
				NumpnsDone++
			}
		}
	}

	fmt.Printf("\nTotal Elapsed Time %v \nNumber of PN's completed %d / Number of PN's", time.Since(start0), NumpnsDone, Numpns)
	goLog.Info.Printf("\nTotal Elapsed Time %v \nNumber of PN's completed %d / Number of PN's", time.Since(start0), NumpnsDone, Numpns)
}
