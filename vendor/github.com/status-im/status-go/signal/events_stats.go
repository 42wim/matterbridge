package signal

const (
	// EventsStats is sent periodically with stats like upload/download rate
	EventStats = "stats"
)

// SendStats sends stats signal.
func SendStats(stats interface{}) {
	send(EventStats, stats)
}
