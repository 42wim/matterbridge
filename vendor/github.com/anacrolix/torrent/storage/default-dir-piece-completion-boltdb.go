// Bolt piece completion is available, and sqlite is not.
//go:build !noboltdb && !wasm && (js || nosqlite)
// +build !noboltdb
// +build !wasm
// +build js nosqlite

package storage

func NewDefaultPieceCompletionForDir(dir string) (PieceCompletion, error) {
	return NewBoltPieceCompletion(dir)
}
