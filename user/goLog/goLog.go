// goLog project goLog.go
package goLog

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLog(logPath string, testname string, application string, action string, trace bool) (bool, *os.File, *os.File, *os.File, *os.File) {

	hostname, _ := os.Hostname()
	pid := os.Getgid()

	var (
		trf, waf, inf, erf *os.File
		defaut             = false
	)

	if logPath == "" {
		fmt.Println("WARNING: Using default logging")
		Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
		defaut = true
	} else {

		// mkAll dir
		logPath = logPath + string(os.PathSeparator) + testname
		/*
			if !files.Exist(logPath) {
				_ = os.MkdirAll(logPath, 0755)

			}
		*/
		_, err := os.Stat(logPath)
		if err != nil {
			if err := os.MkdirAll(logPath, 0755); err != nil {
				fmt.Println("Can't create ", logPath, err)
				os.Exit(3)
			}
		}

		traceLog := logPath + application + "_trace.log"
		infoLog := logPath + application + "_info.log"
		warnLog := logPath + application + "_warning.log"
		errLog := logPath + application + "_error.log"

		trf, err1 := os.OpenFile(traceLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
		inf, err2 := os.OpenFile(infoLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
		waf, err3 := os.OpenFile(warnLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
		erf, err4 := os.OpenFile(errLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
			Warning.Println(err1, err2, err3, err3)
			Warning.Println(hostname, pid, "Using default logging")
		} else {
			if trace == false {
				Init(ioutil.Discard, io.Writer(inf), io.Writer(waf), io.Writer(erf))

			} else {
				Init(io.Writer(trf), io.Writer(inf), io.Writer(waf), io.Writer(erf))
				Trace.Println(hostname, pid, "Start", application, action)
			}
		}
	}
	return defaut, trf, inf, waf, erf
}

func Init(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Time := log.Lmicroseconds

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|Time|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|Time|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|Time|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|Time|log.Lshortfile)

}
