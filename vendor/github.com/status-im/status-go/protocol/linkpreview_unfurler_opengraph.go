package protocol

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"

	"github.com/keighl/metabolize"
	"go.uber.org/zap"

	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

type OpenGraphMetadata struct {
	Title        string `json:"title" meta:"og:title"`
	Description  string `json:"description" meta:"og:description"`
	ThumbnailURL string `json:"thumbnailUrl" meta:"og:image"`
}

// OpenGraphUnfurler should be preferred over OEmbedUnfurler because oEmbed
// gives back a JSON response with a "html" field that's supposed to be embedded
// in an iframe (hardly useful for existing Status' clients).
type OpenGraphUnfurler struct {
	url        *neturl.URL
	logger     *zap.Logger
	httpClient *http.Client
}

func NewOpenGraphUnfurler(URL *neturl.URL, logger *zap.Logger, httpClient *http.Client) *OpenGraphUnfurler {
	return &OpenGraphUnfurler{
		url:        URL,
		logger:     logger,
		httpClient: httpClient,
	}
}

func (u *OpenGraphUnfurler) Unfurl() (*common.LinkPreview, error) {
	preview := newDefaultLinkPreview(u.url)
	preview.Type = protobuf.UnfurledLink_LINK

	headers := map[string]string{
		"accept":          headerAcceptText,
		"accept-language": headerAcceptLanguage,
		"user-agent":      headerUserAgent,
	}
	bodyBytes, err := fetchBody(u.logger, u.httpClient, u.url.String(), headers)
	if err != nil {
		return preview, err
	}

	var ogMetadata OpenGraphMetadata
	err = metabolize.Metabolize(ioutil.NopCloser(bytes.NewBuffer(bodyBytes)), &ogMetadata)
	if err != nil {
		return preview, fmt.Errorf("failed to parse OpenGraph data")
	}

	// There are URLs like https://wikipedia.org/ that don't have an OpenGraph
	// title tag, but article pages do. In the future, we can fallback to the
	// website's title by using the <title> tag.
	if ogMetadata.Title == "" {
		return preview, fmt.Errorf("missing required title in OpenGraph response")
	}

	if ogMetadata.ThumbnailURL != "" {
		t, err := fetchThumbnail(u.logger, u.httpClient, ogMetadata.ThumbnailURL)
		if err != nil {
			// Given we want to fetch thumbnails on a best-effort basis, if an error
			// happens we simply log it.
			u.logger.Info("failed to fetch thumbnail", zap.String("url", u.url.String()), zap.Error(err))
		} else {
			preview.Thumbnail = t
		}
	}

	preview.Title = ogMetadata.Title
	preview.Description = ogMetadata.Description
	return preview, nil
}

func fetchThumbnail(logger *zap.Logger, httpClient *http.Client, url string) (common.LinkPreviewThumbnail, error) {
	var thumbnail common.LinkPreviewThumbnail

	imgBytes, err := fetchBody(logger, httpClient, url, nil)
	if err != nil {
		return thumbnail, fmt.Errorf("could not fetch thumbnail url='%s': %w", url, err)
	}

	width, height, err := images.GetImageDimensions(imgBytes)
	if err != nil {
		return thumbnail, fmt.Errorf("could not get image dimensions url='%s': %w", url, err)
	}
	thumbnail.Width = width
	thumbnail.Height = height

	dataURI, err := images.GetPayloadDataURI(imgBytes)
	if err != nil {
		return thumbnail, fmt.Errorf("could not build data URI url='%s': %w", url, err)
	}
	thumbnail.DataURI = dataURI

	return thumbnail, nil
}
