package main

/*   ./DocumentGet  -action getPage -media pdf   -page p3  -pn /HR/P20020309/A2 -t 1   */
import (
	directory "directory/lib"
	bns "moses/bns/lib"
	/* "encoding/json" */
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	hostpool "github.com/bitly/go-hostpool"
	//base64 "moses/user/base64j"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"
)

type DocumentMetadata struct {
	PubId struct {
		CountryCode string `json: "countryCode`
		PubNumber   string `json: "pubNumber"`
		KindCode    string `json: "kindCode"`
	} `json: "PubId,omitempty"`

	BnsId struct {
		CountryCode string `json: "countryCode`
		PubNumber   string `json: "pubNumber"`
		KindCode    string `json: "kindCode"`
	} `json: "bnsId,omitempty"`

	DocId             string `json:"docId`
	PublicationOffice string `json:"publicationOffice`
	FamilyId          string `json:"familyId"`
	TotalPage         int    `json:totalPage"`
	DocType           string `json:docType"`
	PubDate           string `json:pubDate"`
	LoadDate          string `json:loadDate"`
	Copyright         string `json:"copyright,omitempty"`

	LinkPubId []struct {
		CountryCode string `json: "countryCode`
		PubNumber   string `json: "pubNumber"`
		KindCode    string `json: "kindCode"`
	} `json: "linkPubId,omitemty`

	MultiMedia struct {
		Tiff  bool `json:"tiff"`
		Png   bool `json:"png"`
		Pdf   bool `json:"pdf"`
		Video bool `json:"video"`
	} `json:"multiMedia"`

	AbsRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"absRangePageNumber,omitempty"`

	AmdRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"amdRangePageNumber,omitempty"`

	BibliRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"bibliRangePageNumber,omitempty"`

	ClaimsRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"claimsRangePageNumber,omitempty"`

	DescRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"descRangePageNumber,omitempty"`

	DrawRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"drawRangePageNumber,omitempty"`

	SearchRepRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"searchRepRangePageNumber,omitempty"`

	DnaSequenceRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"dnaSequenceRangePageNumber,omoitempty"`

	ApplicantCitationsRangePageNumber []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"applicantCitationsRangePageNumber,omitempty"`

	Classification []string `json:"classification,omitempty"`
}

type Pagemeta struct {
	DocumentID struct {
		CountryCode  string `json:"countryCode"`
		KindCode     string `json:"kindCode"`
		PatentNumber string `json:"patentNumber"`
	} `json:"documentId"`
	MultiMedia struct {
		Pdf   bool `json:"pdf"`
		Png   bool `json:"png"`
		Tiff  bool `json:"tiff"`
		Video bool `json:"video"`
	} `json:"multiMedia"`
	PageIndicator []string `json:"pageIndicator"`
	PageLength    int      `json:"pageLength"`
	PageNumber    int      `json:"pageNumber"`
	PdfOffset     struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pdfOffset,omitempty"`
	PngOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pngOffset,omitempty"`
	PublicationOffice string `json:"publicationOffice"`
	RotationCode      struct {
		Pdf  int `json:"pdf"`
		Png  int `json:"png"`
		Tiff int `json:"tiff"`
	} `json:"rotationCode"`
	TiffOffset struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"tiffOffset,omitempty"`
}

var (
	action, config, logPath, outDir, application, testname, hostname, pn, page, trace, media string
	Trace                                                                                    bool
	pid                                                                                      int
	timeout                                                                                  time.Duration
)

func usage() {

	usage := "DocumentGet: \n -action <action> -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/storage]\n" +
		"-pn pn -page page"

	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "action", "", "<getPageMeta> <getDocumentMeta> <getPage> <getDocument> <Getpagerange>")
	flag.StringVar(&config, "config", "chord", "Config file")
	flag.StringVar(&trace, "t", "0", "Trace")    // Trace
	flag.StringVar(&testname, "T", "getDoc", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&page, "page", "", "page number")
	flag.StringVar(&media, "media", "tiff", "media type: tiff/png/pdf")
	flag.StringVar(&outDir, "outDir", "/home/paul/outPath", "output directory")
	Trace, _ = strconv.ParseBool(trace)
	flag.Parse()
	if len(action) == 0 {
		usage()
	}

	application = "DocumentGet"
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	if testname != "" {
		testname += string(os.PathSeparator)
	}

	if len(config) != 0 {

		if Config, err := sproxyd.GetConfig(config); err == nil {
			logPath = Config.GetLogPath()
			sproxyd.SetNewProxydHost(Config)
			sproxyd.Driver = Config.GetDriver()
			fmt.Println("INFO: Using config Hosts", sproxyd.Host, sproxyd.Driver, logPath)
		} else {
			sproxyd.HP = hostpool.NewEpsilonGreedy(sproxyd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
			fmt.Println(err, "WARNING: Using default Hosts:", sproxyd.Host)
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

	//goLog.Init0(logPath, testname, application, action, Trace)
	directory.SetCPU("100%")
	client := &http.Client{}
	start := time.Now()
	pathname := "test/" + pn
	switch action {
	case "getPageMeta":
		pathname = pathname + "/" + page
		usermd, err := bns.GetPageMetadata(client, pathname)
		if err == nil {
			goLog.Info.Println(string(usermd))
		} else {
			goLog.Error.Println(err)
		}
	case "getDocumentMeta":
		// the document metatadata is

		usermd, err := bns.GetDocMetadata(client, pathname)
		if err == nil {
			goLog.Info.Println(string(usermd))
		} else {
			goLog.Error.Println(err)
		}
	case "getPage":
		pathname = pathname + "/" + page
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)
		resp, err := bns.GetPageType(client, pathname, getHeader)
		if err == nil {
			defer resp.Body.Close()
			var body []byte
			body, _ = ioutil.ReadAll(resp.Body)
			myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + "." + strings.ToLower(media)
			goLog.Trace.Println("myfile:", myfile)
			err := ioutil.WriteFile(myfile, body, 0644)
			check(err)
			goLog.Info.Println(len(body))
		} else {
			goLog.Error.Println(action, pathname, err)
		}
	default:
		goLog.Info.Println("-action is missing")
	}
	goLog.Info.Println(time.Since(start))
}
