package storage

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

// Determines the filepath to be used for each file in a torrent.
type FilePathMaker func(opts FilePathMakerOpts) string

// Determines the directory for a given torrent within a storage client.
type TorrentDirFilePathMaker func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string

// Info passed to a FilePathMaker.
type FilePathMakerOpts struct {
	Info *metainfo.Info
	File *metainfo.FileInfo
}

// defaultPathMaker just returns the storage client's base directory.
func defaultPathMaker(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
	return baseDir
}

func infoHashPathMaker(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
	return filepath.Join(baseDir, infoHash.HexString())
}

func isSubFilepath(base, sub string) bool {
	rel, err := filepath.Rel(base, sub)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
