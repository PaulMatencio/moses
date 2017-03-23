package files

import (
	"bufio"
	"errors"
	"os"
)

func ScanLines(scanner *bufio.Scanner, num int) ([]string, error) {
	var (
		k   int = 0
		err error
	)
	linea := make([]string, num, num*2)
	stop := false

	for !stop {
		scanner.Scan()
		eof := false
		if text := scanner.Text(); len(text) > 0 {
			linea[k] = scanner.Text()
			k++
		} else {
			eof = true
		}
		err = scanner.Err()
		if k >= num || eof || err != nil {
			stop = true
		}
	}

	return linea[0:k], err
}

func Scanner(pathname string) (*bufio.Scanner, error) {
	var (
		err     error
		scanner *bufio.Scanner
	)
	if len(pathname) > 0 {

		fp, err := os.Open(pathname)
		if err == nil {
			defer fp.Close()
			scanner = bufio.NewScanner(fp)
		} else {
			os.Exit(10)
		}
	} else {
		err = errors.New(pathname + " is empty")
	}
	return scanner, err
}
