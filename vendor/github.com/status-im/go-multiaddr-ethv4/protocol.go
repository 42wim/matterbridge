package ethv4

import (
	"errors"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"

	btcec "github.com/btcsuite/btcd/btcec/v2"
)

const (
	P_ETHv4 = 0x01EA
)

func init() {
	if err := ma.AddProtocol(
		ma.Protocol{
			Name:       "ethv4",
			Code:       P_ETHv4,
			VCode:      ma.CodeToVarint(P_ETHv4),
			Size:       312,
			Path:       false,
			Transcoder: TranscoderETHv4,
		}); err != nil {
		panic(err)
	}
}

var TranscoderETHv4 = ma.NewTranscoderFromFunctions(ethv4StB, ethv4BtS, func([]byte) error { return nil })

func ethv4StB(s string) ([]byte, error) {
	id, err := mh.FromB58String(s)
	if err != nil {
		return nil, err
	}
	return id, err
}

func ethv4BtS(b []byte) (string, error) {
	id, err := mh.Cast(b)
	if err != nil {
		return "", err
	}
	return id.B58String(), err
}

// PeerIDToNodeID casts peer.ID (b58 encoded string) to discover.NodeID
func PeerIDToNodeID(pid string) (n enode.ID, err error) {
	nodeid, err := peer.Decode(pid)
	if err != nil {
		return n, err
	}
	pubkey, err := nodeid.ExtractPublicKey()
	if err != nil {
		return n, err
	}
	seckey, ok := pubkey.(*crypto.Secp256k1PublicKey)
	if !ok {
		return n, errors.New("public key is not on the secp256k1 curve")
	}

	return enode.PubkeyToIDV4((*btcec.PublicKey)(seckey).ToECDSA()), nil
}
