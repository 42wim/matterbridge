package peer_protocol

import (
	"github.com/anacrolix/torrent/bencode"
)

const (
	// http://bittorrent.org/beps/bep_0009.html. Note that there's an
	// LT_metadata, but I've never implemented it.
	ExtensionNameMetadata = "ut_metadata"
)

type (
	ExtendedMetadataRequestMsg struct {
		Piece     int                            `bencode:"piece"`
		TotalSize int                            `bencode:"total_size"`
		Type      ExtendedMetadataRequestMsgType `bencode:"msg_type"`
	}

	ExtendedMetadataRequestMsgType int
)

func MetadataExtensionRequestMsg(peerMetadataExtensionId ExtensionNumber, piece int) Message {
	return Message{
		Type:       Extended,
		ExtendedID: peerMetadataExtensionId,
		ExtendedPayload: bencode.MustMarshal(ExtendedMetadataRequestMsg{
			Piece: piece,
			Type:  RequestMetadataExtensionMsgType,
		}),
	}
}

// Returns the expected piece size for this request message. This is needed to determine the offset
// into an extension message payload that the request metadata piece data starts.
func (me ExtendedMetadataRequestMsg) PieceSize() int {
	ret := me.TotalSize - me.Piece*(1<<14)
	if ret > 1<<14 {
		ret = 1 << 14
	}
	return ret
}
