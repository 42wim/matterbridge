package torrent

import (
	"crypto"
	"expvar"

	pp "github.com/anacrolix/torrent/peer_protocol"
)

const (
	pieceHash        = crypto.SHA1
	defaultChunkSize = 0x4000 // 16KiB

	// Arbitrary maximum of "metadata_size" (see https://www.bittorrent.org/beps/bep_0009.html)
	// This value is 2x what libtorrent-rasterbar uses, which should be plenty
	maxMetadataSize uint32 = 8 * 1024 * 1024
)

// These are our extended message IDs. Peers will use these values to
// select which extension a message is intended for.
const (
	metadataExtendedId = iota + 1 // 0 is reserved for deleting keys
	pexExtendedId
)

func defaultPeerExtensionBytes() PeerExtensionBits {
	return pp.NewPeerExtensionBytes(pp.ExtensionBitDHT, pp.ExtensionBitExtended, pp.ExtensionBitFast)
}

func init() {
	torrent.Set("peers supporting extension", &peersSupportingExtension)
	torrent.Set("chunks received", &chunksReceived)
}

// I could move a lot of these counters to their own file, but I suspect they
// may be attached to a Client someday.
var (
	torrent                  = expvar.NewMap("torrent")
	peersSupportingExtension expvar.Map
	chunksReceived           expvar.Map

	pieceHashedCorrect    = expvar.NewInt("pieceHashedCorrect")
	pieceHashedNotCorrect = expvar.NewInt("pieceHashedNotCorrect")

	completedHandshakeConnectionFlags = expvar.NewMap("completedHandshakeConnectionFlags")
	// Count of connections to peer with same client ID.
	connsToSelf        = expvar.NewInt("connsToSelf")
	receivedKeepalives = expvar.NewInt("receivedKeepalives")
	// Requests received for pieces we don't have.
	requestsReceivedForMissingPieces = expvar.NewInt("requestsReceivedForMissingPieces")
	requestedChunkLengths            = expvar.NewMap("requestedChunkLengths")

	messageTypesReceived = expvar.NewMap("messageTypesReceived")

	// Track the effectiveness of Torrent.connPieceInclinationPool.
	pieceInclinationsReused = expvar.NewInt("pieceInclinationsReused")
	pieceInclinationsNew    = expvar.NewInt("pieceInclinationsNew")
	pieceInclinationsPut    = expvar.NewInt("pieceInclinationsPut")

	concurrentChunkWrites = expvar.NewInt("torrentConcurrentChunkWrites")
)
