package kopuro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type kopuroClient struct {
	baseURL    string
	httpClient *http.Client
}

func newKopuroClient(baseURL string) kopuroClient {
	httpClient := &http.Client{}
	return kopuroClient{baseURL, httpClient}
}

func (kc kopuroClient) writeJSONFile(filename string, content interface{}) error {
	requestMap := map[string]interface{}{
		"filename": filename,
		"content": content,
	}
	requestJson, err := json.Marshal(requestMap)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(requestJson)

	request, err := http.NewRequest("POST", kc.baseURL + "/json/write", buf)
	if err != nil {
		return err
	}

	resp, err := kc.httpClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	errorMap := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&errorMap)
	if err != nil {
		return err
	}

	errorResponse, ok := errorMap["error"]
	if !ok {
		return fmt.Errorf("kopuro returned http status code %v with no error description", resp.StatusCode)
	}

	return fmt.Errorf("http %s from kopuro - %s", resp.Status, errorResponse)
}