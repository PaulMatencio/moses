package directory

import (
	"bufio"
	"bytes"
	"encoding/json"
	sindexd "github.com/paulmatencio/moses/sindexd/lib"
	//	"fmt"
	"io"
	// goLog "github.com/paulmatencio/moses/user/goLog"
	goLog "github.com/paulmatencio/s3/gLog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

type KeyPN struct {
	Docid   string `json:"Docid"`
	Pubdate string `json:"Pubdate"`
}
type KeyPD struct {
	Pubdate string `json:"Pubdate"`
}

func AddDelimiter(pn string, delimiter string) string {
	// Add a delimiter after every byte of the input string and return the result string
	var rec bytes.Buffer
	rec.WriteString(delimiter)
	for i := 0; i < len(pn); i++ {
		rec.WriteString(pn[i : i+1])
		rec.WriteString(delimiter)
	}
	return rec.String()
}

func Createkey(iIndex string, inputDir string) map[string]string {

	// input file/line => country,doc_number,kind,doc_id,
	// output map["key"] = json string
	var (
		file *os.File
		err  error
		f    string
		Key  string
	)
	if len(inputDir) == 0 {
		goLog.Error.Println("The input directory is missing")
		os.Exit(3)
	}
	if file, err = os.Open(inputDir); err != nil {
		goLog.Error.Println(err)
		os.Exit(3)
	}
	f = "CreateKey " + iIndex
	keyObject := make(map[string]string)

	i := 0
	first := true
	in := bufio.NewReader(file)
	var line string
	for err != io.EOF {
		line, err = in.ReadString('\n')
		aLine := strings.Split(line, ",")
		if len(aLine) >= 5 && !first { // skip the first line
			switch iIndex {
			case "PN":
				value := new(KeyPN)
				Key = sindexd.Delimiter + aLine[0] + sindexd.Delimiter + aLine[1] + sindexd.Delimiter + aLine[2]
				if len(aLine[3]) != 0 {
					value.Docid = aLine[3]
				} else {
					value.Docid = Key
				}
				if len(aLine) > 5 {
					value.Pubdate = aLine[4]
				} else {
					value.Pubdate = PubDate
				}
				if valuej, err := json.Marshal(value); err == nil {
					keyObject[Key] = string(valuej)
				} else {
					goLog.Warning.Println("Function:", f, "skipping", Key, "Bad input value at line", i)
				}
				if valuej, err := json.Marshal(value); err == nil {
					keyObject[Key] = string(valuej)
				} else {
					goLog.Warning.Println("Function:", f, "skipping", Key, "Bad input value at line", i)
				}

			case "PD":
				value := new(KeyPD)
				var (
					year  string
					month string
				)
				if len(aLine) > 5 && len(aLine[4]) != 0 {
					year = aLine[4][0:4]
					month = aLine[4][4:6]
					value.Pubdate = aLine[4]
				} else {
					year = PubDate[0:4]
					month = PubDate[4:6]
					value.Pubdate = PubDate
				}
				Key = sindexd.Delimiter + year + sindexd.Delimiter + month + sindexd.Delimiter +
					aLine[0] + sindexd.Delimiter + aLine[1] + sindexd.Delimiter + aLine[2]
				if valuej, err := json.Marshal(value); err == nil {
					keyObject[Key] = string(valuej)
				} else {
					goLog.Warning.Println("Function:", f, "skipping", Key, "Bad input value at line", i)
				}
			}
			i++
		}
		first = false
	}
	return keyObject
}

func AddKeys(client *http.Client, index *sindexd.Index_spec, keyObject map[string]string) (*http.Response, error) {
	// add indexes  (keyObject) to an index table (index)
	l := &sindexd.Load{
		Index_spec: sindexd.Index_spec{
			Index_id: index.Index_id,
			Cos:      index.Cos,
			Vol_id:   index.Vol_id,
			Specific: index.Specific,
		},
	}
	return sindexd.Addkeys(client, l, keyObject)
}

func AddKeys1(HP hostpool.HostPool, client *http.Client, index *sindexd.Index_spec, keyObject map[string]string) (*http.Response, error) {
	// add indexes  (keyObject) to an index table (index)
	l := &sindexd.Load{
		Index_spec: sindexd.Index_spec{
			Index_id: index.Index_id,
			Cos:      index.Cos,
			Vol_id:   index.Vol_id,
			Specific: index.Specific,
		},
	}
	return sindexd.Addkeys1(HP,client, l, keyObject)
}
func AddKeysa(client *http.Client, index *sindexd.Index_spec, pdatep *[]sindexd.PubRecord) (*http.Response, error) {
	//add indexes  (key and values array) to an index table (index)
	l := &sindexd.Load{
		Index_spec: sindexd.Index_spec{
			Index_id: index.Index_id,
			Cos:      index.Cos,
			Vol_id:   index.Vol_id,
			Specific: index.Specific,
		},
	}
	return sindexd.Addkeysa(client, l, pdatep)
}

func AddMap(iIndex string, inputDir string, client *http.Client, index *sindexd.Index_spec, bulk int) {
	// this function is using the input files extracted by Ernst
	var (
		err     error
		file    *os.File
		line    string
		Key     string
		Pubdate []string

		// pubdate   string
		f         string
		keyObject map[string]string
		reclen    int64
		dep       int64
		i         int64
		body      []byte
		response  *sindexd.Response
		start     time.Time
	)
	if len(inputDir) == 0 {
		goLog.Error.Println("The input directory is missing")
		os.Exit(3)
	}
	if file, err = os.Open(inputDir); err != nil {
		goLog.Error.Println(err)
		os.Exit(3)
	}

	keyObject = make(map[string]string)
	f = "Add " + iIndex
	fileinfo, err := file.Stat()
	fsize := fileinfo.Size()
	dep = 0
	stop := false
	i = 1
	in := bufio.NewReader(file)
	for err == nil && !stop {

		line, err = in.ReadString('\n') // read a line
		//fmt.Println(line)
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
				value := new(KeyPN)
				value.Pubdate = pubdate
				Key = sindexd.Delimiter + cc + sindexd.Delimiter + pn + sindexd.Delimiter + kc
				// Key = sindexd.Delimiter + cc + AddDelimiter(pn, sindexd.Delimiter) + kc
				value.Docid = strings.TrimSpace(line[0:10])
				if valuej, err := json.Marshal(value); err == nil {
					keyObject[Key] = string(valuej)
				} else {
					goLog.Warning.Println("Function:", f, "skipping", Key, "Bad input value at line", i)
				}
			case "PD":
				value := new(KeyPD)
				value.Pubdate = pubdate
				Key = cc + sindexd.Delimiter + pubdate[0:4] + sindexd.Delimiter + pubdate[4:6] + sindexd.Delimiter + pubdate[6:8] +
					sindexd.Delimiter + pn + sindexd.Delimiter + kc
				if valuej, err := json.Marshal(value); err == nil {
					keyObject[Key] = string(valuej)
				} else {
					goLog.Warning.Println("Function:", f, "skipping", Key, "Bad input value at line", i)
				}
			}
		}
		if i%int64(bulk) == 0 || dep >= fsize {
			start = time.Now()
			if resp, err := AddKeys(client, index, keyObject); err != nil {
				goLog.Error.Println("Function:", f, "last line:", i, "last key:", Key, err)
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
				// reuse the keyObject map
				resp.Body.Close()
			}
			keyObject = map[string]string{}
		}
		i++
		if sindexd.Maxinput != 0 && i > sindexd.Maxinput {
			break
		}
	}
}

