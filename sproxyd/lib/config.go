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

type Configuration struct {
	Sproxyd       []string `json:"sproxyd"`
	TargetSproxyd []string `json:"targetSproxyd,omitempty"`
	Driver        string   `json:"driver,omitempty"`
	TargetDriver  string   `json:"targetDriver,omitempty"`
	Log           string   `json:"logpath"`
	OutDir        string   `json:"outputDir,omitempty"`
}

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
	// fmt.Println(Config.Sproxyd)
	HP = hostpool.NewEpsilonGreedy(Config.Sproxyd, 0, &hostpool.LinearEpsilonValueCalculator{})
	Driver = Config.Driver
	Host = Host[:0] // reset
	Host = Config.GetProxyd()[0:]

}

func SetNewTargetProxydHost(Config Configuration) {
	// fmt.Println(Config.Sproxyd)
	TargetHP = hostpool.NewEpsilonGreedy(Config.TargetSproxyd, 0, &hostpool.LinearEpsilonValueCalculator{})
	TargetDriver = Config.TargetDriver
	TargetHost = TargetHost[:0] // reset
	TargetHost = Config.GetTargetProxyd()[0:]

}

func (c *Configuration) SetConfig(filename string) error {

	usr, _ := user.Current()
	configdir := path.Join(usr.HomeDir, "sproxyd")
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

func (c Configuration) GetTargetProxyd() (TargetSproxyd []string) {
	return c.TargetSproxyd
}

func (c Configuration) GetProxyd() (Sproxyd []string) {
	return c.Sproxyd
}

func (c Configuration) GetDriver() (Sproxyd string) {
	return c.Driver
}

func (c Configuration) GetTargetDriver() (Sproxyd string) {
	return c.TargetDriver
}

func (c Configuration) GetLog() (Log string) {
	return c.Log
}

func (c Configuration) GetLogPath() (LogPath string) {
	return c.Log
}

func (c Configuration) GetOutputDir() (OutDir string) {
	return c.OutDir
}
