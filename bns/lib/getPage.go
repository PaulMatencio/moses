package bns

import (
	"encoding/json"
	"errors"
	sproxyd "moses/sproxyd/lib"
	goLog "moses/user/goLog"
	"net/http"
	"strconv"
	"strings"
)

// GetPage  will be used by  AsyncHttpGetPage for concurrenr getPage
func GetPage(client *http.Client, path string) (*http.Response, error) {
	header := map[string]string{}
	return sproxyd.GetObject(client, path, header)

}

// GetPageType will be used by  AsyncHttpGetPageType for concurrent Getpagetype
// func GetPageType(client *http.Client, path string, getHeader map[string]string) (*http.Response, error) {
func GetPageType(client *http.Client, path string, media string) (*http.Response, error) {
	//  getrHeader must contain Content-type
	var (
		usermd []byte
		err    error
		resp   *http.Response
	)
	usermd, err = GetMetadata(client, path)
	if err != nil {
		return nil, errors.New("Page metadata is missing or invalid")
	} else {
		// c, _ := base64.Decode64(string(usermd))
		//goLog.Trace.Println("Usermd=", string(usermd))
		if len(usermd) == 0 {
			return nil, errors.New("Page metadata is missing. Please check the warning log for the reason")
		}
		var pagemeta Pagemeta
		if err := json.Unmarshal(usermd, &pagemeta); err != nil {
			return nil, err
		}

		getHeader := map[string]string{}
		getHeader["Content-Type"] = "image/" + strings.ToLower(media)

		if contentType, ok := getHeader["Content-Type"]; ok {

			switch contentType {
			case "image/tiff", "image/tif":
				start := strconv.Itoa(pagemeta.TiffOffset.Start)
				end := strconv.Itoa(pagemeta.TiffOffset.End)
				getHeader["Range"] = "bytes=" + start + "-" + end
				goLog.Trace.Println(getHeader)
				resp, err = sproxyd.GetObject(client, path, getHeader)

			case "image/png":
				start := strconv.Itoa(pagemeta.PngOffset.Start)
				end := strconv.Itoa(pagemeta.PngOffset.End)
				getHeader["Range"] = "bytes=" + start + "-" + end
				resp, err = sproxyd.GetObject(client, path, getHeader)

			case "image/pdf":
				if pagemeta.PdfOffset.Start > 0 {
					start := strconv.Itoa(pagemeta.PdfOffset.Start)
					end := strconv.Itoa(pagemeta.PdfOffset.End)
					getHeader["Range"] = "bytes=" + start + "-" + end
					resp, err = sproxyd.GetObject(client, path, getHeader)
				} else {
					resp = nil
					err = errors.New("Content-type " + contentType + " does not exist")
				}
			default:
				err = errors.New("Content-type is missing or invalid")
			}
		} else {
			err = errors.New("Content-type is missing or invalid")
		}
	}
	return resp, err
}
