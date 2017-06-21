package netutil

import (
	"net/url"
)

func ToUrlValues(m map[string]string) url.Values {
	r := make(url.Values)
	for k, v := range m {
		r.Add(k, v)
	}
	return r
}