func AddArray(iIndex string, inputDir string, client *http.Client, index *sindexd.Index_spec, bulk int) {
	// this function is using the input files extracted by Ernst

	var (
		err      error
		file     *os.File
		line     string
		Key      string
		Pubdate  []string
		Pubdatea []sindexd.PubRecord
		f        string
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
	Pubdatea = make([]sindexd.PubRecord, bulk, bulk*2)
	Pdate := sindexd.PubRecord{}

	f = "Add " + iIndex
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
				pubdate = PubDate // Default publication date
				goLog.Warning.Println("PN:", Key, "has no Publication date, Default to", pubdate)
			}

			switch iIndex {
			case "PN":
				value := new(KeyPN)
				value.Pubdate = pubdate
				Key = /*sindexd.Delimiter + */ cc + sindexd.Delimiter + pn + sindexd.Delimiter + kc

				value.Docid = strings.TrimSpace(line[0:10])
				if valuej, err := json.Marshal(value); err == nil {
					Pdate.Key = Key
					Pdate.Value = string(valuej)
					Pubdatea[j] = Pdate
				} else {
					goLog.Warning.Println("Function:", f, "skipping", Key, "Bad input value at line", i)
				}
			case "PD":
				value := new(KeyPD)
				value.Pubdate = pubdate
				Key = cc + sindexd.Delimiter + pubdate[0:4] + sindexd.Delimiter + pubdate[4:6] + sindexd.Delimiter + pubdate[6:8] +
					sindexd.Delimiter + pn + sindexd.Delimiter + kc
				if valuej, err := json.Marshal(value); err == nil {
					Pdate.Key = Key
					Pdate.Value = string(valuej)
					Pubdatea[j] = Pdate

				} else {
					goLog.Warning.Println("Function:", f, "skipping", Key, "Bad input value at line", i)
				}
			}
			j++
		}
		if j >= bulk || dep >= fsize {
			var n time.Duration
			n = 100
			/* input files are already sort by Publivation number */
			/* sort by publication date when indexing publication date */
			/*
				if iIndex == "PD" {
					sort.Sort(sindexd.ByKey(Pubdatea))
				}
			*/
			sort.Sort(sindexd.ByKey(Pubdatea))
		Here:
			for r := 1; r <= 3; r++ { // retry 3 times

				start = time.Now()
				if resp, err := AddKeysa(client, index, &Pubdatea); err != nil {
					goLog.Error.Println("Function:", f, "last line:", i, "last key:", Key, err)
				} else {
					switch resp.StatusCode {
					case 200:
						response = new(sindexd.Response)
						body = sindexd.GetBody(resp)
						if err := json.Unmarshal(body, &response); err != nil {
							goLog.Error.Println(err)
						} else {
							goLog.Info.Println("Function:", f, i, "sindexd response:", response.Status, response.Reason, "Duration:", time.Since(start))
							if response.Status == 500 || response.Status == 404 {
								time.Sleep(n * time.Millisecond)
								goLog.Error.Println(r, n, " Retry on ", response.Status, response.Reason)
								n = n * 2
							} else {
								break Here
							}
						}
					default:
						goLog.Warning.Println("Function:", f, i, resp.Status, time.Since(start))
						break Here
					}
					defer resp.Body.Close()
				}
			}
			j = 0
		}
		i++
		if sindexd.Maxinput != 0 && i > sindexd.Maxinput {
			break
		}
	}
}

