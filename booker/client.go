package booker

import (
	"encoding/json"
	"net/http"
)

func sendRequest(client *http.Client, req *http.Request, respDto interface{}) error {
	resp, clientErr := client.Do(req)
	if clientErr != nil {
		return clientErr
	}

	decodeErr := json.NewDecoder(resp.Body).Decode(respDto)
	if decodeErr != nil {
		return decodeErr
	}

	return nil
}
