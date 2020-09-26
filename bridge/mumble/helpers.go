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

func (b *Bmumble) decodeImage(uri string, parts *[]MessagePart) error {
	// Decode the data:image/... URI
	image, err := dataurl.DecodeString(uri)
	if err != nil {
		b.Log.WithError(err).Info("No image extracted")
		return err
	}
	// Determine the file extensions for that image
	ext, err := mime.ExtensionsByType(image.MediaType.ContentType())
	if err != nil || len(ext) == 0 {
		b.Log.WithError(err).Infof("No file extension registered for MIME type '%s'", image.MediaType.ContentType())
		return err
	}
	// Add the image to the MessagePart slice
	*parts = append(*parts, MessagePart{"", ext[0], image.Data})
	return nil
}

func (b *Bmumble) tokenize(t *string) ([]MessagePart, error) {
	// `^(.*?)` matches everything before the image
	// `!\[[^\]]*\]\(` matches the `![alt](` part of markdown images
	// `(data:image\/[^)]+)` matches the data: URI used by Mumble
	// `\)` matches the closing parenthesis after the URI
	// `(.*)$` matches the remaining text to be examined in the next iteration
	p := regexp.MustCompile(`^(.*?)!\[[^\]]*\]\((data:image\/[^)]+)\)(.*)$`)
	remaining := *t
	var parts []MessagePart
	for {
		tokens := p.FindStringSubmatch(remaining)

		if tokens == nil {
			// no match -> remaining string is non-image text
			if len(remaining) > 0 {
				parts = append(parts, MessagePart{remaining, "", nil})
			}
			return parts, nil
		}
		// tokens[1] is the text before the image
		if len(tokens[1]) > 0 {
			parts = append(parts, MessagePart{tokens[1], "", nil})
		}
		// tokens[2] is the image URL
		uri, err := dataurl.UnescapeToString(strings.ReplaceAll(tokens[2], " ", ""))
		if err != nil {
			b.Log.WithError(err).Info("URL unescaping failed")
			remaining = tokens[3]
			continue
		}
		err = b.decodeImage(uri, &parts)
		if err != nil {
			b.Log.WithError(err).Info("Decoding the image failed")
		}
		// tokens[3] is the text after the image, processed in the next iteration
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
	// Create a separate message for each file
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		imsg := config.Message{
			Channel:   msg.Channel,
			Username:  msg.Username,
			UserID:    msg.UserID,
			Account:   msg.Account,
			Protocol:  msg.Protocol,
			Timestamp: msg.Timestamp,
			Event:     "mumble_image",
		}
		// If no data is present for the file, send a link instead
		if fi.Data == nil || len(*fi.Data) == 0 {
			if len(fi.URL) > 0 {
				imsg.Text = fmt.Sprintf(`<a href="%s">%s</a>`, fi.URL, fi.URL)
				messages = append(messages, imsg)
			} else {
				b.Log.Infof("Not forwarding file without local data")
			}
			continue
		}
		mimeType := http.DetectContentType(*fi.Data)
		// Mumble only supports images natively, send a link instead
		if !strings.HasPrefix(mimeType, "image/") {
			if len(fi.URL) > 0 {
				imsg.Text = fmt.Sprintf(`<a href="%s">%s</a>`, fi.URL, fi.URL)
				messages = append(messages, imsg)
			} else {
				b.Log.Infof("Not forwarding file of type %s", mimeType)
			}
			continue
		}
		mimeType = strings.TrimSpace(strings.Split(mimeType, ";")[0])
		// Build data:image/...;base64,... style image URL and embed image directly into the message
		du := dataurl.New(*fi.Data, mimeType)
		url, err := du.MarshalText()
		if err != nil {
			b.Log.WithError(err).Infof("Image Serialization into data URL failed (type: %s, length: %d)", mimeType, len(*fi.Data))
			continue
		}
		imsg.Text = fmt.Sprintf(`<img src="%s"/>`, url)
		messages = append(messages, imsg)
	}
	// Remove files from original message
	msg.Extra["file"] = nil
	return messages
}
