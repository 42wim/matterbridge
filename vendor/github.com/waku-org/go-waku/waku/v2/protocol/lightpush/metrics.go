package lightpush

import (
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var lightpushMessages = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "waku_lightpush_messages",
		Help: "The number of messages sent via lightpush protocol",
	})

var lightpushErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_lightpush_errors",
		Help: "The distribution of the lightpush protocol errors",
	},
	[]string{"error_type"},
)

var collectors = []prometheus.Collector{
	lightpushMessages,
	lightpushErrors,
}

// Metrics exposes the functions required to update prometheus metrics for lightpush protocol
type Metrics interface {
	RecordMessage()
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

// RecordMessage is used to increase the counter for the number of messages received via waku lightpush
func (m *metricsImpl) RecordMessage() {
	lightpushMessages.Inc()
}

type metricsErrCategory string

var (
	decodeRPCFailure     metricsErrCategory = "decode_rpc_failure"
	writeRequestFailure  metricsErrCategory = "write_request_failure"
	writeResponseFailure metricsErrCategory = "write_response_failure"
	dialFailure          metricsErrCategory = "dial_failure"
	messagePushFailure   metricsErrCategory = "message_push_failure"
	requestBodyFailure   metricsErrCategory = "request_failure"
	responseBodyFailure  metricsErrCategory = "response_body_failure"
	peerNotFoundFailure  metricsErrCategory = "peer_not_found_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	lightpushErrors.WithLabelValues(string(err)).Inc()
}
