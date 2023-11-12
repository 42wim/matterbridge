package identify

import (
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/metricshelper"

	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_identify"

var (
	pushesTriggered = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "identify_pushes_triggered_total",
			Help:      "Pushes Triggered",
		},
		[]string{"trigger"},
	)
	identify = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "identify_total",
			Help:      "Identify",
		},
		[]string{"dir"},
	)
	identifyPush = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "identify_push_total",
			Help:      "Identify Push",
		},
		[]string{"dir"},
	)
	connPushSupportTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "conn_push_support_total",
			Help:      "Identify Connection Push Support",
		},
		[]string{"support"},
	)
	protocolsCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "protocols_count",
			Help:      "Protocols Count",
		},
	)
	addrsCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "addrs_count",
			Help:      "Address Count",
		},
	)
	numProtocolsReceived = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "protocols_received",
			Help:      "Number of Protocols received",
			Buckets:   buckets,
		},
	)
	numAddrsReceived = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "addrs_received",
			Help:      "Number of addrs received",
			Buckets:   buckets,
		},
	)
	collectors = []prometheus.Collector{
		pushesTriggered,
		identify,
		identifyPush,
		connPushSupportTotal,
		protocolsCount,
		addrsCount,
		numProtocolsReceived,
		numAddrsReceived,
	}
	// 1 to 20 and then up to 100 in steps of 5
	buckets = append(
		prometheus.LinearBuckets(1, 1, 20),
		prometheus.LinearBuckets(25, 5, 16)...,
	)
)

type MetricsTracer interface {
	// TriggeredPushes counts IdentifyPushes triggered by event
	TriggeredPushes(event any)

	// ConnPushSupport counts peers by Push Support
	ConnPushSupport(identifyPushSupport)

	// IdentifyReceived tracks metrics on receiving an identify response
	IdentifyReceived(isPush bool, numProtocols int, numAddrs int)

	// IdentifySent tracks metrics on sending an identify response
	IdentifySent(isPush bool, numProtocols int, numAddrs int)
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

func (t *metricsTracer) TriggeredPushes(ev any) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	typ := "unknown"
	switch ev.(type) {
	case event.EvtLocalProtocolsUpdated:
		typ = "protocols_updated"
	case event.EvtLocalAddressesUpdated:
		typ = "addresses_updated"
	}
	*tags = append(*tags, typ)
	pushesTriggered.WithLabelValues(*tags...).Inc()
}

func (t *metricsTracer) IncrementPushSupport(s identifyPushSupport) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, getPushSupport(s))
	connPushSupportTotal.WithLabelValues(*tags...).Inc()
}

func (t *metricsTracer) IdentifySent(isPush bool, numProtocols int, numAddrs int) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	if isPush {
		*tags = append(*tags, metricshelper.GetDirection(network.DirOutbound))
		identifyPush.WithLabelValues(*tags...).Inc()
	} else {
		*tags = append(*tags, metricshelper.GetDirection(network.DirInbound))
		identify.WithLabelValues(*tags...).Inc()
	}

	protocolsCount.Set(float64(numProtocols))
	addrsCount.Set(float64(numAddrs))
}

func (t *metricsTracer) IdentifyReceived(isPush bool, numProtocols int, numAddrs int) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	if isPush {
		*tags = append(*tags, metricshelper.GetDirection(network.DirInbound))
		identifyPush.WithLabelValues(*tags...).Inc()
	} else {
		*tags = append(*tags, metricshelper.GetDirection(network.DirOutbound))
		identify.WithLabelValues(*tags...).Inc()
	}

	numProtocolsReceived.Observe(float64(numProtocols))
	numAddrsReceived.Observe(float64(numAddrs))
}

func (t *metricsTracer) ConnPushSupport(support identifyPushSupport) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, getPushSupport(support))
	connPushSupportTotal.WithLabelValues(*tags...).Inc()
}

func getPushSupport(s identifyPushSupport) string {
	switch s {
	case identifyPushSupported:
		return "supported"
	case identifyPushUnsupported:
		return "not supported"
	default:
		return "unknown"
	}
}
