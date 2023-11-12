package udp

import (
	"math"
)

type Options struct {
	RequestUri string
}

func (opts Options) Encode() (ret []byte) {
	for {
		l := len(opts.RequestUri)
		if l == 0 {
			break
		}
		if l > math.MaxUint8 {
			l = math.MaxUint8
		}
		ret = append(append(ret, optionTypeURLData, byte(l)), opts.RequestUri[:l]...)
		opts.RequestUri = opts.RequestUri[l:]
	}
	return
}
