package sindexd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"

	hostpool "github.com/bitly/go-hostpool"
)

type Configuration struct {
	Hosts        []string `json:"hosts"`
	TargetHosts  []string `json:"targetHosts"`
	Driver       string   `json:"driver,omitempty"`
	TargetDriver string   `json:"targetDriver,omitempty"`
	Log          string   `json:"logpath"`
}

func (c *Configuration) SetParmConfig(filename string) error {

	usr, _ := user.Current()
	configdir := path.Join(usr.HomeDir, "sindexd")
	configfile := path.Join(configdir, filename)
	cfile, err := os.Open(configfile)
	defer cfile.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	decoder := json.NewDecoder(cfile)
	err = decoder.Decode(&c)
	return err
}

func (c Configuration) GetHost() []string {
	return c.Hosts
}

func (c Configuration) GetTargetHost() []string {
	return c.TargetHosts
}

func (c Configuration) GetLogPath() string {
	return c.Log
}

func GetConfig(c_file string) (Configuration, error) {
	var (
		config     = "sindexd/config"
		usr, _     = user.Current()
		configfile = path.Join(path.Join(usr.HomeDir, config), c_file)
		cfile, err = os.Open(configfile)
	)
	if err != nil {
		fmt.Println("sindexd.GetParmConfig:", err)
		fmt.Println("Trying /etc/moses/" + config)
		configfile = path.Join(path.Join("/etc/moses", config), c_file)
		if cfile, err = os.Open(configfile); err != nil {
			fmt.Println("sproxyd.GetConfig:", err)
			os.Exit(2)
		}

	}
	defer cfile.Close()
	decoder := json.NewDecoder(cfile)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	return configuration, err
}

func SetNewHost(Config Configuration) {
	HP = hostpool.NewEpsilonGreedy(Config.Hosts, 0, &hostpool.LinearEpsilonValueCalculator{})
	TargetHP = hostpool.NewEpsilonGreedy(Config.TargetHosts, 0, &hostpool.LinearEpsilonValueCalculator{})
	Driver = Config.Driver
	TargetDriver = Config.TargetDriver
	Host = Host[:0]
	Host = Config.GetHost()[0:]
	TargetHost = TargetHost[:0]
	TargetHost = Config.GetTargetHost()[0:]
	// fmt.Println(HP, Driver, Host)

}
