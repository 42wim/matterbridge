package peer_exchange

import (
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var peerExchangeErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "peer_exchange_errors",
		Help: "The distribution of the lightpush protocol errors",
	},
	[]string{"error_type"},
)

var collectors = []prometheus.Collector{
	peerExchangeErrors,
}

// Metrics exposes the functions required to update prometheus metrics for peer_exchange protocol
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
	decodeRPCFailure metricsErrCategory = "decode_rpc_failure"
	pxFailure        metricsErrCategory = "px_failure"
	dialFailure      metricsErrCategory = "dial_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	peerExchangeErrors.WithLabelValues(string(err)).Inc()
}
