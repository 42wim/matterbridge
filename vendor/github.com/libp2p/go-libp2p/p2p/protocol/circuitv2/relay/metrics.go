package relay

import (
	"time"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	pbv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/pb"
	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_relaysvc"

var (
	status = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "status",
			Help:      "Relay Status",
		},
	)

	reservationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "reservations_total",
			Help:      "Relay Reservation Request",
		},
		[]string{"type"},
	)
	reservationRequestResponseStatusTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "reservation_request_response_status_total",
			Help:      "Relay Reservation Request Response Status",
		},
		[]string{"status"},
	)
	reservationRejectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "reservation_rejections_total",
			Help:      "Relay Reservation Rejected Reason",
		},
		[]string{"reason"},
	)

	connectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "connections_total",
			Help:      "Relay Connection Total",
		},
		[]string{"type"},
	)
	connectionRequestResponseStatusTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "connection_request_response_status_total",
			Help:      "Relay Connection Request Status",
		},
		[]string{"status"},
	)
	connectionRejectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "connection_rejections_total",
			Help:      "Relay Connection Rejected Reason",
		},
		[]string{"reason"},
	)
	connectionDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "connection_duration_seconds",
			Help:      "Relay Connection Duration",
		},
	)

	dataTransferredBytesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "data_transferred_bytes_total",
			Help:      "Bytes Transferred Total",
		},
	)

	collectors = []prometheus.Collector{
		status,
		reservationsTotal,
		reservationRequestResponseStatusTotal,
		reservationRejectionsTotal,
		connectionsTotal,
		connectionRequestResponseStatusTotal,
		connectionRejectionsTotal,
		connectionDurationSeconds,
		dataTransferredBytesTotal,
	}
)

const (
	requestStatusOK       = "ok"
	requestStatusRejected = "rejected"
	requestStatusError    = "error"
)

// MetricsTracer is the interface for tracking metrics for relay service
type MetricsTracer interface {
	// RelayStatus tracks whether the service is currently active
	RelayStatus(enabled bool)

	// ConnectionOpened tracks metrics on opening a relay connection
	ConnectionOpened()
	// ConnectionClosed tracks metrics on closing a relay connection
	ConnectionClosed(d time.Duration)
	// ConnectionRequestHandled tracks metrics on handling a relay connection request
	ConnectionRequestHandled(status pbv2.Status)

	// ReservationAllowed tracks metrics on opening or renewing a relay reservation
	ReservationAllowed(isRenewal bool)
	// ReservationRequestClosed tracks metrics on closing a relay reservation
	ReservationClosed(cnt int)
	// ReservationRequestHandled tracks metrics on handling a relay reservation request
	ReservationRequestHandled(status pbv2.Status)

	// BytesTransferred tracks the total bytes transferred by the relay service
	BytesTransferred(cnt int)
}

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

func (mt *metricsTracer) RelayStatus(enabled bool) {
	if enabled {
		status.Set(1)
	} else {
		status.Set(0)
	}
}

func (mt *metricsTracer) ConnectionOpened() {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, "opened")

	connectionsTotal.WithLabelValues(*tags...).Add(1)
}

func (mt *metricsTracer) ConnectionClosed(d time.Duration) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, "closed")

	connectionsTotal.WithLabelValues(*tags...).Add(1)
	connectionDurationSeconds.Observe(d.Seconds())
}

func (mt *metricsTracer) ConnectionRequestHandled(status pbv2.Status) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	respStatus := getResponseStatus(status)

	*tags = append(*tags, respStatus)
	connectionRequestResponseStatusTotal.WithLabelValues(*tags...).Add(1)
	if respStatus == requestStatusRejected {
		*tags = (*tags)[:0]
		*tags = append(*tags, getRejectionReason(status))
		connectionRejectionsTotal.WithLabelValues(*tags...).Add(1)
	}
}

func (mt *metricsTracer) ReservationAllowed(isRenewal bool) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	if isRenewal {
		*tags = append(*tags, "renewed")
	} else {
		*tags = append(*tags, "opened")
	}

	reservationsTotal.WithLabelValues(*tags...).Add(1)
}

func (mt *metricsTracer) ReservationClosed(cnt int) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, "closed")

	reservationsTotal.WithLabelValues(*tags...).Add(float64(cnt))
}

func (mt *metricsTracer) ReservationRequestHandled(status pbv2.Status) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	respStatus := getResponseStatus(status)

	*tags = append(*tags, respStatus)
	reservationRequestResponseStatusTotal.WithLabelValues(*tags...).Add(1)
	if respStatus == requestStatusRejected {
		*tags = (*tags)[:0]
		*tags = append(*tags, getRejectionReason(status))
		reservationRejectionsTotal.WithLabelValues(*tags...).Add(1)
	}
}

func (mt *metricsTracer) BytesTransferred(cnt int) {
	dataTransferredBytesTotal.Add(float64(cnt))
}

func getResponseStatus(status pbv2.Status) string {
	responseStatus := "unknown"
	switch status {
	case pbv2.Status_RESERVATION_REFUSED,
		pbv2.Status_RESOURCE_LIMIT_EXCEEDED,
		pbv2.Status_PERMISSION_DENIED,
		pbv2.Status_NO_RESERVATION,
		pbv2.Status_MALFORMED_MESSAGE:

		responseStatus = requestStatusRejected
	case pbv2.Status_UNEXPECTED_MESSAGE, pbv2.Status_CONNECTION_FAILED:
		responseStatus = requestStatusError
	case pbv2.Status_OK:
		responseStatus = requestStatusOK
	}
	return responseStatus
}

func getRejectionReason(status pbv2.Status) string {
	reason := "unknown"
	switch status {
	case pbv2.Status_RESERVATION_REFUSED:
		reason = "ip constraint violation"
	case pbv2.Status_RESOURCE_LIMIT_EXCEEDED:
		reason = "resource limit exceeded"
	case pbv2.Status_PERMISSION_DENIED:
		reason = "permission denied"
	case pbv2.Status_NO_RESERVATION:
		reason = "no reservation"
	case pbv2.Status_MALFORMED_MESSAGE:
		reason = "malformed message"
	}
	return reason
}
