package bns

import (
	"bytes"
	"net/http"
	sproxyd "sproxyd/lib"
	"time"
	goLog "user/goLog"
)

//func UpdatePage(client *http.Client, path string, img *bytes.Buffer, usermd []byte) (error, time.Duration) {
func UpdatePage(client *http.Client, path string, img *bytes.Buffer, putheader map[string]string) (error, time.Duration) {
	/*
		putheader := map[string]string{
			"Usermd":       encoded_usermd := base64.Encode64(usermd),
			"Content-Type": "image/tiff",
		}
	*/
	// url:= base_url+string(pub)+docid+string(kc)+"_"+string(pagenum)
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration
	// defer resp.Body.Close()
	if resp, err = sproxyd.UpdObject(client, path, img.Bytes(), putheader); err != nil {
		goLog.Error.Println(err)
	} else {
		elapse = time.Since(start)
		switch resp.StatusCode {
		case 200:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Key"], elapse)
		case 404:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, " not found", elapse)
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist", elapse)
		case 422:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		default:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, elapse)
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err, elapse

}