func AddPn1(input string, client *http.Client, hp hostpool.HostPool, index *sindexd.Index_spec) (*http.Response, error) {
	//This function is using the input files extracted by Wistre
	// Just for testing alternative to < Createkey + Addkeys >
	var (
		request bytes.Buffer
		value   bytes.Buffer
		key     string
		docid   string
		pubdate string
		line    string
		record  int
		first   bool
		err     error
		file    *os.File
	)

	l := &sindexd.Load{
		Index_spec: sindexd.Index_spec{
			Index_id: index.Index_id,
			Cos:      index.Cos,
			Vol_id:   index.Vol_id,
			Specific: index.Specific,
		},
	}

	if file, err = os.Open(input); err != nil {
		goLog.Error.Println(err)
		os.Exit(3)
	}
	defer file.Close()
	first = true
	in := bufio.NewReader(file)
	record = 0
	request.WriteString(`{"add":{`)
	for err == nil {
		line, err = in.ReadString('\n')
		record++
		aLine := strings.Split(line, ",")

		if len(aLine) == 6 && !first {

			// key = sindexd.Delimiter + aLine[0] + Adddelimiter(aLine[1], sindexd.Delimiter) + aLine[2]
			key = sindexd.Delimiter + aLine[0] + sindexd.Delimiter + aLine[1] + sindexd.Delimiter + aLine[2]
			if len(aLine[3]) != 0 {
				docid = aLine[3]
			} else {
				docid = aLine[0] + aLine[1] + aLine[2]
			}
			pubdate = aLine[4]
			value.Reset()
			value.WriteString(`{"Docid":"`)
			value.WriteString(docid)
			value.WriteString(`","Pubdate":"`)
			value.WriteString(pubdate)
			value.WriteString(`"}`)

			request.WriteString(`"`)
			request.WriteString(key)
			request.WriteString(`":`)
			request.Write(value.Bytes())
			request.WriteString(",")
		}
		first = false
	}

	request.WriteString("}}")
	// goLog.Info.Println(request.String())
	if lj, err1 := json.Marshal(l); err1 != nil {
		goLog.Error.Println(err1)
		return nil, err1
	} else {
		myreq := [][]byte{[]byte(sindexd.AG), []byte(sindexd.HELLO), []byte(sindexd.V), lj, []byte(sindexd.V), request.Bytes(), []byte(sindexd.AD)}
		request := bytes.Join(myreq, []byte(""))
		return sindexd.PostRequest(client, request)
	}
}
