package common

// TimeSource provides a unified way of getting the current time.
// The intention is to always use a synchronized time source
// between all components of the protocol.
//
// This is required by Whisper and Waku protocols
// which rely on a fact that all peers
// have a synchronized time source.
type TimeSource interface {
	GetCurrentTime() uint64
}
