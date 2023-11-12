package shared

import "github.com/anacrolix/torrent/tracker/udp"

const (
	None      udp.AnnounceEvent = iota
	Completed                   // The local peer just completed the torrent.
	Started                     // The local peer has just resumed this torrent.
	Stopped                     // The local peer is leaving the swarm.
)
