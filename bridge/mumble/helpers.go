package bmumble

import (
	"fmt"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/vincent-petithory/dataurl"
)

type MessagePart struct {
	Text          string
	FileExtension string
	Image         []byte
}

func (b *Bmumble) tokenize(t *string) ([]MessagePart, error) {
	// `^(.*?)` matches everyting before the image
	// `!\[[^\]]*\]\(` matches the `![alt](` part of markdown images
	// `(data:image\/[^)]+)` matches the data: URI used by Mumble
	// `\)` matches the closing parenthesis after the URI
	// `(.*)$` matches the remaining text to be examined in the next iteration
	p, err := regexp.Compile(`^(.*?)!\[[^\]]*\]\((data:image\/[^)]+)\)(.*)$`)
	if err != nil {
		return nil, err
	}
	remaining := *t
	var parts []MessagePart
	for {
		tokens := p.FindStringSubmatch(remaining)

		if tokens == nil {
			b.Log.Debugf("Last text token: %s", remaining)
			// no match -> remaining string is non-image text
			if len(remaining) > 0 {
				parts = append(parts, MessagePart{remaining, "", nil})
			}
			return parts, nil
		}
		if len(tokens[1]) > 0 {
			parts = append(parts, MessagePart{tokens[1], "", nil})
		}
		uri, err := dataurl.UnescapeToString(strings.ReplaceAll(tokens[2], " ", ""))
		if err != nil {
			b.Log.WithError(err).Info("URL unescaping failed")
		} else {
			b.Log.Debugf("Raw data: URL: %s", uri)
			image, err := dataurl.DecodeString(uri)
			if err == nil {
				ext, err := mime.ExtensionsByType(image.MediaType.ContentType())
				if ext != nil && len(ext) > 0 {
					parts = append(parts, MessagePart{"", ext[0], image.Data})
				} else {
					b.Log.WithError(err).Infof("No file extension registered for MIME type '%s'", image.MediaType.ContentType())
				}
			} else {
				b.Log.WithError(err).Info("No image extracted")
			}
		}
		remaining = tokens[3]
	}
}

func (b *Bmumble) convertHTMLtoMarkdown(html string) ([]MessagePart, error) {
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return nil, err
	}
	return b.tokenize(&markdown)
}

func (b *Bmumble) extractFiles(msg *config.Message) []config.Message {
	var messages []config.Message
	if msg.Extra == nil || len(msg.Extra["file"]) == 0 {
		return messages
	}
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if fi.Data == nil || len(*fi.Data) == 0 {
			// Mumble needs the raw data
			b.Log.Info("Not forwarding file without local data")
			continue
		}
		mimeType := http.DetectContentType(*fi.Data)
		if !strings.HasPrefix(mimeType, "image/") {
			// Mumble only supports images
			b.Log.Infof("Not forwarding file of type %s", mimeType)
			continue
		}
		mimeType = strings.TrimSpace(strings.Split(mimeType, ";")[0])
		// Build image message
		du := dataurl.New(*fi.Data, mimeType)
		url, err := du.MarshalText()
		if err != nil {
			b.Log.WithError(err).Infof("Image Serialization into data URL failed (type: %s, length: %d)", mimeType, len(*fi.Data))
			continue
		}
		imsg := config.Message{
			Text:      fmt.Sprintf(`<img src="%s"/>`, url),
			Channel:   msg.Channel,
			Username:  msg.Username,
			UserID:    msg.UserID,
			Account:   msg.Account,
			Protocol:  msg.Protocol,
			Timestamp: msg.Timestamp,
			Extra:     make(map[string][]interface{}),
			Event:     "mumble_image",
		}
		messages = append(messages, imsg)
	}
	// Remove files from original message
	msg.Extra["file"] = nil
	return messages
}
