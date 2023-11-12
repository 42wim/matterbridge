package autonat

import (
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/host/autonat/pb"
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_autonat"

var (
	reachabilityStatus = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "reachability_status",
			Help:      "Current node reachability",
		},
	)
	reachabilityStatusConfidence = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "reachability_status_confidence",
			Help:      "Node reachability status confidence",
		},
	)
	receivedDialResponseTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "received_dial_response_total",
			Help:      "Count of dial responses for client",
		},
		[]string{"response_status"},
	)
	outgoingDialResponseTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "outgoing_dial_response_total",
			Help:      "Count of dial responses for server",
		},
		[]string{"response_status"},
	)
	outgoingDialRefusedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "outgoing_dial_refused_total",
			Help:      "Count of dial requests refused by server",
		},
		[]string{"refusal_reason"},
	)
	nextProbeTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "next_probe_timestamp",
			Help:      "Time of next probe",
		},
	)
	collectors = []prometheus.Collector{
		reachabilityStatus,
		reachabilityStatusConfidence,
		receivedDialResponseTotal,
		outgoingDialResponseTotal,
		outgoingDialRefusedTotal,
		nextProbeTimestamp,
	}
)

type MetricsTracer interface {
	ReachabilityStatus(status network.Reachability)
	ReachabilityStatusConfidence(confidence int)
	ReceivedDialResponse(status pb.Message_ResponseStatus)
	OutgoingDialResponse(status pb.Message_ResponseStatus)
	OutgoingDialRefused(reason string)
	NextProbeTime(t time.Time)
}

func getResponseStatus(status pb.Message_ResponseStatus) string {
	var s string
	switch status {
	case pb.Message_OK:
		s = "ok"
	case pb.Message_E_DIAL_ERROR:
		s = "dial error"
	case pb.Message_E_DIAL_REFUSED:
		s = "dial refused"
	case pb.Message_E_BAD_REQUEST:
		s = "bad request"
	case pb.Message_E_INTERNAL_ERROR:
		s = "internal error"
	default:
		s = "unknown"
	}
	return s
}

const (
	rate_limited     = "rate limited"
	dial_blocked     = "dial blocked"
	no_valid_address = "no valid address"
)

type metricsTracer struct{}

var _ MetricsTracer = &metricsTracer{}

type metricsTracerSetting struct {
	reg prometheus.Registerer
}

type MetricsTracerOption func(*metricsTracerSetting)

func WithRegisterer(reg prometheus.Registerer) MetricsTracerOption {
	return func(s *metricsTracerSetting) {
		if reg != nil {
			s.reg = reg
		}
	}
}

func NewMetricsTracer(opts ...MetricsTracerOption) MetricsTracer {
	setting := &metricsTracerSetting{reg: prometheus.DefaultRegisterer}
	for _, opt := range opts {
		opt(setting)
	}
	metricshelper.RegisterCollectors(setting.reg, collectors...)
	return &metricsTracer{}
}

func (mt *metricsTracer) ReachabilityStatus(status network.Reachability) {
	reachabilityStatus.Set(float64(status))
}

func (mt *metricsTracer) ReachabilityStatusConfidence(confidence int) {
	reachabilityStatusConfidence.Set(float64(confidence))
}

func (mt *metricsTracer) ReceivedDialResponse(status pb.Message_ResponseStatus) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, getResponseStatus(status))
	receivedDialResponseTotal.WithLabelValues(*tags...).Inc()
}

func (mt *metricsTracer) OutgoingDialResponse(status pb.Message_ResponseStatus) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, getResponseStatus(status))
	outgoingDialResponseTotal.WithLabelValues(*tags...).Inc()
}

func (mt *metricsTracer) OutgoingDialRefused(reason string) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, reason)
	outgoingDialRefusedTotal.WithLabelValues(*tags...).Inc()
}

func (mt *metricsTracer) NextProbeTime(t time.Time) {
	nextProbeTimestamp.Set(float64(t.Unix()))
}
