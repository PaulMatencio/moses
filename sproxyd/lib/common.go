package sproxyd

import (
	"net/http"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

const Proxy = "proxy"

var (
	Url       = "http://10.12.201.11:81/,http://10.12.201.12:81/,http://10.12.201.21:81/,http://10.12.201.22:81/,http://10.12.201.31:81/,http://10.12.201.32:81/"
	TargetUrl = "http://10.12.202.10:81/,http://10.12.202.11:81/,http://10.12.202.12:81/,http://10.11.202.13:81/,http://10.11.202.20:81/,http://10.11.202.21:81/,http://10.11.202.22:81/, http://10.11.202.23:81/"
	Debug     bool
	HP        hostpool.HostPool
	TargetHP  hostpool.HostPool
	// Driver    = "chord"
	Driver       = "bparc"
	TargetDriver = "chord"
	DummyHost    = "http://0.0.0.0:81/"
	Timeout      = time.Duration(50)
	Host         = []string{"http://10.12.201.11:81/proxy/bparc/", "http://10.12.201.12:81/proxy/bparc/", "http://10.11.201.21:81/proxy/bparc/",
		"http://10.11.201.22:81/proxy/bparc/", "http://10.11.201.31:81/proxy/bparc/", "http://10.11.201.31:81/proxy/bparc/"}

	TargetHost = []string{"http://10.12.202.10:81/proxy/chord/", "http://10.12.202.11:81/proxy/chord/", "http://10.12.202.12:81/proxychord/", "http://10.11.202.13:81/proxy/chord/", "http://10.11.202.20:81/proxy/chord/", "http://10.11.202.21:81/proxy/chord/", "http://10.11.202.22:81/proxy/chord/", "http://10.11.202.23:81/proxy/chord/"}

	//Host = []string{"http://10.12.201.11:81/proxy/bparc/", "http://10.12.201.12:81/proxy/bparc/"}
	// hlist := strings.Split(sproxyd.Url, ",")
	// sproxyd.HP = hostpool.NewEpsilonGreedy(hlist, 0, &hostpool.LinearEpsilonValueCalculator{})
)

type HttpResponse struct {
	Url      string
	Response *http.Response
	Body     *[]byte
	Err      error
}
