// Package besticon includes functions
// finding icons for a given web site.
package besticon

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"image/color"

	// Load supported image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mat/besticon/ico"

	"github.com/mat/besticon/colorfinder"

	"golang.org/x/net/html/charset"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

var defaultFormats []string

const MinIconSize = 0

// TODO: Turn into env var: https://github.com/rendomnet/besticon/commit/c85867cc80c00c898053ce8daf40d51a93b9d39f#diff-37b57e3fdbe4246771791e86deb4d69dL41
const MaxIconSize = 500

// Icon holds icon information.
type Icon struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	Bytes     int    `json:"bytes"`
	Error     error  `json:"error"`
	Sha1sum   string `json:"sha1sum"`
	ImageData []byte `json:",omitempty"`
}

type IconFinder struct {
	FormatsAllowed  []string
	HostOnlyDomains []string
	KeepImageBytes  bool
	icons           []Icon
}

func (f *IconFinder) FetchIcons(url string) ([]Icon, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = "http://" + url
	}

	url = f.stripIfNecessary(url)

	var err error

	if CacheEnabled() {
		f.icons, err = resultFromCache(url)
	} else {
		f.icons, err = fetchIcons(url)
	}

	return f.Icons(), err
}

// stripIfNecessary removes everything from URL but the Scheme and Host
// part if URL.Host is found in HostOnlyDomains.
// This can be used for very popular domains like youtube.com where throttling is
// an issue.
func (f *IconFinder) stripIfNecessary(URL string) string {
	u, e := url.Parse(URL)
	if e != nil {
		return URL
	}

	for _, h := range f.HostOnlyDomains {
		if h == u.Host || h == "*" {
			domainOnlyURL := url.URL{Scheme: u.Scheme, Host: u.Host}
			return domainOnlyURL.String()
		}
	}

	return URL
}

func (f *IconFinder) IconInSizeRange(r SizeRange) *Icon {
	icons := f.Icons()

	// 1. SVG always wins
	for _, ico := range icons {
		if ico.Format == "svg" {
			return &ico
		}
	}

	// 2. Try to return smallest in range perfect..max
	sortIcons(icons, false)
	for _, ico := range icons {
		if (ico.Width >= r.Perfect && ico.Height >= r.Perfect) && (ico.Width <= r.Max && ico.Height <= r.Max) {
			return &ico
		}
	}

	// 3. Try to return biggest in range perfect..min
	sortIcons(icons, true)
	for _, ico := range icons {
		if (ico.Width >= r.Min && ico.Height >= r.Min) && (ico.Width <= r.Perfect && ico.Height <= r.Perfect) {
			return &ico
		}
	}

	return nil
}

func (f *IconFinder) MainColorForIcons() *color.RGBA {
	return MainColorForIcons(f.icons)
}

func (f *IconFinder) Icons() []Icon {
	return discardUnwantedFormats(f.icons, f.FormatsAllowed)
}

func (ico *Icon) Image() (*image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(ico.ImageData))
	return &img, err
}

func discardUnwantedFormats(icons []Icon, wantedFormats []string) []Icon {
	formats := defaultFormats
	if len(wantedFormats) > 0 {
		formats = wantedFormats
	}

	return filterIcons(icons, func(ico Icon) bool {
		return includesString(formats, ico.Format)
	})
}

type iconPredicate func(Icon) bool

func filterIcons(icons []Icon, pred iconPredicate) []Icon {
	var result []Icon
	for _, ico := range icons {
		if pred(ico) {
			result = append(result, ico)
		}
	}
	return result
}

func includesString(arr []string, str string) bool {
	for _, e := range arr {
		if e == str {
			return true
		}
	}
	return false
}

func fetchIcons(siteURL string) ([]Icon, error) {
	var links []string

	html, urlAfterRedirect, e := fetchHTML(siteURL)
	if e == nil {
		// Search HTML for icons
		links, e = findIconLinks(urlAfterRedirect, html)
		if e != nil {
			return nil, e
		}
	} else {
		// Unable to fetch the response or got a bad HTTP status code. Try default
		// icon paths. https://github.com/mat/besticon/discussions/47
		links, e = defaultIconURLs(siteURL)
		if e != nil {
			return nil, e
		}
	}

	icons := fetchAllIcons(links)
	icons = rejectBrokenIcons(icons)
	sortIcons(icons, true)

	return icons, nil
}

