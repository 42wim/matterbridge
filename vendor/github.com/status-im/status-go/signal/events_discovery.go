package signal

const (
	// EventDiscoveryStarted is sent when node discv5 was started.
	EventDiscoveryStarted = "discovery.started"
	// EventDiscoveryStopped is sent when discv5 server was stopped.
	EventDiscoveryStopped = "discovery.stopped"

	// EventDiscoverySummary is sent when peer is added or removed.
	// it will be a map with capability=peer count k/v's.
	EventDiscoverySummary = "discovery.summary"
)

// SendDiscoveryStarted sends discovery.started signal.
func SendDiscoveryStarted() {
	send(EventDiscoveryStarted, nil)
}

// SendDiscoveryStopped sends discovery.stopped signal.
func SendDiscoveryStopped() {
	send(EventDiscoveryStopped, nil)
}

// SendDiscoverySummary sends discovery.summary signal.
func SendDiscoverySummary(summary interface{}) {
	send(EventDiscoverySummary, summary)
}
