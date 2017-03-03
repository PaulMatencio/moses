// CreateTiff project main.go
package main

import (
	"container/ring"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"user/files"
	"user/poc"
	"user/sproxyd"
)

func usage() {
	usage := "\nFunction==> Extract Tiff images extracted from  St33 files to store into files \n\nusage: St33toScality -a -w -i -o\n-a <CreateFile/Test/GetUrl> \n-w <bns/kime> \n-i input_directory \n-o output_directory\n\nDefault Options\n\n"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

//var host = []string{"http://luo001t.internal.epo.org:81/proxy/chord/", "http://luo002t.internal.epo.org:81/proxy/chord/", "http://luo003t.internal.epo.org:81/proxy/chord/"}var host = []string{"http://luo001t.internal.epo.org:81/proxy/chord/", "http://luo002t.internal.epo.org:81/proxy/chord/", "http://luo003t.internal.epo.org:81/proxy/chord/"}
var (
	action    string
	what      string
	InputDir  string
	OutputDir string
	createbns bool
)

func main() {

	flag.Usage = usage
	flag.StringVar(&action, "a", "CreateFile", "")
	flag.StringVar(&what, "w", "bns", "")
	flag.StringVar(&InputDir, "i", "", "")
	flag.StringVar(&OutputDir, "o", "", "")
	flag.Parse()

	if len(InputDir) == 0 || len(OutputDir) == 0 {
		usage()
	}

	if !files.Exist(InputDir) {
		fmt.Println(InputDir, " does not exist")
		os.Exit(3)
	}

	switch what {
	case "bns":
		createbns = true
	case "kime":
		createbns = false
	default:
		fmt.Println("Wrong -w ; should be bns or kime")
		usage()

	}
	/*r := createRing(host)*/
	sep := string(os.PathSeparator)
	fmt.Println(sep)
	r := ring.New(len(sproxyd.Host)) // r  is a pointer to  the ring
	for i := 0; i < r.Len(); i++ {
		r.Value = sproxyd.Host[i]
		fmt.Println(r.Value)
		r = r.Next()
	}
	outputUsermdDir := OutputDir + sep + "Usermd" + sep
	if files.Exist(outputUsermdDir) == false {
		os.MkdirAll(outputUsermdDir, 0755)
	}
	outputTiffDir := OutputDir + sep + "Tiff" + sep
	if files.Exist(outputTiffDir) == false {
		os.MkdirAll(outputTiffDir, 0755)
	}
	outputContainerDir := OutputDir + sep + "Container" + sep
	if files.Exist(outputContainerDir) == false {
		os.MkdirAll(outputContainerDir, 0755)
	}

	dfiles, _ := ioutil.ReadDir(InputDir)

	for _, d := range dfiles {
		base_url := r.Value.(string)
		base_url = base_url + what + "/"

		r = r.Next()
		dInputDir := InputDir + sep + d.Name()
		mfiles, err := ioutil.ReadDir(dInputDir)
		if err == nil {
			for _, f := range mfiles {

				inputFile := dInputDir + sep + f.Name()
				fmt.Println(inputFile)

				err := poc.ST33toTiff(action, base_url, inputFile, outputUsermdDir, outputTiffDir, outputContainerDir, createbns)
				if err != nil {
					fmt.Println(err)
				}
			}
		} else {
			fmt.Println("Processing ", dInputDir, err)
		}
	}
}
