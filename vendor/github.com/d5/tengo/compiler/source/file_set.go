package source

import (
	"sort"
)

// FileSet represents a set of source files.
type FileSet struct {
	Base     int     // base offset for the next file
	Files    []*File // list of files in the order added to the set
	LastFile *File   // cache of last file looked up
}

// NewFileSet creates a new file set.
func NewFileSet() *FileSet {
	return &FileSet{
		Base: 1, // 0 == NoPos
	}
}

// AddFile adds a new file in the file set.
func (s *FileSet) AddFile(filename string, base, size int) *File {
	if base < 0 {
		base = s.Base
	}
	if base < s.Base || size < 0 {
		panic("illegal base or size")
	}

	f := &File{
		set:   s,
		Name:  filename,
		Base:  base,
		Size:  size,
		Lines: []int{0},
	}

	base += size + 1 // +1 because EOF also has a position
	if base < 0 {
		panic("offset overflow (> 2G of source code in file set)")
	}

	// add the file to the file set
	s.Base = base
	s.Files = append(s.Files, f)
	s.LastFile = f

	return f
}

// File returns the file that contains the position p.
// If no such file is found (for instance for p == NoPos),
// the result is nil.
//
func (s *FileSet) File(p Pos) (f *File) {
	if p != NoPos {
		f = s.file(p)
	}

	return
}

// Position converts a Pos p in the fileset into a FilePos value.
func (s *FileSet) Position(p Pos) (pos FilePos) {
	if p != NoPos {
		if f := s.file(p); f != nil {
			return f.position(p)
		}
	}

	return
}

func (s *FileSet) file(p Pos) *File {
	// common case: p is in last file
	if f := s.LastFile; f != nil && f.Base <= int(p) && int(p) <= f.Base+f.Size {
		return f
	}

	// p is not in last file - search all files
	if i := searchFiles(s.Files, int(p)); i >= 0 {
		f := s.Files[i]

		// f.base <= int(p) by definition of searchFiles
		if int(p) <= f.Base+f.Size {
			s.LastFile = f // race is ok - s.last is only a cache
			return f
		}
	}

	return nil
}

func searchFiles(a []*File, x int) int {
	return sort.Search(len(a), func(i int) bool { return a[i].Base > x }) - 1
}
