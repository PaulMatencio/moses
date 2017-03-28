package sindexd

import (
	"net"
	"net/http"
	"time"

	hostpool "github.com/bitly/go-hostpool"
)

const (
	V      = ","
	AG     = "["
	AD     = "]"
	Url    = "http://10.12.201.11:81/sindexd.fcgi,http://10.12.201.12:81/sindexd.fcgi,http://10.12.201.21:81/sindexd.fcgi,http://10.12.201.22:81/sindexd.fcgi,http://10.12.201.31:81/sindexd.fcgi,http://10.12.201.32:81/sindexd.fcgi"
	HELLO  = `{ "hello" : { "protocol" : "sindexd-1" }}`
	CONFIG = `{"config":{}}`
)

var (
	Debug         bool
	Maxinput      int64
	Test          bool
	Memstat       bool
	Delimiter     string
	Host          []string
	TargetHost    []string
	Driver        string
	TargetDriver  string
	HP            hostpool.HostPool
	TargetHP      hostpool.HostPool
	Timeout       = time.Duration(30 * time.Second)
	ReadTimeout   = time.Duration(30 * time.Second)
	DeleteTimeout = time.Duration(3 * time.Minute)
	Transport     = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   100 * time.Millisecond, // connection timeout
			KeepAlive: 20 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
)

type Index_spec struct {
	Index_id  string `json:"index_id"`
	Cos       int    `json:"cos"`
	Vol_id    int    `json:"vol_id"`
	Specific  int    `json:"specific"`
	Read_only int    `json:"readonly,omitempty"`
	Admin     int    `json:"admin,omitempty"`
}

type Load struct {
	Index_spec `json:"load"`
	Version    int `json:"version,omitempty"`
}
