package bns

import (
	sproxyd "moses/sproxyd/lib"
	base64 "moses/user/base64j"
	goLog "moses/user/goLog"
	"net/http"
	"time"
)

func UpdMetadata(client *http.Client, path string, usermd []byte) (error, time.Duration) {
	encoded_usermd := base64.Encode64(usermd)
	updheader := map[string]string{
		"Usermd":       encoded_usermd,
		"Content-Type": "image/tiff",
	}
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration

	if resp, err = sproxyd.UpdMetadata(client, path, updheader); err != nil {
		goLog.Error.Println(err)
	} else {
		elapse := time.Since(start)
		switch resp.StatusCode {
		case 200:
			goLog.Trace.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Key"], elapse)
		case 404:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, " not found", elapse)
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist", elapse)
		case 422:
			goLog.Error.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status, elapse)
		}
		resp.Body.Close() // Sproxyd does  not close the connection
	}
	return err, elapse

}
