package directory

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	sindexd "sindexd/lib"
	"strings"
	"time"
	goLog "user/goLog"
)

func DeleteKeys(client *http.Client, index *sindexd.Index_spec, dKey *[]string) (resp *http.Response, err error) {

	l := &sindexd.Load{
		Index_spec: sindexd.Index_spec{
			Index_id: index.Index_id,
			Cos:      index.Cos,
			Vol_id:   index.Vol_id,
			Specific: index.Specific,
		},
	}
	g := &sindexd.Delete_Keys{
		Key:      *dKey,
		Prefetch: false,
	}
	return g.DeleteKeys(client, l)
}

func DeleteArray(iIndex string, inputDir string, client *http.Client, index *sindexd.Index_spec, bulk int) {
	// this function is using the input files extracted by Ernst

	var (
		err     error
		file    *os.File
		line    string
		Key     string
		Pubdate []string
		// pubdate   string
		f        string
		keya     []string
		reclen   int64
		dep      int64
		i        int64
		j        int
		body     []byte
		response *sindexd.Response
		start    time.Time
	)
	if len(inputDir) == 0 {
		goLog.Error.Println("The input directory is missing")
		os.Exit(3)
	}
	if file, err = os.Open(inputDir); err != nil {
		goLog.Error.Println(err)
		os.Exit(3)
	}
	// url := hp.Get().Host()
	keya = make([]string, bulk, bulk*2)
	f = "Delete " + iIndex
	fileinfo, err := file.Stat()
	fsize := fileinfo.Size()
	dep = 0
	stop := false
	i, j = 1, 0
	in := bufio.NewReader(file)
	for err == nil && !stop {
		line, err = in.ReadString('\n') // read a line
		reclen = int64(len(line))
		dep = dep + reclen
		var pubdate string
		if reclen > 0 {
			cc := strings.TrimSpace(line[11:13])
			pn := strings.TrimSpace(line[18:28])
			kc := strings.TrimSpace(line[28:30])
			Pubdate = strings.Split(line[32:42], "-")
			if len(Pubdate) == 3 {
				pubdate = Pubdate[0] + Pubdate[1] + Pubdate[2]
			} else {
				pubdate = PubDate
				goLog.Warning.Println("PN:", Key, "has no Publication date, Default to", pubdate)
			}
			switch iIndex {
			case "PN":

				keya[j] = cc + sindexd.Delimiter + pn + sindexd.Delimiter + kc

			case "PD":

				keya[j] = cc + sindexd.Delimiter + pubdate[0:4] + sindexd.Delimiter + pubdate[4:6] + sindexd.Delimiter +
					cc + sindexd.Delimiter + pn + sindexd.Delimiter + kc

			}
			j++
		}

		//goLog.Info.Println(i, bulk, i%int64(bulk))

		if j >= bulk || dep >= fsize {
			start = time.Now()
			if resp, err := DeleteKeys(client, index, &keya); err != nil {
				goLog.Error.Println("Function:", f, "last line:", i, "last key:", keya[j-1], err)
			} else {
				switch resp.StatusCode {
				case 200:
					response = new(sindexd.Response)
					body = sindexd.GetBody(resp)
					if err := json.Unmarshal(body, &response); err != nil {
						goLog.Error.Println(err)
					} else {
						goLog.Info.Println("Function:", f, i, "sindexd response:", response.Status, response.Reason, "Duration:", time.Since(start))
					}
				default:
					goLog.Warning.Println("Function:", f, i, resp.Status, time.Since(start))
				}

				resp.Body.Close()
			}

			j = 0
		}

		i++
		if sindexd.Maxinput != 0 && i > sindexd.Maxinput {
			break
		}
	}
}
