package bns

import (
	"encoding/json"
	sproxyd "moses/sproxyd/lib"
	"net/http"
	// base64 "user/base64j"
	// "user/goLog"
	"errors"
	goLog "moses/user/goLog"
	"strconv"
)

func GetPage(client *http.Client, path string) (*http.Response, error) {

	header := map[string]string{}
	/*
		var resp *http.Response
		err := error(nil)
	*/
	// no specific request , just give me an object
	//start := time.Now()
	//var elapse time.Duration
	return sproxyd.GetObject(client, path, header)
	//elapse = time.Since(start)

}

func GetPageType(client *http.Client, path string, getHeader map[string]string) (*http.Response, error) {
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
		//goLog.Trace.Println("PageMeta", pagemeta)
		if contentType, ok := getHeader["Content-Type"]; ok {
			//goLog.Trace.Println("Content-Type", contentType)
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
				start := strconv.Itoa(pagemeta.PdfOffset.Start)
				end := strconv.Itoa(pagemeta.PdfOffset.End)
				getHeader["Range"] = "bytes=" + start + "-" + end
				resp, err = sproxyd.GetObject(client, path, getHeader)
			default:
				err = errors.New("Content-type is missing or invalid")
			}
		} else {
			err = errors.New("Content-type is missing or invalid")
		}
	}
	return resp, err
}
