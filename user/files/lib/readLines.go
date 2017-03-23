package files

import (
	"bufio"
	"fmt"
	"os"
)

func ReadLines(filename string) error {
	fp, err := os.Open(filename)
	if err == nil {
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {

			fmt.Println(scanner.Text())
		}
	}
	return err
}
