package protocol

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

// Invitation represents a group chat invitation request from a user in the application layer, used for persistence, querying and
// signaling
type GroupChatInvitation struct {
	*protobuf.GroupChatInvitation

	// From is a public key of the author of the invitation request.
	From string `json:"from,omitempty"`

	// SigPubKey is the ecdsa encoded public key of the invitation author
	SigPubKey *ecdsa.PublicKey `json:"-"`
}

func NewGroupChatInvitation() *GroupChatInvitation {
	return &GroupChatInvitation{GroupChatInvitation: &protobuf.GroupChatInvitation{}}
}

// ID is the Keccak256() contatenation of From-ChatId
func (g *GroupChatInvitation) ID() string {
	return types.EncodeHex(crypto.Keccak256([]byte(fmt.Sprintf("%s%s", g.From, g.ChatId))))
}

// GetSigPubKey returns an ecdsa encoded public key
// this function is required to implement the ChatEntity interface
func (g *GroupChatInvitation) GetSigPubKey() *ecdsa.PublicKey {
	return g.SigPubKey
}

// GetProtoBuf returns the struct's embedded protobuf struct
// this function is required to implement the ChatEntity interface
func (g *GroupChatInvitation) GetProtobuf() proto.Message {
	return g.GroupChatInvitation
}

func (g *GroupChatInvitation) MarshalJSON() ([]byte, error) {
	item := struct {
		ID                  string                             `json:"id"`
		ChatID              string                             `json:"chatId,omitempty"`
		From                string                             `json:"from"`
		IntroductionMessage string                             `json:"introductionMessage,omitempty"`
		State               protobuf.GroupChatInvitation_State `json:"state,omitempty"`
	}{
		ID:                  g.ID(),
		ChatID:              g.ChatId,
		From:                g.From,
		IntroductionMessage: g.IntroductionMessage,
		State:               g.State,
	}

	return json.Marshal(item)
}
