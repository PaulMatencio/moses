package sproxyd

import (
	"net"
	"net/http"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

const Proxy = "proxy"

var (
	Url          = "http://10.12.201.11:81/,http://10.12.201.12:81/,http://10.12.201.21:81/,http://10.12.201.22:81/,http://10.12.201.31:81/,http://10.12.201.32:81/"
	TargetUrl    = "http://10.12.202.10:81/,http://10.12.202.11:81/,http://10.12.202.12:81/,http://10.11.202.13:81/,http://10.11.202.20:81/,http://10.11.202.21:81/,http://10.11.202.22:81/, http://10.11.202.23:81/"
	Debug        bool                             /* debug mode */
	Test         bool                             /*  test mode  */
	HP           hostpool.HostPool                /* source hosts pool */
	TargetHP     hostpool.HostPool                /* destination hostpool */
	Driver       = "bparc"                        /* default  source sproxyd driver */
	TargetDriver = "bpchord"                      /* destination sproxy driver */
	DummyHost    = "http://0.0.0.0:81/"           /* Used by doRequest.go  to build the url with hostpool */
	Timeout      = time.Duration(50)              /* GET/PUT timeout */
	CopyTimeout  = time.Duration(1 * time.Second) /* Copy PNs TIME OUT */
	ReadTimeout  = time.Duration(5 * time.Second) /* 5sec*/
	WriteTimeout = time.Duration(20 * time.Second)

	DoRetry = 5 /* number of low level sproxyd Retry if errors */
	Host    = []string{"http://10.12.201.11:81/proxy/bparc/", "http://10.12.201.12:81/proxy/bparc/", "http://10.11.201.21:81/proxy/bparc/",
		"http://10.11.201.22:81/proxy/bparc/", "http://10.11.201.31:81/proxy/bparc/", "http://10.11.201.31:81/proxy/bparc/"}
	Env                    = "prod"
	TargetHost             = []string{"http://10.12.202.10:81/proxy/bpchord/", "http://10.12.202.11:81/proxy/bpchord/", "http://10.12.202.12:81/proxy/bpchord/", "http://10.12.202.13:81/proxy/bpchord/", "http://10.12.202.20:81/proxy/bpchord/", "http://10.12.202.21:81/proxy/bpchord/", "http://10.12.202.22:81/proxy/bpchord/", "http://10.12.202.23:81/proxy/bpchord/"}
	TargetEnv              = "moses-prod"
	Num200, Num404, Num412 = 0, 0, 0

	//Host = []string{"http://10.12.201.11:81/proxy/bparc/", "http://10.12.201.12:81/proxy/bparc/"}
	// hlist := strings.Split(sproxyd.Url, ",")
	// sproxyd.HP = hostpool.NewEpsilonGreedy(hlist, 0, &hostpool.LinearEpsilonValueCalculator{})

	Transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   100 * time.Millisecond,
			KeepAlive: 20 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
)

// sproxyd htp request structure
type HttpRequest struct {
	Hspool    hostpool.HostPool
	Client    *http.Client
	Path      string
	ReqHeader map[string]string
	// Buffer    []byte
}

// sproxyd http response structure
type HttpResponse struct {
	Url      string
	Response *http.Response
	Body     *[]byte
	Err      error
}
