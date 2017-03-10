package sproxyd

import (
	"net/http"
)

func DeleteObject(client *http.Client, path string) (*http.Response, error) {

	url := DummyHost + path
	req, _ := http.NewRequest("DELETE", url, nil)
	// resp, err := client.Do(req)
	return DoRequest(HP, client, req, nil)
}
