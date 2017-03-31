package main

import (
	"fmt"
	// sindexd "moses/sindexd/lib"
	//user "moses/user/base64j"
	files "moses/user/files/lib"
	//goLog "moses/user/goLog"
	"encoding/json"
	"os/user"
	"path"
	"strconv"
	"strings"
)

type IndexTab struct {
	Country  string `json:"country"`
	Index_id string `json:"index_id"`
	Cos      int    `json:"cos"`
	Volid    int    `json:"volid"`
	Specific int    `json:"specific"`
}

func main() {
	usr, _ := user.Current()
	sindexdTables := path.Join(usr.HomeDir, "sindexd/config/sindexd-prod-pn")
	var indextab = IndexTab{}
	if scanner, err := files.Scanner(sindexdTables); err != nil {
		fmt.Println(scanner, err)
	} else if linea, err := files.ScanLines(scanner, 40); err == nil {

		for _, v := range linea {
			index := strings.Split(v, " ")
			indextab.Country = index[0]
			indextab.Index_id = index[1]
			indextab.Cos, _ = strconv.Atoi(index[2])
			indextab.Volid, _ = strconv.Atoi(index[3])
			indextab.Specific, _ = strconv.Atoi(index[4])

			if indexJson, err := json.Marshal(&indextab); err == nil {
				fmt.Println(string(indexJson))
			} else {
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println(err)
	}

}
