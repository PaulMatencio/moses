package base64j

import "encoding/base64"

func Encode64(jsonB []byte) string {
	return base64.StdEncoding.EncodeToString(jsonB)
}

func Decode64(usermd_base64 string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(usermd_base64)
}
