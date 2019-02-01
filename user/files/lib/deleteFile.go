package files

import (
	"fmt"
	goLog "github.com/moses/user/goLog"
	"os"
	"time"
)

func DeleteFile(filename string) error {
	var err error
	if err = os.Remove(filename); err != nil {
		hostname, _ := os.Hostname()
		goLog.Warning.Println(hostname, os.Getpid(), err, "Deleting", filename)
	}
	return err

}

func AsyncDeleteFiles(entries []string) []*Responses {
	ch := make(chan *Responses)
	responses := []*Responses{}
	treq := 0
	for _, entry := range entries {
		treq += 1
		go func(entry string) {
			err := os.Remove(entry)
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

func AsyncDeleteAllFiles(entries []string) []*Responses {
	ch := make(chan *Responses)
	responses := []*Responses{}
	treq := 0
	for _, entry := range entries {
		treq += 1
		go func(entry string) {
			err := os.RemoveAll(entry)
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
