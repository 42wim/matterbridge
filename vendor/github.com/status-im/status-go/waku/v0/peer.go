package v0

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	mapset "github.com/deckarep/golang-set"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/status-im/status-go/eth-node/types"

	"github.com/status-im/status-go/waku/common"
)

// Peer is the implementation of the Peer interface and represents a remote Waku client with which the local host Waku
// instance exchanges data / messages.
type Peer struct {
	host    common.WakuHost
	rw      p2p.MsgReadWriter
	p2pPeer *p2p.Peer
	logger  *zap.Logger

	quit chan struct{}

	trusted        bool
	powRequirement float64
	// bloomMu is to allow thread safe access to
	// the bloom filter
	bloomMu     sync.Mutex
	bloomFilter []byte
	// topicInterestMu is to allow thread safe access to
	// the map of topic interests
	topicInterestMu sync.Mutex
	topicInterest   map[common.TopicType]bool
	// fullNode is used to indicate that the node will be accepting any
	// envelope. The opposite is an "empty node" , which is when
	// a bloom filter is all 0s or topic interest is an empty map (not nil).
	// In that case no envelope is accepted.
	fullNode             bool
	confirmationsEnabled bool
	rateLimitsMu         sync.Mutex
	rateLimits           common.RateLimits

	stats *common.StatsTracker

	known mapset.Set // Messages already known by the peer to avoid wasting bandwidth
}

func NewPeer(host common.WakuHost, p2pPeer *p2p.Peer, rw p2p.MsgReadWriter, logger *zap.Logger, stats *common.StatsTracker) common.Peer {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Peer{
		host:           host,
		p2pPeer:        p2pPeer,
		logger:         logger,
		rw:             rw,
		trusted:        false,
		powRequirement: 0.0,
		known:          mapset.NewSet(),
		quit:           make(chan struct{}),
		bloomFilter:    common.MakeFullNodeBloom(),
		fullNode:       true,
		stats:          stats,
	}
}

func (p *Peer) Start() error {
	if err := p.handshake(); err != nil {
		return err
	}
	go p.update()
	p.logger.Debug("starting peer", zap.String("peerID", types.EncodeHex(p.ID())))
	return nil
}

func (p *Peer) Stop() {
	close(p.quit)
	p.logger.Debug("stopping peer", zap.String("peerID", types.EncodeHex(p.ID())))
}

func (p *Peer) NotifyAboutPowRequirementChange(pow float64) error {
	i := math.Float64bits(pow)
	return p2p.Send(p.rw, statusUpdateCode, StatusOptions{PoWRequirement: &i})
}

func (p *Peer) NotifyAboutBloomFilterChange(bloom []byte) error {
	return p2p.Send(p.rw, statusUpdateCode, StatusOptions{BloomFilter: bloom})
}

func (p *Peer) NotifyAboutTopicInterestChange(topics []common.TopicType) error {
	return p2p.Send(p.rw, statusUpdateCode, StatusOptions{TopicInterest: topics})
}

func (p *Peer) SetPeerTrusted(trusted bool) {
	p.trusted = trusted
}

func (p *Peer) RequestHistoricMessages(envelope *common.Envelope) error {
	return p2p.Send(p.rw, p2pRequestCode, envelope)
}

func (p *Peer) SendMessagesRequest(request common.MessagesRequest) error {
	return p2p.Send(p.rw, p2pRequestCode, request)
}

func (p *Peer) SendHistoricMessageResponse(payload []byte) error {
	size, r, err := rlp.EncodeToReader(payload)
	if err != nil {
		return err
	}

	return p.rw.WriteMsg(p2p.Msg{Code: p2pRequestCompleteCode, Size: uint32(size), Payload: r})

}

func (p *Peer) SendP2PMessages(envelopes []*common.Envelope) error {
	return p2p.Send(p.rw, p2pMessageCode, envelopes)
}

func (p *Peer) SendRawP2PDirect(envelopes []rlp.RawValue) error {
	return p2p.Send(p.rw, p2pMessageCode, envelopes)
}

func (p *Peer) SetRWWriter(rw p2p.MsgReadWriter) {
	p.rw = rw
}

// Mark marks an envelope known to the peer so that it won't be sent back.
func (p *Peer) Mark(envelope *common.Envelope) {
	p.known.Add(envelope.Hash())
}

