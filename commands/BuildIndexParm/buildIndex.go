package main

/* Build index parms for sindexd create index table */
/* Input : File containing Scality Ring index_id ( object key )  */
/* Output: a Json file containing the index specification
   index_id Cos Volid Specific
*/
import (
	"encoding/json"
	"flag"
	"fmt"
	directory "moses/directory/lib"
	sindexd "moses/sindexd/lib"
	files "moses/user/files/lib"
	hexa "moses/user/hexa"
	"os"
	// "os/user"
	"path"
	"strconv"
	"strings"
)

func usage() {
	usage := "Buildindex -input <File containers the index id (Object keys) > -v <verbose>"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	var (
		// usr, _                  = user.Current()
		indextab                = sindexd.IndexTab{}
		index_id, input, output string
		cwd, _                  = os.Getwd()
		V                       bool
	)
	flag.Usage = usage
	flag.StringVar(&input, "input", "", "Input file containing the index ids")
	flag.BoolVar(&V, "v", false, "Verbose")
	flag.Parse()
	if len(input) == 0 {
		fmt.Println("input file is  missing")
		usage()
	}
	sindexdTables := path.Join(cwd, input)
	output = input + ".json"
	Output := path.Join(cwd, output)
	fmt.Printf("The output file  is > %s\n", Output)
	f, err := os.Create(Output)
	files.Check(err)
	if scanner, err := files.Scanner(sindexdTables); err != nil {
		fmt.Println(scanner, err)
	} else if linea, err := files.ScanLines(scanner, 40); err == nil {

		for _, v := range linea {
			index := strings.Split(v, " ")
			indextab.Country = index[0]
			index_id = index[1]
			indextab.Index_id = index_id
			long := len(index_id)
			Volid, _ := hexa.HexaByteToInt32(index_id[long-18 : long-10])
			Specific, _ := hexa.HexaByteToInt8(index_id[long-4 : long-2])
			indextab.Cos, _ = strconv.Atoi(index_id[long-2 : long-1])
			indextab.Volid = Volid
			indextab.Specific = Specific
			if indexJson, err := json.Marshal(&indextab); err == nil {
				f.WriteString(string(indexJson) + "\n")
			} else {
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println(err)
	}
	if V {
		for k, v := range directory.BuildIndexspec(Output) {
			fmt.Println(k, *v)
		}
	}

}
