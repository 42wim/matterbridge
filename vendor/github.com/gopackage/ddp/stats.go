package ddp

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Gather statistics about a DDP connection.

// Stats tracks statistics for i/o operations.
type Stats struct {
	// Bytes is the total number of bytes transferred.
	Bytes int64
	// Ops is the total number of i/o operations performed.
	Ops int64
	// Errors is the total number of i/o errors encountered.
	Errors int64
	// Runtime is the duration that stats have been gathered.
	Runtime time.Duration
}

// ClientStats displays combined statistics for the Client.
type ClientStats struct {
	// Reads provides statistics on the raw i/o network reads for the current connection.
	Reads *Stats
	// Reads provides statistics on the raw i/o network reads for the all client connections.
	TotalReads *Stats
	// Writes provides statistics on the raw i/o network writes for the current connection.
	Writes *Stats
	// Writes provides statistics on the raw i/o network writes for all the client connections.
	TotalWrites *Stats
	// Reconnects is the number of reconnections the client has made.
	Reconnects int64
	// PingsSent is the number of pings sent by the client
	PingsSent int64
	// PingsRecv is the number of pings received by the client
	PingsRecv int64
}

// String produces a compact string representation of the client stats.
func (stats *ClientStats) String() string {
	i := stats.Reads
	ti := stats.TotalReads
	o := stats.Writes
	to := stats.TotalWrites
	totalRun := (ti.Runtime * 1000000) / 1000000
	run := (i.Runtime * 1000000) / 1000000
	return fmt.Sprintf("bytes: %d/%d##%d/%d ops: %d/%d##%d/%d err: %d/%d##%d/%d reconnects: %d pings: %d/%d uptime: %v##%v",
		i.Bytes, o.Bytes,
		ti.Bytes, to.Bytes,
		i.Ops, o.Ops,
		ti.Ops, to.Ops,
		i.Errors, o.Errors,
		ti.Errors, to.Errors,
		stats.Reconnects,
		stats.PingsRecv, stats.PingsSent,
		run, totalRun)
}

// CollectionStats combines statistics about a collection.
type CollectionStats struct {
	Name  string // Name of the collection
	Count int    // Count is the total number of documents in the collection
}

// String produces a compact string representation of the collection stat.
func (s *CollectionStats) String() string {
	return fmt.Sprintf("%s[%d]", s.Name, s.Count)
}

// StatsTracker provides the basic tooling for tracking i/o stats.
type StatsTracker struct {
	bytes  int64
	ops    int64
	errors int64
	start  time.Time
	lock   sync.Mutex
}

// NewStatsTracker create a new tracker with start time set to now.
func NewStatsTracker() *StatsTracker {
	return &StatsTracker{start: time.Now()}
}

// Op records an i/o operation. The parameters are passed through to
// allow easy chaining.
func (t *StatsTracker) Op(n int, err error) (int, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.ops++
	if err == nil {
		t.bytes += int64(n)
	} else {
		if err == io.EOF {
			// I don't think we should log EOF stats as an error
		} else {
			t.errors++
		}
	}

	return n, err
}

// Snapshot takes a snapshot of the current Reader statistics.
func (t *StatsTracker) Snapshot() *Stats {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.snap()
}

// Reset all stats to initial values.
func (t *StatsTracker) Reset() *Stats {
	t.lock.Lock()
	defer t.lock.Unlock()

	stats := t.snap()
	t.bytes = 0
	t.ops = 0
	t.errors = 0
	t.start = time.Now()

	return stats
}

func (t *StatsTracker) snap() *Stats {
	return &Stats{Bytes: t.bytes, Ops: t.ops, Errors: t.errors, Runtime: time.Since(t.start)}
}

// ReaderStats tracks statistics on any io.Reader.
// ReaderStats wraps a Reader and passes data to the actual data consumer.
type ReaderStats struct {
	StatsTracker
	Reader io.Reader
}

// NewReaderStats creates a ReaderStats object for the provided Reader.
func NewReaderStats(reader io.Reader) *ReaderStats {
	r := &ReaderStats{Reader: reader}
	r.Reset()
	return r
}

// Read passes through a read collecting statistics and logging activity.
func (r *ReaderStats) Read(p []byte) (int, error) {
	return r.Op(r.Reader.Read(p))
}

// WriterStats tracks statistics on any io.Writer.
// WriterStats wraps a Writer and passes data to the actual data producer.
type WriterStats struct {
	StatsTracker
	Writer io.Writer
}

// NewWriterStats creates a WriterStats object for the provided Writer.
func NewWriterStats(writer io.Writer) *WriterStats {
	w := &WriterStats{Writer: writer}
	w.Reset()
	return w
}

// Write collects Writer statistics.
func (w *WriterStats) Write(p []byte) (int, error) {
	if w.Writer != nil {
		return w.Op(w.Writer.Write(p))
	}
	return 0, nil
}
