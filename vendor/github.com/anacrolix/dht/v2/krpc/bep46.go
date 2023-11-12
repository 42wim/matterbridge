package krpc

import (
	"github.com/anacrolix/torrent/metainfo"
)

type Bep46Payload struct {
	Ih metainfo.Hash `bencode:"ih"`
}
