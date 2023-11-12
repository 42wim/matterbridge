package protocol

import (
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"

	"go.uber.org/zap"

	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

type OEmbedUnfurler struct {
	logger     *zap.Logger
	httpClient *http.Client
	// oembedEndpoint describes where the consumer may request representations for
	// the supported URL scheme. For example, for YouTube, it is
	// https://www.youtube.com/oembed.
	oembedEndpoint string
	// url is the actual URL to be unfurled.
	url *neturl.URL
}

func NewOEmbedUnfurler(oembedEndpoint string,
	url *neturl.URL,
	logger *zap.Logger,
	httpClient *http.Client) *OEmbedUnfurler {
	return &OEmbedUnfurler{
		oembedEndpoint: oembedEndpoint,
		url:            url,
		logger:         logger,
		httpClient:     httpClient,
	}
}

type OEmbedResponse struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
}

func (u *OEmbedUnfurler) newOEmbedURL() (*neturl.URL, error) {
	oembedURL, err := neturl.Parse(u.oembedEndpoint)
	if err != nil {
		return nil, err
	}

	// When format is specified, the provider MUST return data in the requested
	// format, else return an error.
	oembedURL.RawQuery = neturl.Values{
		"url":    {u.url.String()},
		"format": {"json"},
	}.Encode()

	return oembedURL, nil
}

func (u OEmbedUnfurler) Unfurl() (*common.LinkPreview, error) {
	preview := newDefaultLinkPreview(u.url)
	preview.Type = protobuf.UnfurledLink_LINK

	oembedURL, err := u.newOEmbedURL()
	if err != nil {
		return preview, err
	}

	headers := map[string]string{
		"accept":          headerAcceptJSON,
		"accept-language": headerAcceptLanguage,
		"user-agent":      headerUserAgent,
	}
	oembedBytes, err := fetchBody(u.logger, u.httpClient, oembedURL.String(), headers)
	if err != nil {
		return preview, err
	}

	var oembedResponse OEmbedResponse
	if err != nil {
		return preview, err
	}
	err = json.Unmarshal(oembedBytes, &oembedResponse)
	if err != nil {
		return preview, err
	}

	if oembedResponse.Title == "" {
		return preview, fmt.Errorf("missing required title in oEmbed response")
	}

	preview.Title = oembedResponse.Title
	return preview, nil
}
