package filter

import (
	"time"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var filterMessages = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "waku_filter_messages",
		Help: "The number of messages received via filter protocol",
	})

var filterErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_filter_errors",
		Help: "The distribution of the filter protocol errors",
	},
	[]string{"error_type"},
)

var filterRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_filter_requests",
		Help: "The distribution of filter requests",
	},
	[]string{"request_type"},
)

var filterRequestDurationSeconds = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "waku_filter_request_duration_seconds",
		Help: "Duration of Filter Subscribe Requests",
	},
	[]string{"request_type"},
)

var filterHandleMessageDurationSeconds = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name: "waku_filter_handle_message_duration_seconds",
		Help: "Duration to Push Message to Filter Subscribers",
	})

var filterSubscriptions = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "waku_filter_subscriptions",
		Help: "The number of filter subscriptions",
	})

var collectors = []prometheus.Collector{
	filterMessages,
	filterErrors,
	filterRequests,
	filterSubscriptions,
	filterRequestDurationSeconds,
	filterHandleMessageDurationSeconds,
}

// Metrics exposes the functions required to update prometheus metrics for filter protocol
type Metrics interface {
	RecordMessage()
	RecordRequest(requestType string, duration time.Duration)
	RecordPushDuration(duration time.Duration)
	RecordSubscriptions(num int)
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
func (m *metricsImpl) RecordMessage() {
	filterMessages.Inc()
}

type metricsErrCategory string

var (
	unknownPeerMessagePush     metricsErrCategory = "unknown_peer_messagepush"
	decodeRPCFailure           metricsErrCategory = "decode_rpc_failure"
	invalidSubscriptionMessage metricsErrCategory = "invalid_subscription_message"
	dialFailure                metricsErrCategory = "dial_failure"
	writeRequestFailure        metricsErrCategory = "write_request_failure"
	requestIDMismatch          metricsErrCategory = "request_id_mismatch"
	errorResponse              metricsErrCategory = "error_response"
	peerNotFoundFailure        metricsErrCategory = "peer_not_found_failure"
	writeResponseFailure       metricsErrCategory = "write_response_failure"
	pushTimeoutFailure         metricsErrCategory = "push_timeout_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	filterErrors.WithLabelValues(string(err)).Inc()
}

// RecordRequest tracks the duration of each type of filter request received
func (m *metricsImpl) RecordRequest(requestType string, duration time.Duration) {
	filterRequests.WithLabelValues(requestType).Inc()
	filterRequestDurationSeconds.WithLabelValues(requestType).Observe(duration.Seconds())
}

// RecordPushDuration tracks the duration of pushing a message to a filter subscriber
func (m *metricsImpl) RecordPushDuration(duration time.Duration) {
	filterHandleMessageDurationSeconds.Observe(duration.Seconds())
}

// RecordSubscriptions track the current number of filter subscriptions
func (m *metricsImpl) RecordSubscriptions(num int) {
	filterSubscriptions.Set(float64(num))
}
