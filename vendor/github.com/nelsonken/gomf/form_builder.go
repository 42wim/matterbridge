package gomf

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type FormBuilder struct {
	w *multipart.Writer
	b *bytes.Buffer
}

func New() *FormBuilder {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	return &FormBuilder{
		w: writer,
		b: buf,
	}
}

func (ufw *FormBuilder) WriteField(name, value string) error {
	w, err := ufw.w.CreateFormField(name)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		return err
	}

	return nil
}

// WriteFile if contentType is empty-string, will auto convert to application/octet-stream
func (ufw *FormBuilder) WriteFile(fieldName, fileName, contentType string, content []byte) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	wx, err := ufw.w.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{
			contentType,
		},
		"Content-Disposition": []string{
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName),
		},
	})
	if err != nil {
		return err
	}

	_, err = wx.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func (fb *FormBuilder) Close() error {
	return fb.w.Close()
}

func (fb *FormBuilder) GetBuffer() *bytes.Buffer {
	return fb.b
}

func (fb *FormBuilder) GetHTTPRequest(ctx context.Context, reqURL string) (*http.Request, error) {
	err := fb.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", reqURL, fb.b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", fb.w.FormDataContentType())
	req = req.WithContext(ctx)

	return req, nil
}
