// sindexd project main.go
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	directory "github.com/moses/directory/lib"
	sindexd "github.com/moses/sindexd/lib"
	files "github.com/moses/user/files/lib"
	// goLog "github.com/moses/user/goLog"
	goLog "github.com/s3/gLog"
	// "net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"

	hostpool "github.com/bitly/go-hostpool"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	lim, prefix, marker, pubdate, count, config, debug,
	delimiter, force, test, concurrent string
	prefixs, markers                           []string
	Count, Debug, Delimiter, Concurrent        bool
	maxinput, bulkindex, keys, iIndex, logPath string
	action                                     = "GetPrefix"
	usr, _                                     = user.Current()
	homeDir                                    = usr.HomeDir
)

func usage() {
	usage := "\nGetPrefix  -prefix 'p1,p2,p3..'\n-limit n (prefix)\n-debug  0/1 " +
		"\n-Delimiter 0/1" +
		"\n\nDefault Options\n\n"

	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {

	flag.Usage = usage
	flag.StringVar(&lim, "limit", "10000", "Limit the number of fetched keys per Get Prefix")
	flag.StringVar(&marker, "marker", "", "Start with this Marker (Key) for the Get Prefix ")
	flag.StringVar(&debug, "debug", "false", "Debug mode")
	flag.StringVar(&delimiter, "delimiter", "", "Delimiter value")
	flag.StringVar(&prefix, "prefix", "", "Prefix Key")
	flag.StringVar(&iIndex, "index", "PN", "Index Table <PN,PD,BN>")
	flag.StringVar(&concurrent, "C", "true", "Use Goroutine when it is possible")
	flag.StringVar(&count, "count", "false", "Count the number")
	flag.StringVar(&config, "config", "moses-prod", "Default Config file")
	flag.Parse()
	var (
		Limit, _      = strconv.Atoi(lim)
		Concurrent, _ = strconv.ParseBool(concurrent)
	)
	Count, _ = strconv.ParseBool(count)
	if len(prefix) == 0 {
		usage()
	}

	if len(config) != 0 { // always different than

		if Config, err := sindexd.GetConfig(config); err == nil {
			logPath = Config.GetLogPath()
			// hostpool.NewEpsilonGreedy is set by the SetNewHost method as following
			// HP = hostpool.NewEpsilonGreedy(Config.Hosts, 0, &hostpool.LinearEpsilonValueCalculator{})
			if pnOidSpec := Config.GetPnOidSpec(); len(pnOidSpec) != 0 {
				sindexd.PnOidSpec = pnOidSpec
			}
			if pdOidSpec := Config.GetPdOidSpec(); len(pdOidSpec) != 0 {
				sindexd.PdOidSpec = pdOidSpec
			}
			if jsOidSpec := Config.GetJsOidSpec(); len(jsOidSpec) != 0 {
				sindexd.JsOidSpec = jsOidSpec
			}


			sindexd.SetNewHost(Config)
			fmt.Println("INFO: Using config Hosts", sindexd.Host, logPath)
		} else {
			sindexd.HP = hostpool.NewEpsilonGreedy(sindexd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
			fmt.Println(err, "WARNING: Using default Hosts:", sindexd.Host)
			os.Exit(2)
		}
	} else {
		err := errors.New("Config file is missing")
		sindexd.HP = hostpool.NewEpsilonGreedy(sindexd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
		fmt.Println(err, "WARNING: Using default Hosts:", sindexd.Host)
	}

	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated.
	// MaxBackups is the maximum number of old log files to retain
	// Make sure the directory of the log file exists and the application has the write authorization

	logfile := logPath + string(os.PathSeparator) + action + "_" + iIndex + ".log"
	l := &lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    500, // megabytes
		MaxBackups: 2,
		MaxAge:     30, //days
	}
	log.SetOutput(l)
	// Create  Log categories : Trace, Info, Warning, Error
	Debug, _ = strconv.ParseBool(debug)
	Concurrent, _ = strconv.ParseBool(concurrent)
	if Debug {
		goLog.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr,ioutil.Discard,ioutil.Discard)
	} else {
		goLog.Init(os.Stdout, l, l, os.Stderr,ioutil.Discard,ioutil.Discard)
	}

	sindexd.Debug = Debug
	sindexd.Delimiter = "/"
	directory.Action = action

	if err := directory.SetCPU("100%"); err != nil {
		goLog.Error.Printf("Error %v", err)
	}
	var (
		start      = time.Now()
		Ind_Specs  = directory.GetIndexSpec(iIndex)
		response   *directory.HttpResponse
		Nextmarker = true

		pref    = "Prefixs"
		filedir = path.Join(homeDir, pref)
	)

	if !files.Exist(filedir) {
		_ = os.MkdirAll(filedir, 0755)
	}
	filename := filedir + "/" + prefix
	f, _ := os.Create(filename)
	w := bufio.NewWriter(f)
	for Nextmarker {
		//fmt.Println(markers)
		response = directory.GetSerialPrefix(iIndex, prefix, delimiter, marker, Limit, Ind_Specs)
		keys, nextMarker := directory.GetResponse(response)
		for _, v := range keys {
			v += "\n"
			if _, err := w.WriteString(v); err != nil {
				goLog.Error.Printf("\nError writing file %s : %v", filename, err)
				os.Exit(10)
			}
		}
		fmt.Printf("\nNext => %s %d\n", nextMarker, len(keys))

		if len(nextMarker) == 0 {
			Nextmarker = false
		}

		marker = nextMarker
	}
	w.Flush()
	goLog.Info.Println("Concurrent:", Concurrent, "Elasped:", time.Since(start))
	sindexd.HP.Close()
}
