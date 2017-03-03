package bns

import (
	"net/http"
	sproxyd "sproxyd/lib"
	"time"
	goLog "user/goLog"
)

func DeletePage(client *http.Client, path string) (error, time.Duration) {

	// deleteHeader := map[string]string{}
	err := error(nil)
	var resp *http.Response
	start := time.Now()
	var elapse time.Duration
	// defer resp.Body.Close()
	if resp, err = sproxyd.DeleteObject(client, path); err != nil {
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
			goLog.Error.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"], elapse)
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status, elapse)
		}
		resp.Body.Close()
	}
	return err, elapse
}
