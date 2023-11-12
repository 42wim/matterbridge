package identify

type config struct {
	protocolVersion         string
	userAgent               string
	disableSignedPeerRecord bool
	metricsTracer           MetricsTracer
}

// Option is an option function for identify.
type Option func(*config)

// ProtocolVersion sets the protocol version string that will be used to
// identify the family of protocols used by the peer.
func ProtocolVersion(s string) Option {
	return func(cfg *config) {
		cfg.protocolVersion = s
	}
}

// UserAgent sets the user agent this node will identify itself with to peers.
func UserAgent(ua string) Option {
	return func(cfg *config) {
		cfg.userAgent = ua
	}
}

// DisableSignedPeerRecord disables populating signed peer records on the outgoing Identify response
// and ONLY sends the unsigned addresses.
func DisableSignedPeerRecord() Option {
	return func(cfg *config) {
		cfg.disableSignedPeerRecord = true
	}
}

func WithMetricsTracer(tr MetricsTracer) Option {
	return func(cfg *config) {
		cfg.metricsTracer = tr
	}
}
