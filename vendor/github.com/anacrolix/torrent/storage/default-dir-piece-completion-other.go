// Bolt piece completion is not available, and neither is sqlite.
//go:build wasm || noboltdb
// +build wasm noboltdb

package storage

import (
	"errors"
)

func NewDefaultPieceCompletionForDir(dir string) (PieceCompletion, error) {
	return nil, errors.New("y ur OS no have features")
}
