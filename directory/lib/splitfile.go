package directory

import (
	"bufio"
	"os"
	// "os/user"
	// "path"
	// "poc/files"
	// goLog "github.com/paulmatencio/moses/user/goLog"
	goLog "github.com/paulmatencio/s3/gLog"
	"strings"
)

var (
	err    error
	file   *os.File
	line   string
	reclen int
	f      *os.File
)

func SplitFile(inputFile string, outputDir string) {

	if len(inputFile) == 0 || len(outputDir) == 0 {
		goLog.Error.Println("The input or output directory are missing")
		os.Exit(3)
	}
	if file, err = os.Open(inputFile); err != nil {
		goLog.Error.Println(err)
		os.Exit(3)
	}
	defer file.Close()
	//
	// read the input file and create multiple output files based of the country code
	i := 0
	in := bufio.NewReader(file)
	suf0 := ""
	for err == nil {
		line, err = in.ReadString('\n') // read a line
		reclen := len(line)
		i++
		if reclen > 0 {
			cc := strings.TrimSpace(line[11:13])

			suf := "OTHER"
			switch cc {
			case "CN", "CA", "DE", "EP", "FR", "GB", "JP", "KR", "US", "WO", "IT", "RU", "TW", "SU", "NL", "NO", "PL", "AT", "MX", "IL", "ZA", "NZ", "FI", "ES", "DK", "DD", "CH", "BE", "AU", "BR":
				suf = cc
			default:
			}

			fn := outputDir + string(os.PathSeparator) + suf

			if suf != suf0 {
				suf0 = suf
				_ = f.Close() // Close the previous File
				goLog.Info.Println("Opening File", fn)
				if f, err = os.OpenFile(fn, os.O_CREATE+os.O_WRONLY+os.O_APPEND, 0660); err != nil {
					goLog.Error.Println("Open File", fn, err)
					os.Exit(3)
				}
			}
			if _, err := f.WriteString(line); err != nil {
				goLog.Error.Println("Error Writing line", line)
				os.Exit(3)
			}

		} else {
			goLog.Warning.Println("Line:", i, " is empty")
		}
	}
}
