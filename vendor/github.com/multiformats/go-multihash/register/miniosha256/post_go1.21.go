//go:build go1.21

// This package has no purpose except to perform registration of multihashes.
//
// It is meant to be used as a side-effecting import, e.g.
//
//	import (
//		_ "github.com/multiformats/go-multihash/register/miniosha256"
//	)
//
// This package registers alternative implementations for sha2-256, using
// the github.com/minio/sha256-simd library for go1.20 and bellow. Go 1.21 and
// later fallback to [github.com/multiformats/go-multihash/register/sha256].
//
// Deprecated: please switch to [github.com/multiformats/go-multihash/register/sha256]
// as of go1.21 the go std has a SHANI implementation that is just as fast. See https://go.dev/issue/50543.
// This will be removed shortly after go1.22 is released.
package miniosha256

import (
	_ "github.com/multiformats/go-multihash/register/sha256"
)
