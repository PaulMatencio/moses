// sproxyd project sproxyd.go
package sproxyd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"

	hostpool "github.com/bitly/go-hostpool"
)

// var Host = []string{"http://luo001t.internal.epo.org:81/proxy/chord/", "http://luo002t.internal.epo.org:81/proxy/chord/", "http://luo003t.internal.epo.org:81/proxy/chord/"}

func GetConfig(c_file string) (Configuration, error) {

	usr, _ := user.Current()
	configdir := path.Join(usr.HomeDir, "sproxyd/config")
	configfile := path.Join(configdir, c_file)
	cfile, err := os.Open(configfile)
	defer cfile.Close()
	if err != nil {
		fmt.Println("sproxyd.GetConfig", err)
		os.Exit(2)
	}
	decoder := json.NewDecoder(cfile)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	return configuration, err
}

func SetProxydHost(config string) (err error) {
	if Config, err := GetConfig(config); err == nil {

		HP = hostpool.NewEpsilonGreedy(Config.Sproxyd, 0, &hostpool.LinearEpsilonValueCalculator{})
		// for compatibility with old setProxydHost but Host[]
		Host = Host[:0]
		Host = Config.GetProxyd()[0:]
	}
	return err
}

func SetNewProxydHost(Config Configuration) {
	fmt.Println(Config.Sproxyd)
	HP = hostpool.NewEpsilonGreedy(Config.Sproxyd, 0, &hostpool.LinearEpsilonValueCalculator{})
	Driver = Config.Driver
	Host = Host[:0] // reset
	Host = Config.GetProxyd()[0:]

}
