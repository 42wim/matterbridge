package storage

import (
	"errors"
	"path/filepath"
	"strings"
)

// Get the first file path component. We can't use filepath.Split because that breaks off the last
// one. We could optimize this to avoid allocating a slice down the track.
func firstComponent(filePath string) string {
	return strings.SplitN(filePath, string(filepath.Separator), 2)[0]
}

// Combines file info path components, ensuring the result won't escape into parent directories.
func ToSafeFilePath(fileInfoComponents ...string) (string, error) {
	safeComps := make([]string, 0, len(fileInfoComponents))
	for _, comp := range fileInfoComponents {
		safeComps = append(safeComps, filepath.Clean(comp))
	}
	safeFilePath := filepath.Join(safeComps...)
	fc := firstComponent(safeFilePath)
	switch fc {
	case "..":
		return "", errors.New("escapes root dir")
	default:
		return safeFilePath, nil
	}
}
