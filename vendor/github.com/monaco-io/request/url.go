package request

import "net/url"

type requestURL struct {
	httpURL    *url.URL
	urlString  string
	parameters map[string]string
}

// EncodeURL add and encoded parameters.
func (ru *requestURL) EncodeURL() (err error) {
	ru.httpURL, err = url.Parse(ru.urlString)
	if err != nil {
		return
	}
	query := ru.httpURL.Query()
	for k := range ru.parameters {
		query.Set(k, ru.parameters[k])
	}
	ru.httpURL.RawQuery = query.Encode()
	return
}

// String return example: https://www.google.com/search?a=1&b=2
func (ru requestURL) string() string {
	return ru.httpURL.String()
}

func (ru requestURL) scheme() string {
	return ru.httpURL.Scheme
}

func (ru requestURL) host() string {
	return ru.httpURL.Host
}
