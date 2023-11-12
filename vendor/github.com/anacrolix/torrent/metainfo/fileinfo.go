package metainfo

import "strings"

// Information specific to a single file inside the MetaInfo structure.
type FileInfo struct {
	Length   int64    `bencode:"length"` // BEP3
	Path     []string `bencode:"path"`   // BEP3
	PathUTF8 []string `bencode:"path.utf-8,omitempty"`
}

func (fi *FileInfo) DisplayPath(info *Info) string {
	if info.IsDir() {
		return strings.Join(fi.Path, "/")
	} else {
		return info.Name
	}
}

func (me FileInfo) Offset(info *Info) (ret int64) {
	for _, fi := range info.UpvertedFiles() {
		if me.DisplayPath(info) == fi.DisplayPath(info) {
			return
		}
		ret += fi.Length
	}
	panic("not found")
}
