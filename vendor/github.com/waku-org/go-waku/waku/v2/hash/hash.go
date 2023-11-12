package hash

import (
	"crypto/sha256"
	"hash"
	"sync"
)

var sha256Pool = sync.Pool{New: func() interface{} {
	return sha256.New()
}}

// SHA256 generates the SHA256 hash from the input data
func SHA256(data ...[]byte) []byte {
	h, ok := sha256Pool.Get().(hash.Hash)
	if !ok {
		h = sha256.New()
	}
	defer sha256Pool.Put(h)
	h.Reset()
	for i := range data {
		h.Write(data[i])
	}
	return h.Sum(nil)
}
