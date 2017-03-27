package main

import (
	bns "bns/lib"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	directory "moses/directory/lib"
	base64 "moses/user/base64j"
	files "moses/user/files/lib"
	goLog "moses/user/goLog"
	"net/http"
	"os"
	"path"
	sproxyd "sproxyd/lib"
	"strconv"
	"strings"
	"time"

	hostpool "github.com/bitly/go-hostpool"
	//imaging "github.com/disintegration/imaging"
	"image"

	tiff "golang.org/x/image/tiff"
)

func usage() {
	usage := "DocumentCUD -action <A> -input <Directory> -config <Config file name> -testname <testname> "
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	action, inputDir, config, logPath, trace, application, testname, hostname string
	Trace                                                                     bool
	pid                                                                       int
	timeout                                                                   time.Duration
	err                                                                       error
)

func main() {

	const (
		LAYOUT    = "Jan 2, 2006 at 3:04pm (MST)"
		TIFF      = "Tiff"
		CONTAINER = "Container"
		USERMD    = "Usermd"
	)
	var (
		buf []byte
		/*    total_size, treq, tok, doc_size                 int64 */
		/* t_pages, req, ok, pages_to_write, pages_written int */
	)

	flag.Usage = usage
	flag.StringVar(&action, "action", "Test", "<Test/Create/Update/Delete>")
	flag.StringVar(&inputDir, "input", "", "<input directory to upload>")
	flag.StringVar(&config, "config", "bparc", "<Contain host name and drivers>")
	flag.StringVar(&testname, "testname", "test0", "")
	flag.StringVar(&trace, "trace", "0", "")

	flag.Parse()
	if len(inputDir) == 0 {
		usage()
	}
	if Trace, err = strconv.ParseBool(trace); err != nil {
		Trace = false
	}
	directory.SetCPU("100%")
	application = "DocumentCUD"
	ct := "application/binary"
	input_dir := inputDir
	tiff_path := path.Join(input_dir, TIFF)
	container_path := path.Join(input_dir, CONTAINER)
	container_exist, _ := files.Exists(container_path)
	pid = os.Getpid()
	hostname, _ = os.Hostname()

	if testname != "" {
		testname += string(os.PathSeparator)
	}
	if len(config) != 0 {

		if Config, err := sproxyd.GetConfig(config); err == nil {
			logPath = Config.GetLogPath()
			sproxyd.SetNewProxydHost(Config)
			fmt.Println("INFO: Using config Hosts", sproxyd.Host, logPath)
		} else {
			sproxyd.HP = hostpool.NewEpsilonGreedy(sproxyd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
			fmt.Println(err, "WARNING: Using default Hosts:", sproxyd.Host)
		}
	}
	// Replace following cod by
	// goLog.Init0(logPath, testname, application, action, Trace)

	if logPath == "" {
		fmt.Println("WARNING: Using default logging")
		goLog.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	} else {

		//
		logPath = logPath + string(os.PathSeparator) + testname
		if !files.Exist(logPath) {
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

	//
	// THE CONTAINER DIRECTORY MUST EXIST
	// CONTAINER DIRECTORY SHOULD HAVE BEEN  CREATED BY ST33toFiles.go
	//
	if container_exist {
		goLog.Info.Println("Start", application)
		cont_ent, _ := ioutil.ReadDir(container_path)
		/* lop on the number of documents directories */
		Begin := time.Now()
		for _, container_file := range cont_ent {
			var (
				err       error
				base_path string
				pathx     string
				fulldoc   string
				docmeta   bns.Documentmeta
				header    map[string]string
			)
			begin := time.Now()
			container_fn := container_file.Name()
			split := strings.Split(container_fn, ".")
			docid := split[0]
			doc_fn := path.Join(container_path, container_fn)
			if split[1] != "json" {
				goLog.Error.Println(container_fn, " is likely not a container file, not a json file?.")
				continue // SKIP THIS DOCUMENT
			}
			goLog.Trace.Println(doc_fn)
			// READ THE DOCUMENT META DATA

			buf, err = ioutil.ReadFile(doc_fn)

			if err != nil {
				goLog.Warning.Println(hostname, pid, err, "Reading", doc_fn)
				continue // SKIP THIS DOCUMENT
			}

			base_path = sproxyd.Proxy + "/" + sproxyd.Driver + "/BNS"

			// UNMARSHAL THE DOCUMENT METATA To GET THE DOCID STRUCT

			if err = json.Unmarshal(buf, &docmeta); err == nil {
				fulldoc = docmeta.DocumentID.CC + "/" + docmeta.DocumentID.PN + "/" + docmeta.DocumentID.KC
				pathx = base_path + "/" + fulldoc
			} else {
				goLog.Warning.Println(hostname, pid, err, "Unmarshalling", string(buf))
				continue
			}
			// ENCODED THE CONTAINER METADATA
			header = map[string]string{
				"Usermd":       base64.Encode64(buf),
				"Content-Type": ct,
			}

			client := &http.Client{}
			hok := false

			switch action {

			case "Create":
				buf0 := new(bytes.Buffer)
				if err, elapse := bns.PutPage(client, pathx, buf0, header); err == nil {
					goLog.Trace.Printf("%s path %s in %v", action, pathx, elapse)
					hok = true
				}

			case "Update":
				buf0 := new(bytes.Buffer)
				if err, elapse := bns.UpdatePage(client, pathx, buf0, header); err == nil {
					goLog.Trace.Printf("%s path %s in %v", action, pathx, elapse)
					hok = true
				}
			case "Delete":
				if err, elapse := bns.DeletePage(client, pathx); err == nil {
					goLog.Trace.Printf("%s path %s in %v", action, pathx, elapse)
					hok = true
				}
			case "Test":
				hpool := sproxyd.HP.Get()
				curl := hpool.Host() + pathx
				goLog.Info.Printf("Hostname: %s pid: %v Action: %s Url: %s", hostname, pid, action, curl)
				hpool.Mark(nil)
				hok = true
			default:
				goLog.Info.Printf("Wrong action:%s\n", action)
				os.Exit(2)
			}

			if !hok { // skip if document meta data is not ok
				continue
			}

			t_pages := docmeta.TotalPages // GET THE NUMBER OF PAGES FROM THE DOCUMENT META DARA
			// READ THE IMAGES DIRECTORY CORRESPONDING TO THE CONTAINER
			doc_path := path.Join(tiff_path, docid)
			doc_entry, _ := ioutil.ReadDir(doc_path)
			t1_pages := len(doc_entry)
			if t1_pages != t_pages {
				goLog.Warning.Printf("Hostname:%s Pid:%v #Tiff images %v not same %v", hostname, pid, t1_pages, t1_pages)
			}
			patha := make([]string, t1_pages)
			doca := make([][]byte, t1_pages)
			heada := make([]map[string]string, t1_pages)
			filea := make([]string, t1_pages)
			bread := time.Now()

			// READ  THE DOCUMENT IMAGES DIRECTORY
			for n, doc_files := range doc_entry {
				fulldoc = docmeta.DocumentID.CC + "/" + docmeta.DocumentID.PN + "/" + docmeta.DocumentID.KC
				page := doc_files.Name()
				page0 := strings.Split(page, ".")[0] // remove .tiff type
				patha[n] = base_path + "/" + fulldoc + "/" + page0
				filea[n] = path.Join(doc_path, page)
			}
			if action != "Delete" {
				responses := files.AsyncReadFiles(filea)

				goLog.Info.Printf("Time to read document %s %v %v", fulldoc, time.Since(bread), t1_pages)

				for k, v := range responses {

					var page = bns.PAGE{} // SHOULD HAVE BEEN CREATED BY ST33toFiles

					if v.Err == nil {

						if err := json.Unmarshal(v.Body, &page); err != nil {
							goLog.Warning.Println(err)
							continue

						}
						//fmt.Println(page.Tiff.Size, page.Metadata)

						var usermd []byte
						var header map[string]string

						if usermd, err = json.Marshal(page.Metadata); err == nil {

							header = map[string]string{
								"Usermd":       base64.Encode64(usermd),
								"Content-Type": ct,
							}
							heada[k] = header

							//var Blob []byte
							var img = new(bytes.Buffer)

							if page.Metadata.MultiMedia.TIFF == true {
								if page.Tiff.Size > 0 {
									img.Write(page.Tiff.Image)
								} else {
									goLog.Warning.Println("Wrong TIFF images in ", fulldoc, "pages", page.Metadata.PageNumber)
								}

							}
							if page.Metadata.MultiMedia.PNG == true {
								if page.Png.Size > 0 {
									img.Write(page.Png.Image)
								} else {
									goLog.Warning.Println("Wrong TIFF images in ", fulldoc, "pages", page.Metadata.PageNumber)
								}

							}
							doca[k] = img.Bytes()
							goLog.Trace.Println("Image length:", len(doca[k]), len(page.Tiff.Image), len(page.Png.Image))
							img.Reset()

						} else {
							goLog.Warning.Println(hostname, pid, err, "Marshalling", string(usermd))
						}
					}
				}
			}

			start := time.Now()

			var (
				duration, elapse time.Duration
				results          []*sproxyd.HttpResponse
			)
			switch action {
			case "Create":
				results = bns.AsyncHttpPuts(patha, doca, heada)
			case "Update":
				results = bns.AsyncHttpUpdates(patha, doca, heada)
			case "Delete":
				results = bns.AsyncHttpDeletes(patha)
			case "Test":
				for k, v := range doca {

					hpool := sproxyd.HP.Get()
					curl := hpool.Host() + patha[k]
					goLog.Info.Printf("Hostname: %s pid: %v Action: %s Url: %s Image size:", hostname, pid, action, curl, len(v))
					hpool.Mark(nil)
					hok = true
				}
			default:
			}
			duration = time.Since(start)

			if action != "Test" {
				for _, v := range results {
					if v.Err != nil {
						goLog.Error.Println("URL:", v.Url, "Error", v.Err)
					} else {
						elapse = duration / time.Duration(t1_pages)
						if v.Response.StatusCode == 200 || v.Response.StatusCode == 204 {
							goLog.Info.Printf("URL:%s Action %s Status:%v Elapse:%v %v\n", v.Url, action, v.Response.StatusCode, duration, elapse)
						} else {
							goLog.Warning.Printf("URL:%s Action %s Status:%v %v Elapse:%v %v\n", v.Url, action, v.Response.StatusCode, v.Response.Status, duration, elapse)
						}
					}
				}

			}

			goLog.Info.Printf("Action: %s Fulldoc: %s in %v %v", action, fulldoc, duration, time.Since(begin))
		} // loop on all pages of a document
		goLog.Info.Printf("Total Elapse: %v\n", time.Since(Begin))
	} else {
		goLog.Warning.Println("Container directory does not exist")
	}

}

// Check if PNG image is correct
func openImage(filename string) (image.Image, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return tiff.Decode(f)
}
