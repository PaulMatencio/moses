/* */
package main

import (
	"fmt"
	"user/originalpdf"
	"flag"
	"os"
	
)

func usage() {
	usage := "usage: CreateOriginalPdf  -a [createFile/CreateObject] -i inputdir -o"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

 
func main() {

	const (
		 
		
	)
	var (

	inputDir string 
	outputDir string
	action string
	)
	flag.Usage = usage
	flag.StringVar(&action, "a", "", "")
	flag.StringVar(&inputDir, "i", "", "")
	flag.StringVar(&outputDir, "o", "", "")
	flag.Parse()
	
	if len(action) == 0 {
		usage()
	}

	if err:= originalpdf.CreateFromFolder(inputDir,outputDir); err != nil {
		fmt.Println("DDD",err)
	}
	
	
}	
