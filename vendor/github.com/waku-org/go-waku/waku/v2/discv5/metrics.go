package discv5

import (
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var discV5Errors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_discv5_errors",
		Help: "The distribution of the discv5 protocol errors",
	},
	[]string{"error_type"},
)

var collectors = []prometheus.Collector{
	discV5Errors,
}

// Metrics exposes the functions required to update prometheus metrics for discv5 protocol
type Metrics interface {
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

type metricsErrCategory string

var (
	peerInfoFailure metricsErrCategory = "peer_info_failure"
	iteratorFailure metricsErrCategory = "iterator_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	discV5Errors.WithLabelValues(string(err)).Inc()
}
