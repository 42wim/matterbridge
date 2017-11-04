package helper

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

func DownloadFile(url string) (*[]byte, error) {
	var buf bytes.Buffer
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	io.Copy(&buf, resp.Body)
	data := buf.Bytes()
	resp.Body.Close()
	return &data, nil
}
