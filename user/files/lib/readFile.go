package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"user/goLog"
)

func ReadFile(filename string) ([]byte, error) {
	var (
		buf []byte
		err error
	)
	if buf, err = ioutil.ReadFile(filename); err != nil {
		hostname, _ := os.Hostname()
		goLog.Warning.Println(hostname, os.Getpid(), err, "Reading", filename)
	}

	return buf, err
}

func AsyncReadFiles(entries []string) []*Responses {

	ch := make(chan *Responses)
	responses := []*Responses{}
	treq := 0

	for _, entry := range entries {
		treq += 1
		go func(entry string) {
			buf, err := ReadFile(entry)
			// fmt.Println(entry, len(buf))
			ch <- &Responses{err, buf}
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
