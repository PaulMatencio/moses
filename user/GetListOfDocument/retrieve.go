package main

import (
  "fmt"
  "net/http"
  "time"
  "io/ioutil" 
  "flag"
  "os"
  "strings"
  "os/user"
  "user/files"
  "path"	
  "user/sproxyd"	
  "container/ring"
)

/*
var docs = []string {
    
 
   
}
*/
//var urls = []string

func usage() {
	usage := "GetListOfDocument -c config -a Test/get  -w <pdf/xpdf> -l <d1,d2....dn> -s yes/no" 		 
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

type HttpResponse struct {
  url      string
  response *http.Response
  size     int
  err      error
}

func asyncHttpGets(urls []string, save string) []*HttpResponse {
  // create http response channel
  ch := make(chan *HttpResponse)
  responses := []*HttpResponse{}
  // for every url, start a goroutine ()

  usr,_ := user.Current()
  homeDir := usr.HomeDir  	
  downloadDir:="Downloads"
  var (   	 
	mypath string
  )

  for _, url := range urls {

    	  if save == "yes" {
			                 
		mypath = path.Join(homeDir,downloadDir)	
		if !files.Exist(mypath) {	
		_ = os.MkdirAll(mypath, 0755)
      	 	}
    	  }

      go func(url string) {
         fmt.Printf("Fetching %s \n", url)	  
         resp, err := http.Get(url)
         body,_ := ioutil.ReadAll(resp.Body)	  
	 if save == "yes" {
		filename := strings.Split(url, "/")
	 	myfile := filename[len(filename)-1]	
           	file := path.Join(mypath,myfile)	 
	 	if  ioutil.WriteFile(file, body, 0644) != nil {
		fmt.Println("can't save",err)
		}
          }
          ch <- &HttpResponse{url, resp,len(body), err}
      }(url)
  }

  // wait for http response  message 
  for {
      select {
      case r := <-ch:
          fmt.Printf("%s was fetched\n", r.url)
          responses = append(responses, r)
          if len(responses) == len(urls) {
              return responses
          }
      case <-time.After(50 * time.Millisecond):
          fmt.Printf(".")
      }
  }
  return responses
}


func main() {

   // Read the list of document to retrieve	
   
   // create a ring and populate it 

	 var (
		source string	
		action string	
		config string
		listdoc string
		save string
	)

	flag.Usage = usage
	flag.StringVar(&source, "w", "xpdf", "")
	flag.StringVar(&action, "a", "Test", "")	
	flag.StringVar(&save, "s", "no", "")	
	flag.StringVar(&config, "c", "", "")
	flag.StringVar(&listdoc, "l", "", "")
	flag.Parse()

	if  listdoc == "?" {
		usage()
	}
	

  	if len(config) != 0 {
		if err := sproxyd.SetProxydHost(config); err == nil {
			fmt.Println(sproxyd.Host)
		} else { fmt.Println(err,"Use default:", sproxyd.Host)}
		
  	  }

  
   	 // parse the list of documents 	
   	
   	 //  A ring of sproxyd nodes is used to balance (Rounf robin) requests among all sproxyd nodes		 
   	r := ring.New(len(sproxyd.Host)) // r  is a pointer to  the ring
    	for i := 0; i < r.Len(); i++ {
		r.Value = sproxyd.Host[i] + source + "/"
		//fmt.Println(r.Value)
		r = r.Next()
     	}   

   	// fill urls
	docs := strings.Split(listdoc,",")
        if len(docs) == 0  {
	 usage()
	}
  	urls := make ([]string ,len(docs))
   	for d:= range docs {
   
    		host :=  r.Value
     		urls[d] = host.(string)   + docs[d] + ".pdf"
    		r = r.Next()
  	 } 


  switch action {

   case "Test": fmt.Println(urls)
   case  "get":	

	 time1 := time.Now()	
 	 size :=0 

  	results := asyncHttpGets(urls,save)

 	for _, result := range results {
   	   size += result.size
    	  result.response.Body.Close()
    	  fmt.Printf("%s status: %s%s%d\n", result.url,
                 result.response.Status,"Size(KB)=",result.size/1024)
 	 }
  	duration:= time.Since(time1)
 	fmt.Println("OK","MB/sec=",1000*float64(size)/float64(duration),"Duration=",duration)
   case "?": usage()
   default: 
	fmt.Println("Wrong request! should be <Test/get>" )
  }

}

