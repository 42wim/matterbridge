package rcmgr

import (
	"strings"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_rcmgr"

var (

	// Conns
	conns = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Name:      "connections",
		Help:      "Number of Connections",
	}, []string{"dir", "scope"})

	connsInboundSystem     = conns.With(prometheus.Labels{"dir": "inbound", "scope": "system"})
	connsInboundTransient  = conns.With(prometheus.Labels{"dir": "inbound", "scope": "transient"})
	connsOutboundSystem    = conns.With(prometheus.Labels{"dir": "outbound", "scope": "system"})
	connsOutboundTransient = conns.With(prometheus.Labels{"dir": "outbound", "scope": "transient"})

	oneTenThenExpDistributionBuckets = []float64{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 16, 32, 64, 128, 256,
	}

	// PeerConns
	peerConns = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "peer_connections",
		Buckets:   oneTenThenExpDistributionBuckets,
		Help:      "Number of connections this peer has",
	}, []string{"dir"})
	peerConnsInbound  = peerConns.With(prometheus.Labels{"dir": "inbound"})
	peerConnsOutbound = peerConns.With(prometheus.Labels{"dir": "outbound"})

	// Lets us build a histogram of our current state. See https://github.com/libp2p/go-libp2p-resource-manager/pull/54#discussion_r911244757 for more information.
	previousPeerConns = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "previous_peer_connections",
		Buckets:   oneTenThenExpDistributionBuckets,
		Help:      "Number of connections this peer previously had. This is used to get the current connection number per peer histogram by subtracting this from the peer_connections histogram",
	}, []string{"dir"})
	previousPeerConnsInbound  = previousPeerConns.With(prometheus.Labels{"dir": "inbound"})
	previousPeerConnsOutbound = previousPeerConns.With(prometheus.Labels{"dir": "outbound"})

	// Streams
	streams = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Name:      "streams",
		Help:      "Number of Streams",
	}, []string{"dir", "scope", "protocol"})

	peerStreams = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "peer_streams",
		Buckets:   oneTenThenExpDistributionBuckets,
		Help:      "Number of streams this peer has",
	}, []string{"dir"})
	peerStreamsInbound  = peerStreams.With(prometheus.Labels{"dir": "inbound"})
	peerStreamsOutbound = peerStreams.With(prometheus.Labels{"dir": "outbound"})

	previousPeerStreams = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "previous_peer_streams",
		Buckets:   oneTenThenExpDistributionBuckets,
		Help:      "Number of streams this peer has",
	}, []string{"dir"})
	previousPeerStreamsInbound  = previousPeerStreams.With(prometheus.Labels{"dir": "inbound"})
	previousPeerStreamsOutbound = previousPeerStreams.With(prometheus.Labels{"dir": "outbound"})

	// Memory
	memoryTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Name:      "memory",
		Help:      "Amount of memory reserved as reported to the Resource Manager",
	}, []string{"scope", "protocol"})

	// PeerMemory
	peerMemory = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "peer_memory",
		Buckets:   memDistribution,
		Help:      "How many peers have reserved this bucket of memory, as reported to the Resource Manager",
	})
	previousPeerMemory = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "previous_peer_memory",
		Buckets:   memDistribution,
		Help:      "How many peers have previously reserved this bucket of memory, as reported to the Resource Manager",
	})

	// ConnMemory
	connMemory = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "conn_memory",
		Buckets:   memDistribution,
		Help:      "How many conns have reserved this bucket of memory, as reported to the Resource Manager",
	})
	previousConnMemory = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "previous_conn_memory",
		Buckets:   memDistribution,
		Help:      "How many conns have previously reserved this bucket of memory, as reported to the Resource Manager",
	})

	// FDs
	fds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Name:      "fds",
		Help:      "Number of file descriptors reserved as reported to the Resource Manager",
	}, []string{"scope"})

	fdsSystem    = fds.With(prometheus.Labels{"scope": "system"})
	fdsTransient = fds.With(prometheus.Labels{"scope": "transient"})

	// Blocked resources
	blockedResources = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Name:      "blocked_resources",
		Help:      "Number of blocked resources",
	}, []string{"dir", "scope", "resource"})
)

var (
	memDistribution = []float64{
		1 << 10,   // 1KB
		4 << 10,   // 4KB
		32 << 10,  // 32KB
		1 << 20,   // 1MB
		32 << 20,  // 32MB
		256 << 20, // 256MB
		512 << 20, // 512MB
		1 << 30,   // 1GB
		2 << 30,   // 2GB
		4 << 30,   // 4GB
	}
)

