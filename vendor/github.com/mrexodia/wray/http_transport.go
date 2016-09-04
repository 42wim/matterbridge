package wray

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HttpTransport struct {
	url      string
	SendHook func(data map[string]interface{})
}

func (self HttpTransport) isUsable(clientUrl string) bool {
	parsedUrl, err := url.Parse(clientUrl)
	if err != nil {
		return false
	}
	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		return true
	}
	return false
}

func (self HttpTransport) connectionType() string {
	return "long-polling"
}

func (self HttpTransport) send(data map[string]interface{}) (Response, error) {
	if self.SendHook != nil {
		self.SendHook(data)
	}
	dataBytes, _ := json.Marshal(data)
	buffer := bytes.NewBuffer(dataBytes)
	responseData, err := http.Post(self.url, "application/json", buffer)
	if err != nil {
		return Response{}, err
	}
	if responseData.StatusCode != 200 {
		return Response{}, errors.New(responseData.Status)
	}
	readData, _ := ioutil.ReadAll(responseData.Body)
	responseData.Body.Close()
	var jsonData []interface{}
	json.Unmarshal(readData, &jsonData)
	response := newResponse(jsonData)
	return response, nil
}

func (self *HttpTransport) setUrl(url string) {
	self.url = url
}
