package httptoo

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bradfitz/iter"

	"github.com/anacrolix/missinggo"
)

func OriginatingProtocol(r *http.Request) string {
	if fp := r.Header.Get("X-Forwarded-Proto"); fp != "" {
		return fp
	} else if r.TLS != nil {
		return "https"
	} else {
		return "http"
	}
}

// Clears the named cookie for every domain that leads to the current one.
func NukeCookie(w http.ResponseWriter, r *http.Request, name, path string) {
	parts := strings.Split(missinggo.SplitHostMaybePort(r.Host).Host, ".")
	for i := range iter.N(len(parts) + 1) { // Include the empty domain.
		http.SetCookie(w, &http.Cookie{
			Name:   name,
			MaxAge: -1,
			Path:   path,
			Domain: strings.Join(parts[i:], "."),
		})
	}
}

// Performs quoted-string from http://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html
func EncodeQuotedString(s string) string {
	return strconv.Quote(s)
}

// https://httpstatuses.com/499
const StatusClientCancelledRequest = 499
