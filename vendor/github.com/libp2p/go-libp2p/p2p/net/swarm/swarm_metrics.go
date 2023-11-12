package swarm

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/metricshelper"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_swarm"

var (
	connsOpened = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "connections_opened_total",
			Help:      "Connections Opened",
		},
		[]string{"dir", "transport", "security", "muxer", "early_muxer", "ip_version"},
	)
	keyTypes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "key_types_total",
			Help:      "key type",
		},
		[]string{"dir", "key_type"},
	)
	connsClosed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "connections_closed_total",
			Help:      "Connections Closed",
		},
		[]string{"dir", "transport", "security", "muxer", "early_muxer", "ip_version"},
	)
	dialError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "dial_errors_total",
			Help:      "Dial Error",
		},
		[]string{"transport", "error", "ip_version"},
	)
	connDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "connection_duration_seconds",
			Help:      "Duration of a Connection",
			Buckets:   prometheus.ExponentialBuckets(1.0/16, 2, 25), // up to 24 days
		},
		[]string{"dir", "transport", "security", "muxer", "early_muxer", "ip_version"},
	)
	connHandshakeLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "handshake_latency_seconds",
			Help:      "Duration of the libp2p Handshake",
			Buckets:   prometheus.ExponentialBuckets(0.001, 1.3, 35),
		},
		[]string{"transport", "security", "muxer", "early_muxer", "ip_version"},
	)
	dialsPerPeer = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "dials_per_peer_total",
			Help:      "Number of addresses dialed per peer",
		},
		[]string{"outcome", "num_dials"},
	)
	dialRankingDelay = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "dial_ranking_delay_seconds",
			Help:      "delay introduced by the dial ranking logic",
			Buckets:   []float64{0.001, 0.01, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.75, 1, 2},
		},
	)
	blackHoleFilterState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "black_hole_filter_state",
			Help:      "State of the black hole filter",
		},
		[]string{"name"},
	)
	blackHoleFilterSuccessFraction = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "black_hole_filter_success_fraction",
			Help:      "Fraction of successful dials among the last n requests",
		},
		[]string{"name"},
	)
	blackHoleFilterNextRequestAllowedAfter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "black_hole_filter_next_request_allowed_after",
			Help:      "Number of requests after which the next request will be allowed",
		},
		[]string{"name"},
	)
	collectors = []prometheus.Collector{
		connsOpened,
		keyTypes,
		connsClosed,
		dialError,
		connDuration,
		connHandshakeLatency,
		dialsPerPeer,
		dialRankingDelay,
		blackHoleFilterSuccessFraction,
		blackHoleFilterState,
		blackHoleFilterNextRequestAllowedAfter,
	}
)

type MetricsTracer interface {
	OpenedConnection(network.Direction, crypto.PubKey, network.ConnectionState, ma.Multiaddr)
	ClosedConnection(network.Direction, time.Duration, network.ConnectionState, ma.Multiaddr)
	CompletedHandshake(time.Duration, network.ConnectionState, ma.Multiaddr)
	FailedDialing(ma.Multiaddr, error)
	DialCompleted(success bool, totalDials int)
	DialRankingDelay(d time.Duration)
	UpdatedBlackHoleFilterState(name string, state blackHoleState, nextProbeAfter int, successFraction float64)
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

func appendConnectionState(tags []string, cs network.ConnectionState) []string {
	if cs.Transport == "" {
		// This shouldn't happen, unless the transport doesn't properly set the Transport field in the ConnectionState.
		tags = append(tags, "unknown")
	} else {
		tags = append(tags, string(cs.Transport))
	}
	// These might be empty, depending on the transport.
	// For example, QUIC doesn't set security nor muxer.
	tags = append(tags, string(cs.Security))
	tags = append(tags, string(cs.StreamMultiplexer))

	earlyMuxer := "false"
	if cs.UsedEarlyMuxerNegotiation {
		earlyMuxer = "true"
	}
	tags = append(tags, earlyMuxer)
	return tags
}

func (m *metricsTracer) OpenedConnection(dir network.Direction, p crypto.PubKey, cs network.ConnectionState, laddr ma.Multiaddr) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, metricshelper.GetDirection(dir))
	*tags = appendConnectionState(*tags, cs)
	*tags = append(*tags, metricshelper.GetIPVersion(laddr))
	connsOpened.WithLabelValues(*tags...).Inc()

	*tags = (*tags)[:0]
	*tags = append(*tags, metricshelper.GetDirection(dir))
	*tags = append(*tags, p.Type().String())
	keyTypes.WithLabelValues(*tags...).Inc()
}

func (m *metricsTracer) ClosedConnection(dir network.Direction, duration time.Duration, cs network.ConnectionState, laddr ma.Multiaddr) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, metricshelper.GetDirection(dir))
	*tags = appendConnectionState(*tags, cs)
	*tags = append(*tags, metricshelper.GetIPVersion(laddr))
	connsClosed.WithLabelValues(*tags...).Inc()
	connDuration.WithLabelValues(*tags...).Observe(duration.Seconds())
}

func (m *metricsTracer) CompletedHandshake(t time.Duration, cs network.ConnectionState, laddr ma.Multiaddr) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = appendConnectionState(*tags, cs)
	*tags = append(*tags, metricshelper.GetIPVersion(laddr))
	connHandshakeLatency.WithLabelValues(*tags...).Observe(t.Seconds())
}

func (m *metricsTracer) FailedDialing(addr ma.Multiaddr, err error) {
	transport := metricshelper.GetTransport(addr)
	e := "other"
	if errors.Is(err, context.Canceled) {
		e = "canceled"
	} else if errors.Is(err, context.DeadlineExceeded) {
		e = "deadline"
	} else {
		nerr, ok := err.(net.Error)
		if ok && nerr.Timeout() {
			e = "timeout"
		} else if strings.Contains(err.Error(), "connect: connection refused") {
			e = "connection refused"
		}
	}

	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, transport, e)
	*tags = append(*tags, metricshelper.GetIPVersion(addr))
	dialError.WithLabelValues(*tags...).Inc()
}

func (m *metricsTracer) DialCompleted(success bool, totalDials int) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	if success {
		*tags = append(*tags, "success")
	} else {
		*tags = append(*tags, "failed")
	}

	numDialLabels := [...]string{"0", "1", "2", "3", "4", "5", ">=6"}
	var numDials string
	if totalDials < len(numDialLabels) {
		numDials = numDialLabels[totalDials]
	} else {
		numDials = numDialLabels[len(numDialLabels)-1]
	}
	*tags = append(*tags, numDials)
	dialsPerPeer.WithLabelValues(*tags...).Inc()
}

func (m *metricsTracer) DialRankingDelay(d time.Duration) {
	dialRankingDelay.Observe(d.Seconds())
}

func (m *metricsTracer) UpdatedBlackHoleFilterState(name string, state blackHoleState,
	nextProbeAfter int, successFraction float64) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, name)

	blackHoleFilterState.WithLabelValues(*tags...).Set(float64(state))
	blackHoleFilterSuccessFraction.WithLabelValues(*tags...).Set(successFraction)
	blackHoleFilterNextRequestAllowedAfter.WithLabelValues(*tags...).Set(float64(nextProbeAfter))
}
