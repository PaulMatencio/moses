package main

import (
	"encoding/json"
	"errors"
	"fmt"
	files "moses/user/files/lib"
	"strings"
)

type Stats struct {
	Node  string
	Start int
	End   int
	Ok    int
	Nok   int
	Skip  int
	Qty   int
}

func main() {

	var (
		file               = "/home/paul/go/src/moses/scloned/data"
		Qty, Skip, Ok, Nok = 0, 0, 0, 0
		stats              = Stats{}
	)

	if scanner, err := files.Scanner(file); err == nil {
		if lines, err := files.ScanLines(scanner, 50); err == nil {
			for k, v := range lines {
				if err = parse(v, &stats); err == nil {
					Qty += stats.Qty
					Skip += stats.Skip
					Ok += stats.Ok
					Nok += stats.Nok
				} else {
					fmt.Printf("Parsing error line %d: %v\n", k, err)
				}
			}
			fmt.Printf("Qty:%d , Skip: %d , OK: %d , Nok:%d Completion=%0.2f%s\n", Qty, Skip, Ok, Nok, float32(Skip+Ok)/float32(Qty), "%")
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}
}

func parse(line string, stats *Stats) error {
	line = strings.TrimSpace(line)
	if len(line) > 0 {
		Words := strings.Split(line, " ")
		if len(Words) == 14 {
			line1 := fmt.Sprintf("{%q:%q,%q:%s,%q:%s,%q:%s,%q:%s,%q:%s,%q:%s}", "Node", Words[1], Words[2], Words[3], Words[4], Words[5], Words[6], Words[7], Words[8], Words[9], Words[10], Words[11], Words[12], Words[13])
			return json.Unmarshal([]byte(line1), &stats)
		} else {
			return errors.New("Invalid Input line")
		}
	} else {
		return errors.New("Empty line")
	}

}
