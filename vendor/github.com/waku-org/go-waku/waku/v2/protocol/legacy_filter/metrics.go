package legacy_filter

import (
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var filterMessages = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "legacy_filter_messages",
		Help: "The number of messages received via legacy filter protocol",
	})

var filterErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "legacy_filter_errors",
		Help: "The distribution of the legacy filter protocol errors",
	},
	[]string{"error_type"},
)

var filterSubscribers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "legacy_filter_subscriptions",
		Help: "The number of legacy filter subscribers",
	})

var collectors = []prometheus.Collector{
	filterMessages,
	filterErrors,
	filterSubscribers,
}

// Metrics exposes the functions required to update prometheus metrics for legacy filter protocol
type Metrics interface {
	RecordMessages(num int)
	RecordSubscribers(num int)
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

// RecordMessage is used to increase the counter for the number of messages received via waku filter
func (m *metricsImpl) RecordMessages(num int) {
	filterMessages.Add(float64(num))
}

type metricsErrCategory string

var (
	decodeRPCFailure    metricsErrCategory = "decode_rpc_failure"
	dialFailure         metricsErrCategory = "dial_failure"
	pushWriteError      metricsErrCategory = "push_write_error"
	peerNotFoundFailure metricsErrCategory = "peer_not_found_failure"
	writeRequestFailure metricsErrCategory = "write_request_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	filterErrors.WithLabelValues(string(err)).Inc()
}

// RecordSubscribers track the current number of filter subscribers
func (m *metricsImpl) RecordSubscribers(num int) {
	filterSubscribers.Set(float64(num))
}