// Marked checks if an envelope is already known to the remote peer.
func (p *Peer) Marked(envelope *common.Envelope) bool {
	return p.known.Contains(envelope.Hash())
}

func (p *Peer) BloomFilter() []byte {
	p.bloomMu.Lock()
	defer p.bloomMu.Unlock()

	bloomFilterCopy := make([]byte, len(p.bloomFilter))
	copy(bloomFilterCopy, p.bloomFilter)
	return bloomFilterCopy
}

func (p *Peer) PoWRequirement() float64 {
	return p.powRequirement
}

func (p *Peer) ConfirmationsEnabled() bool {
	return p.confirmationsEnabled
}

// ID returns a peer's id
func (p *Peer) ID() []byte {
	id := p.p2pPeer.ID()
	return id[:]
}

func (p *Peer) EnodeID() enode.ID {
	return p.p2pPeer.ID()
}

func (p *Peer) IP() net.IP {
	return p.p2pPeer.Node().IP()
}

func (p *Peer) Run() error {
	logger := p.logger.Named("Run")

	for {
		// fetch the next packet
		packet, err := p.rw.ReadMsg()
		if err != nil {
			logger.Info("failed to read a message", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}

		if packet.Size > p.host.MaxMessageSize() {
			logger.Warn("oversize message received", zap.String("peerID", types.EncodeHex(p.ID())), zap.Uint32("size", packet.Size))
			return errors.New("oversize message received")
		}

		if err := p.handlePacket(packet); err != nil {
			logger.Warn("failed to handle packet message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
		}
		_ = packet.Discard()
	}
}

func (p *Peer) handlePacket(packet p2p.Msg) error {
	switch packet.Code {
	case messagesCode:
		if err := p.handleMessagesCode(packet); err != nil {
			p.logger.Warn("failed to handle messagesCode message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}
	case messageResponseCode:
		if err := p.handleMessageResponseCode(packet); err != nil {
			p.logger.Warn("failed to handle messageResponseCode message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}
	case batchAcknowledgedCode:
		if err := p.handleBatchAcknowledgeCode(packet); err != nil {
			p.logger.Warn("failed to handle batchAcknowledgedCode message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}
	case statusUpdateCode:
		if err := p.handleStatusUpdateCode(packet); err != nil {
			p.logger.Warn("failed to decode status update message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}
	case p2pMessageCode:
		if err := p.handleP2PMessageCode(packet); err != nil {
			p.logger.Warn("failed to decode direct message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}
	case p2pRequestCode:
		if err := p.handleP2PRequestCode(packet); err != nil {
			p.logger.Warn("failed to decode p2p request message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}
	case p2pRequestCompleteCode:
		if err := p.handleP2PRequestCompleteCode(packet); err != nil {
			p.logger.Warn("failed to decode p2p request complete message, peer will be disconnected", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
			return err
		}
	default:
		// New message common might be implemented in the future versions of Waku.
		// For forward compatibility, just ignore.
		p.logger.Debug("ignored packet with message code", zap.Uint64("code", packet.Code))
	}

	return nil
}

func (p *Peer) handleMessagesCode(packet p2p.Msg) error {
	// decode the contained envelopes
	data, err := ioutil.ReadAll(packet.Payload)
	if err != nil {
		common.EnvelopesRejectedCounter.WithLabelValues("failed_read").Inc()
		return fmt.Errorf("failed to read packet payload: %v", err)
	}

	var envelopes []*common.Envelope
	if err := rlp.DecodeBytes(data, &envelopes); err != nil {
		common.EnvelopesRejectedCounter.WithLabelValues("invalid_data").Inc()
		return fmt.Errorf("invalid payload: %v", err)
	}

	envelopeErrors, err := p.host.OnNewEnvelopes(envelopes, p)

	if p.host.ConfirmationsEnabled() {
		go p.sendConfirmation(data, envelopeErrors) // nolint: errcheck
	}

	return err
}

func (p *Peer) handleMessageResponseCode(packet p2p.Msg) error {
	var resp MultiVersionResponse
	if err := packet.Decode(&resp); err != nil {
		common.EnvelopesRejectedCounter.WithLabelValues("failed_read").Inc()
		return fmt.Errorf("invalid response message: %v", err)
	}
	if resp.Version != 1 {
		p.logger.Info("received unsupported version of MultiVersionResponse for messageResponseCode packet", zap.Uint("version", resp.Version))
		return nil
	}

	response, err := resp.DecodeResponse1()
	if err != nil {
		common.EnvelopesRejectedCounter.WithLabelValues("invalid_data").Inc()
		return fmt.Errorf("failed to decode response message: %v", err)
	}

	return p.host.OnMessagesResponse(response, p)
}

func (p *Peer) handleP2PRequestCode(packet p2p.Msg) error {
	// Must be processed if mail server is implemented. Otherwise ignore.
	if !p.host.Mailserver() {
		return nil
	}

	// Read all data as we will try to decode it possibly twice.
	data, err := ioutil.ReadAll(packet.Payload)
	if err != nil {
		return fmt.Errorf("invalid p2p request messages: %v", err)
	}
	r := bytes.NewReader(data)
	packet.Payload = r

	var requestDeprecated common.Envelope
	errDepReq := packet.Decode(&requestDeprecated)
	if errDepReq == nil {
		return p.host.OnDeprecatedMessagesRequest(&requestDeprecated, p)
	}
	p.logger.Info("failed to decode p2p request message (deprecated)", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(errDepReq))

	// As we failed to decode the request, let's set the offset
	// to the beginning and try decode it again.
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("invalid p2p request message: %v", err)
	}

	var request common.MessagesRequest
	errReq := packet.Decode(&request)
	if errReq == nil {
		return p.host.OnMessagesRequest(request, p)
	}
	p.logger.Info("failed to decode p2p request message", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(errReq))

	return errors.New("invalid p2p request message")
}

func (p *Peer) handleBatchAcknowledgeCode(packet p2p.Msg) error {
	var batchHash gethcommon.Hash
	if err := packet.Decode(&batchHash); err != nil {
		return fmt.Errorf("invalid batch ack message: %v", err)
	}
	return p.host.OnBatchAcknowledged(batchHash, p)
}

func (p *Peer) handleStatusUpdateCode(packet p2p.Msg) error {
	var StatusOptions StatusOptions
	err := packet.Decode(&StatusOptions)
	if err != nil {
		p.logger.Error("failed to decode status-options", zap.Error(err))
		common.EnvelopesRejectedCounter.WithLabelValues("invalid_settings_changed").Inc()
		return err
	}

	return p.setOptions(StatusOptions)

}

func (p *Peer) handleP2PMessageCode(packet p2p.Msg) error {
	// peer-to-peer message, sent directly to peer bypassing PoW checks, etc.
	// this message is not supposed to be forwarded to other peers, and
	// therefore might not satisfy the PoW, expiry and other requirements.
	// these messages are only accepted from the trusted peer.
	if !p.trusted {
		return nil
	}

	var (
		envelopes []*common.Envelope
		err       error
	)

	if err = packet.Decode(&envelopes); err != nil {
		return fmt.Errorf("invalid direct message payload: %v", err)
	}

	return p.host.OnNewP2PEnvelopes(envelopes)
}

func (p *Peer) handleP2PRequestCompleteCode(packet p2p.Msg) error {
	if !p.trusted {
		return nil
	}

	var payload []byte
	if err := packet.Decode(&payload); err != nil {
		return fmt.Errorf("invalid p2p request complete message: %v", err)
	}
	return p.host.OnP2PRequestCompleted(payload, p)
}

// sendConfirmation sends messageResponseCode and batchAcknowledgedCode messages.
func (p *Peer) sendConfirmation(data []byte, envelopeErrors []common.EnvelopeError) (err error) {
	batchHash := crypto.Keccak256Hash(data)
	err = p2p.Send(p.rw, messageResponseCode, NewMessagesResponse(batchHash, envelopeErrors))
	if err != nil {
		return
	}
	err = p2p.Send(p.rw, batchAcknowledgedCode, batchHash) // DEPRECATED
	return
}

// handshake sends the protocol initiation status message to the remote peer and
// verifies the remote status too.
func (p *Peer) handshake() error {
	// Send the handshake status message asynchronously
	errc := make(chan error, 1)
	opts := StatusOptionsFromHost(p.host)
	go func() {
		errc <- p2p.SendItems(p.rw, statusCode, Version, opts)
	}()

	// Fetch the remote status packet and verify protocol match
	packet, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if packet.Code != statusCode {
		return fmt.Errorf("p [%x] sent packet %x before status packet", p.ID(), packet.Code)
	}

	var (
		peerProtocolVersion uint64
		peerOptions         StatusOptions
	)
	s := rlp.NewStream(packet.Payload, uint64(packet.Size))
	if _, err := s.List(); err != nil {
		return fmt.Errorf("p [%x]: failed to decode status packet: %v", p.ID(), err)
	}
	// Validate protocol version.
	if err := s.Decode(&peerProtocolVersion); err != nil {
		return fmt.Errorf("p [%x]: failed to decode peer protocol version: %v", p.ID(), err)
	}
	if peerProtocolVersion != Version {
		return fmt.Errorf("p [%x]: protocol version mismatch %d != %d", p.ID(), peerProtocolVersion, Version)
	}
	// Decode and validate other status packet options.
	if err := s.Decode(&peerOptions); err != nil {
		return fmt.Errorf("p [%x]: failed to decode status options: %v", p.ID(), err)
	}
	if err := s.ListEnd(); err != nil {
		return fmt.Errorf("p [%x]: failed to decode status packet: %v", p.ID(), err)
	}
	if err := p.setOptions(peerOptions.WithDefaults()); err != nil {
		return fmt.Errorf("p [%x]: failed to set options: %v", p.ID(), err)
	}
	if err := <-errc; err != nil {
		return fmt.Errorf("p [%x] failed to send status packet: %v", p.ID(), err)
	}
	return nil
}

// update executes periodic operations on the peer, including message transmission
// and expiration.
func (p *Peer) update() {
	// Start the tickers for the updates
	expire := time.NewTicker(common.ExpirationCycle)
	transmit := time.NewTicker(common.TransmissionCycle)

	// Loop and transmit until termination is requested
	for {
		select {
		case <-expire.C:
			p.expire()

		case <-transmit.C:
			if err := p.broadcast(); err != nil {
				p.logger.Debug("broadcasting failed", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
				return
			}

		case <-p.quit:
			return
		}
	}
}

func (p *Peer) setOptions(peerOptions StatusOptions) error {

	p.logger.Debug("settings options", zap.String("peerID", types.EncodeHex(p.ID())), zap.Any("Options", peerOptions))

	if err := peerOptions.Validate(); err != nil {
		return fmt.Errorf("p [%x]: sent invalid options: %v", p.ID(), err)
	}
	// Validate and save peer's PoW.
	pow := peerOptions.PoWRequirementF()
	if pow != nil {
		if math.IsInf(*pow, 0) || math.IsNaN(*pow) || *pow < 0.0 {
			return fmt.Errorf("p [%x]: sent bad status message: invalid pow", p.ID())
		}
		p.powRequirement = *pow
	}

	if peerOptions.TopicInterest != nil {
		p.setTopicInterest(peerOptions.TopicInterest)
	} else if peerOptions.BloomFilter != nil {
		// Validate and save peer's bloom filters.
		bloom := peerOptions.BloomFilter
		bloomSize := len(bloom)
		if bloomSize != 0 && bloomSize != common.BloomFilterSize {
			return fmt.Errorf("p [%x] sent bad status message: wrong bloom filter size %d", p.ID(), bloomSize)
		}
		p.setBloomFilter(bloom)
	}

	if peerOptions.LightNodeEnabled != nil {
		// Validate and save other peer's options.
		if *peerOptions.LightNodeEnabled && p.host.LightClientMode() && p.host.LightClientModeConnectionRestricted() {
			return fmt.Errorf("p [%x] is useless: two light client communication restricted", p.ID())
		}
	}
	if peerOptions.ConfirmationsEnabled != nil {
		p.confirmationsEnabled = *peerOptions.ConfirmationsEnabled
	}
	if peerOptions.RateLimits != nil {
		p.setRateLimits(*peerOptions.RateLimits)
	}

	return nil
}

// expire iterates over all the known envelopes in the host and removes all
// expired (unknown) ones from the known list.
func (p *Peer) expire() {
	unmark := make(map[gethcommon.Hash]struct{})
	p.known.Each(func(v interface{}) bool {
		if !p.host.IsEnvelopeCached(v.(gethcommon.Hash)) {
			unmark[v.(gethcommon.Hash)] = struct{}{}
		}
		return true
	})
	// Dump all known but no longer cached
	for hash := range unmark {
		p.known.Remove(hash)
	}
}

// broadcast iterates over the collection of envelopes and transmits yet unknown
// ones over the network.
func (p *Peer) broadcast() error {
	envelopes := p.host.Envelopes()
	bundle := make([]*common.Envelope, 0, len(envelopes))
	for _, envelope := range envelopes {
		if !p.Marked(envelope) && envelope.PoW() >= p.powRequirement && p.topicOrBloomMatch(envelope) {
			bundle = append(bundle, envelope)
		}
	}

	if len(bundle) == 0 {
		return nil
	}

	batchHash, err := p.SendBundle(bundle)
	if err != nil {
		p.logger.Debug("failed to deliver envelopes", zap.String("peerID", types.EncodeHex(p.ID())), zap.Error(err))
		return err
	}

	// mark envelopes only if they were successfully sent
	for _, e := range bundle {
		p.Mark(e)
		event := common.EnvelopeEvent{
			Event: common.EventEnvelopeSent,
			Hash:  e.Hash(),
			Peer:  p.EnodeID(),
		}
		if p.confirmationsEnabled {
			event.Batch = batchHash
		}
		p.host.SendEnvelopeEvent(event)
	}
	p.logger.Debug("broadcasted bundles successfully", zap.String("peerID", types.EncodeHex(p.ID())), zap.Int("count", len(bundle)))
	return nil
}

func (p *Peer) SendBundle(bundle []*common.Envelope) (rst gethcommon.Hash, err error) {
	data, err := rlp.EncodeToBytes(bundle)
	if err != nil {
		return
	}
	err = p.rw.WriteMsg(p2p.Msg{
		Code:    messagesCode,
		Size:    uint32(len(data)),
		Payload: bytes.NewBuffer(data),
	})
	if err != nil {
		return
	}
	return crypto.Keccak256Hash(data), nil
}

func (p *Peer) setBloomFilter(bloom []byte) {
	p.bloomMu.Lock()
	defer p.bloomMu.Unlock()
	p.bloomFilter = bloom
	p.fullNode = common.IsFullNode(bloom)
	if p.fullNode && p.bloomFilter == nil {
		p.bloomFilter = common.MakeFullNodeBloom()
	}
	p.topicInterest = nil
}

func (p *Peer) setTopicInterest(topicInterest []common.TopicType) {
	p.topicInterestMu.Lock()
	defer p.topicInterestMu.Unlock()
	if topicInterest == nil {
		p.topicInterest = nil
		return
	}
	p.topicInterest = make(map[common.TopicType]bool)
	for _, topic := range topicInterest {
		p.topicInterest[topic] = true
	}
	p.fullNode = false
	p.bloomFilter = nil
}

func (p *Peer) setRateLimits(r common.RateLimits) {
	p.rateLimitsMu.Lock()
	p.rateLimits = r
	p.rateLimitsMu.Unlock()
}

// topicOrBloomMatch matches against topic-interest if topic interest
// is not nil. Otherwise it will match against the bloom-filter.
// If the bloom-filter is nil, or full, the node is considered a full-node
// and any envelope will be accepted. An empty topic-interest (but not nil)
// signals that we are not interested in any envelope.
func (p *Peer) topicOrBloomMatch(env *common.Envelope) bool {
	p.topicInterestMu.Lock()
	topicInterestMode := p.topicInterest != nil
	p.topicInterestMu.Unlock()

	if topicInterestMode {
		return p.topicInterestMatch(env)
	}
	return p.bloomMatch(env)
}

func (p *Peer) topicInterestMatch(env *common.Envelope) bool {
	p.topicInterestMu.Lock()
	defer p.topicInterestMu.Unlock()

	if p.topicInterest == nil {
		return false
	}

	return p.topicInterest[env.Topic]
}

func (p *Peer) bloomMatch(env *common.Envelope) bool {
	p.bloomMu.Lock()
	defer p.bloomMu.Unlock()
	return p.fullNode || common.BloomFilterMatch(p.bloomFilter, env.Bloom())
}
