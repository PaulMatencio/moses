// RetrievTiff project main.go
package main

import (
	"container/ring"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	//"runtime"
	"strconv"
	"strings"
	"time"
	"user/bns"
	"user/sproxyd"
)

var Section = map[string]string{
	"bib": "B",
	"abs": "A",
	"cla": "C",
	"dra": "D",
	"amd": "U",
	"des": "V",
	"srp": "S",
}

func usage() {
	usage := "\nFunction ==> Get tiff/png images\n\nUsage: GetDocument -c -w -d -p -s\n-w <bns/png/kime> \n-d <document id>\n-p <All/Abs/Amd/Bib/Cla/Des/Dra/Srp/i-j,p-q>\n-s <yes/no> " + "\n\n-c <both/connector/storage>" +
		"\n\nOptions for  -p\n All => Full document\n Abs => Abstrac\n Amd => Amendement\n Bib => Bibliography\n Cla => Claims\n Des => Description\n Dra => Drawing\n Srp => Search Report\n i-j,p-q are pages i to j and p to q" +
		"\n\nOptions for  -s\n yes=>save the results\n\nDefault Options:"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Response struct {
	Err  error
	Resp *http.Response
}

//var urls = []string

type HttpResponse struct {
	url      string
	response *http.Response
	size     int
	err      error
}

func asyncHttpGets(urls []string, ftype string, save string) []*HttpResponse {
	// create http response channel
	ch := make(chan *HttpResponse)
	responses := []*HttpResponse{}
	// for every valid url, start a goroutine ()
	treq := 0 /* count the number of requests */
	usr,_ := user.Current()
	homeDir := usr.HomeDir
	downloadDir:="Downloads"
	
	for _, url := range urls {
		var (
			mypath string
			myfile string
		)
		/* just in case, the requested page number is beyond the max number of pages */
		if len(url) == 0 {
			break
		} else {
			treq += 1
		}

		if save == "yes" {
			filename := strings.Split(url, "/")
			myfile = filename[len(filename)-1]
			dir := strings.Split(myfile, "_")
			mydir := dir[0]                     
			mypath = path.Join(homeDir,downloadDir, mydir)
			//fmt.Println(mypath)
			//mypath = os.TempDir() + string(os.PathSeparator) + mydir
			//fmt.Println(mypath)
			_ = os.RemoveAll(mypath)
			_ = os.MkdirAll(mypath, 0755)

		}
		go func(url string) {
			// fmt.Printf("Fetching %s \n", url)
			client := &http.Client{}
			resp, err := bns.GetPage(client, url)
			var body []byte
			if err == nil {
				body, _ = ioutil.ReadAll(resp.Body)
				// Write to tmp files if save = yes
				if save == "yes" {
					file := strings.Split(myfile, "_")
					if len(file) > 1 {
						//myfile := path.Join(mypath, file[1])
						myfile := mypath + string(os.PathSeparator) + file[1]
						/*
						if runtime.GOOS == "windows" {
							myfile += ".tif"
						}
						*/
						myfile += "."+ftype
						err := ioutil.WriteFile(myfile, body, 0644)
						check(err)
					} else {
						fmt.Println("Can't save", mypath, file)
					}

				}
			}
                        
			ch <- &HttpResponse{url, resp, len(body), err}

		}(url)
	}

	// wait for http response  message
	for {
		select {
		case r := <-ch:
			// fmt.Printf("%s was fetched\n", r.url)
			responses = append(responses, r)
			if len(responses) == treq /*len(urls)*/ {
				return responses
			}
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func buildPageTable(plist string) []int {
	page_tab := make([]int, 0, bns.Max_page)
	npage := strings.Split(plist, ",")
	for d := range npage {
		//fmt.Println(npage[d])
		dpage := strings.Split(npage[d], "-")
		for k := range dpage {
			if len(dpage) > 1 {
				debut, _ := strconv.Atoi(dpage[0])
				fin, _ := strconv.Atoi(dpage[1])
				if k == len(dpage)-1 {
					for i := debut; i <= fin; i++ {
						page_tab = append(page_tab, i)

					}
				}
			}
		}
	}
	return page_tab
}

func main() {
	//get document id

	var (
		source string
		doc    string
		plist  string
		save   string
		trace  string
		config string
		ftype  string
	)

	flag.Usage = usage
	flag.StringVar(&source, "w", "bns", "")
	flag.StringVar(&doc, "d", "", "")
	flag.StringVar(&plist, "p", "1-10", "")
	flag.StringVar(&save, "s", "no", "")
	flag.StringVar(&trace, "t", "no", "")
	flag.StringVar(&config, "c", "", "")
	flag.Parse()

	if len(doc) == 0 || plist == "?" {
		usage()
	}
	switch source {
	
		case "bns":  ftype="tiff"
		case "kime": ftype="tiff"
		case "png":  ftype="png"
		default: 
			{  fmt.Println("Invalid -w options") 
			 usage() }
	}	

	// check if there is a config file to override the default list of sproxyd hosts sproxyd.Host[]
	if len(config) != 0 {
		if err := sproxyd.SetProxydHost(config); err == nil {
			fmt.Println(sproxyd.Host)
		} else { fmt.Println(err,"Use default:", sproxyd.Host)}
		
	}

	// create a ring ( circular list) and populate it with the sproxyd.Host[]
	//  A ring of sproxyd nodes is used to balance (Rounf robin) requests among all sproxyd nodes
		 
	r := ring.New(len(sproxyd.Host)) // r  is a pointer to  the ring
	for i := 0; i < r.Len(); i++ {
		r.Value = sproxyd.Host[i] + source + "/"
		//fmt.Println(r.Value)
		r = r.Next()
	}
	
	// Retrieve document
	var page_tab []int
	total_pages := 0
	t_pages := 0
	//date:= bns.Date{}

	if source == "bns" || source== "png" {
		// Always get the metadata of a BNS document
		// BNS document must have metadata
		var usermd map[string]interface{}
		var err error
		murl := r.Value.(string) + doc
		if usermd, err = bns.GetDocMetadata(murl);  len(usermd) == 0{
			if err != nil {fmt.Println(err)}
                        goto nometadata
		}
		//if date,err = bns.GetPubDate(usermd); err != nil { fmt.Println(err) }

		if trace == "yes" {
			fmt.Println(usermd)
		}
		if err == nil {
			t_pages, _ = bns.GetPageNumber(usermd) // Get total number of pages
			if plist == "All" {
				end_page := strconv.Itoa(t_pages)
				plist = "1-" + end_page
				page_tab = buildPageTable(plist)
			} else if strings.Contains(plist, "-") {
				page_tab = buildPageTable(plist)
			} else {
				content, _ := bns.GetContent(usermd) // get the Content layout of a document
				vlist := Section[strings.ToLower(plist)]
				page_tab = bns.BuildSubtable(content, vlist)
			}
		}
	nometadata:  
	}
	if source == "kime" {
		
		// NOT YET IMPLEMENTED 
		fmt.Println("Not yet implemented")
 		os.Exit(5)
		if strings.Contains(plist, "-") {
			page_tab = buildPageTable(plist)
		} else {
			fmt.Println("You should provide pages range")
			os.Exit(1)
		}
	}
       
	total_pages = len(page_tab)
	// create all the requests and store them in the table urls
	urls := make([]string, total_pages)
	for k, d := range page_tab {
		if d > t_pages {
			break
		}
		host := r.Value
		var page string
		switch {
		case d < 10:
			page = "000" + strconv.Itoa(d)
		case d >= 10 && d < 100:
			page = "00" + strconv.Itoa(d)
		case d >= 100 && d < 1000:
			page = "0" + strconv.Itoa(d)
		default:
		}
		urls[k] = host.(string) + doc + "_" + page
		r = r.Next()
	}
        // fmt.Println(urls)
	if len(urls) > 0 {
		time1 := time.Now()
		size := 0
		// Concurrency
		results := asyncHttpGets(urls, ftype,save)
		for _, result := range results {
			size += result.size
			fmt.Printf("%s status: %s\n", result.url,
				result.response.Status)
		}
		duration := time.Since(time1) // in nano sec
		fmt.Println("OK", "MB/sec=", 1000*float64(size)/float64(duration), "Duration=", duration)
	}
         
}
