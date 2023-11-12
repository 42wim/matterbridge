// TODO: These types should be defined using protobuf, but protoc can only emit []byte instead of types.HexBytes,
// which causes issues when marshaling to JSON on the react side. Let's do that once the chat protocol is moved to the go repo.

package ext

import (
	"crypto/ecdsa"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
)

// SendPublicMessageRPC represents the RPC payload for the SendPublicMessage RPC method
type SendPublicMessageRPC struct {
	Sig     string // TODO: remove
	Chat    string
	Payload types.HexBytes
}

// TODO: implement with accordance to https://github.com/status-im/status-go/protocol/issues/28.
func (m SendPublicMessageRPC) ID() string { return m.Chat }

func (m SendPublicMessageRPC) PublicName() string { return m.Chat }

func (m SendPublicMessageRPC) PublicKey() *ecdsa.PublicKey { return nil }

// SendDirectMessageRPC represents the RPC payload for the SendDirectMessage RPC method
type SendDirectMessageRPC struct {
	Sig     string // TODO: remove
	Chat    string
	Payload types.HexBytes
	PubKey  types.HexBytes
	DH      bool // TODO: make sure to remove safely
}

// TODO: implement with accordance to https://github.com/status-im/status-go/protocol/issues/28.
func (m SendDirectMessageRPC) ID() string { return "" }

func (m SendDirectMessageRPC) PublicName() string { return "" }

func (m SendDirectMessageRPC) PublicKey() *ecdsa.PublicKey {
	publicKey, _ := crypto.UnmarshalPubkey(m.PubKey)
	return publicKey
}

type JoinRPC struct {
	Chat    string
	PubKey  types.HexBytes
	Payload types.HexBytes
}

func (m JoinRPC) ID() string { return m.Chat }

func (m JoinRPC) PublicName() string {
	if len(m.PubKey) > 0 {
		return ""
	}
	return m.Chat
}

func (m JoinRPC) PublicKey() *ecdsa.PublicKey {
	if len(m.PubKey) > 0 {
		return nil
	}
	publicKey, _ := crypto.UnmarshalPubkey(m.PubKey)
	return publicKey
}
