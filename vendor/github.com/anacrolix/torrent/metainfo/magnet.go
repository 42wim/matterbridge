package metainfo

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Magnet link components.
type Magnet struct {
	InfoHash    Hash       // Expected in this implementation
	Trackers    []string   // "tr" values
	DisplayName string     // "dn" value, if not empty
	Params      url.Values // All other values, such as "x.pe", "as", "xs" etc.
}

const xtPrefix = "urn:btih:"

func (m Magnet) String() string {
	// Deep-copy m.Params
	vs := make(url.Values, len(m.Params)+len(m.Trackers)+2)
	for k, v := range m.Params {
		vs[k] = append([]string(nil), v...)
	}

	for _, tr := range m.Trackers {
		vs.Add("tr", tr)
	}
	if m.DisplayName != "" {
		vs.Add("dn", m.DisplayName)
	}

	// Transmission and Deluge both expect "urn:btih:" to be unescaped. Deluge wants it to be at the
	// start of the magnet link. The InfoHash field is expected to be BitTorrent in this
	// implementation.
	u := url.URL{
		Scheme:   "magnet",
		RawQuery: "xt=" + xtPrefix + m.InfoHash.HexString(),
	}
	if len(vs) != 0 {
		u.RawQuery += "&" + vs.Encode()
	}
	return u.String()
}

// Deprecated: Use ParseMagnetUri.
var ParseMagnetURI = ParseMagnetUri

// ParseMagnetUri parses Magnet-formatted URIs into a Magnet instance
func ParseMagnetUri(uri string) (m Magnet, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("error parsing uri: %w", err)
		return
	}
	if u.Scheme != "magnet" {
		err = fmt.Errorf("unexpected scheme %q", u.Scheme)
		return
	}
	q := u.Query()
	xt := q.Get("xt")
	m.InfoHash, err = parseInfohash(q.Get("xt"))
	if err != nil {
		err = fmt.Errorf("error parsing infohash %q: %w", xt, err)
		return
	}
	dropFirst(q, "xt")
	m.DisplayName = q.Get("dn")
	dropFirst(q, "dn")
	m.Trackers = q["tr"]
	delete(q, "tr")
	if len(q) == 0 {
		q = nil
	}
	m.Params = q
	return
}

func parseInfohash(xt string) (ih Hash, err error) {
	if !strings.HasPrefix(xt, xtPrefix) {
		err = errors.New("bad xt parameter prefix")
		return
	}
	encoded := xt[len(xtPrefix):]
	decode := func() func(dst, src []byte) (int, error) {
		switch len(encoded) {
		case 40:
			return hex.Decode
		case 32:
			return base32.StdEncoding.Decode
		}
		return nil
	}()
	if decode == nil {
		err = fmt.Errorf("unhandled xt parameter encoding (encoded length %d)", len(encoded))
		return
	}
	n, err := decode(ih[:], []byte(encoded))
	if err != nil {
		err = fmt.Errorf("error decoding xt: %w", err)
		return
	}
	if n != 20 {
		panic(n)
	}
	return
}

func dropFirst(vs url.Values, key string) {
	sl := vs[key]
	switch len(sl) {
	case 0, 1:
		vs.Del(key)
	default:
		vs[key] = sl[1:]
	}
}
