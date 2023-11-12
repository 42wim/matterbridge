package perf

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"text/tabwriter"
)

var (
	mu     sync.RWMutex
	events = map[string]*Event{}
)

func init() {
	http.HandleFunc("/debug/perf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		switch r.FormValue("sort") {
		case "desc":
			writeEventsTableCustomSort(w, func(l, r NamedEvent) bool {
				return l.Name < r.Name
			})
		default:
			WriteEventsTable(w)
		}
	})
}

type NamedEvent struct {
	Name string
	Event
}

func WriteEventsTable(w io.Writer) {
	writeEventsTableCustomSort(w, func(l, r NamedEvent) bool {
		return l.Total > r.Total
	})
}

func writeEventsTableCustomSort(w io.Writer, less func(l, r NamedEvent) bool) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprint(tw, "description\ttotal\tcount\tmin\tmean\tmax\n")
	mu.RLock()
	es := make([]NamedEvent, 0, len(events))
	for d, e := range events {
		e.Mu.RLock()
		es = append(es, NamedEvent{d, *e})
		e.Mu.RUnlock()
	}
	mu.RUnlock()
	sort.Slice(es, func(i, j int) bool {
		return less(es[i], es[j])
	})
	for _, ne := range es {
		fmt.Fprintf(tw, "%s\t%v\t%v\t%v\t%v\t%v\n", ne.Name, ne.Total, ne.Count, ne.Min, ne.MeanTime(), ne.Max)
	}
	tw.Flush()
}
