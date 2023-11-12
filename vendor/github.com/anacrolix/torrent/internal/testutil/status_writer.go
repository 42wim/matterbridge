package testutil

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

	_ "github.com/anacrolix/envpprof"
)

type StatusWriter interface {
	WriteStatus(io.Writer)
}

// The key is the route pattern. The value is nil when the resource is released.
var (
	mu  sync.Mutex
	sws = map[string]StatusWriter{}
)

func ExportStatusWriter(sw StatusWriter, path string, t *testing.T) (release func()) {
	pattern := fmt.Sprintf("/%s/%s", t.Name(), path)
	t.Logf("exporting status path %q", pattern)
	release = func() {
		mu.Lock()
		defer mu.Unlock()
		sws[pattern] = nil
	}
	mu.Lock()
	defer mu.Unlock()
	if curSw, ok := sws[pattern]; ok {
		if curSw != nil {
			panic(fmt.Sprintf("%q still in use", pattern))
		}
		sws[pattern] = sw
		return
	}
	http.HandleFunc(
		pattern,
		func(w http.ResponseWriter, r *http.Request) {
			sw := sws[pattern]
			if sw == nil {
				http.NotFound(w, r)
				return
			}
			sw.WriteStatus(w)
		},
	)
	sws[pattern] = sw
	return
}
