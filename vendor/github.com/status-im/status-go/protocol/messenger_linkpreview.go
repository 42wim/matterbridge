package protocol

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/net/publicsuffix"

	"github.com/status-im/markdown"

	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/protocol/common"
)

const UnfurledLinksPerMessageLimit = 5

type URLUnfurlPermission int

const (
	URLUnfurlingAllowed URLUnfurlPermission = iota
	URLUnfurlingAskUser
	URLUnfurlingForbiddenBySettings
	URLUnfurlingNotSupported
)

type URLUnfurlingMetadata struct {
	URL               string              `json:"url"`
	Permission        URLUnfurlPermission `json:"permission"`
	IsStatusSharedURL bool                `json:"isStatusSharedURL"`
}

type URLsUnfurlPlan struct {
	URLs []URLUnfurlingMetadata `json:"urls"`
}

func URLUnfurlingSupported(url string) bool {
	return !strings.HasSuffix(url, ".gif")
}

type UnfurlURLsResponse struct {
	LinkPreviews       []*common.LinkPreview       `json:"linkPreviews,omitempty"`
	StatusLinkPreviews []*common.StatusLinkPreview `json:"statusLinkPreviews,omitempty"`
}

func normalizeHostname(hostname string) string {
	hostname = strings.ToLower(hostname)
	re := regexp.MustCompile(`^www\.(.*)$`)
	return re.ReplaceAllString(hostname, "$1")
}

func (m *Messenger) newURLUnfurler(httpClient *http.Client, url *neturl.URL) Unfurler {

	if IsSupportedImageURL(url) {
		return NewImageUnfurler(
			url,
			m.logger,
			httpClient)
	}

	switch normalizeHostname(url.Hostname()) {
	case "reddit.com":
		return NewOEmbedUnfurler(
			"https://www.reddit.com/oembed",
			url,
			m.logger,
			httpClient)
	default:
		return NewOpenGraphUnfurler(
			url,
			m.logger,
			httpClient)
	}
}

func (m *Messenger) unfurlURL(httpClient *http.Client, url string) (*common.LinkPreview, error) {
	preview := new(common.LinkPreview)

	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return preview, err
	}

	unfurler := m.newURLUnfurler(httpClient, parsedURL)
	preview, err = unfurler.Unfurl()
	if err != nil {
		return preview, err
	}
	preview.Hostname = strings.ToLower(parsedURL.Hostname())

	return preview, nil
}

// parseValidURL is a stricter version of url.Parse that performs additional
// checks to ensure the URL is valid for clients to request a link preview.
func parseValidURL(rawURL string) (*neturl.URL, error) {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL failed: %w", err)
	}

	if u.Scheme == "" {
		return nil, errors.New("missing URL scheme")
	}

	_, err = publicsuffix.EffectiveTLDPlusOne(u.Hostname())
	if err != nil {
		return nil, fmt.Errorf("missing known URL domain: %w", err)
	}

	return u, nil
}

func (m *Messenger) GetTextURLsToUnfurl(text string) *URLsUnfurlPlan {
	s, err := m.getSettings()
	if err != nil {
		// log the error and keep parsing the text
		m.logger.Error("GetTextURLsToUnfurl: failed to get settings", zap.Error(err))
		s.URLUnfurlingMode = settings.URLUnfurlingDisableAll
	}

	indexedUrls := map[string]struct{}{}
	result := &URLsUnfurlPlan{
		// The usage of `UnfurledLinksPerMessageLimit` is quite random here. I wanted to allocate
		// some not-zero place here, using the limit number is at least some binding.
		URLs: make([]URLUnfurlingMetadata, 0, UnfurledLinksPerMessageLimit),
	}
	parsedText := markdown.Parse([]byte(text), nil)
	visitor := common.RunLinksVisitor(parsedText)

	for _, rawURL := range visitor.Links {
		parsedURL, err := parseValidURL(rawURL)
		if err != nil {
			continue
		}
		// Lowercase the host so the URL can be used as a cache key. Particularly on
		// mobile clients it is common that the first character in a text input is
		// automatically uppercased. In WhatsApp they incorrectly lowercase the
		// URL's path, but this is incorrect. For instance, some URL shorteners are
		// case-sensitive, some websites encode base64 in the path, etc.
		parsedURL.Host = strings.ToLower(parsedURL.Host)

		url := parsedURL.String()
		url = strings.TrimRight(url, "/") // Removes the spurious trailing forward slash.
		if _, exists := indexedUrls[url]; exists {
			continue
		}

		metadata := URLUnfurlingMetadata{
			URL:               url,
			IsStatusSharedURL: IsStatusSharedURL(url),
		}

		if !URLUnfurlingSupported(rawURL) {
			metadata.Permission = URLUnfurlingNotSupported
		} else if metadata.IsStatusSharedURL {
			metadata.Permission = URLUnfurlingAllowed
		} else {
			switch s.URLUnfurlingMode {
			case settings.URLUnfurlingAlwaysAsk:
				metadata.Permission = URLUnfurlingAskUser
			case settings.URLUnfurlingEnableAll:
				metadata.Permission = URLUnfurlingAllowed
			case settings.URLUnfurlingDisableAll:
				metadata.Permission = URLUnfurlingForbiddenBySettings
			default:
				metadata.Permission = URLUnfurlingForbiddenBySettings
			}
		}

		result.URLs = append(result.URLs, metadata)
	}

	return result
}

// Deprecated: GetURLs is deprecated in favor of more generic GetTextURLsToUnfurl.
//
// This is a wrapper around GetTextURLsToUnfurl that returns the list of URLs found in the text
// without any additional information.
func (m *Messenger) GetURLs(text string) []string {
	plan := m.GetTextURLsToUnfurl(text)
	limit := int(math.Min(UnfurledLinksPerMessageLimit, float64(len(plan.URLs))))
	urls := make([]string, 0, limit)
	for _, metadata := range plan.URLs {
		urls = append(urls, metadata.URL)
		if len(urls) == limit {
			break
		}
	}
	return urls
}

func NewDefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: DefaultRequestTimeout}
}

// UnfurlURLs assumes clients pass URLs verbatim that were validated and
// processed by GetURLs.
func (m *Messenger) UnfurlURLs(httpClient *http.Client, urls []string) (UnfurlURLsResponse, error) {
	response := UnfurlURLsResponse{}

	// Unfurl in a loop

	response.LinkPreviews = make([]*common.LinkPreview, 0, len(urls))
	response.StatusLinkPreviews = make([]*common.StatusLinkPreview, 0, len(urls))

	if httpClient == nil {
		httpClient = NewDefaultHTTPClient()
	}

	for _, url := range urls {
		m.logger.Debug("unfurling", zap.String("url", url))

		if IsStatusSharedURL(url) {
			unfurler := NewStatusUnfurler(url, m, m.logger)
			preview, err := unfurler.Unfurl()
			if err != nil {
				m.logger.Warn("failed to unfurl status link", zap.String("url", url), zap.Error(err))
				continue
			}
			response.StatusLinkPreviews = append(response.StatusLinkPreviews, preview)
			continue
		}

		p, err := m.unfurlURL(httpClient, url)
		if err != nil {
			m.logger.Warn("failed to unfurl", zap.String("url", url), zap.Error(err))
			continue
		}
		response.LinkPreviews = append(response.LinkPreviews, p)
	}

	return response, nil
}
