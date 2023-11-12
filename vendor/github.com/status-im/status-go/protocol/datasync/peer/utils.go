package peer

import (
	"crypto/ecdsa"

	"github.com/status-im/mvds/state"

	"github.com/status-im/status-go/eth-node/crypto"
)

func PublicKeyToPeerID(k ecdsa.PublicKey) state.PeerID {
	var p state.PeerID
	copy(p[:], crypto.FromECDSAPub(&k))
	return p
}

func IDToPublicKey(p state.PeerID) (*ecdsa.PublicKey, error) {
	return crypto.UnmarshalPubkey(p[:])
}
