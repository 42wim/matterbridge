package protocol

import (
	"crypto/ecdsa"
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	utils "github.com/status-im/status-go/common"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/encryption"
	"github.com/status-im/status-go/protocol/encryption/multidevice"
	"github.com/status-im/status-go/protocol/encryption/sharedsecret"
	"github.com/status-im/status-go/protocol/protobuf"
)

// TransportLayer is the lowest layer and represents waku message.
type TransportLayer struct {
	// Payload as received from the transport layer
	Payload   []byte           `json:"-"`
	Hash      []byte           `json:"-"`
	SigPubKey *ecdsa.PublicKey `json:"-"`
	Dst       *ecdsa.PublicKey
	Message   *types.Message `json:"message"`
}

// EncryptionLayer handles optional encryption.
// It is not mandatory and can be omitted,
// also its presence does not guarantee encryption.
type EncryptionLayer struct {
	// Payload after having been processed by the encryption layer
	Payload         []byte `json:"-"`
	Installations   []*multidevice.Installation
	SharedSecrets   []*sharedsecret.Secret
	HashRatchetInfo []*encryption.HashRatchetInfo
}

// ApplicationLayer is the topmost layer and represents the application message.
type ApplicationLayer struct {
	// Payload after having been unwrapped from the application layer
	Payload   []byte                                   `json:"-"`
	ID        types.HexBytes                           `json:"id"`
	SigPubKey *ecdsa.PublicKey                         `json:"-"`
	Type      protobuf.ApplicationMetadataMessage_Type `json:"-"`
}

// StatusMessage encapsulates all layers of the protocol
type StatusMessage struct {
	TransportLayer   TransportLayer   `json:"transportLayer"`
	EncryptionLayer  EncryptionLayer  `json:"encryptionLayer"`
	ApplicationLayer ApplicationLayer `json:"applicationLayer"`
}

// Temporary JSON marshaling for those messages that are not yet processed
// by the go code
func (m *StatusMessage) MarshalJSON() ([]byte, error) {
	item := struct {
		ID        types.HexBytes `json:"id"`
		Payload   string         `json:"payload"`
		From      types.HexBytes `json:"from"`
		Timestamp uint32         `json:"timestamp"`
	}{
		ID:        m.ApplicationLayer.ID,
		Payload:   string(m.ApplicationLayer.Payload),
		Timestamp: m.TransportLayer.Message.Timestamp,
		From:      m.TransportLayer.Message.Sig,
	}
	return json.Marshal(item)
}

// SigPubKey returns the most important signature, from the application layer to transport
func (m *StatusMessage) SigPubKey() *ecdsa.PublicKey {
	if m.ApplicationLayer.SigPubKey != nil {
		return m.ApplicationLayer.SigPubKey
	}

	return m.TransportLayer.SigPubKey
}

func (m *StatusMessage) Clone() (*StatusMessage, error) {
	copy := &StatusMessage{}

	err := copier.Copy(&copy, m)
	return copy, err
}

func (m *StatusMessage) HandleTransportLayer(wakuMessage *types.Message) error {
	publicKey, err := crypto.UnmarshalPubkey(wakuMessage.Sig)
	if err != nil {
		return errors.Wrap(err, "failed to get signature")
	}

	m.TransportLayer.Message = wakuMessage
	m.TransportLayer.Hash = wakuMessage.Hash
	m.TransportLayer.SigPubKey = publicKey
	m.TransportLayer.Payload = wakuMessage.Payload

	if wakuMessage.Dst != nil {
		publicKey, err := crypto.UnmarshalPubkey(wakuMessage.Dst)
		if err != nil {
			return err
		}
		m.TransportLayer.Dst = publicKey
	}

	return nil
}

func (m *StatusMessage) HandleEncryptionLayer(myKey *ecdsa.PrivateKey, senderKey *ecdsa.PublicKey, enc *encryption.Protocol, skipNegotiation bool) error {
	// As we handle non-encrypted messages, we make sure that DecryptPayload
	// is set regardless of whether this step is successful
	m.EncryptionLayer.Payload = m.TransportLayer.Payload
	// Nothing to do
	if skipNegotiation {
		return nil
	}

	var protocolMessage encryption.ProtocolMessage
	err := proto.Unmarshal(m.TransportLayer.Payload, &protocolMessage)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal ProtocolMessage")
	}

	response, err := enc.HandleMessage(
		myKey,
		senderKey,
		&protocolMessage,
		m.TransportLayer.Hash,
	)

	if err == encryption.ErrHashRatchetGroupIDNotFound {

		if response != nil {
			m.EncryptionLayer.HashRatchetInfo = response.HashRatchetInfo
		}
		return err
	}

	if err != nil {
		return errors.Wrap(err, "failed to handle Encryption message")
	}

	m.EncryptionLayer.Payload = response.DecryptedMessage
	m.EncryptionLayer.Installations = response.Installations
	m.EncryptionLayer.SharedSecrets = response.SharedSecrets
	m.EncryptionLayer.HashRatchetInfo = response.HashRatchetInfo
	return nil
}

func (m *StatusMessage) HandleApplicationLayer() error {

	message, err := protobuf.Unmarshal(m.EncryptionLayer.Payload)
	if err != nil {
		return err
	}

	recoveredKey, err := utils.RecoverKey(message)
	if err != nil {
		return err
	}
	m.ApplicationLayer.SigPubKey = recoveredKey
	// Calculate ID using the wrapped record
	m.ApplicationLayer.ID = MessageID(recoveredKey, m.EncryptionLayer.Payload)
	log.Debug("calculated ID for envelope", "envelopeHash", hexutil.Encode(m.TransportLayer.Hash), "messageId", hexutil.Encode(m.ApplicationLayer.ID))

	m.ApplicationLayer.Payload = message.Payload
	m.ApplicationLayer.Type = message.Type
	return nil

}
