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

package common

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// EventType used to define known waku events.
type EventType string

const (
	// EventEnvelopeSent fires when envelope was sent to a peer.
	EventEnvelopeSent EventType = "envelope.sent"

	// EventEnvelopeExpired fires when envelop expired
	EventEnvelopeExpired EventType = "envelope.expired"

	// EventEnvelopeReceived is sent once envelope was received from a peer.
	// EventEnvelopeReceived must be sent to the feed even if envelope was previously in the cache.
	// And event, ideally, should contain information about peer that sent envelope to us.
	EventEnvelopeReceived EventType = "envelope.received"

	// EventBatchAcknowledged is sent when batch of envelopes was acknowledged by a peer.
	EventBatchAcknowledged EventType = "batch.acknowledged"

	// EventEnvelopeAvailable fires when envelop is available for filters
	EventEnvelopeAvailable EventType = "envelope.available"

	// EventMailServerRequestSent fires when such request is sent.
	EventMailServerRequestSent EventType = "mailserver.request.sent"

	// EventMailServerRequestCompleted fires after mailserver sends all the requested messages
	EventMailServerRequestCompleted EventType = "mailserver.request.completed"

	// EventMailServerRequestExpired fires after mailserver the request TTL ends.
	// This event is independent and concurrent to EventMailServerRequestCompleted.
	// Request should be considered as expired only if expiry event was received first.
	EventMailServerRequestExpired EventType = "mailserver.request.expired"

	// EventMailServerEnvelopeArchived fires after an envelope has been archived
	EventMailServerEnvelopeArchived EventType = "mailserver.envelope.archived"

	// EventMailServerSyncFinished fires when the sync of messages is finished.
	EventMailServerSyncFinished EventType = "mailserver.sync.finished"
)

// EnvelopeEvent represents an envelope event.
type EnvelopeEvent struct {
	Event EventType
	Topic TopicType
	Hash  common.Hash
	Batch common.Hash
	Peer  enode.ID
	Data  interface{}
}
