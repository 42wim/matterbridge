// Package testutil contains stuff for testing torrent-related behaviour.
//
// "greeting" is a single-file torrent of a file called "greeting" that
// "contains "hello, world\n".

package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent/metainfo"
)

var Greeting = Torrent{
	Files: []File{{
		Data: GreetingFileContents,
	}},
	Name: GreetingFileName,
}

const (
	// A null in the middle triggers an error if SQLite stores data as text instead of blob.
	GreetingFileContents = "hello,\x00world\n"
	GreetingFileName     = "greeting"
)

func CreateDummyTorrentData(dirName string) string {
	f, _ := os.Create(filepath.Join(dirName, "greeting"))
	defer f.Close()
	f.WriteString(GreetingFileContents)
	return f.Name()
}

func GreetingMetaInfo() *metainfo.MetaInfo {
	return Greeting.Metainfo(5)
}

// Gives a temporary directory containing the completed "greeting" torrent,
// and a corresponding metainfo describing it. The temporary directory can be
// cleaned away with os.RemoveAll.
func GreetingTestTorrent() (tempDir string, metaInfo *metainfo.MetaInfo) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		panic(err)
	}
	CreateDummyTorrentData(tempDir)
	metaInfo = GreetingMetaInfo()
	return
}
