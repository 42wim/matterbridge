package protocol

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/keighl/metabolize"
	"go.uber.org/zap"
	"golang.org/x/net/html"

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

func GetFavicon(bodyBytes []byte) string {
	htmlTokens := html.NewTokenizer(bytes.NewBuffer(bodyBytes))
loop:
	for {
		tt := htmlTokens.Next()
		switch tt {
		case html.ErrorToken:
			break loop
		case html.StartTagToken:
			t := htmlTokens.Token()
			if t.Data != "link" {
				continue
			}

			isIcon := false
			href := ""
			for _, attr := range t.Attr {
				k := attr.Key
				v := attr.Val
				if k == "rel" && (v == "icon" || v == "shortcut icon") {
					isIcon = true
				} else if k == "href" &&
					(strings.Contains(v, ".ico") ||
						strings.Contains(v, ".png") ||
						strings.Contains(v, ".svg")) {
					href = v
				}
			}

			if isIcon && href != "" {
				return href
			}
		}
	}
	return ""
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

	faviconPath := GetFavicon(bodyBytes)
	t, err := fetchImage(u.logger, u.httpClient, faviconPath, false)
	if err != nil {
		u.logger.Info("failed to fetch favicon", zap.String("url", u.url.String()), zap.Error(err))
	} else {
		preview.Favicon.DataURI = t.DataURI
	}
	// There are URLs like https://wikipedia.org/ that don't have an OpenGraph
	// title tag, but article pages do. In the future, we can fallback to the
	// website's title by using the <title> tag.
	if ogMetadata.Title == "" {
		return preview, fmt.Errorf("missing required title in OpenGraph response")
	}

	if ogMetadata.ThumbnailURL != "" {
		t, err := fetchImage(u.logger, u.httpClient, ogMetadata.ThumbnailURL, true)
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

func fetchImage(logger *zap.Logger, httpClient *http.Client, url string, getDimensions bool) (common.LinkPreviewThumbnail, error) {
	var thumbnail common.LinkPreviewThumbnail

	imgBytes, err := fetchBody(logger, httpClient, url, nil)
	if err != nil {
		return thumbnail, fmt.Errorf("could not fetch thumbnail url='%s': %w", url, err)
	}
	if getDimensions {
		width, height, err := images.GetImageDimensions(imgBytes)
		if err != nil {
			return thumbnail, fmt.Errorf("could not get image dimensions url='%s': %w", url, err)
		}
		thumbnail.Width = width
		thumbnail.Height = height
	}
	dataURI, err := images.GetPayloadDataURI(imgBytes)
	if err != nil {
		return thumbnail, fmt.Errorf("could not build data URI url='%s': %w", url, err)
	}
	thumbnail.DataURI = dataURI

	return thumbnail, nil
}
