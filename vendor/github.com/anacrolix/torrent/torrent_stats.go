package torrent

// Due to ConnStats, may require special alignment on some platforms. See
// https://github.com/anacrolix/torrent/issues/383.
type TorrentStats struct {
	// Aggregates stats over all connections past and present. Some values may not have much meaning
	// in the aggregate context.
	ConnStats

	// Ordered by expected descending quantities (if all is well).
	TotalPeers       int
	PendingPeers     int
	ActivePeers      int
	ConnectedSeeders int
	HalfOpenPeers    int
	PiecesComplete   int
}
