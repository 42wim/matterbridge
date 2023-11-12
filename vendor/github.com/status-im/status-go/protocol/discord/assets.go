package discord

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/images"
)

func DownloadAvatarAsset(url string) ([]byte, error) {
	imgs, err := images.GenerateIdentityImagesFromURL(url)
	if err != nil {
		return nil, err
	}
	payload := imgs[0].Payload
	return payload, nil
}

func DownloadAsset(url string) ([]byte, string, error) {
	client := http.Client{Timeout: time.Minute}
	res, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("failed to close message asset http request body", "err", err)
		}
	}()

	contentType := res.Header.Get("Content-Type")
	bodyBytes, err := ioutil.ReadAll(res.Body)
	return bodyBytes, contentType, err
}
