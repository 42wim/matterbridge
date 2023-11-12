package dnsdisc

import (
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var dnsDiscoveredNodes = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "waku_dnsdisc_discovered",
		Help: "The number of nodes discovered via DNS discovery",
	},
)

var dnsDiscoveryErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_dnsdisc_errors",
		Help: "The distribution of the dns discovery protocol errors",
	},
	[]string{"error_type"},
)

var collectors = []prometheus.Collector{
	dnsDiscoveredNodes,
	dnsDiscoveryErrors,
}

// Metrics exposes the functions required to update prometheus metrics for dnsdisc protocol
type Metrics interface {
	RecordDiscoveredNodes(numNodes int)
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
	treeSyncFailure metricsErrCategory = "tree_sync_failure"
	peerInfoFailure metricsErrCategory = "peer_info_failure"
)

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err metricsErrCategory) {
	dnsDiscoveryErrors.WithLabelValues(string(err)).Inc()
}

func (m *metricsImpl) RecordDiscoveredNodes(numNodes int) {
	dnsDiscoveredNodes.Add(float64(numNodes))
}
