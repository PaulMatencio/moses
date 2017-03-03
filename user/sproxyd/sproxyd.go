// sproxyd project sproxyd.go
package sproxyd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"
	//"path/filepath"
	"strconv"
)

var Host = []string{"http://luo001t.internal.epo.org:81/proxy/chord/", "http://luo002t.internal.epo.org:81/proxy/chord/", "http://luo003t.internal.epo.org:81/proxy/chord/"}

type Configuration struct {
	Sproxyd []string
	LogPath string
}

func GetConfig(c_file string) (Configuration, error) {

	usr, _ := user.Current()
	configdir := path.Join(usr.HomeDir, "sproxyd")
	configfile := path.Join(configdir, c_file)
	cfile, err := os.Open(configfile)
	defer cfile.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	decoder := json.NewDecoder(cfile)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	return configuration, err
}

func (c Configuration) GetProxyd() (Sproxyd []string) {
	return c.Sproxyd
}

func (c Configuration) GetLogPath() (LogPath string) {
	return c.LogPath
}

func SetProxydHost(config string) (err error) {

	//var conffile string
	/*
		currentdir, _ := os.Getwd()
		configdir := filepath.Join(currentdir, "..", "config")
		configfile := filepath.Join(configdir, config)
	*/

	if Config, err := GetConfig(config); err == nil {
		Host = Host[:0]
		Host = Config.GetProxyd()[0:]
	}
	return err
}

func SetNewProxydHost(Config Configuration) {
	Host = Host[:0]
	Host = Config.GetProxyd()[0:]

}

func PutObject(client *http.Client, url string, object []byte, putHeader map[string]string) (*http.Response, error) {

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(object))
	if usermd, ok := putHeader["Usermd"]; ok {
		req.Header.Add("X-Scal-Usermd", usermd)
	}
	if contentType, ok := putHeader["Content-Type"]; ok {
		req.Header.Add("Content-Type", contentType)
	}
	if contentLength, ok := putHeader["Content-Length"]; ok {
		req.Header.Add("Content-Length", contentLength)
	} else {
		req.Header.Add("Content-Length", strconv.Itoa(len(object)))
	}
	req.Header.Add("If-None-Match", "*")
	//fmt.Println(req)
	resp, err := client.Do(req)

	return resp, err
}

func UpdObject(client *http.Client, url string, object []byte, putHeader map[string]string) (*http.Response, error) {

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(object))
	if usermd, ok := putHeader["Usermd"]; ok {
		req.Header.Add("X-Scal-Usermd", usermd)
	}
	if contentType, ok := putHeader["Content-Type"]; ok {
		req.Header.Add("Content-Type", contentType)
	}
	if contentLength, ok := putHeader["Content-Length"]; ok {
		req.Header.Add("Content-Length", contentLength)
	} else {
		req.Header.Add("Content-Length", strconv.Itoa(len(object)))
	}
	req.Header.Add("If-Match", "*")
	resp, err := client.Do(req)
	return resp, err
}

func DeleteObject(client *http.Client, url string, deleteHeader map[string]string) (*http.Response, error) {

	req, _ := http.NewRequest("DELETE", url, nil)
	//fmt.Println(req)
	resp, err := client.Do(req)
	return resp, err
}

func GetObject(client *http.Client, url string, getHeader map[string]string) (*http.Response, error) {

	req, _ := http.NewRequest("GET", url, nil)
	if Range, ok := getHeader["Range"]; ok {
		req.Header.Add("Range", Range)
	}
	if ifmod, ok := getHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := getHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	resp, err := client.Do(req)
	return resp, err
}

func GetMetadata(client *http.Client, url string, getHeader map[string]string) (*http.Response, error) {

	req, _ := http.NewRequest("HEAD", url, nil)
	if Range, ok := getHeader["Range"]; ok {
		req.Header.Add("Range", Range)
	}
	if ifmod, ok := getHeader["If-Modified-Since"]; ok {
		req.Header.Add("If-Modified-Since", ifmod)
	}
	if ifunmod, ok := getHeader["If-Unmodified-Since"]; ok {
		req.Header.Add("If-Unmodified-Since", ifunmod)
	}
	resp, err := client.Do(req)

	return resp, err
}

func UpdMetadata(client *http.Client, url string, updHeader map[string]string) (*http.Response, error) {

	req, _ := http.NewRequest("PUT", url, nil)
	if usermd, ok := updHeader["Usermd"]; ok {
		req.Header.Add("X-Scal-Usermd", usermd)
		/* update the metadata if the object exist */
		req.Header.Add("x-scal-cmd", "update-usermd") // tell Scality Ring to Update only the metadata
		req.Header.Add("If-Match", "*")
		resp, err := client.Do(req)
		return resp, err
	} else {
		// custom http response
		resp := new(http.Response)
		resp.StatusCode = 1000
		resp.Status = "1000 Metadata is missing"
		err := error(nil)
		return resp, err
	}

}
