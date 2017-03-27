package main

/*   ./DocumentGet  -action getPage -media pdf   -page p3  -pn /HR/P20020309/A2 -t 1   */
import (
	directory "directory/lib"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"io/ioutil"
	bns "moses/bns/lib"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	base64 "moses/user/base64j"
	file "moses/user/files/lib"
	goLog "moses/user/goLog"

	"github.com/bradfitz/slice"
)

var (
	action, config, env, logPath, outDir, runname,
	hostname, pn, page, trace, test, meta, image, media, pagesranges string
	Trace, Meta, Image, CopyObject, Test bool
	pid                                  int
	timeout                              time.Duration
	application                          = "moses"
	Config                               sproxyd.Configuration
	err                                  error
)

func usage() {

	usage := "DocumentGet: \n -action <action> -config  <config>, sproxyd configfile;default file is [$HOME/sproxyd/moses-dev]\n" +
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

func writeMeta(outDir string, page string, metadata []byte) {
	if !Meta {
		return
	}

	if err := checkOutdir(outDir); err != nil {
		goLog.Error.Println(err)
		return
	}

	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + ".md"
	goLog.Trace.Println("myfile:", myfile)
	err := ioutil.WriteFile(myfile, metadata, 0644)
	check(err)
}

func writeImage(outDir string, page string, media string, body *[]byte) {

	if !Image {
		return
	}

	if err := checkOutdir(outDir); err != nil {
		goLog.Error.Println(err)
		return
	}

	myfile := outDir + string(os.PathSeparator) + bns.RemoveSlash(pn) + page + "." + strings.ToLower(media)
	goLog.Trace.Println("myfile:", myfile)
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
	/*
		if docmd, err, statusCode := bns.GetDocMetadata(bnsRequest, pathname); err == nil {
			goLog.Trace.Println("Document Metadata=>", string(docmd))
			if len(docmd) != 0 {
				if err = json.Unmarshal(docmd, &docmeta); err != nil {
					goLog.Error.Println(docmd, docmeta, err)
					return "", err
				}
			} else if statusCode == 404 {
				goLog.Warning.Printf("Document %s is not found", pathname)
				return "", errors.New("Document not found")
			} else {
				goLog.Warning.Printf("Document's %s metadata is missing", pathname)
				return "", errors.New("Document metadata is missing")
			}
		} else {
			goLog.Error.Println(err)
			return "", err
		}
	*/
	if err = docmeta.GetMetadata(bnsRequest, pathname); err != nil {
		//  Compute pages ranges based on the action value
		pagesranges = docmeta.GetPagesRanges(action)
	}

	return pagesranges, err
}

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "action", "", "<getPageMeta> <getPageType> <getDocumentMeta> <getDocumentType> <agesRange>")
	flag.StringVar(&config, "config", "moses-dev", "Config file")
	flag.StringVar(&env, "env", "", "Environment")
	flag.StringVar(&trace, "trace", "0", "Trace")   // Trace
	flag.StringVar(&test, "test", "0", "Test mode") // Test mode
	flag.StringVar(&meta, "meta", "0", "Save object meta in output Directory")
	flag.StringVar(&image, "image", "0", "Save object image  type in output Directory")
	flag.StringVar(&runname, "runname", "", "") // Test name
	flag.StringVar(&pn, "pn", "", "Publication number")
	flag.StringVar(&page, "page", "1", "page number")
	flag.StringVar(&pagesranges, "pagesranges", "", "multiple pages ranges")
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
	client := &http.Client{}
	start := time.Now()
	page = "p" + page
	if len(env) == 0 {
		env = sproxyd.Env
	}
	pathname := env + "/" + pn
	bnsRequest := bns.HttpRequest{
		Hspool: sproxyd.HP,
		Client: client,
		Media:  media,
	}
	n := 0
	switch action {
	case "getPageMeta":
		Meta = true
		pathname = pathname + "/" + page
		// bnsRequest.Path = pathname
		if pagemd, err, _ := bns.GetPageMetadata(&bnsRequest, pathname); err == nil {
			writeMeta(outDir, page, pagemd)
		} else {
			goLog.Error.Println(err)
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
		docmeta := bns.DocumentMetadata{}
		if err = docmeta.GetMetadata(&bnsRequest, pathname); err != nil {
			goLog.Error.Println(err)
			os.Exit(2)
		}
		// build []urls of pages  of the document to be fecthed
		num := docmeta.TotalPage

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
				page := bnsResponse.Page
				Page := "p" + strconv.Itoa(page)
				if Image {
					writeImage(outDir, Page, media, bnsResponse.Image)
				}
				if Meta {
					writeMeta(outDir, Page, bnsResponse.Pagemd)
				}
				defer resp.Body.Close()

			}
		}
		// Sort the bnsResponse array by page number
		slice.Sort(bnsResponses[:], func(i, j int) bool {
			return bnsResponses[i].Page < bnsResponses[j].Page
		})
	case "PagesRanges", "Abstract", "Descrition", "Claims", "Drawings", "Citations", "DNASequence", "Biblio":

		var (
			Page   string
			pagesa []string
		)
		if action == "getPagesRanges" {
			// pagesranges := "5:7,17:25"
			pagesa, _ = bns.BuildPagesRanges(pagesranges)
		} else {
			pagesranges, _ = BuildSubPagesRanges(action, &bnsRequest, pathname)
			pagesa, _ = bns.BuildPagesRanges(pagesranges)
		}
		num := len(pagesa)
		urls := make([]string, num, num)
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)

		for i, page := range pagesa {
			urls[i] = pathname + "/p" + page
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
				//page = bnsResponse.Page
				// Page = "p" + strconv.Itoa(page)
				/*
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
		slice.Sort(bnsResponses[:], func(i, j int) bool {
			return bnsResponses[i].Page < bnsResponses[j].Page
		})
		for _, bnsResponse := range bnsResponses {
			Page = "p" + strconv.Itoa(bnsResponse.Page)
			if Image {
				writeImage(outDir, Page, media, bnsResponse.Image)
			}
			if Meta {
				writeMeta(outDir, Page, bnsResponse.Pagemd)
			}
		}

	case "getObject":
		var (
			err           error
			encoded_docmd string
			docmd         []byte
			statusCode    int
		)
		media = "binary"
		url := pathname

		// Get the document metadata
		if encoded_docmd, err, statusCode = bns.GetEncodedMetadata(&bnsRequest, url); err == nil {
			if len(encoded_docmd) > 0 {
				if docmd, err = base64.Decode64(encoded_docmd); err != nil {
					goLog.Error.Println(err)
					os.Exit(2)
				}
				goLog.Trace.Println("Document Metadata=>", string(docmd))
			} else if statusCode == 404 {
				goLog.Error.Printf("Document %s is not found", pathname)
			} else {
				goLog.Error.Printf("Document's %s metadata is missing", pathname)
			}

		} else {
			goLog.Error.Println(err)
			os.Exit(2)
		}

		docmeta := bns.DocumentMetadata{}
		if err := json.Unmarshal(docmd, &docmeta); err != nil {
			goLog.Error.Println("Metadata is invalid ", pathname)
			goLog.Error.Println(string(docmd), docmeta, err)
			os.Exit(2)
		} else {
			writeMeta(outDir, "", docmd)
		}

		num := docmeta.TotalPage
		urls := make([]string, num, num)
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "application/binary"

		for i := 0; i < num; i++ {
			urls[i] = pathname + "/p" + strconv.Itoa(i+1)
		}
		bnsRequest.Urls = urls
		bnsRequest.Hspool = sproxyd.HP
		sproxyResponses := bns.AsyncHttpGetBlobs(&bnsRequest, getHeader)
		bnsResponses := make([]bns.BnsResponse, num, num)
		bnsRequest.Client = &http.Client{}
		for i, v := range sproxyResponses {

			if err := v.Err; err == nil { //
				n++
				resp := v.Response
				body := *v.Body
				bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body) // bnsImage is a Go structure
				page := bnsResponse.PageNumber

				if Image {
					writeImage(outDir, page, media, &bnsResponse.Image)
				}
				if Meta {
					writeMeta(outDir, page, bnsResponse.Pagemd)
				}
				bnsResponses[i] = bnsResponse
			}
		}

	case "getPageType":

		pathname = pathname + "/" + page
		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)
		bnsRequest := bns.HttpRequest{
			Hspool: sproxyd.HP,
			Client: client,
		}
		bnsRequest.Media = media
		if resp, err := bns.GetPageType(&bnsRequest, pathname); err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			bnsResponse := bns.BuildBnsResponse(resp, getHeader["Content-Type"], &body)
			page = bnsResponse.PageNumber
			writeImage(outDir, page, media, &bnsResponse.Image)
			writeMeta(outDir, page, bnsResponse.Pagemd)

		} else {
			goLog.Error.Println(action, pathname, err)
		}

	default:
		goLog.Info.Println("-action <action value> is missing")
	}
	duration := time.Since(start)
	fmt.Println("Total get elapsed time:", duration, " to get ", n, " pages")
	goLog.Info.Println("Total get elapsed time:", duration, " to get ", n, " pages")
}
