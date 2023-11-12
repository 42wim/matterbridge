package bep44

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/anacrolix/torrent/bencode"
)

var Empty32ByteArray [32]byte

type Item struct {
	// time when this object was added to storage
	created time.Time

	// Value to be stored
	V interface{}

	// 32 byte ed25519 public key
	K    [32]byte
	Salt []byte
	Sig  [64]byte
	Cas  int64
	Seq  int64
}

func (i *Item) ToPut() Put {
	p := Put{
		V:    i.V,
		Salt: i.Salt,
		Sig:  i.Sig,
		Cas:  i.Cas,
		Seq:  i.Seq,
	}
	if i.K != Empty32ByteArray {
		p.K = &i.K
	}
	return p
}

// NewItem creates a new arbitrary DHT element. The distinction between storing mutable
// and immutable items is the inclusion of a public key, a sequence number, signature
// and an optional salt.
//
// cas parameter is short for compare and swap, it has similar semantics as CAS CPU
// instructions. It is used to avoid race conditions when multiple nodes are writing
// to the same slot in the DHT. It is optional. If present it specifies the sequence
// number of the data blob being overwritten by the put.
//
// salt parameter is used to make possible several targets using the same private key.
//
// The optional seq field specifies that an item's value should only be sent if its
// sequence number is greater than the given value.
func NewItem(value interface{}, salt []byte, seq, cas int64, k ed25519.PrivateKey) (*Item, error) {
	v, err := bencode.Marshal(value)
	if err != nil {
		return nil, err
	}

	var kk [32]byte
	var sig [64]byte

	if k != nil {
		pk := []byte(k.Public().(ed25519.PublicKey))
		copy(kk[:], pk)
		copy(sig[:], Sign(k, salt, seq, v))
	}

	return &Item{
		V:    value,
		Salt: salt,
		Cas:  cas,
		Seq:  seq,

		K:   kk,
		Sig: sig,
	}, nil
}

func (i *Item) Target() Target {
	if i.IsMutable() {
		return sha1.Sum(append(i.K[:], i.Salt...))
	}

	return sha1.Sum(bencode.MustMarshal(i.V))
}

func (i *Item) Modify(value interface{}, k ed25519.PrivateKey) bool {
	if !i.IsMutable() {
		return false
	}

	v, err := bencode.Marshal(value)
	if err != nil {
		return false
	}

	i.V = value
	i.Seq++
	var sig [64]byte
	copy(sig[:], Sign(k, i.Salt, i.Seq, v))
	i.Sig = sig

	return true
}

func (s *Item) IsMutable() bool {
	return s.K != Empty32ByteArray
}

func bufferToSign(salt, bv []byte, seq int64) []byte {
	var bts []byte
	if len(salt) != 0 {
		bts = append(bts, []byte("4:salt")...)
		x := bencode.MustMarshal(salt)
		bts = append(bts, x...)
	}
	bts = append(bts, []byte(fmt.Sprintf("3:seqi%de1:v", seq))...)
	bts = append(bts, bv...)
	return bts
}

func Check(i *Item) error {
	bv, err := bencode.Marshal(i.V)
	if err != nil {
		return err
	}

	if len(bv) > 1000 {
		return ErrValueFieldTooBig
	}

	if !i.IsMutable() {
		return nil
	}

	if len(i.Salt) > 64 {
		return ErrSaltFieldTooBig
	}

	if !Verify(i.K[:], i.Salt, i.Seq, bv, i.Sig[:]) {
		return ErrInvalidSignature
	}

	return nil
}

func CheckIncoming(stored, incoming *Item) error {
	// If the sequence number is equal, and the value is also the same,
	// the node SHOULD reset its timeout counter.
	if stored.Seq == incoming.Seq {
		if bytes.Equal(
			bencode.MustMarshal(stored.V),
			bencode.MustMarshal(incoming.V),
		) {
			return nil
		}
	}

	if stored.Seq >= incoming.Seq {
		return ErrSequenceNumberLessThanCurrent
	}

	// Cas should be ignored if not present
	if stored.Cas == 0 {
		return nil
	}

	if stored.Cas != incoming.Cas {
		return ErrCasHashMismatched
	}

	return nil
}
