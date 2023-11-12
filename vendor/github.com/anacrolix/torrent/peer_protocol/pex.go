package peer_protocol

import (
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/torrent/bencode"
)

type PexMsg struct {
	Added       krpc.CompactIPv4NodeAddrs `bencode:"added"`
	AddedFlags  []PexPeerFlags            `bencode:"added.f"`
	Added6      krpc.CompactIPv6NodeAddrs `bencode:"added6"`
	Added6Flags []PexPeerFlags            `bencode:"added6.f"`
	Dropped     krpc.CompactIPv4NodeAddrs `bencode:"dropped"`
	Dropped6    krpc.CompactIPv6NodeAddrs `bencode:"dropped6"`
}

func (m *PexMsg) Len() int {
	return len(m.Added) + len(m.Added6) + len(m.Dropped) + len(m.Dropped6)
}

func (m *PexMsg) Message(pexExtendedId ExtensionNumber) Message {
	payload := bencode.MustMarshal(m)
	return Message{
		Type:            Extended,
		ExtendedID:      pexExtendedId,
		ExtendedPayload: payload,
	}
}

func LoadPexMsg(b []byte) (ret PexMsg, err error) {
	err = bencode.Unmarshal(b, &ret)
	return
}

type PexPeerFlags byte

func (me PexPeerFlags) Get(f PexPeerFlags) bool {
	return me&f == f
}

const (
	PexPrefersEncryption PexPeerFlags = 1 << iota
	PexSeedUploadOnly
	PexSupportsUtp
	PexHolepunchSupport
	PexOutgoingConn
)
