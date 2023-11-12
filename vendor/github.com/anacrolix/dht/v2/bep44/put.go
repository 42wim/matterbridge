package bep44

import (
	"crypto/ed25519"
	"crypto/sha1"

	"github.com/anacrolix/torrent/bencode"
)

type Put struct {
	V    interface{}
	K    *[32]byte
	Salt []byte
	Sig  [64]byte
	Cas  int64
	Seq  int64
}

func (p *Put) ToItem() *Item {
	i := &Item{
		V:    p.V,
		Salt: p.Salt,
		Sig:  p.Sig,
		Cas:  p.Cas,
		Seq:  p.Seq,
	}
	if p.K != nil {
		i.K = *p.K
	}
	return i
}

func (p *Put) Sign(k ed25519.PrivateKey) {
	copy(p.Sig[:], Sign(k, p.Salt, p.Seq, bencode.MustMarshal(p.V)))
}

func (i *Put) Target() Target {
	if i.IsMutable() {
		return MakeMutableTarget(*i.K, i.Salt)
	} else {
		return sha1.Sum(bencode.MustMarshal(i.V))
	}
}

func (s *Put) IsMutable() bool {
	return s.K != nil
}
