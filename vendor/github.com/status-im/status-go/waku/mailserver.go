// Copyright 2019 The Waku Library Authors.
//
// The Waku library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Waku library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty off
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Waku library. If not, see <http://www.gnu.org/licenses/>.
//
// This software uses the go-ethereum library, which is licensed
// under the GNU Lesser General Public Library, version 3 or any later.

package waku

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/status-im/status-go/waku/common"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	mailServerFailedPayloadPrefix = "ERROR="
	cursorSize                    = 36
)

// MailServer represents a mail server, capable of
// archiving the old messages for subsequent delivery
// to the peers. Any implementation must ensure that both
// functions are thread-safe. Also, they must return ASAP.
// DeliverMail should use p2pMessageCode for delivery,
// in order to bypass the expiry checks.
type MailServer interface {
	Archive(env *common.Envelope)
	DeliverMail(peerID []byte, request *common.Envelope) // DEPRECATED; use Deliver()
	Deliver(peerID []byte, request common.MessagesRequest)
}

// MailServerResponse is the response payload sent by the mailserver.
type MailServerResponse struct {
	LastEnvelopeHash gethcommon.Hash
	Cursor           []byte
	Error            error
}

func invalidResponseSizeError(size int) error {
	return fmt.Errorf("unexpected payload size: %d", size)
}

// CreateMailServerRequestCompletedPayload creates a payload representing
// a successful request to mailserver
func CreateMailServerRequestCompletedPayload(requestID, lastEnvelopeHash gethcommon.Hash, cursor []byte) []byte {
	payload := make([]byte, len(requestID))
	copy(payload, requestID[:])
	payload = append(payload, lastEnvelopeHash[:]...)
	payload = append(payload, cursor...)
	return payload
}

// CreateMailServerRequestFailedPayload creates a payload representing
// a failed request to a mailserver
func CreateMailServerRequestFailedPayload(requestID gethcommon.Hash, err error) []byte {
	payload := []byte(mailServerFailedPayloadPrefix)
	payload = append(payload, requestID[:]...)
	payload = append(payload, []byte(err.Error())...)
	return payload
}

// CreateMailServerEvent returns EnvelopeEvent with correct data
// if payload corresponds to any of the know mailserver events:
// * request completed successfully
// * request failed
// If the payload is unknown/unparseable, it returns `nil`
func CreateMailServerEvent(nodeID enode.ID, payload []byte) (*common.EnvelopeEvent, error) {
	if len(payload) < gethcommon.HashLength {
		return nil, invalidResponseSizeError(len(payload))
	}

	event, err := tryCreateMailServerRequestFailedEvent(nodeID, payload)
	if err != nil {
		return nil, err
	} else if event != nil {
		return event, nil
	}

	return tryCreateMailServerRequestCompletedEvent(nodeID, payload)
}

func tryCreateMailServerRequestFailedEvent(nodeID enode.ID, payload []byte) (*common.EnvelopeEvent, error) {
	if len(payload) < gethcommon.HashLength+len(mailServerFailedPayloadPrefix) {
		return nil, nil
	}

	prefix, remainder := extractPrefix(payload, len(mailServerFailedPayloadPrefix))

	if !bytes.Equal(prefix, []byte(mailServerFailedPayloadPrefix)) {
		return nil, nil
	}

	var (
		requestID gethcommon.Hash
		errorMsg  string
	)

	requestID, remainder = extractHash(remainder)
	errorMsg = string(remainder)

	event := common.EnvelopeEvent{
		Peer:  nodeID,
		Hash:  requestID,
		Event: common.EventMailServerRequestCompleted,
		Data: &MailServerResponse{
			Error: errors.New(errorMsg),
		},
	}

	return &event, nil

}

func tryCreateMailServerRequestCompletedEvent(nodeID enode.ID, payload []byte) (*common.EnvelopeEvent, error) {
	// check if payload is
	// - requestID or
	// - requestID + lastEnvelopeHash or
	// - requestID + lastEnvelopeHash + cursor
	// requestID is the hash of the request envelope.
	// lastEnvelopeHash is the last envelope sent by the mail server
	// cursor is the db key, 36 bytes: 4 for the timestamp + 32 for the envelope hash.
	if len(payload) > gethcommon.HashLength*2+cursorSize {
		return nil, invalidResponseSizeError(len(payload))
	}

	var (
		requestID        gethcommon.Hash
		lastEnvelopeHash gethcommon.Hash
		cursor           []byte
	)

	requestID, remainder := extractHash(payload)

	if len(remainder) >= gethcommon.HashLength {
		lastEnvelopeHash, remainder = extractHash(remainder)
	}

	if len(remainder) >= cursorSize {
		cursor = remainder
	}

	event := common.EnvelopeEvent{
		Peer:  nodeID,
		Hash:  requestID,
		Event: common.EventMailServerRequestCompleted,
		Data: &MailServerResponse{
			LastEnvelopeHash: lastEnvelopeHash,
			Cursor:           cursor,
		},
	}

	return &event, nil
}

func extractHash(payload []byte) (gethcommon.Hash, []byte) {
	prefix, remainder := extractPrefix(payload, gethcommon.HashLength)
	return gethcommon.BytesToHash(prefix), remainder
}

func extractPrefix(payload []byte, size int) ([]byte, []byte) {
	return payload[:size], payload[size:]
}
