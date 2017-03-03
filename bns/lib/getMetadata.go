package bns

import (
	"net/http"
	sproxyd "sproxyd/lib"
	base64 "user/base64j"
	goLog "user/goLog"
)

func GetMetadata(client *http.Client, path string) ([]byte, error) {
	// client := &http.Client{}
	getHeader := map[string]string{}
	var usermd []byte
	var resp *http.Response
	err := error(nil)

	if resp, err = sproxyd.GetMetadata(client, path, getHeader); err == nil {
		switch resp.StatusCode {
		case 200:
			encoded_usermd := resp.Header["X-Scal-Usermd"]
			usermd, err = base64.Decode64(encoded_usermd[0])
		case 404:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status)
		case 412:
			goLog.Warning.Println(resp.Request.URL.Path, resp.Status, "key=", resp.Header["X-Scal-Ring-Key"], " does not exist")
		case 422:
			goLog.Error.Println(resp.Request.URL.Path, resp.Status, resp.Header["X-Scal-Ring-Status"])
		default:
			goLog.Info.Println(resp.Request.URL.Path, resp.Status)
		}
	}
	/* the resp,Body is closed by sproxyd.getMetadata */
	return usermd, err
}
