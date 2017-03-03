// Config project main.go
package main

import (
	"flag"
	"fmt"
	"os"
 	 
	 
	 
	"user/sproxyd"
)

func usage() {
	usage := "usage: ConfigTest -c config_file"
	fmt.Println(usage)
	flag.PrintDefaults()
	os.Exit(2)
}
func main() {
	var config string
	flag.Usage = usage
	flag.StringVar(&config, "c", "", "")
	flag.Parse()
	if len(config) != 0 {
		if err:= sproxyd.SetProxydHost(config);err == nil {
			fmt.Println(sproxyd.Host)
		} else { fmt.Println(err)}

		/*
		if Config,err := sproxyd.GetConfig(config); err == nil {
			sproxyd.Host = sproxyd.Host[:0]
			sproxyd.Host = Config.GetProxyd()[0:]
			fmt.Println(sproxyd.Host)

		} else {
			fmt.Println(err)
		}
		
		/*
		usr,_ := user.Current()
		configdir := path.Join(usr.HomeDir,"sproxyd")
		 
		configfile := filepath.Join(configdir, config)
		 
		if Config, err := sproxyd.GetConfig(configfile); err == nil {
			sproxyd.Host = sproxyd.Host[:0]
			sproxyd.Host = Config.GetProxyd()[0:]
			fmt.Println(sproxyd.Host)
		} else {
			fmt.Println(err)
		}
		*/
	} else {
		usage()
	}
      

}
