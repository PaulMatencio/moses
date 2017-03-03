// base64 project base64.go
package base64

import (
	"encoding/base64"
	"encoding/json"
)

func Encode64(directory map[string]string) (string, error) {
	jsonB, err := json.Marshal(directory)
	out_base64 := base64.StdEncoding.EncodeToString(jsonB)
	return out_base64, err
}

func Decode64(usermd_base64 string) (map[string]interface{}, error) {
	// decode the encoded base64 user metadata
	usermd, _ := base64.StdEncoding.DecodeString(usermd_base64)
	/* We need to provide a variable where the JSON package can put the decoded data.
	   This map[string]interface{} will hold a map of strings to arbitrary data types. */
	var bnsdirectory map[string]interface{}
	err := json.Unmarshal(usermd, &bnsdirectory)
	return bnsdirectory, err
}

func Decode64_1(usermd_base64 string) (map[string]string, error) {

	usermd, _ := base64.StdEncoding.DecodeString(usermd_base64)

	var bnsdirectory map[string]string
	err := json.Unmarshal(usermd, &bnsdirectory)
	return bnsdirectory, err
}
