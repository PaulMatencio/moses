package files

import (
	"fmt"
	"io/ioutil"
	goLog "github.com/moses/user/goLog"
	"os"
	"time"
)

func WriteFile(filename string, buf []byte, mode os.FileMode) error {
	var err error
	if err = ioutil.WriteFile(filename, buf, mode); err != nil {
		hostname, _ := os.Hostname()
		goLog.Warning.Println(hostname, os.Getpid(), err, "Writing", filename)
	}
	return err
}

func AsyncWriteFiles(entries []string, buf [][]byte, mode os.FileMode) []*Responses {
	ch := make(chan *Responses)
	responses := []*Responses{}
	treq := 0
	for k, entry := range entries {
		treq += 1
		go func(entry string) {
			err := WriteFile(entry, buf[k], mode)
			ch <- &Responses{err, nil}
		}(entry)
	}
	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == treq {
				return responses
			}
		case <-time.After(Timeout * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}
