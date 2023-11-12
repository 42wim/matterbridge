// This package has no purpose except to perform registration of multihashes.
//
// It is meant to be used as a side-effecting import, e.g.
//
//	import (
//		_ "github.com/multiformats/go-multihash/register/sha256"
//	)
//
// This package an implementation of sha256 using the go std, this is recomanded
// if you are using go1.21 or above.
package sha256

import (
	"crypto/sha256"

	multihash "github.com/multiformats/go-multihash/core"
)

func init() {
	multihash.Register(multihash.SHA2_256, sha256.New)
}
