// +build !nomatrix

package bridgemap

import (
	bmatrix "github.com/42wim/matterbridge/bridge/matrix"
)

func init() {
	FullMap["matrix"] = bmatrix.New
}
