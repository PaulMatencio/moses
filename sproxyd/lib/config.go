// sproxyd project sproxyd.go
package sproxyd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

// var Host = []string{"http://luo001t.internal.epo.org:81/proxy/chord/", "http://luo002t.internal.epo.org:81/proxy/chord/", "http://luo003t.internal.epo.org:81/proxy/chord/"}

type Configuration struct {
	Sproxyd           []string `json:"sproxyd"`
	TargetSproxyd     []string `json:"targetSproxyd,omitempty"`
	Driver            string   `json:"driver,omitempty"`
	TargetDriver      string   `json:"targetDriver,omitempty"`
	Env               string   `json:"env,omitempty`
	TargetEnv         string   `json:"targetEnv,omitempty`
	ReplicaPolicy     string   `json:"replicaPolicy,omitempty`
	Log               string   `json:"logpath"`
	OutDir            string   `json:"outputDir,omitempty"`
	Timeout           int      `json:"timeout,omitempty"`
	CopyTimeout       int      `json:"copyTimeout,omitempty"`
	WriteTimeout      int      `json:"writeTimeout,omitempty"`
	ConnectionTimeout int      `json:"connectionTimeout,omitempty"`
}

func InitConfig(config string) (Configuration, error) {
	var (
		err    error
		Config Configuration
	)
	if Config, err = GetConfig(config); err == nil {
		SetNewProxydHost(Config)
		Driver = Config.GetDriver()
		Env = Config.GetEnv()
		SetNewTargetProxydHost(Config)
		TargetDriver = Config.GetTargetDriver()
		TargetEnv = Config.GetTargetEnv()
		ReplicaPolicy = Config.GetReplicaPolicy()
		Timeout = Config.GetTimeout()
		CopyTimeout = Config.GetCopyTimeout()
		WriteTimeout = Config.GetWriteTimeout()
		ConnectionTimeout = Config.GetConnectionTimeout()
		fmt.Printf("INFO: Using config Hosts=>%s %s %s\n", Host, Driver, Env)
		fmt.Printf("INFO: Using config target Hosts=> %s %s %s\n", TargetHost, TargetDriver, TargetEnv)
		fmt.Printf("INFO: Timeout => Read:%v , Write:%v , Connection:%v\n", Timeout, WriteTimeout, ConnectionTimeout)
		// fmt.Printf("INFO: &HP %v\n HP: %v\n", &HP, HP)
	} else {
		// sproxyd.HP = hostpool.NewEpsilonGreedy(sproxyd.Host, 0, &hostpool.LinearEpsilonValueCalculator{})
		fmt.Printf("%v WARNING: Using defaults : Hosts=>%s %s Env %s %s\n", err, Host, TargetHost, Env, TargetEnv)
		fmt.Printf("$HOME/sproxyd/config/%s  must exist and well formed\n", config)
		Config = Configuration{}
	}
	return Config, err
}

func GetConfig(c_file string) (Configuration, error) {

	var (
		usr, _     = user.Current()
		config     = path.Join("sproxyd", "config")
		configfile = path.Join(path.Join(usr.HomeDir, config), c_file)
		cfile, err = os.Open(configfile)
	)
	if err != nil {
		fmt.Printf("sproxyd.GetConfig:%v\n", err)
		fmt.Printf("Trying /etc/moses/%v\n" + config)
		configfile = path.Join(path.Join("/etc/moses", config), c_file)
		if cfile, err = os.Open(configfile); err != nil {
			fmt.Printf("sproxyd.GetConfig:%v\n", err)
			os.Exit(2)
		}
	}
	defer cfile.Close()

	decoder := json.NewDecoder(cfile)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	return configuration, err
}

/*
func SetProxydHost(config string) (err error) {
	if Config, err := GetConfig(config); err == nil {
		HP = hostpool.NewEpsilonGreedy(Config.Sproxyd, 0, &hostpool.LinearEpsilonValueCalculator{})
		Host = Host[:0]
		Host = Config.GetProxyd()[0:]
	}
	return err
}
*/

func SetNewProxydHost(Config Configuration) {
	// fmt.Println(Config.Sproxyd)
	HP = hostpool.NewEpsilonGreedy(Config.Sproxyd, 0, &hostpool.LinearEpsilonValueCalculator{})
	Driver = Config.Driver
	Env = Config.Env
	Host = Host[:0] // reset
	Host = Config.GetProxyd()[0:]

}

func SetNewTargetProxydHost(Config Configuration) {
	// fmt.Println(Config.Sproxyd)
	TargetHP = hostpool.NewEpsilonGreedy(Config.TargetSproxyd, 0, &hostpool.LinearEpsilonValueCalculator{})
	TargetDriver = Config.TargetDriver
	TargetEnv = Config.TargetEnv
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

func (c Configuration) GetTargetProxyd() []string {
	return c.TargetSproxyd
}

func (c Configuration) GetProxyd() []string {
	return c.Sproxyd
}

func (c Configuration) GetEnv() string {
	return c.Env
}

func (c Configuration) GetTargetEnv() string {
	return c.TargetEnv
}

func (c Configuration) GetDriver() string {
	return c.Driver
}

func (c Configuration) GetTargetDriver() string {
	return c.TargetDriver
}

func (c Configuration) GetReplicaPolicy() string {
	replicaPolicy := strings.ToLower(c.ReplicaPolicy)
	if replicaPolicy != "immutable" && replicaPolicy != "consistent" {
		c.ReplicaPolicy = ""
	}
	return c.ReplicaPolicy
}

func (c Configuration) GetLog() string {
	return c.Log
}

func (c Configuration) GetLogPath() string {
	return c.Log
}

func (c Configuration) GetOutputDir() string {
	return c.OutDir
}

/* time out */

func (c Configuration) GetTimeout() time.Duration {
	return time.Duration(time.Duration(c.Timeout) * time.Second)

}

func (c Configuration) GetCopyTimeout() time.Duration {
	return time.Duration(time.Duration(c.CopyTimeout) * time.Second)
}

func (c Configuration) GetWriteTimeout() time.Duration {
	return time.Duration(time.Duration(c.WriteTimeout) * time.Second)
}

func (c Configuration) GetConnectionTimeout() time.Duration {
	return time.Duration(time.Duration(c.ConnectionTimeout) * time.Millisecond)
}
