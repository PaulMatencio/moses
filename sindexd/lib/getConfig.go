package sindexd

import (
	"bytes"
	"encoding/json"
	// goLog "github.com/moses/user/goLog"
	goLog "github.com/s3/gLog"
	"net/http"
)

type Get_Config struct {
	Config interface{} `json:"config"`
}

type Config struct {
	Status   int         `json:"status"`
	Protocol string      `json:"protocol"`
	Index_id string      `json:"indexd_id"`
	Config   interface{} `json:"config"`
}

func GetSindexdConfig(client *http.Client) (*http.Response, error) {
	/*
	   [{ "hello":{ "protocol": "sindexd-1"} },
	   {"config": { }} ]
	*/
	myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), []byte(CONFIG), []byte(AD)}
	request := bytes.Join(myreq, []byte(""))
	return PostRequest(client, request)
}

func GetConfigResponse(resp *http.Response) interface{} {
	var v interface{}
	response := new(Config)
	if err := json.Unmarshal(GetBody(resp), &response); err != nil {
		goLog.Error.Println(err)
	} else {
		v = response.Config

	}
	return v
}

func PrintConfig(f string, resp *http.Response) {
	goLog.Info.Println(GetConfigResponse(resp))

}
