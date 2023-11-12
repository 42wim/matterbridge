package common

import (
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
)

// Peer represents a remote Waku client with which the local host waku instance exchanges data / messages.
type Peer interface {
	// Start performs the handshake and initialize the broadcasting of messages
	Start() error
	Stop()
	// Run start the polling loop
	Run() error

	// NotifyAboutPowRequirementChange notifies the peer that POW for the host has changed
	NotifyAboutPowRequirementChange(float64) error
	// NotifyAboutBloomFilterChange notifies the peer that bloom filter for the host has changed
	NotifyAboutBloomFilterChange([]byte) error
	// NotifyAboutTopicInterestChange notifies the peer that topics for the host have changed
	NotifyAboutTopicInterestChange([]TopicType) error

	// SetPeerTrusted sets the value of trusted, meaning we will
	// allow p2p messages from them, which is necessary to interact
	// with mailservers.
	SetPeerTrusted(bool)
	// SetRWWriter sets the socket to read/write
	SetRWWriter(p2p.MsgReadWriter)

	RequestHistoricMessages(*Envelope) error
	SendMessagesRequest(MessagesRequest) error
	SendHistoricMessageResponse([]byte) error
	SendP2PMessages([]*Envelope) error
	SendRawP2PDirect([]rlp.RawValue) error

	SendBundle(bundle []*Envelope) (rst common.Hash, err error)

	// Mark marks an envelope known to the peer so that it won't be sent back.
	Mark(*Envelope)
	// Marked checks if an envelope is already known to the remote peer.
	Marked(*Envelope) bool

	ID() []byte
	IP() net.IP
	EnodeID() enode.ID

	PoWRequirement() float64
	BloomFilter() []byte
	ConfirmationsEnabled() bool
}

// WakuHost is the local instance of waku, which both interacts with remote clients
// (peers) and local clients (through RPC API)
type WakuHost interface {
	// HandlePeer handles the connection of a new peer
	HandlePeer(Peer, p2p.MsgReadWriter) error
	// MaxMessageSize returns the maximum accepted message size.
	MaxMessageSize() uint32
	// LightClientMode returns whether the host is running in light client mode
	LightClientMode() bool
	// Mailserver returns whether the host is running a mailserver
	Mailserver() bool
	// LightClientModeConnectionRestricted indicates that connection to light client in light client mode not allowed
	LightClientModeConnectionRestricted() bool
	// ConfirmationsEnabled returns true if message confirmations are enabled.
	ConfirmationsEnabled() bool
	// PacketRateLimits returns the current rate limits for the host
	PacketRateLimits() RateLimits
	// BytesRateLimits returns the current rate limits for the host
	BytesRateLimits() RateLimits
	// MinPow returns the MinPow for the host
	MinPow() float64
	// BloomFilterMode returns whether the host is using bloom filter
	BloomFilterMode() bool
	// BloomFilter returns the bloom filter for the host
	BloomFilter() []byte
	//TopicInterest returns the topics for the host
	TopicInterest() []TopicType
	// IsEnvelopeCached checks if envelope with specific hash has already been received and cached.
	IsEnvelopeCached(common.Hash) bool
	// Envelopes returns all the envelopes queued
	Envelopes() []*Envelope
	SendEnvelopeEvent(EnvelopeEvent) int
	// OnNewEnvelopes handles newly received envelopes from a peer
	OnNewEnvelopes([]*Envelope, Peer) ([]EnvelopeError, error)
	// OnNewP2PEnvelopes handles envelopes received though the P2P
	// protocol (i.e from a mailserver in most cases)
	OnNewP2PEnvelopes([]*Envelope) error
	// OnMessagesResponse handles when the peer receive a message response
	// from a mailserver
	OnMessagesResponse(MessagesResponse, Peer) error
	// OnMessagesRequest handles when the peer receive a message request
	// this only works if the peer is a mailserver
	OnMessagesRequest(MessagesRequest, Peer) error
	// OnDeprecatedMessagesRequest handles when the peer receive a message request
	// using the *Envelope format. Currently the only production client (status-mobile)
	// is exclusively using this one.
	OnDeprecatedMessagesRequest(*Envelope, Peer) error

	OnBatchAcknowledged(common.Hash, Peer) error
	OnP2PRequestCompleted([]byte, Peer) error
}
