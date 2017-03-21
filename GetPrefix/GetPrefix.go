// sindexd project main.go
package main

import (
	"flag"
	"fmt"
	"log"
	goLog "moses/user/goLog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	directory "moses/directory/lib"

	sindexd "moses/sindexd/lib"

	hostpool "github.com/bitly/go-hostpool"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	action, lim, prefix, marker, pubdate, count, config, debug, delimiter, force, test, concurrent, memstat, reset, lo string
	prefixs, markers                                                                                                   []string
	Count, Debug, Delimiter, Force, Memstat                                                                            bool
	//Test       bool
	maxinput, bulkindex, keys, iIndex, inputFile, outputDir, logPath string
)

/*
type HttpResponse struct { // used for get prefix
	pref     string
	response *sindexd.Response
	err      error
}
*/

// sindexd -action Gp -prefix "GB" -index PN -delimiter "/"  -debug true
//
func usage() {
	usage := "\nFunction==> Directory functions\n\nUsage: \nsindexd -action  Ci/Di/AMe/UMe/Uea/Ge/De/Gp/Gc/Sp " +
		"\n-prefix 'p1,p2,p3..'\n-limit n (prefix)\n-debug  0/1 " +
		"\n-Delimiter 0/1" +
		"\n\nDefault Options\n\n"

	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {

	flag.Usage = usage
	flag.StringVar(&lim, "limit", "500", "Limit the number of fetched keys per Get Prefix")
	flag.StringVar(&marker, "marker", "", "Start with this Marker (Key) for the Get Prefix ")
	flag.StringVar(&pubdate, "pd", "18000101", "Default Publication date")
	flag.StringVar(&debug, "debug", "false", "Debug mode")
	flag.StringVar(&delimiter, "delimiter", "", "Delimiter value")
	flag.StringVar(&prefix, "prefix", "", "Prefix Key")
	flag.StringVar(&iIndex, "index", "", "Index Table <PN or PD>")
	flag.StringVar(&concurrent, "C", "true", "Use Goroutine when it is possible")
	flag.StringVar(&count, "count", "false", "Count the number")
	flag.StringVar(&config, "config", "s11", "Default Config file")
	flag.Parse()
	if len(action) == 0 {
		usage()
	}
	if len(iIndex) == 0 {
		if action != "St" && action != "Gc" && action != "Sp" {
			fmt.Println("-index table is missing")
			usage()
		}
	}
	if len(config) != 0 {

		if Config, err := sindexd.GetParmConfig(config); err == nil {
			logPath = Config.GetLogPath()
			// hostpool.NewEpsilonGreedy is set by the SetNewHost method as following
			// HP = hostpool.NewEpsilonGreedy(Config.Hosts, 0, &hostpool.LinearEpsilonValueCalculator{})
			sindexd.SetNewHost(Config)
			fmt.Println("INFO: Using config Hosts", sindexd.Host, logPath)
		} else {
			sindexd.HP = hostpool.NewEpsilonGreedy(sindexd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
			fmt.Println(err, "WARNING: Using default Hosts:", sindexd.Host)
		}
	}

	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated.
	// MaxBackups is the maximum number of old log files to retain
	// Make sure the directory of the log file exists and the application has the write autorization
	action := "GetPrefix"
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
	if Debug {
		goLog.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		goLog.Init(os.Stdout, l, l, os.Stderr)
	}
	//  Create hosts pool
	/*
		hlist := strings.Split(sindexd.Url, ",")
		sindexd.HP = hostpool.NewEpsilonGreedy(hlist, 0, &hostpool.LinearEpsilonValueCalculator{})
	*/
	keys = strings.TrimSpace(keys)
	Limit, _ := strconv.Atoi(lim)
	Max, _ := strconv.Atoi(maxinput)
	sindexd.Maxinput = int64(Max)
	Memstat, _ = strconv.ParseBool(memstat)
	Force, _ = strconv.ParseBool(force)
	Count, _ = strconv.ParseBool(count)
	Concurrent, _ := strconv.ParseBool(concurrent)
	sindexd.Debug = Debug
	sindexd.Delimiter = "/"
	directory.PubDate = pubdate
	directory.Action = action
	sindexd.Memstat = Memstat
	sindexd.Test, _ = strconv.ParseBool(test)
	if err := directory.SetCPU("100%"); err != nil {
		goLog.Error.Println(err)
	}
	//client := &http.Client{}
	start := time.Now()
	//
	// Buid the index Specification based on the country code
	//
	Ind_Specs := directory.GetIndexSpec(iIndex)

	// Get prefix
	var (
		responses []*directory.HttpResponse
		markers   []string
	)
	prefixs := strings.Split(prefix, ",")
	if len(marker) != 0 {
		markers = strings.Split(marker, ",")
	}
	numpref := len(prefixs)
	if numpref == 0 {
		goLog.Error.Println("prefix keys  are missing")
		os.Exit(3)
	} else if numpref == 1 {
		Concurrent = false
	}
	if !Concurrent {
		responses = directory.GetSerialPrefix(iIndex, prefixs, delimiter, markers, Limit, Ind_Specs)
	} else {
		responses = directory.GetAsyncPrefix(iIndex, prefixs, delimiter, markers, Limit, Ind_Specs)
	}
	time1 := time.Since(start)
	if Count {
		m, nextMarker := directory.CountResponse(responses)
		for k, v := range m {
			goLog.Info.Println("Count:", k, v)
		}
		goLog.Info.Println("Next marker:", nextMarker)
	} else {
		directory.PrintResponse(responses)
	}
	goLog.Info.Println("Concurrent:", Concurrent, "Elasped:", time1)

	sindexd.HP.Close()
}

func check(f string, start time.Time, resp *http.Response, err error) {
	if err != nil {
		goLog.Error.Println("Function:", f, err)
	} else {
		response := sindexd.GetResponse(resp)
		goLog.Info.Println("Function:", f, "sindexd.Response:", response.Status, response.Reason, "Duration:", time.Since(start))

	}
}
