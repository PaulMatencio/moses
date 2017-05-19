package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	bns "moses/bns/lib"
	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	"net/http"
	"os"
	"time"
	// "github.com/pkg/errors"
)

const command = "KeytoPath"

var (
	keyid, env, driver, bpdriver, ip string
	get, write                       bool
)

func usage() {
	fmt.Printf("Usage %s -keyid <ObjectId> \n -ip <ip:port> -driver <by key driver>  -bpdriver <by path driver> -get <true/false>", command)
	flag.PrintDefaults()
	os.Exit(1)
}

func getPath(req *http.Request, client *http.Client) (string, error) {

	var (
		docmeta  = bns.DocumentMetadata{}
		pagemeta = bns.Pagemeta{}
	)
	start := time.Now()
	if resp, err := client.Do(req); err == nil {

		if resp.StatusCode == 200 {
			usermd := resp.Header["X-Scal-Usermd"][0]
			fmt.Printf("Time to getPath %v\n", time.Since(start))
			if err = pagemeta.UsermdToStruct(usermd); err == nil {
				if pagemeta.PageNumber > 0 {
					return pagemeta.GetPathName(), nil
				} else {
					err = docmeta.UsermdToStruct(usermd)
					return docmeta.GetPathName(), err
				}
			} else if err = docmeta.UsermdToStruct(usermd); err == nil {
				return docmeta.GetPathName(), nil
			} else {
				usermd_decoded, _ := base64.Decode64(usermd)
				fmt.Printf("Bad user metata : %v\n", string(usermd_decoded))
				return "", err
			}

		} else {
			return "", errors.New(resp.Status)
		}
	} else {
		return "", err
	}

}

func getObject(req *http.Request, client *http.Client) ([]byte, error) {
	var (
		body []byte
		err  error
		resp *http.Response
	)
	if resp, err = client.Do(req); err == nil {
		defer resp.Body.Close()
		if err == nil {
			switch resp.StatusCode {
			case 200:
				body, _ = ioutil.ReadAll(resp.Body)
			case 404:
				fmt.Println("... Try the command with the option  -env test")
				err = errors.New(resp.Status)
			default:
				err = errors.New(resp.Status)
			}
		} else {
			resp.Body.Close()
		}
	}
	return body, err

}

func main() {

	flag.Usage = usage
	flag.StringVar(&env, "env", "prod", "environment")
	flag.StringVar(&driver, "driver", "chord", "by key driver")
	flag.StringVar(&bpdriver, "bpdriver", "bpchord", "by key driver")
	flag.StringVar(&ip, "ip", "10.12.202.10:81", "by key driver")
	flag.StringVar(&keyid, "keyid", "", "Object Id")
	flag.BoolVar(&get, "get", true, "get the object by path name")
	flag.Parse()

	var (
		req      *http.Request
		proxy    = sproxyd.Proxy
		pathname string
		err      error
		hostbk   = fmt.Sprintf("http://%s/%s/%s/", ip, proxy, driver)
		hostbp   = fmt.Sprintf("http://%s/%s/%s/%s/", ip, proxy, bpdriver, env)
	)

	client := &http.Client{
		Timeout:   sproxyd.WriteTimeout,
		Transport: sproxyd.Transport,
	}

	if len(keyid) > 0 {
		req, _ = http.NewRequest("HEAD", hostbk+keyid, nil)
	} else {
		fmt.Printf("%s\n\n", "-keyid is missing")
		usage()
	}

	if pathname, err = getPath(req, client); err == nil {
		fmt.Printf("Hostname: %s\nPathname:%s\n", hostbp, pathname)
		if get {
			start := time.Now()
			req, _ = http.NewRequest("GET", hostbp+pathname, nil)
			var body []byte
			if body, err = getObject(req, client); err == nil {
				fmt.Printf("Object length: %d\n", len(body))
			} else {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Printf("Time for getting=> %s%s : %v", hostbp, pathname, time.Since(start))
		}
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}
