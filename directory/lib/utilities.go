package directory

import (
	"encoding/json"
	"errors"
	"github.com/bitly/go-hostpool"
	sindexd "github.com/moses/sindexd/lib"
	// goLog "github.com/moses/user/goLog"
	goLog "github.com/s3/gLog"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func Print(iIndex string, body []byte) int {

	response := new(sindexd.Response)
	goLog.Info.Println(response)
	if err := json.Unmarshal(body, &response); err != nil {
		goLog.Error.Println(err)
	} else {
		response.PrintFetched()
	}
	if len(response.Not_found) != 0 {
		goLog.Info.Println(response.Not_found, "NOT FOUND")
	}
	if len(response.Common_prefix) != 0 {
		goLog.Info.Println("Common Prefix:", response.Common_prefix)
	}
	if response.Truncated == true {
		goLog.Info.Println("Response is truncated, Next_marker is:", response.GetNMarker())
	}
	return len(response.Fetched)
}

func PrintResponse(responses []*HttpResponse) {
	for i := range responses {
		var (
			iresponse sindexd.Response
			err       error
			pref      string
		)
		pref = responses[i].Pref
		err = responses[i].Err
		//index_id := responses[i].indexId
		if err == nil {
			iresponse = *responses[i].Response
			iresponse.PrintFetched()
			if iresponse.Next_marker != "" {
				goLog.Info.Println("Next marker:", iresponse.Next_marker)
			}
			if len(iresponse.Not_found) != 0 {
				iresponse.PrintNotFound()
			}
			if len(iresponse.Common_prefix) != 0 {
				iresponse.PrintCommonPrefix()
			}
		} else {
			goLog.Error.Println(pref, err)
		}
	}
}



func CountResponse(responses []*HttpResponse) (map[string]int, string) {
	m := make(map[string]int)
	var nextMarker string
	for i := range responses {
		var (
			iresponse sindexd.Response
			err       error
			pref      string
		)
		pref = responses[i].Pref
		err = responses[i].Err
		if err == nil {
			iresponse = *responses[i].Response
			m[pref] = len(iresponse.Fetched)
			if iresponse.Next_marker != "" {
				nextMarker = iresponse.Next_marker
			}
		} else {
			goLog.Error.Println(pref, err)
		}
	}
	return m, nextMarker
}

func GetResponse(response *HttpResponse) ([]string, string) {

	var (
		keys       []string
		nextMarker string
		iresponse  sindexd.Response
		err        error
		pref       string
	)
	pref = response.Pref
	err = response.Err

	if err == nil {
		iresponse = *response.Response
		if iresponse.Status == 200 {
			keys, nextMarker = iresponse.GetFetchedKeys()
			if sindexd.Debug {
				goLog.Trace.Println(iresponse)
			}
			if iresponse.Next_marker != "" {
				goLog.Info.Printf("Next marker: %s\n", iresponse.Next_marker)
			} else {
				goLog.Info.Printf("No more marker\n")
			}
		} else {
			goLog.Error.Printf("Sindexd Status: %d  Sindexd Reason: %s", iresponse.Status, iresponse.Reason)
		}
	} else {
		goLog.Error.Println(pref, err)
	}
	return keys, nextMarker
}


func Check(iIndex string, start time.Time, resp *http.Response) {

	time0 := time.Since(start)
	var num int
	switch iIndex {
	case "PN":
		num = Print(iIndex, sindexd.GetBody(resp))

	case "PD":
		num = Print(iIndex, sindexd.GetBody(resp))
	}
	goLog.Info.Println("Elapsed time:", time.Since(start), "Retrieve", num, "indexes in", time0)
}

func SetCPU(cpu string) error {
	var numCPU int

	availCPU := runtime.NumCPU()

	if strings.HasSuffix(cpu, "%") {
		// Percent
		var percent float32
		pctStr := cpu[:len(cpu)-1]
		pctInt, err := strconv.Atoi(pctStr)
		if err != nil || pctInt < 1 || pctInt > 100 {
			return errors.New("Invalid CPU value: percentage must be between 1-100")
		}
		percent = float32(pctInt) / 100
		numCPU = int(float32(availCPU) * percent)
	} else {
		// Number
		num, err := strconv.Atoi(cpu)
		if err != nil || num < 1 {
			return errors.New("Invalid CPU value: provide a number or percent greater than 0")
		}
		numCPU = num
	}

	if numCPU > availCPU {
		numCPU = availCPU
	}

	runtime.GOMAXPROCS(numCPU)
	return nil
}

/*
func BuildIndexSpec(id_spec map[string][]string) map[string]*sindexd.Index_spec {
	m := make(map[string]*sindexd.Index_spec)
	for k, v := range id_spec {

		cos, _ := strconv.Atoi(v[1])
		volid, _ := strconv.Atoi(v[2])
		specific, _ := strconv.Atoi(v[3])
		m[k] = &sindexd.Index_spec{
			Index_id: v[0],
			Cos:      cos,
			Vol_id:   volid,
			Specific: specific,
		}
	}
	return m
}
*/

func GetAsyncPrefixs(iIndex string, prefixs []string, delimiter string, markers []string, Limit int, Ind_Specs map[string]*sindexd.Index_spec) []*HttpResponse {

	//prefixs := strings.Split(prefix, ",")
	// url = hp.Get().Host()
	var j, marker string
	treq := len(prefixs)
	ch := make(chan *HttpResponse)
	responses := []*HttpResponse{}
	if treq > 0 {
		for i := range prefixs {

			pref := prefixs[i]
			if len(pref) > 2 {
				j = pref[0:2]
			} else {
				j = pref[0:]
			}
			index := Ind_Specs[j]
			if index == nil {
				index = Ind_Specs["OTHER"]
			}
			if len(markers) > i {
				marker = markers[i]
			}
			index.Read_only = 1
			go func(index *sindexd.Index_spec, pref string, marker string) {
				var (
					iresponse *sindexd.Response
					resp      *http.Response
					err       error
				)
				client := &http.Client{}
				if resp, err = GetPrefix(client, index, pref, delimiter, marker, Limit); err == nil {
					if resp.StatusCode == 200 {
						iresponse, err = sindexd.GetResponse(resp)
					} else {
						iresponse = nil
						err = errors.New(resp.Status)
					}
				}
				ch <- &HttpResponse{pref, iresponse, index,err}
			}(index, pref, marker)
		}
		// wait for Http response message
		for {
			select {
			case r := <-ch:
				// fmt.Printf("%s was fetched\n", r.err)
				responses = append(responses, r)

				if len(responses) == treq {
					return responses
				}
			case <-time.After(150 * time.Millisecond):
				goLog.Info.Printf(".")
			}
		}
	}
	return responses
}

func GetSerialPrefixs(iIndex string, prefixs []string, delimiter string, markers []string, Limit int, Ind_Specs map[string]*sindexd.Index_spec) []*HttpResponse {
	/* Does not work for XX and NP index tables*/
	var (
		iresponse *sindexd.Response
		resp      *http.Response
		err       error
		marker    string
		j         string
		r         *HttpResponse
		index     *sindexd.Index_spec
		pref	 string
	)
	responses := []*HttpResponse{}
	client := &http.Client{}
	//prefixs = strings.Split(prefix, ",")
	for i := range prefixs {
		pref = prefixs[i]

		if len(pref) > 2 {
			j = pref[0:2]
		} else {
			j = pref[0:]
		}
		index = Ind_Specs[j]
		if index == nil {
			index = Ind_Specs["OTHER"]
		}

		if len(markers) > i {
			marker = markers[i]
		}
		// goLog.Info.Println(index, pref, delimiter, marker, Limit)
		if resp, err = GetPrefix(client, index, pref, delimiter, marker, Limit); err == nil {
			// goLog.Info.Println("Status Code ===>", resp.StatusCode)
			if resp.StatusCode == 200 {
				iresponse, err = sindexd.GetResponse(resp)
			} else {
				iresponse = nil
				err = errors.New(resp.Status)
			}
		}
		// iresponse is nil if err != nil
		r = &HttpResponse{pref, iresponse, index,err}
		responses = append(responses, r)
	}
	return responses
}

func GetSerialPrefix(iIndex string, prefix string, delimiter string, marker string, Limit int, Ind_Specs map[string]*sindexd.Index_spec) *HttpResponse {

	var (
		iresponse *sindexd.Response
		resp      *http.Response
		err       error
		j string
		index     *sindexd.Index_spec
	)
	responses := &HttpResponse{}
	client := &http.Client{}

	// goLog.Info.Printf("Index: %s - Index Specifiaction %v",iIndex,Ind_Specs)

	switch (iIndex) {
		case "XX","NP":  /* recently loaded document */
			index=Ind_Specs[iIndex]
		default:    /* all other cases  PN, PD, OM , OB */
			if len(prefix) > 2 {
				j = prefix[0:2]
			} else {
				j = prefix[0:]
			}
			index = Ind_Specs[j]
			if index == nil {
				index = Ind_Specs["OTHER"]
			}
	}

	// goLog.Info.Println(index, pref, delimiter, marker, Limit)
	if resp, err = GetPrefix(client, index, prefix, delimiter, marker, Limit); err == nil {
		// goLog.Info.Println("Status Code ===>", resp.StatusCode)
		if resp.StatusCode == 200 {
			iresponse, err = sindexd.GetResponse(resp)
		} else {
			iresponse = nil
			err = errors.New(resp.Status)
		}
	}
	// iresponse is nil if err != nil
	responses = &HttpResponse{prefix, iresponse, index,err}

	return responses
}

func GetSerialKeys(specs map[string][]string, Ind_Specs map[string]*sindexd.Index_spec) []*HttpResponse {

	var (
		// iresponse *sindexd.Response
		err      error
		resp     *http.Response
		response *sindexd.Response
	)
	responses := []*HttpResponse{}
	client := &http.Client{}

	for k, v := range specs {
		index := Ind_Specs[k]
		AKey := v
		if Action == "Ge" {
			resp, err = GetKeys(client, index, &AKey)
		} else {
			resp, err = DeleteKeys(client, index, &AKey)
		}
		response, err = sindexd.GetResponse(resp)
		r := &HttpResponse{"", response, index,err}
		responses = append(responses, r)
	}
	return responses
}

func GetAsyncKeys(specs map[string][]string, Ind_Specs map[string]*sindexd.Index_spec) []*HttpResponse {

	//prefixs := strings.Split(prefix, ",")
	// url = hp.Get().Host()

	treq := len(specs)
	ch := make(chan *HttpResponse)
	responses := []*HttpResponse{}
	if treq > 0 {
		for k, v := range specs {
			index := Ind_Specs[k]
			AKey := v
			index.Read_only = 1
			go func(index *sindexd.Index_spec, Akey []string) {
				var (
					iresponse *sindexd.Response
					resp      *http.Response
					err       error
				)
				client := &http.Client{}
				if resp, err = GetKeys(client, index, &AKey); err == nil {
					if resp.StatusCode == 200 {
						iresponse, err = sindexd.GetResponse(resp)
					} else {
						iresponse = nil
						err = errors.New(resp.Status)
					}
				}
				ch <- &HttpResponse{"", iresponse, index,err}
			}(index, AKey)
		}
		// wait for Http response message
		for {
			select {
			case r := <-ch:
				// fmt.Printf("%s was fetched\n", r.err)
				responses = append(responses, r)

				if len(responses) == treq {
					return responses
				}
			case <-time.After(150 * time.Millisecond):
				goLog.Info.Printf(".")
			}
		}
	}
	return responses
}





func AddSerialPrefix1(HP hostpool.HostPool,iIndex string, prefix string, Ind_Specs map[string]*sindexd.Index_spec, keyObj map[string]string) *HttpResponse {

	var (
		iresponse *sindexd.Response
		resp      *http.Response
		err       error
		j string
		index     *sindexd.Index_spec
	)
	responses := &HttpResponse{}
	client := &http.Client{}

	switch (iIndex) {
		case "XX","NP":
			index=Ind_Specs[iIndex]
		default:    /* all other cases */
			if len(prefix) > 2 {
				j = prefix[0:2]
			} else {
				j = prefix[0:]
			}
			index = Ind_Specs[j]
			if index == nil {
				index = Ind_Specs["OTHER"]
			}
	}

	// goLog.Info.Println(index, pref, delimiter, marker, Limit)
	if resp,err = AddKeys1(HP,client,index,keyObj);err == nil  {
		if resp.StatusCode == 200 {
			iresponse, err = sindexd.GetResponse(resp)
		} else {
			iresponse = nil
			err = errors.New(resp.Status)
		}

	}
	// iresponse is nil if err != nil
	responses = &HttpResponse{prefix, iresponse, index,err}
	return responses
}


func AddSerialPrefix(prefix string, iIndex string, Ind_Specs map[string]*sindexd.Index_spec, keyObj map[string]string) *HttpResponse {

	var (
		iresponse *sindexd.Response
		resp      *http.Response
		err       error
		j string
		index     *sindexd.Index_spec
	)
	responses := &HttpResponse{}
	client := &http.Client{}
	switch (iIndex) {
	case "XX","NP":  /* recently loaded document */
		index=Ind_Specs[iIndex]
	default:    /* all other cases */
		if len(prefix) > 2 {
			j = prefix[0:2]
		} else {
			j = prefix[0:]
		}
		index = Ind_Specs[j]
		if index == nil {
			index = Ind_Specs["OTHER"]
		}
	}


	// goLog.Info.Println(index, pref, delimiter, marker, Limit)
	if resp,err = AddKeys(client,index,keyObj);err == nil  {
		if resp.StatusCode == 200 {
			iresponse, err = sindexd.GetResponse(resp)
		} else {
			iresponse = nil
			err = errors.New(resp.Status)
		}

	}
	// iresponse is nil if err != nil
	responses = &HttpResponse{prefix, iresponse, index,err}
	return responses
}
