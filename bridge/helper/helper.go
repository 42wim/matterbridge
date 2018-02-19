package helper

import (
	"bytes"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
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
		return nil, err
	}
	defer resp.Body.Close()
	io.Copy(&buf, resp.Body)
	data := buf.Bytes()
	return &data, nil
}

func SplitStringLength(input string, length int) string {
	a := []rune(input)
	str := ""
	for i, r := range a {
		str = str + string(r)
		if i > 0 && (i+1)%length == 0 {
			str += "\n"
		}
	}
	return str
}

// handle all the stuff we put into extra
func HandleExtra(msg *config.Message, general *config.Protocol) []config.Message {
	extra := msg.Extra
	rmsg := []config.Message{}
	if len(extra[config.EVENT_FILE_FAILURE_SIZE]) > 0 {
		for _, f := range extra[config.EVENT_FILE_FAILURE_SIZE] {
			fi := f.(config.FileInfo)
			text := fmt.Sprintf("file %s too big to download (%#v > allowed size: %#v)", fi.Name, fi.Size, general.MediaDownloadSize)
			rmsg = append(rmsg, config.Message{Text: text, Username: "<system> ", Channel: msg.Channel})
		}
		return rmsg
	}
	return rmsg
}

func GetAvatar(av map[string]string, userid string, general *config.Protocol) string {
	if sha, ok := av[userid]; ok {
		return general.MediaServerUpload + "/" + sha + "/" + userid + ".png"
	}
	return ""
}
