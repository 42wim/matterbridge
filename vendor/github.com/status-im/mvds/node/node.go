// Package node contains node logic.
package node

// @todo this is a very rough implementation that needs cleanup

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/status-im/mvds/peers"
	"github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"
	"github.com/status-im/mvds/store"
	"github.com/status-im/mvds/transport"
)

// Mode represents the synchronization mode.
type Mode int

const (
	INTERACTIVE Mode = iota
	BATCH
)

// CalculateNextEpoch is a function used to calculate the next `SendEpoch` for a given message.
type CalculateNextEpoch func(count uint64, epoch int64) int64

// Node represents an MVDS node, it runs all the logic like sending and receiving protocol messages.
type Node struct {
	// This needs to be declared first: https://github.com/golang/go/issues/9959
	epoch  int64
	ctx    context.Context
	cancel context.CancelFunc

	store     store.MessageStore
	transport transport.Transport

	syncState state.SyncState

	peers peers.Persistence

	payloads payloads

	nextEpoch CalculateNextEpoch

	ID state.PeerID

	epochPersistence *epochSQLitePersistence
	mode             Mode

	subscription chan protobuf.Message

	logger *zap.Logger
}

func NewPersistentNode(
	db *sql.DB,
	st transport.Transport,
	id state.PeerID,
	mode Mode,
	nextEpoch CalculateNextEpoch,
	logger *zap.Logger,
) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = zap.NewNop()
	}

	node := Node{
		ID:               id,
		ctx:              ctx,
		cancel:           cancel,
		store:            store.NewPersistentMessageStore(db),
		transport:        st,
		peers:            peers.NewSQLitePersistence(db),
		syncState:        state.NewPersistentSyncState(db),
		payloads:         newPayloads(),
		epochPersistence: newEpochSQLitePersistence(db),
		nextEpoch:        nextEpoch,
		logger:           logger.With(zap.Namespace("mvds")),
		mode:             mode,
	}
	if currentEpoch, err := node.epochPersistence.Get(id); err != nil {
		return nil, err
	} else {
		node.epoch = currentEpoch
	}
	return &node, nil
}

func NewEphemeralNode(
	id state.PeerID,
	t transport.Transport,
	nextEpoch CalculateNextEpoch,
	currentEpoch int64,
	mode Mode,
	logger *zap.Logger,
) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Node{
		ID:        id,
		ctx:       ctx,
		cancel:    cancel,
		store:     store.NewDummyStore(),
		transport: t,
		syncState: state.NewSyncState(),
		peers:     peers.NewMemoryPersistence(),
		payloads:  newPayloads(),
		nextEpoch: nextEpoch,
		epoch:     currentEpoch,
		logger:    logger.With(zap.Namespace("mvds")),
		mode:      mode,
	}
}

// NewNode returns a new node.
func NewNode(
	ms store.MessageStore,
	st transport.Transport,
	ss state.SyncState,
	nextEpoch CalculateNextEpoch,
	currentEpoch int64,
	id state.PeerID,
	mode Mode,
	pp peers.Persistence,
	logger *zap.Logger,
) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Node{
		ctx:       ctx,
		cancel:    cancel,
		store:     ms,
		transport: st,
		syncState: ss,
		peers:     pp,
		payloads:  newPayloads(),
		nextEpoch: nextEpoch,
		ID:        id,
		epoch:     currentEpoch,
		logger:    logger.With(zap.Namespace("mvds")),
		mode:      mode,
	}
}

func (n *Node) CurrentEpoch() int64 {
	return atomic.LoadInt64(&n.epoch)
}

// Start listens for new messages received by the node and sends out those required every epoch.
func (n *Node) Start(duration time.Duration) {
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				n.logger.Info("Watch stopped")
				return
			default:
				p := n.transport.Watch()
				go n.onPayload(p.Sender, p.Payload)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-n.ctx.Done():
				n.logger.Info("Epoch processing stopped")
				return
			default:
				n.logger.Debug("Epoch processing", zap.String("node", hex.EncodeToString(n.ID[:4])), zap.Int64("epoch", n.epoch))
				time.Sleep(duration)
				err := n.sendMessages()
				if err != nil {
					n.logger.Error("Error sending messages.", zap.Error(err))
				}
				atomic.AddInt64(&n.epoch, 1)
				// When a persistent node is used, the epoch needs to be saved.
				if n.epochPersistence != nil {
					if err := n.epochPersistence.Set(n.ID, n.epoch); err != nil {
						n.logger.Error("Failed to persisten epoch", zap.Error(err))
					}
				}
			}
		}
	}()
}

// Stop message reading and epoch processing
func (n *Node) Stop() {
	n.logger.Info("Stopping node")
	n.Unsubscribe()
	n.cancel()
}

// Subscribe subscribes to incoming messages.
func (n *Node) Subscribe() chan protobuf.Message {
	n.subscription = make(chan protobuf.Message)
	return n.subscription
}

