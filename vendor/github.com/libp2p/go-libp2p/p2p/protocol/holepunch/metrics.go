package holepunch

import (
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_holepunch"

var (
	directDialsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "direct_dials_total",
			Help:      "Direct Dials Total",
		},
		[]string{"outcome"},
	)
	hpAddressOutcomesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "address_outcomes_total",
			Help:      "Hole Punch outcomes by Transport",
		},
		[]string{"side", "num_attempts", "ipv", "transport", "outcome"},
	)
	hpOutcomesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "outcomes_total",
			Help:      "Hole Punch outcomes overall",
		},
		[]string{"side", "num_attempts", "outcome"},
	)

	collectors = []prometheus.Collector{
		directDialsTotal,
		hpAddressOutcomesTotal,
		hpOutcomesTotal,
	}
)

type MetricsTracer interface {
	HolePunchFinished(side string, attemptNum int, theirAddrs []ma.Multiaddr, ourAddr []ma.Multiaddr, directConn network.ConnMultiaddrs)
	DirectDialFinished(success bool)
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
	// initialise metrics's labels so that the first data point is handled correctly
	for _, side := range []string{"initiator", "receiver"} {
		for _, numAttempts := range []string{"1", "2", "3", "4"} {
			for _, outcome := range []string{"success", "failed", "cancelled", "no_suitable_address"} {
				for _, ipv := range []string{"ip4", "ip6"} {
					for _, transport := range []string{"quic", "quic-v1", "tcp", "webtransport"} {
						hpAddressOutcomesTotal.WithLabelValues(side, numAttempts, ipv, transport, outcome)
					}
				}
				if outcome == "cancelled" {
					// not a valid outcome for the overall holepunch metric
					continue
				}
				hpOutcomesTotal.WithLabelValues(side, numAttempts, outcome)
			}
		}
	}
	return &metricsTracer{}
}

// HolePunchFinished tracks metrics completion of a holepunch. Metrics are tracked on
// a holepunch attempt level and on individual addresses involved in a holepunch.
//
// outcome for an address is computed as:
//
//   - success:
//     A direct connection was established with the peer using this address
//   - cancelled:
//     A direct connection was established with the peer but not using this address
//   - failed:
//     No direct connection was made to the peer and the peer reported an address
//     with the same transport as this address
//   - no_suitable_address:
//     The peer reported no address with the same transport as this address
func (mt *metricsTracer) HolePunchFinished(side string, numAttempts int,
	remoteAddrs []ma.Multiaddr, localAddrs []ma.Multiaddr, directConn network.ConnMultiaddrs) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, side, getNumAttemptString(numAttempts))
	var dipv, dtransport string
	if directConn != nil {
		dipv = metricshelper.GetIPVersion(directConn.LocalMultiaddr())
		dtransport = metricshelper.GetTransport(directConn.LocalMultiaddr())
	}

	matchingAddressCount := 0
	// calculate holepunch outcome for all the addresses involved
	for _, la := range localAddrs {
		lipv := metricshelper.GetIPVersion(la)
		ltransport := metricshelper.GetTransport(la)

		matchingAddress := false
		for _, ra := range remoteAddrs {
			ripv := metricshelper.GetIPVersion(ra)
			rtransport := metricshelper.GetTransport(ra)
			if ripv == lipv && rtransport == ltransport {
				// the peer reported an address with the same transport
				matchingAddress = true
				matchingAddressCount++

				*tags = append(*tags, ripv, rtransport)
				if directConn != nil && dipv == ripv && dtransport == rtransport {
					// the connection was made using this address
					*tags = append(*tags, "success")
				} else if directConn != nil {
					// connection was made but not using this address
					*tags = append(*tags, "cancelled")
				} else {
					// no connection was made
					*tags = append(*tags, "failed")
				}
				hpAddressOutcomesTotal.WithLabelValues(*tags...).Inc()
				*tags = (*tags)[:2] // 2 because we want to keep (side, numAttempts)
				break
			}
		}
		if !matchingAddress {
			*tags = append(*tags, lipv, ltransport, "no_suitable_address")
			hpAddressOutcomesTotal.WithLabelValues(*tags...).Inc()
			*tags = (*tags)[:2] // 2 because we want to keep (side, numAttempts)
		}
	}

	outcome := "failed"
	if directConn != nil {
		outcome = "success"
	} else if matchingAddressCount == 0 {
		// there were no matching addresses, this attempt was going to fail
		outcome = "no_suitable_address"
	}

	*tags = append(*tags, outcome)
	hpOutcomesTotal.WithLabelValues(*tags...).Inc()
}

func getNumAttemptString(numAttempt int) string {
	var attemptStr = [...]string{"0", "1", "2", "3", "4", "5"}
	if numAttempt > 5 {
		return "> 5"
	}
	return attemptStr[numAttempt]
}

func (mt *metricsTracer) DirectDialFinished(success bool) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	if success {
		*tags = append(*tags, "success")
	} else {
		*tags = append(*tags, "failed")
	}
	directDialsTotal.WithLabelValues(*tags...).Inc()
}
