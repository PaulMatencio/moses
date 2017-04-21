package main

import (
	"flag"
	"fmt"
	sproxyd "moses/sproxyd/lib"
	user "moses/user/base64j"

	"net/http"
)

func usage() {

}

var (
	config, env, keyid, defaultConfig string
)

func main() {

	flag.Usage = usage
	flag.StringVar(&config, "config", defaultConfig, "Config file")
	flag.StringVar(&env, "env", "", "Environment")
	flag.StringVar(&keyid, "keyid", "", "Keyid  to be decoded")
	flag.Parse()
	host := "http://10.12.202.10:81/proxy/chord/"
	client := &http.Client{
		Timeout:   sproxyd.WriteTimeout,
		Transport: sproxyd.Transport,
	}
	keyid = "7AB17AD5F32DA99E32024CAA581C2EC1EBC22920"
	req, _ := http.NewRequest("HEAD", host+keyid, nil)

	if resp, err := client.Do(req); err == nil {
		if resp.StatusCode == 200 {
			usermd := resp.Header["X-Scal-Usermd"][0]
			if meta, err := user.Decode64(usermd); err == nil {
				fmt.Println(string(meta))
			} else {
				fmt.Println(err)
			}
		}

	} else {
		fmt.Println(err)
	}

}
