package http

import (
	"fmt"

	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/torrent/bencode"
)

type HttpResponse struct {
	FailureReason string `bencode:"failure reason"`
	Interval      int32  `bencode:"interval"`
	TrackerId     string `bencode:"tracker id"`
	Complete      int32  `bencode:"complete"`
	Incomplete    int32  `bencode:"incomplete"`
	Peers         Peers  `bencode:"peers"`
	// BEP 7
	Peers6 krpc.CompactIPv6NodeAddrs `bencode:"peers6"`
}

type Peers []Peer

func (me *Peers) UnmarshalBencode(b []byte) (err error) {
	var _v interface{}
	err = bencode.Unmarshal(b, &_v)
	if err != nil {
		return
	}
	switch v := _v.(type) {
	case string:
		vars.Add("http responses with string peers", 1)
		var cnas krpc.CompactIPv4NodeAddrs
		err = cnas.UnmarshalBinary([]byte(v))
		if err != nil {
			return
		}
		for _, cp := range cnas {
			*me = append(*me, Peer{
				IP:   cp.IP[:],
				Port: int(cp.Port),
			})
		}
		return
	case []interface{}:
		vars.Add("http responses with list peers", 1)
		for _, i := range v {
			var p Peer
			p.FromDictInterface(i.(map[string]interface{}))
			*me = append(*me, p)
		}
		return
	default:
		vars.Add("http responses with unhandled peers type", 1)
		err = fmt.Errorf("unsupported type: %T", _v)
		return
	}
}