const maxResponseBodySize = 10485760 // 10MB

func fetchHTML(url string) ([]byte, *url.URL, error) {
	r, e := Get(url)
	if e != nil {
		return nil, nil, e
	}

	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return nil, nil, errors.New("besticon: not found")
	}

	b, e := GetBodyBytes(r)
	if e != nil {
		return nil, nil, e
	}
	if len(b) == 0 {
		return nil, nil, errors.New("besticon: empty response")
	}

	reader := bytes.NewReader(b)
	contentType := r.Header.Get("Content-Type")
	utf8reader, e := charset.NewReader(reader, contentType)
	if e != nil {
		return nil, nil, e
	}
	utf8bytes, e := ioutil.ReadAll(utf8reader)
	if e != nil {
		return nil, nil, e
	}

	return utf8bytes, r.Request.URL, nil
}

func MainColorForIcons(icons []Icon) *color.RGBA {
	if len(icons) == 0 {
		return nil
	}

	var icon *Icon
	// Prefer gif, jpg, png
	for _, ico := range icons {
		if ico.Format == "gif" || ico.Format == "jpg" || ico.Format == "png" {
			icon = &ico
			break
		}
	}
	// Try .ico else
	if icon == nil {
		for _, ico := range icons {
			if ico.Format == "ico" {
				icon = &ico
				break
			}
		}
	}

	if icon == nil {
		return nil
	}

	img, err := icon.Image()
	if err != nil {
		return nil
	}

	cf := colorfinder.ColorFinder{}
	mainColor, err := cf.FindMainColor(*img)
	if err != nil {
		return nil
	}

	return &mainColor
}

// Construct default icon URLs. A fallback if we can't fetch the HTML.
func defaultIconURLs(siteURL string) ([]string, error) {
	baseURL, e := url.Parse(siteURL)
	if e != nil {
		return nil, e
	}

	var links []string
	for _, path := range iconPaths {
		absoluteURL, e := absoluteURL(baseURL, path)
		if e != nil {
			return nil, e
		}
		links = append(links, absoluteURL)
	}

	return links, nil
}

func fetchAllIcons(urls []string) []Icon {
	ch := make(chan Icon)

	for _, u := range urls {
		go func(u string) { ch <- fetchIconDetails(u) }(u)
	}

	var icons []Icon
	for range urls {
		icon := <-ch
		icons = append(icons, icon)
	}
	return icons
}

func fetchIconDetails(url string) Icon {
	i := Icon{URL: url}

	response, e := Get(url)
	if e != nil {
		i.Error = e
		return i
	}

	b, e := GetBodyBytes(response)
	if e != nil {
		i.Error = e
		return i
	}

	if isSVG(b) {
		// Special handling for svg, which golang can't decode with
		// image.DecodeConfig. Fill in an absurdly large width/height so SVG always
		// wins size contests.
		i.Format = "svg"
		i.Width = 9999
		i.Height = 9999
	} else {
		cfg, format, e := image.DecodeConfig(bytes.NewReader(b))
		if e != nil {
			i.Error = fmt.Errorf("besticon: unknown image format: %s", e)
			return i
		}

		// jpeg => jpg
		if format == "jpeg" {
			format = "jpg"
		}

		i.Width = cfg.Width
		i.Height = cfg.Height
		i.Format = format
	}

	i.Bytes = len(b)
	i.Sha1sum = sha1Sum(b)
	if keepImageBytes {
		i.ImageData = b
	}

	return i
}

// SVG detector. We can't use image.RegisterFormat, since RegisterFormat is
// limited to a simple magic number check. It's easy to confuse the first few
// bytes of HTML with SVG.
func isSVG(body []byte) bool {
	// is it long enough?
	if len(body) < 10 {
		return false
	}

	// does it start with something reasonable?
	switch {
	case bytes.Equal(body[0:2], []byte("<!")):
	case bytes.Equal(body[0:2], []byte("<?")):
	case bytes.Equal(body[0:4], []byte("<svg")):
	default:
		return false
	}

	// is there an <svg in the first 300 bytes?
	if off := bytes.Index(body, []byte("<svg")); off == -1 || off > 300 {
		return false
	}

	return true
}

