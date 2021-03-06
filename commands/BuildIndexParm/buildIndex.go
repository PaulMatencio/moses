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
	directory "github.com/moses/directory/lib"
	sindexd "github.com/moses/sindexd/lib"
	files "github.com/moses/user/files/lib"
	hexa "github.com/moses/user/hexa"
	"os"
	// "os/user"
	"path"
	"strconv"
	"strings"
)

func usage() {
	usage := "BuildIndexParm -input <File containers the index id (Object keys) > -v <verbose>"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	var (
		// usr, _                  = user.Current()
		indextab                             = sindexd.IndexTab{}
		index_id, input, output, indexTables string
		cwd, _                               = os.Getwd()
		V                                    bool
	)
	flag.Usage = usage
	flag.StringVar(&input, "input", "", "Input file containing the index ids")
	flag.BoolVar(&V, "v", false, "Verbose")
	flag.Parse()

	if len(input) == 0 {
		fmt.Println("The input file is missing")
		usage()
	}

	if input[0:1] != string(os.PathSeparator) {
		indexTables = path.Join(cwd, input)
	} else {
		indexTables = input
	}

	output = path.Base(indexTables) + ".json"
	Output := path.Join(cwd, output)
	fmt.Printf("The output file is > %s\n", Output)
	f, err := os.Create(Output)
	files.Check(err)
	if scanner, err := files.Scanner(indexTables); err != nil {
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
