package sproxyd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

const Proxy = "proxy"

var (
	Url   = "http://10.12.201.11:81/,http://10.12.201.12:81/,http://10.12.201.21:81/,http://10.12.201.22:81/,http://10.12.201.31:81/,http://10.12.201.32:81/"
	Debug bool
	HP    hostpool.HostPool
	// Driver    = "chord"
	Driver    = "bparc"
	DummyHost = "http://0.0.0.0:81/"
	Timeout   = time.Duration(50)
	Host      = []string{"http://10.12.201.11:81/proxy/bparc/", "http://10.12.201.12:81/proxy/bparc/", "http://10.11.201.21:81/proxy/bparc/",
		"http://10.11.201.22:81/proxy/bparc/", "http://10.11.201.31:81/proxy/bparc/", "http://10.11.201.31:81/proxy/bparc/"}

	//Host = []string{"http://10.12.201.11:81/proxy/bparc/", "http://10.12.201.12:81/proxy/bparc/"}
	// hlist := strings.Split(sproxyd.Url, ",")
	// sproxyd.HP = hostpool.NewEpsilonGreedy(hlist, 0, &hostpool.LinearEpsilonValueCalculator{})
)

type Configuration struct {
	Sproxyd []string `json:"sproxyd"`
	Driver  string   `json:"driver,omitempty"`
	Log     string   `json:"logpath"`
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
func (c Configuration) GetProxyd() (Sproxyd []string) {
	return c.Sproxyd
}

func (c Configuration) GetDriver() (Sproxyd string) {
	return c.Driver
}

func (c Configuration) GetLog() (Log string) {
	return c.Log
}

func (c Configuration) GetLogPath() (LogPath string) {
	return c.Log
}

type HttpResponse struct {
	Url      string
	Response *http.Response
	Body     []byte
	Err      error
}
