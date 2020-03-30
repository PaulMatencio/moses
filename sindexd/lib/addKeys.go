package sindexd

import (
	"bytes"
	"encoding/json"
	"github.com/bitly/go-hostpool"
	"net/http"
	"runtime"
	// goLog "github.com/moses/user/goLog"
	goLog "github.com/s3/gLog"
)

type PubRecord struct {
	Key   string
	Value string
}

type ByKey []PubRecord

func (k ByKey) Len() int {
	return len(k)
}
func (k ByKey) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}
func (k ByKey) Less(i, j int) bool {
	return k[i].Key < k[j].Key
}

func Addkeys(client *http.Client, l *Load, keyObject map[string]string) (*http.Response, error) {
	/*
		l is a pointer to a Load (sindexd) structure
		keyObject is a map of "key" = obj  pair to be indexed
		[ { "hello":{ "protocol": "sindexd-1"} },
		{ "load":   {  "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x}},
		{ "add": { "k1": obj1, "k2": obj2, ..., "kn": objn }}]
	*/
	var keyobj bytes.Buffer
	keyobj.WriteString(`{"add":{`)
	i := 0
	for k, v := range keyObject {
		keyobj.WriteString(`"`)
		keyobj.WriteString(k) // key
		keyobj.WriteString(`":`)
		keyobj.WriteString(v) // value
		i++
		if i < len(keyObject) {
			keyobj.WriteString(V)
		}
	}
	// keyobj.WriteString("}}")
	keyobj.WriteString("},")
	keyobj.WriteString(`"prefetch":false`)
	keyobj.WriteString("}")

	pj := keyobj.Bytes()
	if lj, err := json.Marshal(l); err != nil {
		return nil, err
	} else {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), lj, []byte(V), pj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		if Memstat {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			goLog.Info.Println(m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)
		}
		return PostRequest(client, request)
	}

}


func Addkeys1(HP hostpool.HostPool, client *http.Client, l *Load, keyObject map[string]string) (*http.Response, error) {
	/*
		l is a pointer to a Load (sindexd) structure
		keyObject is a map of "key" = obj  pair to be indexed
		[ { "hello":{ "protocol": "sindexd-1"} },
		{ "load":   {  "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x}},
		{ "add": { "k1": obj1, "k2": obj2, ..., "kn": objn }}]
	*/
	var keyobj bytes.Buffer
	keyobj.WriteString(`{"add":{`)
	i := 0
	for k, v := range keyObject {
		keyobj.WriteString(`"`)
		keyobj.WriteString(k) // key
		keyobj.WriteString(`":`)
		keyobj.WriteString(v) // value
		i++
		if i < len(keyObject) {
			keyobj.WriteString(V)
		}
	}
	// keyobj.WriteString("}}")
	keyobj.WriteString("},")
	keyobj.WriteString(`"prefetch":false`)
	keyobj.WriteString("}")

	pj := keyobj.Bytes()
	if lj, err := json.Marshal(l); err != nil {
		return nil, err
	} else {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), lj, []byte(V), pj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		if Memstat {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			goLog.Info.Println(m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)
		}
		return PostRequest1(HP,client, request)
	}

}
/*
func Addkeysa(client *http.Client, url string, l *Load, keyp *[]string, valuep *[]string) (*http.Response, error) {
	/*
		l is a pointer to a Load (sindexd) structure
		keyObject is a map of "key" = obj  pair to be indexed
		[ { "hello":{ "protocol": "sindexd-1"} },
		{ "load":   {  "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x}},
		{ "add": { "k1": obj1, "k2": obj2, ..., "kn": objn }}]

	keya := *keyp
	valuea := *valuep
	var keyobj bytes.Buffer
	keyobj.WriteString(`{"add":{`)

	for i := range keya {
		keyobj.WriteString(`"`)
		keyobj.WriteString(keya[i]) // key
		keyobj.WriteString(`":`)
		keyobj.WriteString(valuea[i]) // value
		i++
		if i < len(keya) {
			keyobj.WriteString(V)
		}
	}
	// keyobj.WriteString("}}")
	keyobj.WriteString("},")
	keyobj.WriteString(`"prefetch":false`)
	keyobj.WriteString("}")

	pj := keyobj.Bytes()
	if lj, err := json.Marshal(l); err != nil {
		return nil, err
	} else {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), lj, []byte(V), pj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		if Memstat {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			goLog.Info.Println(m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)
		}
		return PostRequest(client, url, request)
	}

}
*/

func Addkeysa(client *http.Client, l *Load, pubdatep *[]PubRecord) (*http.Response, error) {
	/*
		l is a pointer to a Load (sindexd) structure
		keyObject is a map of "key" = obj  pair to be indexed
		[ { "hello":{ "protocol": "sindexd-1"} },
		{ "load":   {  "index_id": "xxxx", "cos": x, "vol_id": x, "specific": x}},
		{ "add": { "k1": obj1, "k2": obj2, ..., "kn": objn }}]
	*/
	pubdatea := *pubdatep
	var keyobj bytes.Buffer
	keyobj.WriteString(`{"add":{`)

	for i := range pubdatea {
		keyobj.WriteString(`"`)
		//keya := pubdatea[i].Key
		keyobj.WriteString(pubdatea[i].Key) // key
		keyobj.WriteString(`":`)
		keyobj.WriteString(pubdatea[i].Value) // value
		i++
		if i < len(pubdatea) {
			keyobj.WriteString(V)
		}
	}
	// keyobj.WriteString("}}")
	keyobj.WriteString("},")
	keyobj.WriteString(`"prefetch":false`)
	keyobj.WriteString("}")

	pj := keyobj.Bytes()
	if lj, err := json.Marshal(l); err != nil {
		return nil, err
	} else {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), lj, []byte(V), pj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		if Memstat {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			goLog.Info.Println(m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)
		}
		return PostRequest(client, request)
	}

}
