package main

import (
	directory "github.com/paulmatencio/moses/directory/lib"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	// "github.com/emicklei/go-restful/log"
	// "github.com/s3/gLog"

	"io/ioutil"
	bns "github.com/paulmatencio/moses/bns/lib"
	sproxyd "github.com/paulmatencio/moses/sproxyd/lib"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	base64 "github.com/paulmatencio/moses/user/base64j"
	file "github.com/paulmatencio/moses/user/files/lib"
	goLog "github.com/paulmatencio/moses/user/goLog"

	"github.com/bradfitz/slice"
)

var (
	action, config, env, logPath, outDir, runname, subpages,
	hostname, pn, page, trace, test, meta, image, media, ranges string
	Trace, Meta, Image, CopyObject, Test ,async bool
	pid                             int
	Size                            int64 =0
	timeout                              time.Duration
	application                          = "moses"
	Config                               sproxyd.Configuration
	err                                  error
)

func usage() {

	usage := "DocumentGet: \n -action [action] -config  <config>  [sproxyd configfile]\n" +
		"-pn Document path [/CC/Pn/KC]\n" +
		"-page [page number]\n"

	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeMeta(outDir string, page string, metadata []byte) {

	if !Meta {

		return
	}

	if err := checkOutdir(outDir); err != nil {
		os.MkdirAll(outDir,7644)
		goLog.Error.Println(err)
		return
	}

	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + ".md"
	goLog.Trace.Println("myfile:", myfile)

	err := ioutil.WriteFile(myfile, metadata, 0644)
	check(err)
}


func writeMetaError(outDir string, page string) {

	if !Meta {

		return
	}

	if err := checkOutdir(outDir); err != nil {
		os.MkdirAll(outDir,7644)
		goLog.Error.Println(err)
		return
	}

	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + ".404"
	goLog.Trace.Println("myfile:", myfile)

	err := ioutil.WriteFile(myfile,[]byte{}, 0644)
	check(err)
}

func writeImage(outDir string, page string, media string, body *[]byte) {

	if !Image {
		return
	}

	if body == nil {
		// 404 or something wrong
		goLog.Error.Printf("Body of %s %s is nil", page, media)
		return
	}

	if err := checkOutdir(outDir); err != nil {
		goLog.Error.Println(err)
		return
	}

	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + "." + strings.ToLower(media)
	// goLog.Trace.Println("myfile:", myfile)
	err := ioutil.WriteFile(myfile, *body, 0644)
	check(err)
}

func checkOutdir(outDir string) (err error) {

	if len(outDir) == 0 {
		err = errors.New("Please specify an output directory with -outDir argument")
	} else if !file.Exist(outDir) {
		err = os.MkdirAll(outDir, 0755)
	}
	return err
}

func BuildSubPagesRanges(action string, bnsRequest *bns.HttpRequest, pathname string) (string, error) {
	var (
		pagesranges string
		err         error
	)
	/* Get the meta data of the document */
	docmeta := bns.DocumentMetadata{}
	if err = docmeta.GetMetadata(bnsRequest, pathname); err == nil {
		//  Compute pages ranges based on the action value
		pagesranges = docmeta.GetPagesRanges(action)
	}
	return pagesranges, err
}

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "action", "", "<getObject> <chkPageMeta> <getPageMeta> <getPageType> <getDocumentMeta> <getDocumentType> <getPagesRanges> <getSubpages><getAbstract>, <getDescription>, <getClaims>, <getDrawings>, <getCitations>, <getDNASequence>, <getBiblio>, <getAmendement")
	flag.StringVar(&config, "config", "moses-prod", "Config file")
	flag.StringVar(&env, "env", "", "Environment")
	flag.StringVar(&trace, "trace", "0", "Trace")   // Trace
	flag.StringVar(&test, "test", "0", "Test mode") // Test mode
	flag.StringVar(&meta, "meta", "0", "Save object meta in output Directory")
	flag.StringVar(&image, "image", "0", "Save object image  type in output Directory")
	flag.StringVar(&runname, "runname", "", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&page, "page", "1", "page number")
	flag.StringVar(&ranges, "ranges", "", "multiple pages ranges")
	flag.BoolVar(&async, "async", true, "asynchronous request")
	// flag.StringVar(&subpages, "subpages", "biblio", "multiple pages ranges")
	flag.StringVar(&media, "media", "tiff", "media type: tiff/png/pdf")
	flag.StringVar(&outDir, "outDir", "", "output directory")

	flag.Parse()
	Trace, _ = strconv.ParseBool(trace)
	Meta, _ = strconv.ParseBool(meta)
	Image, _ = strconv.ParseBool(image)
	Test, _ = strconv.ParseBool(test)
	sproxyd.Test = Test

	if len(action) == 0 {
		usage()
	}
	if len(pn) == 0 {
		fmt.Println("-pn <DocumentId> is missing")
	}
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	// Check input parameters
	if runname == "" {
		runname = action + "_" + env + "_"
		runname += time.Now().Format("2006-01-02:15:04:05.00")
	}
	runname += string(os.PathSeparator)

	/* INIT CONFIG */
	if Config, err = sproxyd.InitConfig(config); err != nil {
		os.Exit(12)
	}
	if len(outDir) == 0 {
		outDir = path.Join(homeDir, Config.GetOutputDir())
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

	directory.SetCPU("100%")
	start := time.Now()
	page = "p" + page
	if len(env) == 0 {
		env = sproxyd.Env
	}
	pathname := env + "/" + pn
	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP,
		Client: &http.Client{
			Timeout:   sproxyd.ReadTimeout,
			Transport: sproxyd.Transport,
		},
		Media: media,
	}
	n := 0
	switch action {
	case "getPageMeta":
		Meta = true
		pathname = pathname + "/" + page
		if pagemd, err, status := bns.GetPageMetadata(&bnsRequest, pathname); err == nil {

			writeMeta(outDir, page, pagemd)
		} else {
			goLog.Error.Println(err,status)
		}
		n++

	case "chkPageMeta":
		Meta = true
		pathname = pathname + "/" + page
		if pagemd, err, status := bns.ChkPageMetadata(&bnsRequest, pathname); err == nil {
			writeMeta(outDir, page, pagemd)
		} else {
			goLog.Error.Printf("Err: %v  Status: %d",err,status)
			writeMetaError(outDir, page)
		}
		n++
	case "getDocumentMeta":
		// the document's  metatadata is the metadata the object given <pathname>
		// bnsRequest.Path = pathname
		Meta = true
		if docmd, err, statusCode := bns.GetDocMetadata(&bnsRequest, pathname); err == nil {
			goLog.Info.Println("Document Metadata=>\n", string(docmd))
			if len(docmd) != 0 {
				docmeta := bns.DocumentMetadata{}
				if err := json.Unmarshal(docmd, &docmeta); err != nil {
					goLog.Error.Println(err, docmd, &docmeta)
				} else {
					writeMeta(outDir, "", docmd)
				}
			} else if statusCode == 404 {
				goLog.Warning.Printf("Document %s is not found", pathname)
			} else {
				goLog.Warning.Printf("Document's %s metadata is missing", pathname)
			}
		} else {
			goLog.Error.Println(err)
		}
		n++

	case "getDocumentType":
		var (
			num  int
			err  error
			Page string
		)
		docmeta := bns.DocumentMetadata{}
		if err = docmeta.GetMetadata(&bnsRequest, pathname); err != nil {
			goLog.Error.Println(err)
			os.Exit(2)
		}
		// build []urls of pages  of the document to be fecthed
		//num := docmeta.TotalPage
		if num, err = docmeta.GetPageNumber(); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		urls := make([]string, num, num)

		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)

		for i := 0; i < num; i++ {
			urls[i] = pathname + "/p" + strconv.Itoa(i+1)
		}

		bnsRequest.Urls = urls
		sproxyResponses := bns.AsyncHttpGetpageType(&bnsRequest)
		bnsResponses := make([]bns.BnsResponseLi, num, num)

		for i, v := range sproxyResponses {
			if err := v.Err; err == nil {
				n++
				resp := v.Response
				body := *v.Body
				bnsResponse := bns.BuildBnsResponseLi(resp, getHeader["Content-Type"], &body)
				bnsResponses[i] = bnsResponse
				/*
					page := bnsResponse.Page

						Page := "p" + strconv.Itoa(page)
						if Image {
							writeImage(outDir, Page, media, bnsResponse.Image)
						}
						if Meta {
							writeMeta(outDir, Page, bnsResponse.Pagemd)
						}
				*/
				defer resp.Body.Close()
			}
		}
		// Sort the bnsResponse array by page number

		slice.SortInterface(bnsResponses[:], func(i, j int) bool {
			return bnsResponses[i].Page < bnsResponses[j].Page
		})
		for _, bnsResponse := range bnsResponses {
			Page = "p" + strconv.Itoa(bnsResponse.Page)
			if *bnsResponse.Image != nil {
				Size += int64(len(*bnsResponse.Image))
			}
			if Image {
				writeImage(outDir, Page, media, bnsResponse.Image)
			}
			if Meta {
				writeMeta(outDir, Page, bnsResponse.Pagemd)
			}
		}

	case "getPagesRanges", "getAbstract", "getDescription", "getClaims", "getDrawings", "getCitations", "getDNASequence", "getBiblio", "getAmendement":
		var (
			Page   string
			pagesa []string
		)
		section := action[3:]
		if section == "PagesRanges" {
			// pagesranges := "5:7,17:25"
			if len(ranges) != 0 {
				pagesa, _ = bns.BuildPagesRanges(ranges)
			} else {
				goLog.Warning.Println("-ranges is missing")
				os.Exit(2)
			}
		} else {
			fmt.Println("....", section)
			if ranges, _ = BuildSubPagesRanges(section, &bnsRequest, pathname); len(ranges) > 0 {

				pagesa, _ = bns.BuildPagesRanges(ranges)
			} else {
				goLog.Warning.Println("ranges is empty")
				os.Exit(2)
			}
		}
		var (
			num       = len(pagesa)
			urls      = make([]string, num, num)
			getHeader = map[string]string{
				"Content-Type": "image/" + strings.ToLower(media),
			}
		)

		for i, page := range pagesa {
			urls[i] = pathname + "/p" + page
		}

		bnsRequest.Urls = urls
		sproxyResponses := bns.AsyncHttpGetpageType(&bnsRequest)
		bnsResponses := make([]bns.BnsResponseLi, num, num)

		for k, v := range sproxyResponses {
			if err := v.Err; err == nil {
				n++
				resp := v.Response
				body := *v.Body
				bnsResponse := bns.BuildBnsResponseLi(resp, getHeader["Content-Type"], &body)
				bnsResponses[k] = bnsResponse
				defer resp.Body.Close()
			}
		}
		// Sort the bnsResponse array by page number

		slice.SortInterface(bnsResponses[:], func(i, j int) bool {
			return bnsResponses[i].Page < bnsResponses[j].Page
		})

		for _, v := range bnsResponses {
			if v.Image != nil {
				Size += int64(len(*v.Image))
			}
			Page = "p" + strconv.Itoa(v.Page)
			if Image {
				writeImage(outDir, Page, media, v.Image)
			}
			if Meta {
				writeMeta(outDir, Page, v.Pagemd)
			}
		}

	case "getObject":
		var (
			err           error
			encoded_docmd string
			docmd         []byte
			statusCode    int
			url           = pathname
		)
		media = "binary"

		// Get the document metadata
		if encoded_docmd, err, statusCode = bns.GetEncodedMetadata(&bnsRequest, url); err == nil {
			if len(encoded_docmd) > 0 {
				if docmd, err = base64.Decode64(encoded_docmd); err != nil {
					goLog.Error.Println(err)
					os.Exit(2)
				}
				goLog.Trace.Println("Document Metadata=>", string(docmd))
			} else if statusCode == 404 {
				goLog.Error.Printf("Document %s is not found/n", pathname)
			} else {
				goLog.Error.Printf("Document's %s metadata is missing/n", pathname)
			}

		} else {
			goLog.Error.Println(err)
			os.Exit(2)
		}

		docmeta := bns.DocumentMetadata{}
		if err := json.Unmarshal(docmd, &docmeta); err != nil {
			goLog.Error.Printf("Metadata is of %s invalid /n", pathname)
			goLog.Error.Println(string(docmd), docmeta, err)
			os.Exit(2)
		} else {
			writeMeta(outDir, "", docmd)
		}

		var (
			num       = docmeta.TotalPage
			urls      = make([]string, num, num)
			getHeader = map[string]string{
				"Content-Type": "application/binary",
			}
		)

		for i := 0; i < num; i++ {
			urls[i] = pathname + "/p" + strconv.Itoa(i+1)
		}
		bnsRequest.Urls = urls
		bnsRequest.Hspool = sproxyd.HP
		sproxyResponses := bns.AsyncHttpGetBlobs(&bnsRequest, getHeader)
		bnsResponses := make([]bns.BnsResponse, num, num)

		for k, v := range sproxyResponses {

			if err := v.Err; err == nil { //
				n++
				resp := v.Response
				body := *v.Body
				bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body) // bnsImage is a Go structure
				page := bnsResponse.PageNumber
				Size += int64(len(bnsResponse.Image))
				if Image {
					writeImage(outDir, page, media, &bnsResponse.Image)
				}
				if Meta {
					writeMeta(outDir, page, bnsResponse.Pagemd)
				}
				bnsResponses[k] = bnsResponse
			}
		}

	case "getPageType":

		pathname = pathname + "/" + page
		getHeader := map[string]string{
			"Content-Type": "image/" + strings.ToLower(media),
		}
		bnsRequest.Media = media
		if resp, err := bns.GetPageType(&bnsRequest, pathname); err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body)
			page = bnsResponse.PageNumber
			Size += int64(len(bnsResponse.Image))
			writeImage(outDir, page, media, &bnsResponse.Image)
			writeMeta(outDir, page, bnsResponse.Pagemd)

		} else {
			goLog.Error.Println(action, pathname, err)
		}

	default:
		goLog.Info.Println("-action <action value> is missing")
	}
	duration := time.Since(start)
	fmt.Printf("\nTotal elapsed time: %v - Total number of pages: %d - Total size(KB): %d\n", duration,n, Size/1024.0)
	goLog.Info.Println("Total Get pages elapsed times", duration, " for ", n, " pages ")
}
