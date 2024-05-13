package types

import (
	"context"
	"crypto/ecdsa"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pborman/uuid"

	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/connection"
)

type ConnStatus struct {
	IsOnline   bool                  `json:"isOnline"`
	HasHistory bool                  `json:"hasHistory"`
	Peers      map[string]WakuV2Peer `json:"peers"`
}

type WakuV2Peer struct {
	Protocols []protocol.ID `json:"protocols"`
	Addresses []string      `json:"addresses"`
}

type ConnStatusSubscription struct {
	sync.RWMutex

	ID     string
	C      chan ConnStatus
	active bool
}

func NewConnStatusSubscription() *ConnStatusSubscription {
	return &ConnStatusSubscription{
		ID:     uuid.NewRandom().String(),
		C:      make(chan ConnStatus, 100),
		active: true,
	}
}

func (u *ConnStatusSubscription) Active() bool {
	u.RLock()
	defer u.RUnlock()
	return u.active
}

func (u *ConnStatusSubscription) Unsubscribe() {
	u.Lock()
	defer u.Unlock()
	close(u.C)
	u.active = false
}

func (u *ConnStatusSubscription) Send(s ConnStatus) bool {
	u.RLock()
	defer u.RUnlock()
	if !u.active {
		return false
	}
	u.C <- s
	return true
}

type WakuKeyManager interface {
	// GetPrivateKey retrieves the private key of the specified identity.
	GetPrivateKey(id string) (*ecdsa.PrivateKey, error)
	// AddKeyPair imports a asymmetric private key and returns a deterministic identifier.
	AddKeyPair(key *ecdsa.PrivateKey) (string, error)
	// DeleteKeyPair deletes the key with the specified ID if it exists.
	DeleteKeyPair(keyID string) bool
	// DeleteKeyPairs deletes all the keys
	DeleteKeyPairs() error
	AddSymKeyDirect(key []byte) (string, error)
	AddSymKeyFromPassword(password string) (string, error)
	DeleteSymKey(id string) bool
	GetSymKey(id string) ([]byte, error)
}

// Whisper represents a dark communication interface through the Ethereum
// network, using its very own P2P communication layer.
type Waku interface {
	PublicWakuAPI() PublicWakuAPI

	// Waku protocol version
	Version() uint

	// PeerCount
	PeerCount() int

	ListenAddresses() ([]string, error)

	Peers() map[string]WakuV2Peer

	StartDiscV5() error

	StopDiscV5() error

	SubscribeToPubsubTopic(topic string, optPublicKey *ecdsa.PublicKey) error

	UnsubscribeFromPubsubTopic(topic string) error

	StorePubsubTopicKey(topic string, privKey *ecdsa.PrivateKey) error

	RetrievePubsubTopicKey(topic string) (*ecdsa.PrivateKey, error)

	RemovePubsubTopicKey(topic string) error

	AddStorePeer(address string) (peer.ID, error)

	AddRelayPeer(address string) (peer.ID, error)

	DialPeer(address string) error

	DialPeerByID(peerID string) error

	DropPeer(peerID string) error

	SubscribeToConnStatusChanges() (*ConnStatusSubscription, error)

	// MinPow returns the PoW value required by this node.
	MinPow() float64
	// BloomFilter returns the aggregated bloom filter for all the topics of interest.
	// The nodes are required to send only messages that match the advertised bloom filter.
	// If a message does not match the bloom, it will tantamount to spam, and the peer will
	// be disconnected.
	BloomFilter() []byte

	// GetCurrentTime returns current time.
	GetCurrentTime() time.Time

	// GetPrivateKey retrieves the private key of the specified identity.
	GetPrivateKey(id string) (*ecdsa.PrivateKey, error)

	SubscribeEnvelopeEvents(events chan<- EnvelopeEvent) Subscription

	// AddKeyPair imports a asymmetric private key and returns a deterministic identifier.
	AddKeyPair(key *ecdsa.PrivateKey) (string, error)
	// DeleteKeyPair deletes the key with the specified ID if it exists.
	DeleteKeyPair(keyID string) bool
	AddSymKeyDirect(key []byte) (string, error)
	AddSymKeyFromPassword(password string) (string, error)
	DeleteSymKey(id string) bool
	GetSymKey(id string) ([]byte, error)
	MaxMessageSize() uint32

	GetStats() StatsSummary

	Subscribe(opts *SubscriptionOptions) (string, error)
	GetFilter(id string) Filter
	Unsubscribe(ctx context.Context, id string) error
	UnsubscribeMany(ids []string) error

	// RequestHistoricMessages sends a message with p2pRequestCode to a specific peer,
	// which is known to implement MailServer interface, and is supposed to process this
	// request and respond with a number of peer-to-peer messages (possibly expired),
	// which are not supposed to be forwarded any further.
	// The whisper protocol is agnostic of the format and contents of envelope.
	// A timeout of 0 never expires.
	RequestHistoricMessagesWithTimeout(peerID []byte, envelope Envelope, timeout time.Duration) error
	// SendMessagesRequest sends a MessagesRequest. This is an equivalent to RequestHistoricMessages
	// in terms of the functionality.
	SendMessagesRequest(peerID []byte, request MessagesRequest) error

	// RequestStoreMessages uses the WAKU2-STORE protocol to request historic messages
	RequestStoreMessages(ctx context.Context, peerID []byte, request MessagesRequest, processEnvelopes bool) (*StoreRequestCursor, int, error)

	// ProcessingP2PMessages indicates whether there are in-flight p2p messages
	ProcessingP2PMessages() bool

	// MarkP2PMessageAsProcessed tells the waku layer that a P2P message has been processed
	MarkP2PMessageAsProcessed(common.Hash)

	// ConnectionChanged is called whenever the client knows its connection status has changed
	ConnectionChanged(connection.State)

	// ClearEnvelopesCache clears waku envelopes cache
	ClearEnvelopesCache()
}
