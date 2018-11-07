package helper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/sirupsen/logrus"
)

func DownloadFile(url string) (*[]byte, error) {
	return DownloadFileAuth(url, "")
}

func DownloadFileAuth(url string, auth string) (*[]byte, error) {
	var buf bytes.Buffer
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest("GET", url, nil)
	if auth != "" {
		req.Header.Add("Authorization", auth)
	}
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
	for _, f := range extra[config.EVENT_FILE_FAILURE_SIZE] {
		fi := f.(config.FileInfo)
		text := fmt.Sprintf("file %s too big to download (%#v > allowed size: %#v)", fi.Name, fi.Size, general.MediaDownloadSize)
		rmsg = append(rmsg, config.Message{Text: text, Username: "<system> ", Channel: msg.Channel, Account: msg.Account})
	}
	return rmsg
}

func GetAvatar(av map[string]string, userid string, general *config.Protocol) string {
	if sha, ok := av[userid]; ok {
		return general.MediaServerDownload + "/" + sha + "/" + userid + ".png"
	}
	return ""
}

func HandleDownloadSize(flog *log.Entry, msg *config.Message, name string, size int64, general *config.Protocol) error {
	// check blacklist here
	for _, entry := range general.MediaDownloadBlackList {
		if entry != "" {
			re, err := regexp.Compile(entry)
			if err != nil {
				flog.Errorf("incorrect regexp %s for %s", entry, msg.Account)
				continue
			}
			if re.MatchString(name) {
				return fmt.Errorf("Matching blacklist %s. Not downloading %s", entry, name)
			}
		}
	}
	flog.Debugf("Trying to download %#v with size %#v", name, size)
	if int(size) > general.MediaDownloadSize {
		msg.Event = config.EVENT_FILE_FAILURE_SIZE
		msg.Extra[msg.Event] = append(msg.Extra[msg.Event], config.FileInfo{Name: name, Comment: msg.Text, Size: size})
		return fmt.Errorf("File %#v to large to download (%#v). MediaDownloadSize is %#v", name, size, general.MediaDownloadSize)
	}
	return nil
}

func HandleDownloadData(flog *log.Entry, msg *config.Message, name, comment, url string, data *[]byte, general *config.Protocol) {
	var avatar bool
	flog.Debugf("Download OK %#v %#v", name, len(*data))
	if msg.Event == config.EVENT_AVATAR_DOWNLOAD {
		avatar = true
	}
	msg.Extra["file"] = append(msg.Extra["file"], config.FileInfo{Name: name, Data: data, URL: url, Comment: comment, Avatar: avatar})
}

func RemoveEmptyNewLines(msg string) string {
	lines := ""
	for _, line := range strings.Split(msg, "\n") {
		if line != "" {
			lines += line + "\n"
		}
	}
	lines = strings.TrimRight(lines, "\n")
	return lines
}

func ClipMessage(text string, length int) string {
	// clip too long messages
	if len(text) > length {
		text = text[:length-len(" *message clipped*")]
		if r, size := utf8.DecodeLastRuneInString(text); r == utf8.RuneError {
			text = text[:len(text)-size]
		}
		text += " *message clipped*"
	}
	return text
}
