package signal

const (
	localPairingEvent = "localPairing"
)

// SendLocalPairingEvent sends event from services/pairing/events.
func SendLocalPairingEvent(event interface{}) {
	send(localPairingEvent, event)
}