func Get(urlstring string) (*http.Response, error) {
	u, e := url.Parse(urlstring)
	if e != nil {
		return nil, e
	}
	// Maybe we can get rid of this conversion someday
	// https://github.com/golang/go/issues/13835
	u.Host, e = idna.ToASCII(u.Host)
	if e != nil {
		return nil, e
	}

	req, e := http.NewRequest("GET", u.String(), nil)
	if e != nil {
		return nil, e
	}

	setDefaultHeaders(req)

	start := time.Now()
	resp, err := client.Do(req)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Printf("Error: %s %s %s %.2fms",
			req.Method,
			req.URL,
			err,
			float64(duration)/float64(time.Millisecond),
		)
	} else {
		logger.Printf("%s %s %d %.2fms %d",
			req.Method,
			req.URL,
			resp.StatusCode,
			float64(duration)/float64(time.Millisecond),
			resp.ContentLength,
		)
	}

	return resp, err
}

func GetBodyBytes(r *http.Response) ([]byte, error) {
	limitReader := io.LimitReader(r.Body, maxResponseBodySize)
	b, e := ioutil.ReadAll(limitReader)
	r.Body.Close()

	if len(b) >= maxResponseBodySize {
		return nil, errors.New("body too large")
	}
	return b, e
}

func setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", getenvOrFallback("HTTP_USER_AGENT", "Mozilla/5.0 (iPhone; CPU iPhone OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A5297c Safari/602.1"))
}

func mustInitCookieJar() *cookiejar.Jar {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, e := cookiejar.New(&options)
	if e != nil {
		panic(e)
	}

	return jar
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	setDefaultHeaders(req)

	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

func absoluteURL(baseURL *url.URL, path string) (string, error) {
	u, e := url.Parse(path)
	if e != nil {
		return "", e
	}

	u.Scheme = baseURL.Scheme
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	if u.Host == "" {
		u.Host = baseURL.Host
	}
	return baseURL.ResolveReference(u).String(), nil
}

func urlFromBase(baseURL *url.URL, path string) string {
	u := *baseURL
	u.Path = path
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	return u.String()
}

func rejectBrokenIcons(icons []Icon) []Icon {
	var result []Icon
	for _, img := range icons {
		if img.Error == nil && (img.Width > 1 && img.Height > 1) {
			result = append(result, img)
		}
	}
	return result
}

func sha1Sum(b []byte) string {
	hash := sha1.New()
	hash.Write(b)
	bs := hash.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

var client *http.Client
var keepImageBytes bool

func init() {
	duration, e := time.ParseDuration(getenvOrFallback("HTTP_CLIENT_TIMEOUT", "5s"))
	if e != nil {
		panic(e)
	}
	setHTTPClient(&http.Client{Timeout: duration})

	// see
	// https://github.com/mat/besticon/pull/52/commits/208e9dcbdbdeb7ef7491bb42f1bc449e87e084a2
	// when we are ready to add support for the FORMATS env variable

	defaultFormats = []string{"gif", "ico", "jpg", "png"}
}

func setHTTPClient(c *http.Client) {
	c.Jar = mustInitCookieJar()
	c.CheckRedirect = checkRedirect
	client = c
}

var logger *log.Logger

// SetLogOutput sets the output for the package's logger.
func SetLogOutput(w io.Writer) {
	logger = log.New(w, "http:  ", log.LstdFlags|log.Lmicroseconds)
}

func init() {
	SetLogOutput(os.Stdout)
	keepImageBytes = true
}

func getenvOrFallback(key string, fallbackValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if len(value) != 0 {
		return value
	}
	return fallbackValue
}

func getenvOrFallbackArray(key string, fallbackValue []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if len(value) != 0 {
		return strings.Split(value, ",")
	}
	return fallbackValue
}

var BuildDate string // set via ldflags on Make