// Unsubscribe closes the listening channels
func (n *Node) Unsubscribe() {
	if n.subscription != nil {
		close(n.subscription)
	}
	n.subscription = nil
}

// AppendMessage sends a message to a given group.
func (n *Node) AppendMessage(groupID state.GroupID, data []byte) (state.MessageID, error) {
	m := protobuf.Message{
		GroupId:   groupID[:],
		Timestamp: time.Now().Unix(),
		Body:      data,
	}

	id := m.ID()

	peers, err := n.peers.GetByGroupID(groupID)
	if err != nil {
		return state.MessageID{}, fmt.Errorf("trying to send to unknown group %x", groupID[:4])
	}

	err = n.store.Add(&m)
	if err != nil {
		return state.MessageID{}, err
	}

	for _, p := range peers {
		t := state.OFFER
		if n.mode == BATCH {
			t = state.MESSAGE
		}

		n.insertSyncState(&groupID, id, p, t)
	}

	n.logger.Debug("Sending message",
		zap.String("node", hex.EncodeToString(n.ID[:4])),
		zap.String("groupID", hex.EncodeToString(groupID[:4])),
		zap.String("id", hex.EncodeToString(id[:4])))
	// @todo think about a way to insta trigger send messages when send was selected, we don't wanna wait for ticks here

	return id, nil
}

// RequestMessage adds a REQUEST record to the next payload for a given message ID.
func (n *Node) RequestMessage(group state.GroupID, id state.MessageID) error {
	peers, err := n.peers.GetByGroupID(group)
	if err != nil {
		return fmt.Errorf("trying to request from an unknown group %x", group[:4])
	}

	for _, p := range peers {
		exist, err := n.IsPeerInGroup(group, p)
		if err != nil {
			return err
		}

		if exist {
			continue
		}

		n.insertSyncState(&group, id, p, state.REQUEST)
	}

	return nil
}

// AddPeer adds a peer to a specific group making it a recipient of messages.
func (n *Node) AddPeer(group state.GroupID, id state.PeerID) error {
	return n.peers.Add(group, id)
}

// IsPeerInGroup checks whether a peer is in the specified group.
func (n *Node) IsPeerInGroup(g state.GroupID, p state.PeerID) (bool, error) {
	return n.peers.Exists(g, p)
}

func (n *Node) sendMessages() error {
	err := n.syncState.Map(n.epoch, func(s state.State) state.State {
		m := s.MessageID
		p := s.PeerID
		switch s.Type {
		case state.OFFER:
			n.payloads.AddOffers(p, m[:])
		case state.REQUEST:
			n.payloads.AddRequests(p, m[:])
			n.logger.Debug("sending REQUEST",
				zap.String("from", hex.EncodeToString(n.ID[:4])),
				zap.String("to", hex.EncodeToString(p[:4])),
				zap.String("messageID", hex.EncodeToString(m[:4])),
			)

		case state.MESSAGE:
			g := *s.GroupID
			//  TODO: Handle errors
			exist, err := n.IsPeerInGroup(g, p)
			if err != nil {
				return s
			}

			if !exist {
				return s
			}

			msg, err := n.store.Get(m)
			if err != nil {
				n.logger.Error("Failed to retreive message",
					zap.String("messageID", hex.EncodeToString(m[:4])),
					zap.Error(err),
				)

				return s
			}

			n.payloads.AddMessages(p, msg)
			n.logger.Debug("sending MESSAGE",
				zap.String("groupID", hex.EncodeToString(g[:4])),
				zap.String("from", hex.EncodeToString(n.ID[:4])),
				zap.String("to", hex.EncodeToString(p[:4])),
				zap.String("messageID", hex.EncodeToString(m[:4])),
			)

		}

		return n.updateSendEpoch(s)
	})

	if err != nil {
		n.logger.Error("error while mapping sync state", zap.Error(err))
		return err
	}

	return n.payloads.MapAndClear(func(peer state.PeerID, payload *protobuf.Payload) error {
		err := n.transport.Send(n.ID, peer, payload)
		if err != nil {
			n.logger.Error("error sending message", zap.Error(err))
			return err
		}
		return nil
	})

}

func (n *Node) onPayload(sender state.PeerID, payload *protobuf.Payload) {
	// Acks, Requests and Offers are all arrays of bytes as protobuf doesn't allow type aliases otherwise arrays of messageIDs would be nicer.
	if err := n.onAck(sender, payload.Acks); err != nil {
		n.logger.Error("error processing acks", zap.Error(err))
	}
	if err := n.onRequest(sender, payload.Requests); err != nil {
		n.logger.Error("error processing requests", zap.Error(err))
	}
	if err := n.onOffer(sender, payload.Offers); err != nil {
		n.logger.Error("error processing offers", zap.Error(err))
	}
	messageIds := n.onMessages(sender, payload.Messages)
	n.payloads.AddAcks(sender, messageIds)
}

