package metainfo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/anacrolix/missinggo/slices"
)

// The info dictionary.
type Info struct {
	PieceLength int64  `bencode:"piece length"`      // BEP3
	Pieces      []byte `bencode:"pieces"`            // BEP3
	Name        string `bencode:"name"`              // BEP3
	Length      int64  `bencode:"length,omitempty"`  // BEP3, mutually exclusive with Files
	Private     *bool  `bencode:"private,omitempty"` // BEP27
	// TODO: Document this field.
	Source string     `bencode:"source,omitempty"`
	Files  []FileInfo `bencode:"files,omitempty"` // BEP3, mutually exclusive with Length
}

// The Info.Name field is "advisory". For multi-file torrents it's usually a suggested directory
// name. There are situations where we don't want a directory (like using the contents of a torrent
// as the immediate contents of a directory), or the name is invalid. Transmission will inject the
// name of the torrent file if it doesn't like the name, resulting in a different infohash
// (https://github.com/transmission/transmission/issues/1775). To work around these situations, we
// will use a sentinel name for compatibility with Transmission and to signal to our own client that
// we intended to have no directory name. By exposing it in the API we can check for references to
// this behaviour within this implementation.
const NoName = "-"

// This is a helper that sets Files and Pieces from a root path and its children.
func (info *Info) BuildFromFilePath(root string) (err error) {
	info.Name = func() string {
		b := filepath.Base(root)
		switch b {
		case ".", "..", string(filepath.Separator):
			return NoName
		default:
			return b
		}
	}()
	info.Files = nil
	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			// Directories are implicit in torrent files.
			return nil
		} else if path == root {
			// The root is a file.
			info.Length = fi.Size()
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %s", err)
		}
		info.Files = append(info.Files, FileInfo{
			Path:   strings.Split(relPath, string(filepath.Separator)),
			Length: fi.Size(),
		})
		return nil
	})
	if err != nil {
		return
	}
	slices.Sort(info.Files, func(l, r FileInfo) bool {
		return strings.Join(l.Path, "/") < strings.Join(r.Path, "/")
	})
	err = info.GeneratePieces(func(fi FileInfo) (io.ReadCloser, error) {
		return os.Open(filepath.Join(root, strings.Join(fi.Path, string(filepath.Separator))))
	})
	if err != nil {
		err = fmt.Errorf("error generating pieces: %s", err)
	}
	return
}

// Concatenates all the files in the torrent into w. open is a function that
// gets at the contents of the given file.
func (info *Info) writeFiles(w io.Writer, open func(fi FileInfo) (io.ReadCloser, error)) error {
	for _, fi := range info.UpvertedFiles() {
		r, err := open(fi)
		if err != nil {
			return fmt.Errorf("error opening %v: %s", fi, err)
		}
		wn, err := io.CopyN(w, r, fi.Length)
		r.Close()
		if wn != fi.Length {
			return fmt.Errorf("error copying %v: %s", fi, err)
		}
	}
	return nil
}

// Sets Pieces (the block of piece hashes in the Info) by using the passed
// function to get at the torrent data.
func (info *Info) GeneratePieces(open func(fi FileInfo) (io.ReadCloser, error)) (err error) {
	if info.PieceLength == 0 {
		return errors.New("piece length must be non-zero")
	}
	pr, pw := io.Pipe()
	go func() {
		err := info.writeFiles(pw, open)
		pw.CloseWithError(err)
	}()
	defer pr.Close()
	info.Pieces, err = GeneratePieces(pr, info.PieceLength, nil)
	return
}

func (info *Info) TotalLength() (ret int64) {
	if info.IsDir() {
		for _, fi := range info.Files {
			ret += fi.Length
		}
	} else {
		ret = info.Length
	}
	return
}

func (info *Info) NumPieces() int {
	return len(info.Pieces) / 20
}

func (info *Info) IsDir() bool {
	return len(info.Files) != 0
}

// The files field, converted up from the old single-file in the parent info
// dict if necessary. This is a helper to avoid having to conditionally handle
// single and multi-file torrent infos.
func (info *Info) UpvertedFiles() []FileInfo {
	if len(info.Files) == 0 {
		return []FileInfo{{
			Length: info.Length,
			// Callers should determine that Info.Name is the basename, and
			// thus a regular file.
			Path: nil,
		}}
	}
	return info.Files
}

func (info *Info) Piece(index int) Piece {
	return Piece{info, pieceIndex(index)}
}
