package quicreuse

type Option func(*ConnManager) error

func DisableReuseport() Option {
	return func(m *ConnManager) error {
		m.enableReuseport = false
		return nil
	}
}

// DisableDraft29 disables support for QUIC draft-29.
// This option should be set, unless support for this legacy QUIC version is needed for backwards compatibility.
// Support for QUIC draft-29 is already deprecated and will be removed in the future, see https://github.com/libp2p/go-libp2p/issues/1841.
func DisableDraft29() Option {
	return func(m *ConnManager) error {
		m.enableDraft29 = false
		return nil
	}
}

// EnableMetrics enables Prometheus metrics collection.
func EnableMetrics() Option {
	return func(m *ConnManager) error {
		m.enableMetrics = true
		return nil
	}
}
