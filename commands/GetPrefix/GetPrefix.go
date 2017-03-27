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
	// "strings"

	"bufio"
	directory "moses/directory/lib"
	files "moses/user/files/lib"
	"os/user"
	"path"
	"time"

	sindexd "moses/sindexd/lib"

	hostpool "github.com/bitly/go-hostpool"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	action, lim, prefix, marker, pubdate, count, config, debug, delimiter, force, test, concurrent string
	prefixs, markers                                                                               []string
	Count, Debug, Delimiter, Concurrent                                                            bool
	//Test       bool
	maxinput, bulkindex, keys, iIndex, logPath string
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
	flag.StringVar(&iIndex, "index", "PN", "Index Table <PN or PD>")
	flag.StringVar(&concurrent, "C", "true", "Use Goroutine when it is possible")
	flag.StringVar(&count, "count", "false", "Count the number")
	flag.StringVar(&config, "config", "moses-dev", "Default Config file")
	flag.Parse()
	if len(prefix) == 0 {
		usage()
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
	Concurrent, _ = strconv.ParseBool(concurrent)
	if Debug {
		goLog.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		goLog.Init(os.Stdout, l, l, os.Stderr)
	}

	Limit, _ := strconv.Atoi(lim)
	Count, _ = strconv.ParseBool(count)
	Concurrent, _ := strconv.ParseBool(concurrent)
	sindexd.Debug = Debug
	sindexd.Delimiter = "/"
	directory.Action = action

	if err := directory.SetCPU("100%"); err != nil {
		goLog.Error.Println(err)
	}

	start := time.Now()
	Ind_Specs := directory.GetIndexSpec(iIndex)
	var response *directory.HttpResponse
	Nextmarker := true

	usr, _ := user.Current()
	homeDir := usr.HomeDir
	pref := "Prefixs"
	filedir := path.Join(homeDir, pref)
	if !files.Exist(filedir) {
		_ = os.MkdirAll(filedir, 0755)
	}
	filename := filedir + "/" + prefix

	f, _ := os.Create(filename)

	w := bufio.NewWriter(f)
	for Nextmarker {
		fmt.Println(markers)
		response = directory.GetSerialPrefix(iIndex, prefix, delimiter, marker, Limit, Ind_Specs)
		keys, nextMarker := directory.GetResponse(response)

		for _, v := range keys {
			v = v + "\n"
			if _, err := w.WriteString(v); err != nil {
				goLog.Error.Println("Error writing file", filename, err)
				os.Exit(10)
			}
		}
		fmt.Println("Next =>", nextMarker, len(keys))

		if len(nextMarker) == 0 {
			Nextmarker = false
		}
		marker = nextMarker
	}
	w.Flush()

	goLog.Info.Println("Concurrent:", Concurrent, "Elasped:", time.Since(start))
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
