package webseed

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

func trailingPath(infoName string, pathComps []string) string {
	return path.Join(
		func() (ret []string) {
			for _, comp := range append([]string{infoName}, pathComps...) {
				ret = append(ret, url.QueryEscape(comp))
			}
			return
		}()...,
	)
}

// Creates a request per BEP 19.
func NewRequest(url_ string, fileIndex int, info *metainfo.Info, offset, length int64) (*http.Request, error) {
	fileInfo := info.UpvertedFiles()[fileIndex]
	if strings.HasSuffix(url_, "/") {
		// BEP specifies that we append the file path. We need to escape each component of the path
		// for things like spaces and '#'.
		url_ += trailingPath(info.Name, fileInfo.Path)
	}
	req, err := http.NewRequest(http.MethodGet, url_, nil)
	if err != nil {
		return nil, err
	}
	if offset != 0 || length != fileInfo.Length {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+length-1))
	}
	return req, nil
}