func MustRegisterWith(reg prometheus.Registerer) {
	metricshelper.RegisterCollectors(reg,
		conns,
		peerConns,
		previousPeerConns,
		streams,
		peerStreams,

		previousPeerStreams,

		memoryTotal,
		peerMemory,
		previousPeerMemory,
		connMemory,
		previousConnMemory,
		fds,
		blockedResources,
	)
}

func WithMetricsDisabled() Option {
	return func(r *resourceManager) error {
		r.disableMetrics = true
		return nil
	}
}

// StatsTraceReporter reports stats on the resource manager using its traces.
type StatsTraceReporter struct{}

func NewStatsTraceReporter() (StatsTraceReporter, error) {
	// TODO tell prometheus the system limits
	return StatsTraceReporter{}, nil
}

func (r StatsTraceReporter) ConsumeEvent(evt TraceEvt) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	r.consumeEventWithLabelSlice(evt, tags)
}

// Separate func so that we can test that this function does not allocate. The syncPool may allocate.
func (r StatsTraceReporter) consumeEventWithLabelSlice(evt TraceEvt, tags *[]string) {
	switch evt.Type {
	case TraceAddStreamEvt, TraceRemoveStreamEvt:
		if p := PeerStrInScopeName(evt.Name); p != "" {
			// Aggregated peer stats. Counts how many peers have N number of streams open.
			// Uses two buckets aggregations. One to count how many streams the
			// peer has now. The other to count the negative value, or how many
			// streams did the peer use to have. When looking at the data you
			// take the difference from the two.

			oldStreamsOut := int64(evt.StreamsOut - evt.DeltaOut)
			peerStreamsOut := int64(evt.StreamsOut)
			if oldStreamsOut != peerStreamsOut {
				if oldStreamsOut != 0 {
					previousPeerStreamsOutbound.Observe(float64(oldStreamsOut))
				}
				if peerStreamsOut != 0 {
					peerStreamsOutbound.Observe(float64(peerStreamsOut))
				}
			}

			oldStreamsIn := int64(evt.StreamsIn - evt.DeltaIn)
			peerStreamsIn := int64(evt.StreamsIn)
			if oldStreamsIn != peerStreamsIn {
				if oldStreamsIn != 0 {
					previousPeerStreamsInbound.Observe(float64(oldStreamsIn))
				}
				if peerStreamsIn != 0 {
					peerStreamsInbound.Observe(float64(peerStreamsIn))
				}
			}
		} else {
			if evt.DeltaOut != 0 {
				if IsSystemScope(evt.Name) || IsTransientScope(evt.Name) {
					*tags = (*tags)[:0]
					*tags = append(*tags, "outbound", evt.Name, "")
					streams.WithLabelValues(*tags...).Set(float64(evt.StreamsOut))
				} else if proto := ParseProtocolScopeName(evt.Name); proto != "" {
					*tags = (*tags)[:0]
					*tags = append(*tags, "outbound", "protocol", proto)
					streams.WithLabelValues(*tags...).Set(float64(evt.StreamsOut))
				} else {
					// Not measuring service scope, connscope, servicepeer and protocolpeer. Lots of data, and
					// you can use aggregated peer stats + service stats to infer
					// this.
					break
				}
			}

			if evt.DeltaIn != 0 {
				if IsSystemScope(evt.Name) || IsTransientScope(evt.Name) {
					*tags = (*tags)[:0]
					*tags = append(*tags, "inbound", evt.Name, "")
					streams.WithLabelValues(*tags...).Set(float64(evt.StreamsIn))
				} else if proto := ParseProtocolScopeName(evt.Name); proto != "" {
					*tags = (*tags)[:0]
					*tags = append(*tags, "inbound", "protocol", proto)
					streams.WithLabelValues(*tags...).Set(float64(evt.StreamsIn))
				} else {
					// Not measuring service scope, connscope, servicepeer and protocolpeer. Lots of data, and
					// you can use aggregated peer stats + service stats to infer
					// this.
					break
				}
			}
		}

	case TraceAddConnEvt, TraceRemoveConnEvt:
		if p := PeerStrInScopeName(evt.Name); p != "" {
			// Aggregated peer stats. Counts how many peers have N number of connections.
			// Uses two buckets aggregations. One to count how many streams the
			// peer has now. The other to count the negative value, or how many
			// conns did the peer use to have. When looking at the data you
			// take the difference from the two.

			oldConnsOut := int64(evt.ConnsOut - evt.DeltaOut)
			connsOut := int64(evt.ConnsOut)
			if oldConnsOut != connsOut {
				if oldConnsOut != 0 {
					previousPeerConnsOutbound.Observe(float64(oldConnsOut))
				}
				if connsOut != 0 {
					peerConnsOutbound.Observe(float64(connsOut))
				}
			}

			oldConnsIn := int64(evt.ConnsIn - evt.DeltaIn)
			connsIn := int64(evt.ConnsIn)
			if oldConnsIn != connsIn {
				if oldConnsIn != 0 {
					previousPeerConnsInbound.Observe(float64(oldConnsIn))
				}
				if connsIn != 0 {
					peerConnsInbound.Observe(float64(connsIn))
				}
			}
		} else {
			if IsConnScope(evt.Name) {
				// Not measuring this. I don't think it's useful.
				break
			}

			if IsSystemScope(evt.Name) {
				connsInboundSystem.Set(float64(evt.ConnsIn))
				connsOutboundSystem.Set(float64(evt.ConnsOut))
			} else if IsTransientScope(evt.Name) {
				connsInboundTransient.Set(float64(evt.ConnsIn))
				connsOutboundTransient.Set(float64(evt.ConnsOut))
			}

			// Represents the delta in fds
			if evt.Delta != 0 {
				if IsSystemScope(evt.Name) {
					fdsSystem.Set(float64(evt.FD))
				} else if IsTransientScope(evt.Name) {
					fdsTransient.Set(float64(evt.FD))
				}
			}
		}

	case TraceReserveMemoryEvt, TraceReleaseMemoryEvt:
		if p := PeerStrInScopeName(evt.Name); p != "" {
			oldMem := evt.Memory - evt.Delta
			if oldMem != evt.Memory {
				if oldMem != 0 {
					previousPeerMemory.Observe(float64(oldMem))
				}
				if evt.Memory != 0 {
					peerMemory.Observe(float64(evt.Memory))
				}
			}
		} else if IsConnScope(evt.Name) {
			oldMem := evt.Memory - evt.Delta
			if oldMem != evt.Memory {
				if oldMem != 0 {
					previousConnMemory.Observe(float64(oldMem))
				}
				if evt.Memory != 0 {
					connMemory.Observe(float64(evt.Memory))
				}
			}
		} else {
			if IsSystemScope(evt.Name) || IsTransientScope(evt.Name) {
				*tags = (*tags)[:0]
				*tags = append(*tags, evt.Name, "")
				memoryTotal.WithLabelValues(*tags...).Set(float64(evt.Memory))
			} else if proto := ParseProtocolScopeName(evt.Name); proto != "" {
				*tags = (*tags)[:0]
				*tags = append(*tags, "protocol", proto)
				memoryTotal.WithLabelValues(*tags...).Set(float64(evt.Memory))
			} else {
				// Not measuring connscope, servicepeer and protocolpeer. Lots of data, and
				// you can use aggregated peer stats + service stats to infer
				// this.
				break
			}
		}

	case TraceBlockAddConnEvt, TraceBlockAddStreamEvt, TraceBlockReserveMemoryEvt:
		var resource string
		if evt.Type == TraceBlockAddConnEvt {
			resource = "connection"
		} else if evt.Type == TraceBlockAddStreamEvt {
			resource = "stream"
		} else {
			resource = "memory"
		}

		scopeName := evt.Name
		// Only the top scopeName. We don't want to get the peerid here.
		// Using indexes and slices to avoid allocating.
		scopeSplitIdx := strings.IndexByte(scopeName, ':')
		if scopeSplitIdx != -1 {
			scopeName = evt.Name[0:scopeSplitIdx]
		}
		// Drop the connection or stream id
		idSplitIdx := strings.IndexByte(scopeName, '-')
		if idSplitIdx != -1 {
			scopeName = scopeName[0:idSplitIdx]
		}

		if evt.DeltaIn != 0 {
			*tags = (*tags)[:0]
			*tags = append(*tags, "inbound", scopeName, resource)
			blockedResources.WithLabelValues(*tags...).Add(float64(evt.DeltaIn))
		}

		if evt.DeltaOut != 0 {
			*tags = (*tags)[:0]
			*tags = append(*tags, "outbound", scopeName, resource)
			blockedResources.WithLabelValues(*tags...).Add(float64(evt.DeltaOut))
		}

		if evt.Delta != 0 && resource == "connection" {
			// This represents fds blocked
			*tags = (*tags)[:0]
			*tags = append(*tags, "", scopeName, "fd")
			blockedResources.WithLabelValues(*tags...).Add(float64(evt.Delta))
		} else if evt.Delta != 0 {
			*tags = (*tags)[:0]
			*tags = append(*tags, "", scopeName, resource)
			blockedResources.WithLabelValues(*tags...).Add(float64(evt.Delta))
		}
	}
}
