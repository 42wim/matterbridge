package testutil

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/anacrolix/missinggo/expect"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

type File struct {
	Name string
	Data string
}

type Torrent struct {
	Files []File
	Name  string
}

func (t *Torrent) IsDir() bool {
	return len(t.Files) == 1 && t.Files[0].Name == ""
}

func (t *Torrent) GetFile(name string) *File {
	if t.IsDir() && t.Name == name {
		return &t.Files[0]
	}
	for _, f := range t.Files {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func (t *Torrent) Info(pieceLength int64) metainfo.Info {
	info := metainfo.Info{
		Name:        t.Name,
		PieceLength: pieceLength,
	}
	if t.IsDir() {
		info.Length = int64(len(t.Files[0].Data))
	}
	err := info.GeneratePieces(func(fi metainfo.FileInfo) (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader(t.GetFile(strings.Join(fi.Path, "/")).Data)), nil
	})
	expect.Nil(err)
	return info
}

func (t *Torrent) Metainfo(pieceLength int64) *metainfo.MetaInfo {
	mi := metainfo.MetaInfo{}
	var err error
	mi.InfoBytes, err = bencode.Marshal(t.Info(pieceLength))
	expect.Nil(err)
	return &mi
}
