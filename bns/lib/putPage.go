package bns

import (
	"bytes"
	"net/http"
	sproxyd "sproxyd/lib"
	"time"
	goLog "user/goLog"
)

//func PutPage(client *http.Client, path string, img *bytes.Buffer, usermd []byte) (error, time.Duration) {
func PutPage(client *http.Client, path string, img *bytes.Buffer, putheader map[string]string) (error, time.Duration) {
	/*
		putheader := map[string]string{
			"Usermd":       base64.Encode64(usermd),
			"Content-Type": "image/tiff",
		}
	*/
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration
	if resp, err = sproxyd.PutObject(client, path, img.Bytes(), putheader); err != nil {
		goLog.Error.Println(err)
	} else {
		elapse = time.Since(start)
		switch resp.StatusCode {
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], "already exist", elapse)
		case 422:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		case 200:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, "Elapse:", elapse)
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status, "Elapse:", elapse)
		}
		resp.Body.Close() // Sproxyd did  not close the connection
	}
	return err, elapse
}
