// sindexd project main.go
package main

import (
	"flag"
	"fmt"
	"log"
	goLog "github.com/moses/user/goLog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	directory "github.com/moses/directory/lib"

	sindexd "github.com/moses/sindexd/lib"

	hostpool "github.com/bitly/go-hostpool"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	action, lim, prefix, marker, pubdate, config,
	delimiter, reset, lo string
	prefixs, markers                                                 []string
	Count, Debug, Test, Delimiter, Force, Memstat, Concurrent        bool
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
		"\n-prefix 'p1,p2,p3..'\n-limit n (prefix)\n-key 'k1,k2,k3,...'\n-input inputfile \n-marker key\n-debug  0/1 " +
		"\n-Delimiter 0/1 \n-force 0/1" +
		"\n-max  m (input lines)\n-bulk b (size bulk insert)\n-reset  0/1 " +
		"\n\nListe of actions:" +
		"\n Ci:Create indexes \n Di:Drop indexes\n AMe:Add entries(input files)\n UMe:Update entries (input files)" +
		"\n De:Delete entries(K)\n Dme:Delete entries(input file)" +
		"\n Ge:Get entries(K)\n Gp:Get prefix (-prefix)\n Gc:Get sindexd Config" +
		"\n\nDefault Options\n\n"

	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "action", "", "Value: Ci/Di/AMe/UMe/Uea/Ge/De/DMe/Gp/Gc/Sp")
	flag.StringVar(&lim, "limit", "500", "Limit the number of fetched keys per Get Prefix")
	flag.StringVar(&marker, "marker", "", "Start with this Marker (Key) for the Get Prefix ")
	flag.StringVar(&pubdate, "pd", "18000101", "Default Publication date")
	flag.BoolVar(&Debug, "debug", false, "Debug mode")
	flag.StringVar(&delimiter, "delimiter", "", "Delimiter value")
	flag.BoolVar(&Memstat, "memstat", false, "Print memory used stats")
	flag.StringVar(&reset, "reset", "0", "Reset the sindexd stats counter")
	flag.StringVar(&lo, "lo", "0", "Provide sindexd Low level stats")
	flag.BoolVar(&Force, "force", false, "Drop the directory even if it is not empty")
	flag.StringVar(&maxinput, "max", "10000", "maximum number of keys to index; 0 => no limit")
	flag.StringVar(&bulkindex, "bulk", "1000", "number of indexes per Add/Update indexes operations")
	flag.StringVar(&prefix, "prefix", "", "Prefix Key")
	flag.BoolVar(&Test, "test", false, "Test mode")
	flag.StringVar(&keys, "key", "", "Keys <seprated by comma>  to be fetched")
	flag.StringVar(&iIndex, "index", "", "Index Table <PN or PD>")
	flag.StringVar(&inputFile, "input", "", "Input file for indexing or Splitting")
	flag.StringVar(&outputDir, "outputDir", "", "Output directory for spliiting")
	flag.BoolVar(&Concurrent, "con", true, "Use Goroutine when it is possible")
	flag.BoolVar(&Count, "count", false, "Count the number")
	flag.StringVar(&config, "config", "moses-prod", "Default Config file")
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
			sindexd.SetNewHost(Config)
			fmt.Println("INFO: Using config Hosts", sindexd.Host, logPath)
		} else {
			sindexd.HP = hostpool.NewEpsilonGreedy(sindexd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
			fmt.Printf("errors %s parsing Config file %s", err, config)
			os.Exit(1)
		}
	}

	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated.
	// MaxBackups is the maximum number of old log files to retain
	// Make sure the directory of the log file exists and the application has the write autorization
	var CC string
	if len(inputFile) != 0 {
		inputa := strings.Split(inputFile, "/")
		CC = inputa[len(inputa)-1]
	}
	logfile := logPath + string(os.PathSeparator) + action + "_" + CC + "_" + iIndex + ".log"
	l := &lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    500, // megabytes
		MaxBackups: 2,
		MaxAge:     30, //days
	}
	log.SetOutput(l)
	// Create  Log categories : Trace, Info, Warning, Error

	if Debug {
		goLog.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		goLog.Init(os.Stdout, l, l, os.Stderr)
	}

	var (
		Limit, _    = strconv.Atoi(lim)
		Reset, _    = strconv.Atoi(reset)
		Lowlevel, _ = strconv.Atoi(lo)
		Bulk, _     = strconv.Atoi(bulkindex)
		Max, _      = strconv.Atoi(maxinput)
		client      = &http.Client{
			Timeout:   sindexd.ReadTimeout,
			Transport: sindexd.Transport,
		}
		resp  *http.Response
		err   error
		f     string
		start = time.Now()
	)

	keys = strings.TrimSpace(keys)
	sindexd.Maxinput = int64(Max)

	sindexd.Debug = Debug
	sindexd.Delimiter = "/"
	directory.PubDate = pubdate
	directory.Action = action
	sindexd.Memstat = Memstat
	sindexd.Test = Test
	if err := directory.SetCPU("100%"); err != nil {
		goLog.Error.Println(err)
	}

	//
	// Buid the index Specification based on the country code
	//
	Ind_Specs := directory.GetIndexSpec(iIndex)

	switch action {
	case "Ci":
		f = "Create directory" + iIndex
		for _, v := range Ind_Specs {
			// index := v
			if Debug {
				goLog.Info.Println("Create id:", v.Index_id, v.Vol_id, v.Specific, v.Cos)
			}

			resp, err = directory.Create(client, v)
			check(f, start, resp, err)
		}

	case "Di":
		f = "Drop directory" + iIndex
		client.Timeout = sindexd.DeleteTimeout
		for _, v := range Ind_Specs {
			// index := v
			if Debug {
				goLog.Info.Println("Delete id:", v.Index_id, v.Vol_id, v.Specific, v.Cos)
			}
			resp, err = directory.Drop(client, v, Force, true)
			check(f, start, resp, err)
		}

	case "AMe", "UMe":
		// Add keys , inputs are files extracted from DocDB
		var (
			// index *sindexd.Index_spec
			key string
		)
		start = time.Now()
		f = "Add/Update keys"
		pn := strings.Split(inputFile, "/")
		key = pn[len(pn)-1]
		index := Ind_Specs[key]
		if index == nil {
			goLog.Warning.Println("Could not find the Index Specication for key:", key)
			os.Exit(3)
		}
		directory.AddArray(iIndex, inputFile, client, index, Bulk)

	case "Ge", "De":
		var (
			specs     map[string][]string
			aKey      []string
			responses []*directory.HttpResponse
			start     = time.Now()
		)
		if aKey = strings.Split(keys, ","); len(keys) > 0 && len(aKey) > 0 {
			// sort the array of string
			sort.Strings(aKey)
		} else {
			goLog.Warning.Println("Input Keys are missing, Please specify the string of  keys separated by comma")
		}
		// Build an index
		specs = make(map[string][]string)
		for _, v := range aKey {
			// index := aKey[i][0:2]
			index := v[0:2]
			if Ind_Specs[index] == nil {
				index = "OTHER"
			}
			specs[index] = append(specs[index], v)
		}

		if !Concurrent {
			responses = directory.GetSerialKeys(specs, Ind_Specs)
		} else {
			responses = directory.GetAsyncKeys(specs, Ind_Specs)
		}
		time1 := time.Since(start)

		directory.PrintResponse(responses)
		goLog.Info.Println("Concurrent:", Concurrent, "Elasped:", time1)

	case "Gp":
		start := time.Now()
		var (
			responses []*directory.HttpResponse
			markers   []string
		)
		prefixs := strings.Split(prefix, ",")
		if len(marker) != 0 {
			markers = strings.Split(marker, ",")
		}
		l := len(prefixs)
		if l == 0 {
			goLog.Error.Println("prefix keys  are missing")
			os.Exit(3)
		} else if l == 1 {
			Concurrent = false
		}
		if !Concurrent {
			responses = directory.GetSerialPrefixs(iIndex, prefixs, delimiter, markers, Limit, Ind_Specs)
		} else {
			responses = directory.GetAsyncPrefixs(iIndex, prefixs, delimiter, markers, Limit, Ind_Specs)
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

	case "DMe":
		// Keys to be deleted are in inputFile
		f = "Delete keys"
		pn := strings.Split(inputFile, "/")
		key := pn[len(pn)-1]
		index := Ind_Specs[key]
		if index == nil {
			goLog.Warning.Println("Could not find the Index Specication for key:", key)
			os.Exit(3)
		}
		directory.DeleteArray(iIndex, inputFile, client, index, Bulk)

	case "St":
		f = "Get_Stats"
		if resp, err = directory.GetStats(client, Reset, Lowlevel); err != nil {
			goLog.Error.Println(f, err)
		} else {
			sindexd.PrintStats(f, resp)
		}

	case "Gc":
		f = "Get_Config"
		if resp, err = sindexd.GetSindexdConfig(client); err != nil {
			goLog.Error.Println(f, err)
		} else {
			sindexd.PrintConfig(f, resp)
		}
	case "Sp":
		if len(inputFile) != 0 && len(outputDir) != 0 {
			f = "Split_file"
			directory.SplitFile(inputFile, outputDir)
		} else {
			goLog.Info.Println("Please enter t -i <inut File to splitted> and -o <Output directory>  ")
		}

	default:
		goLog.Error.Println("You request a wrong action")
	}
	sindexd.HP.Close()
}

func check(f string, start time.Time, resp *http.Response, err error) {
	if err != nil {
		goLog.Error.Println("Function:", f, err)
	} else {
		response,_ := sindexd.GetResponse(resp)
		goLog.Info.Println("Function:", f, "sindexd.Response:", response.Status, response.Reason, "Duration:", time.Since(start))

	}
}
