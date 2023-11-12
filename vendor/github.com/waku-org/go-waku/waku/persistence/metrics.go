package persistence

import (
	"time"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var archiveMessages = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "waku_archive_messages",
		Help: "The number of messages stored via archive protocol",
	})

var archiveErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_archive_errors",
		Help: "The distribution of the archive protocol errors",
	},
	[]string{"error_type"},
)

var archiveInsertDurationSeconds = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name: "waku_archive_insert_duration_seconds",
		Help: "Message insertion duration",
	})

var archiveQueryDurationSeconds = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name: "waku_archive_query_duration_seconds",
		Help: "History query duration",
	})

var collectors = []prometheus.Collector{
	archiveMessages,
	archiveErrors,
	archiveInsertDurationSeconds,
	archiveQueryDurationSeconds,
}

// Metrics exposes the functions required to update prometheus metrics for archive protocol
type Metrics interface {
	RecordMessage(num int)
	RecordError(err metricsErrCategory)
	RecordInsertDuration(duration time.Duration)
	RecordQueryDuration(duration time.Duration)
}

type metricsImpl struct {
	reg prometheus.Registerer
}

func newMetrics(reg prometheus.Registerer) Metrics {
	metricshelper.RegisterCollectors(reg, collectors...)
	return &metricsImpl{
		reg: reg,
	}
}

// RecordMessage increases the counter for the number of messages stored in the archive
func (m *metricsImpl) RecordMessage(num int) {
	archiveMessages.Add(float64(num))
}

type metricsErrCategory string

var (
	retPolicyFailure metricsErrCategory = "retpolicy_failure"
	insertFailure    metricsErrCategory = "retpolicy_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	archiveErrors.WithLabelValues(string(err)).Inc()
}

// RecordInsertDuration tracks the duration for inserting a record in the archive database
func (m *metricsImpl) RecordInsertDuration(duration time.Duration) {
	archiveInsertDurationSeconds.Observe(duration.Seconds())
}

// RecordQueryDuration tracks the duration for executing a query in the archive database
func (m *metricsImpl) RecordQueryDuration(duration time.Duration) {
	archiveQueryDurationSeconds.Observe(duration.Seconds())
}
