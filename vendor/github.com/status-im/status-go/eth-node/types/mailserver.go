package types

import (
	"time"
)

const (
	// MaxLimitInMessagesRequest represents the maximum number of messages
	// that can be requested from the mailserver
	MaxLimitInMessagesRequest = 1000
)

// MessagesRequest contains details of a request of historic messages.
type MessagesRequest struct {
	// ID of the request. The current implementation requires ID to be 32-byte array,
	// however, it's not enforced for future implementation.
	ID []byte `json:"id"`
	// From is a lower bound of time range.
	From uint32 `json:"from"`
	// To is a upper bound of time range.
	To uint32 `json:"to"`
	// Limit determines the number of messages sent by the mail server
	// for the current paginated request.
	Limit uint32 `json:"limit"`
	// Cursor is used as starting point for paginated requests.
	Cursor []byte `json:"cursor"`
	// StoreCursor is used as starting point for WAKUV2 paginatedRequests
	StoreCursor *StoreRequestCursor `json:"storeCursor"`
	// Bloom is a filter to match requested messages.
	Bloom []byte `json:"bloom"`
	// PubsubTopic is the gossipsub topic on which the message was broadcasted
	PubsubTopic string `json:"pubsubTopic"`
	// ContentTopics is a list of topics. A returned message should
	// belong to one of the topics from the list.
	ContentTopics [][]byte `json:"contentTopics"`
}

type StoreRequestCursor struct {
	Digest       []byte `json:"digest"`
	ReceiverTime int64  `json:"receiverTime"`
	SenderTime   int64  `json:"senderTime"`
	PubsubTopic  string `json:"pubsubTopic"`
}

// SetDefaults sets the From and To defaults
func (r *MessagesRequest) SetDefaults(now time.Time) {
	// set From and To defaults
	if r.To == 0 {
		r.To = uint32(now.UTC().Unix())
	}

	if r.From == 0 {
		oneDay := uint32(86400) // -24 hours
		if r.To < oneDay {
			r.From = 0
		} else {
			r.From = r.To - oneDay
		}
	}
}

// MailServerResponse is the response payload sent by the mailserver.
type MailServerResponse struct {
	LastEnvelopeHash Hash
	Cursor           []byte
	Error            error
}

// SyncMailRequest contains details which envelopes should be synced
// between Mail Servers.
type SyncMailRequest struct {
	// Lower is a lower bound of time range for which messages are requested.
	Lower uint32
	// Upper is a lower bound of time range for which messages are requested.
	Upper uint32
	// Bloom is a bloom filter to filter envelopes.
	Bloom []byte
	// Limit is the max number of envelopes to return.
	Limit uint32
	// Cursor is used for pagination of the results.
	Cursor []byte
}

// SyncEventResponse is a response from the Mail Server
// form which the peer received envelopes.
type SyncEventResponse struct {
	Cursor []byte
	Error  string
}
