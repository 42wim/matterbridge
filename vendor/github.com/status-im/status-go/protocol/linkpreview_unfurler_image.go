package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"path"
	"regexp"

	"go.uber.org/zap"

	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

const (
	maxImageSize = 1024 * 350
)

var imageURLRegexp = regexp.MustCompile(`(?i)^.+(png|jpg|jpeg|webp)$`)

type ImageUnfurler struct {
	url        *neturl.URL
	logger     *zap.Logger
	httpClient *http.Client
}

func NewImageUnfurler(URL *neturl.URL, logger *zap.Logger, httpClient *http.Client) *ImageUnfurler {
	return &ImageUnfurler{
		url:        URL,
		logger:     logger,
		httpClient: httpClient,
	}
}

func compressImage(imgBytes []byte) ([]byte, error) {
	smallest := imgBytes

	img, err := images.DecodeImageData(imgBytes, bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}

	compressed := bytes.NewBuffer([]byte{})
	err = images.CompressToFileLimits(compressed, img, images.DefaultBounds)
	if err != nil {
		return nil, err
	}

	if len(compressed.Bytes()) < len(smallest) {
		smallest = compressed.Bytes()
	}

	if len(smallest) > maxImageSize {
		return nil, errors.New("image too large")
	}

	return smallest, nil
}

// IsSupportedImageURL detects whether a URL ends with one of the
// supported image extensions. It provides a quick way to identify whether URLs
// should be unfurled as images without needing to retrieve the full response
// body first.
func IsSupportedImageURL(url *neturl.URL) bool {
	return imageURLRegexp.MatchString(url.Path)
}

// isSupportedImage returns true when payload is one of the supported image
// types. In the future, we should differentiate between animated and
// non-animated WebP because, currently, only static WebP can be processed by
// functions in the status-go/images package.
func isSupportedImage(payload []byte) bool {
	return images.IsJpeg(payload) || images.IsPng(payload) || images.IsWebp(payload)
}

func (u *ImageUnfurler) Unfurl() (*common.LinkPreview, error) {
	preview := newDefaultLinkPreview(u.url)
	preview.Type = protobuf.UnfurledLink_IMAGE

	headers := map[string]string{"user-agent": headerUserAgent}
	imgBytes, err := fetchBody(u.logger, u.httpClient, u.url.String(), headers)
	if err != nil {
		return preview, err
	}

	if !isSupportedImage(imgBytes) {
		return preview, fmt.Errorf("unsupported image type url='%s'", u.url.String())
	}

	compressedBytes, err := compressImage(imgBytes)
	if err != nil {
		return preview, fmt.Errorf("failed to compress image url='%s': %w", u.url.String(), err)
	}

	width, height, err := images.GetImageDimensions(compressedBytes)
	if err != nil {
		return preview, fmt.Errorf("could not get image dimensions url='%s': %w", u.url.String(), err)
	}

	dataURI, err := images.GetPayloadDataURI(compressedBytes)
	if err != nil {
		return preview, fmt.Errorf("could not build data URI url='%s': %w", u.url.String(), err)
	}

	preview.Title = path.Base(u.url.Path)
	preview.Thumbnail.Width = width
	preview.Thumbnail.Height = height
	preview.Thumbnail.DataURI = dataURI

	return preview, nil
}
