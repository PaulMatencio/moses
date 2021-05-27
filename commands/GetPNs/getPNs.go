package main

//  Copy  all the objects (pages + document metatdata) of a document  from one Ring  ( moses-dev) to another Ring (Moses-Prod)
//
//   GET he document metatada from the source  Ring
//   PUT  the document metadata  to the destination Ring
//     For every object ( header+ tiff+ png + pdf) of the document
//         GET The Object  from the source Ring
//
//  Check the config file sproxyd/conf/<default config file> moses-dev for more detail before running this program
//  The <default config file>  can be changed via the -config parm
//

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	bns "github.com/paulmatencio/moses/bns/lib"
	sproxyd "github.com/paulmatencio/moses/sproxyd/lib"
	file "github.com/paulmatencio/moses/user/files/lib"
	goLog "github.com/paulmatencio/moses/user/goLog"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	config, env, logPath, outDir, runname,
	hostname, pns, cpn, pnfile, trace, meta, image,
	media, doconly string
	Trace, Meta, Image, CopyObject, Doconly           bool
	pid, Cpn                                          int
	timeout, duration                                 time.Duration
	scanner                                           *bufio.Scanner
	action, application                               = "CopyPNs", "Moses"
	numloop, Numpns, NumpnsDone, Numpns404, NumpnsErr = 0, 0, 0, 0, 0
	Config                                            sproxyd.Configuration
	err                                               error
	defaultConfig                                     = "moses-dev"
	start, start0                                     time.Time
	usr, _                                            = user.Current()
	homeDir                                           = usr.HomeDir
)

func usage() {
	default_config := "moses-dev"
	usage := "\n\nUsage:\n\nGetPNs -config  <config file>  default is  $HOME/sproxyd/config/moses-dev]" +
		"\n -pns <string> [List of PN separated by a comma] " +
		"\n -pnfile <string> [a filename containing the Pns]" +
		"\n -cpn <number> [concurrent number of PNs to be procesed form the pnfile]" +
		"\n -env <string> [the Ring environment]" +
		"\n -trace <bool> [true/false]"

	what := "\nFunction:\n Get PN's (Publication Numbers) of a specific Ring(Ex:moses-dev))" +
		"\n For every PN of a list { " +
		"\n     GET he PN's metatada (TOC) " +
		"\n     For every object of the TOC {" +
		"\n      	GET the Object" +
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
	flag.StringVar(&env, "envv", "", "Environment")
	flag.BoolVar(&Trace, "trace", false, "Trace")       // Trace
	flag.StringVar(&runname, "runname", "", "Run name") // Test name
	flag.StringVar(&pns, "pns", "", "Publication numbers -pns PN1,PN2,PN3,PN4")
	flag.StringVar(&pnfile, "pnfile", "", "File of publication numbers, one PN per line  -pnfile filename")
	flag.StringVar(&cpn, "cpn", "10", "Concurrent number of PN's reading from -pnfile")
	flag.BoolVar(&Doconly, "doconly", false, "Copy only the document meta")
	flag.Parse()
	Cpn, _ = strconv.Atoi(cpn)
	cwd, _ := os.Getwd()

	// Check input parameters
	action = os.Args[0]
	if len(runname) == 0 {
		runname = action + "_"
		runname += time.Now().Format("2006-01-02:15:04:05.00")
	}
	runname += string(os.PathSeparator)
	if len(pnfile) > 0 {
		if pnfile[0:1] != string(os.PathSeparator) {
			pnfile = path.Join(cwd, pnfile)
		}
		if scanner, err = file.Scanner(pnfile); err != nil {

			fmt.Println(err)
			os.Exit(10)
		}
	} else if len(pns) == 0 {
		fmt.Println("Error:\n-pn <DocumentId list separated by comma>  or -pnfile <file name> is missing ?")
		usage()
	}

	/* INIT CONFIG */
	if Config, err = sproxyd.InitConfig(config); err != nil {
		fmt.Println(err)
		os.Exit(12)
	}
	logPath = path.Join(homeDir, Config.GetLogPath())
	fmt.Printf("INFO: Logs Path=>%s", logPath)

	if len(outDir) == 0 {
		outDir = path.Join(homeDir, Config.GetOutputDir())
	}

	// Init logging
	if defaut, trf, inf, waf, erf := goLog.InitLog(logPath, runname, application, action, Trace); !defaut {
		defer trf.Close()
		defer inf.Close()
		defer waf.Close()
		defer erf.Close()
	}
	var (
		pna    = strings.Split(pns, ",")
		start0 = time.Now()
		stop   = false
	)

	if len(pns) == 0 {
		//  Read PNs from a file
		//  Cpn is the number of concuurent PN's to be processed
		for !stop {
			if linea, _ := file.ScanLines(scanner, Cpn); len(linea) > 0 {
				start = time.Now()
				r, doc404, docErr := bns.AsyncGetPns(linea, env)
				Numpns404 += doc404
				NumpnsErr += docErr
				duration = time.Since(start)
				for _, v := range r {
					fmt.Printf("\nSource Url=%s,Error=%v,#Input=%d, #200=%d,#404= %d,#Err=%d,Duration=%v", v.SrcUrl, v.Err, v.Num, v.Num200, doc404, docErr, duration)
					goLog.Info.Printf("\nSource Url=%s,Error=%v,#Input=%d,#200=%d,#404=%d,#Err=%d,Duration=%v", v.SrcUrl, v.Err, v.Num, v.Num200, doc404, docErr, duration)
					if v.Num > 0 && v.Num200 == v.Num {
						NumpnsDone++
					}
				}
				numloop++
				Numpns += len(linea)
			} else {
				stop = true
			}
		}
	} else {
		// take the PN's from the pna ( -pns PN1,PN2,PN3,PN4 )
		start = time.Now()
		r, doc404, docErr := bns.AsyncGetPns(pna, env)
		Numpns = len(pna)
		Numpns404 += doc404
		NumpnsErr += docErr
		duration = time.Since(start)
		for _, v := range r {
			fmt.Printf("\nSource Url=%s,Error=%v,#Pages=%d, #200=%d,#404= %d,#Err=%d,Duration=%v", v.SrcUrl, v.Err, v.Num, v.Num200, doc404, docErr, duration)
			goLog.Info.Printf("\nSource Url=%s,Error=%v,#Pages=%d, #200=%d,#404=%d,#Err=%d,Duration=%v", v.SrcUrl, v.Err, v.Num, v.Num200, doc404, docErr, duration)
			if v.Num > 0 && v.Num200 == v.Num {
				NumpnsDone++
			}
		}
	}

	fmt.Printf("\nTotal Elapsed Time %v \n Total PN=%d,Done=%d,404=%d,Err=%d", time.Since(start0), Numpns, NumpnsDone, Numpns404, NumpnsErr)
	goLog.Info.Printf("\nTotal Elapsed Time %v \n Total PN=%d,Done=%d,404=%d,Err=%d", time.Since(start0), Numpns, NumpnsDone, Numpns404, NumpnsErr)
}