func (n *Node) onOffer(sender state.PeerID, offers [][]byte) error {
	for _, raw := range offers {
		id := toMessageID(raw)
		n.logger.Debug("OFFER received",
			zap.String("from", hex.EncodeToString(sender[:4])),
			zap.String("to", hex.EncodeToString(n.ID[:4])),
			zap.String("messageID", hex.EncodeToString(id[:4])),
		)

		exist, err := n.store.Has(id)
		// @todo maybe ack?
		if err != nil {
			return err
		}

		if exist {
			continue
		}

		n.insertSyncState(nil, id, sender, state.REQUEST)
	}
	return nil
}

func (n *Node) onRequest(sender state.PeerID, requests [][]byte) error {
	for _, raw := range requests {
		id := toMessageID(raw)
		n.logger.Debug("REQUEST received",
			zap.String("from", hex.EncodeToString(sender[:4])),
			zap.String("to", hex.EncodeToString(n.ID[:4])),
			zap.String("messageID", hex.EncodeToString(id[:4])),
		)

		message, err := n.store.Get(id)
		if err != nil {
			return err
		}

		if message == nil {
			n.logger.Error("message does not exist", zap.String("messageID", hex.EncodeToString(id[:4])))
			continue
		}

		groupID := toGroupID(message.GroupId)

		exist, err := n.IsPeerInGroup(groupID, sender)
		if err != nil {
			return err
		}

		if !exist {
			n.logger.Error("peer is not in group",
				zap.String("groupID", hex.EncodeToString(groupID[:4])),
				zap.String("peer", hex.EncodeToString(sender[:4])),
			)
			continue
		}

		n.insertSyncState(&groupID, id, sender, state.MESSAGE)
	}

	return nil
}

func (n *Node) onAck(sender state.PeerID, acks [][]byte) error {
	for _, ack := range acks {
		id := toMessageID(ack)

		err := n.syncState.Remove(id, sender)
		if err != nil {
			n.logger.Error("Error while removing sync state.", zap.Error(err))
			return err
		}

		n.logger.Debug("ACK received",
			zap.String("from", hex.EncodeToString(sender[:4])),
			zap.String("to", hex.EncodeToString(n.ID[:4])),
			zap.String("messageID", hex.EncodeToString(id[:4])),
		)

	}
	return nil
}

func (n *Node) onMessages(sender state.PeerID, messages []*protobuf.Message) [][]byte {
	a := make([][]byte, 0)

	for _, m := range messages {
		groupID := toGroupID(m.GroupId)
		err := n.onMessage(sender, *m)
		if err != nil {
			n.logger.Error("Error processing message", zap.Error(err))
			continue
		}

		id := m.ID()
		n.logger.Debug("sending ACK",
			zap.String("groupID", hex.EncodeToString(groupID[:4])),
			zap.String("from", hex.EncodeToString(n.ID[:4])),
			zap.String("", hex.EncodeToString(sender[:4])),
			zap.String("messageID", hex.EncodeToString(id[:4])),
		)

		a = append(a, id[:])
	}

	return a
}

func (n *Node) onMessage(sender state.PeerID, msg protobuf.Message) error {
	id := msg.ID()
	groupID := toGroupID(msg.GroupId)
	n.logger.Debug("MESSAGE received",
		zap.String("from", hex.EncodeToString(sender[:4])),
		zap.String("to", hex.EncodeToString(n.ID[:4])),
		zap.String("messageID", hex.EncodeToString(id[:4])),
	)

	err := n.syncState.Remove(id, sender)
	if err != nil && err != state.ErrStateNotFound {
		return err
	}

	err = n.store.Add(&msg)
	if err != nil {
		return err
		// @todo process, should this function ever even have an error?
	}

	peers, err := n.peers.GetByGroupID(groupID)
	if err != nil {
		return err
	}

	for _, peer := range peers {
		if peer == sender {
			continue
		}

		n.insertSyncState(&groupID, id, peer, state.OFFER)
	}

	if n.subscription != nil {
		n.subscription <- msg
	}

	return nil
}

func (n *Node) insertSyncState(groupID *state.GroupID, messageID state.MessageID, peerID state.PeerID, t state.RecordType) {
	s := state.State{
		GroupID:   groupID,
		MessageID: messageID,
		PeerID:    peerID,
		Type:      t,
		SendEpoch: n.epoch + 1,
	}

	err := n.syncState.Add(s)
	if err != nil {
		n.logger.Error("error setting sync states",
			zap.Error(err),
			zap.String("groupID", hex.EncodeToString(groupID[:4])),
			zap.String("messageID", hex.EncodeToString(messageID[:4])),
			zap.String("peerID", hex.EncodeToString(peerID[:4])),
		)

	}
}

func (n *Node) updateSendEpoch(s state.State) state.State {
	s.SendCount += 1
	s.SendEpoch = n.nextEpoch(s.SendCount, n.epoch)
	return s
}

func toMessageID(b []byte) state.MessageID {
	var id state.MessageID
	copy(id[:], b)
	return id
}

func toGroupID(b []byte) state.GroupID {
	var id state.GroupID
	copy(id[:], b)
	return id
}
