package store

import (
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var storeQueries = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "waku_store_queries",
		Help: "The number of the store queries received",
	})

var storeErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_store_errors",
		Help: "The distribution of the store protocol errors",
	},
	[]string{"error_type"},
)

var collectors = []prometheus.Collector{
	storeQueries,
	storeErrors,
}

// Metrics exposes the functions required to update prometheus metrics for store protocol
type Metrics interface {
	RecordQuery()
	RecordError(err metricsErrCategory)
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

func (m *metricsImpl) RecordQuery() {
	storeQueries.Inc()
}

type metricsErrCategory string

var (
	dialFailure          metricsErrCategory = "dial_failure"
	decodeRPCFailure     metricsErrCategory = "decode_rpc_failure"
	writeRequestFailure  metricsErrCategory = "write_request_failure"
	writeResponseFailure metricsErrCategory = "write_response_failure"
	storeFailure         metricsErrCategory = "store_failure"
	emptyRPCQueryFailure metricsErrCategory = "empty_rpc_query_failure"
	peerNotFoundFailure  metricsErrCategory = "peer_not_found_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	storeErrors.WithLabelValues(string(err)).Inc()
}
