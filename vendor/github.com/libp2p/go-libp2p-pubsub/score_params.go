package pubsub

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerScoreThresholds struct {
	// whether it is allowed to just set some params and not all of them.
	SkipAtomicValidation bool

	// GossipThreshold is the score threshold below which gossip propagation is suppressed;
	// should be negative.
	GossipThreshold float64

	// PublishThreshold is the score threshold below which we shouldn't publish when using flood
	// publishing (also applies to fanout and floodsub peers); should be negative and <= GossipThreshold.
	PublishThreshold float64

	// GraylistThreshold is the score threshold below which message processing is suppressed altogether,
	// implementing an effective gray list according to peer score; should be negative and <= PublishThreshold.
	GraylistThreshold float64

	// AcceptPXThreshold is the score threshold below which PX will be ignored; this should be positive
	// and limited to scores attainable by bootstrappers and other trusted nodes.
	AcceptPXThreshold float64

	// OpportunisticGraftThreshold is the median mesh score threshold before triggering opportunistic
	// grafting; this should have a small positive value.
	OpportunisticGraftThreshold float64
}

func (p *PeerScoreThresholds) validate() error {

	if !p.SkipAtomicValidation || p.PublishThreshold != 0 || p.GossipThreshold != 0 || p.GraylistThreshold != 0 {
		if p.GossipThreshold > 0 || isInvalidNumber(p.GossipThreshold) {
			return fmt.Errorf("invalid gossip threshold; it must be <= 0 and a valid number")
		}
		if p.PublishThreshold > 0 || p.PublishThreshold > p.GossipThreshold || isInvalidNumber(p.PublishThreshold) {
			return fmt.Errorf("invalid publish threshold; it must be <= 0 and <= gossip threshold and a valid number")
		}
		if p.GraylistThreshold > 0 || p.GraylistThreshold > p.PublishThreshold || isInvalidNumber(p.GraylistThreshold) {
			return fmt.Errorf("invalid graylist threshold; it must be <= 0 and <= publish threshold and a valid number")
		}
	}

	if !p.SkipAtomicValidation || p.AcceptPXThreshold != 0 {
		if p.AcceptPXThreshold < 0 || isInvalidNumber(p.AcceptPXThreshold) {
			return fmt.Errorf("invalid accept PX threshold; it must be >= 0 and a valid number")
		}
	}

	if !p.SkipAtomicValidation || p.OpportunisticGraftThreshold != 0 {
		if p.OpportunisticGraftThreshold < 0 || isInvalidNumber(p.OpportunisticGraftThreshold) {
			return fmt.Errorf("invalid opportunistic grafting threshold; it must be >= 0 and a valid number")
		}
	}

	return nil
}

type PeerScoreParams struct {
	// whether it is allowed to just set some params and not all of them.
	SkipAtomicValidation bool

	// Score parameters per topic.
	Topics map[string]*TopicScoreParams

	// Aggregate topic score cap; this limits the total contribution of topics towards a positive
	// score. It must be positive (or 0 for no cap).
	TopicScoreCap float64

	// P5: Application-specific peer scoring
	AppSpecificScore  func(p peer.ID) float64
	AppSpecificWeight float64

	// P6: IP-colocation factor.
	// The parameter has an associated counter which counts the number of peers with the same IP.
	// If the number of peers in the same IP exceeds IPColocationFactorThreshold, then the value
	// is the square of the difference, ie (PeersInSameIP - IPColocationThreshold)^2.
	// If the number of peers in the same IP is less than the threshold, then the value is 0.
	// The weight of the parameter MUST be negative, unless you want to disable for testing.
	// Note: In order to simulate many IPs in a managable manner when testing, you can set the weight to 0
	//       thus disabling the IP colocation penalty.
	IPColocationFactorWeight    float64
	IPColocationFactorThreshold int
	IPColocationFactorWhitelist []*net.IPNet

	// P7: behavioural pattern penalties.
	// This parameter has an associated counter which tracks misbehaviour as detected by the
	// router. The router currently applies penalties for the following behaviors:
	// - attempting to re-graft before the prune backoff time has elapsed.
	// - not following up in IWANT requests for messages advertised with IHAVE.
	//
	// The value of the parameter is the square of the counter over the threshold, which decays with
	// BehaviourPenaltyDecay.
	// The weight of the parameter MUST be negative (or zero to disable).
	BehaviourPenaltyWeight, BehaviourPenaltyThreshold, BehaviourPenaltyDecay float64

	// the decay interval for parameter counters.
	DecayInterval time.Duration

	// counter value below which it is considered 0.
	DecayToZero float64

	// time to remember counters for a disconnected peer.
	RetainScore time.Duration

	// time to remember a message delivery for. Default to global TimeCacheDuration if 0.
	SeenMsgTTL time.Duration
}

type TopicScoreParams struct {
	// whether it is allowed to just set some params and not all of them.
	SkipAtomicValidation bool

	// The weight of the topic.
	TopicWeight float64

	// P1: time in the mesh
	// This is the time the peer has been grafted in the mesh.
	// The value of the parameter is the time/TimeInMeshQuantum, capped by TimeInMeshCap.
	// The weight of the parameter MUST be positive (or zero to disable).
	TimeInMeshWeight  float64
	TimeInMeshQuantum time.Duration
	TimeInMeshCap     float64

	// P2: first message deliveries
	// This is the number of message deliveries in the topic.
	// The value of the parameter is a counter, decaying with FirstMessageDeliveriesDecay, and capped
	// by FirstMessageDeliveriesCap.
	// The weight of the parameter MUST be positive (or zero to disable).
	FirstMessageDeliveriesWeight, FirstMessageDeliveriesDecay float64
	FirstMessageDeliveriesCap                                 float64

	// P3: mesh message deliveries
	// This is the number of message deliveries in the mesh, within the MeshMessageDeliveriesWindow of
	// message validation; deliveries during validation also count and are retroactively applied
	// when validation succeeds.
	// This window accounts for the minimum time before a hostile mesh peer trying to game the score
	// could replay back a valid message we just sent them.
	// It effectively tracks first and near-first deliveries, i.e., a message seen from a mesh peer
	// before we have forwarded it to them.
	// The parameter has an associated counter, decaying with MeshMessageDeliveriesDecay.
	// If the counter exceeds the threshold, its value is 0.
	// If the counter is below the MeshMessageDeliveriesThreshold, the value is the square of
	// the deficit, ie (MessageDeliveriesThreshold - counter)^2
	// The penalty is only activated after MeshMessageDeliveriesActivation time in the mesh.
	// The weight of the parameter MUST be negative (or zero to disable).
	MeshMessageDeliveriesWeight, MeshMessageDeliveriesDecay      float64
	MeshMessageDeliveriesCap, MeshMessageDeliveriesThreshold     float64
	MeshMessageDeliveriesWindow, MeshMessageDeliveriesActivation time.Duration

	// P3b: sticky mesh propagation failures
	// This is a sticky penalty that applies when a peer gets pruned from the mesh with an active
	// mesh message delivery penalty.
	// The weight of the parameter MUST be negative (or zero to disable)
	MeshFailurePenaltyWeight, MeshFailurePenaltyDecay float64

	// P4: invalid messages
	// This is the number of invalid messages in the topic.
	// The value of the parameter is the square of the counter, decaying with
	// InvalidMessageDeliveriesDecay.
	// The weight of the parameter MUST be negative (or zero to disable).
	InvalidMessageDeliveriesWeight, InvalidMessageDeliveriesDecay float64
}

// peer score parameter validation
func (p *PeerScoreParams) validate() error {
	for topic, params := range p.Topics {
		err := params.validate()
		if err != nil {
			return fmt.Errorf("invalid score parameters for topic %s: %w", topic, err)
		}
	}

	if !p.SkipAtomicValidation || p.TopicScoreCap != 0 {
		// check that the topic score is 0 or something positive
		if p.TopicScoreCap < 0 || isInvalidNumber(p.TopicScoreCap) {
			return fmt.Errorf("invalid topic score cap; must be positive (or 0 for no cap) and a valid number")
		}
	}

	// check that we have an app specific score; the weight can be anything (but expected positive)
	if p.AppSpecificScore == nil {
		if p.SkipAtomicValidation {
			p.AppSpecificScore = func(p peer.ID) float64 {
				return 0
			}
		} else {
			return fmt.Errorf("missing application specific score function")
		}
	}

	if !p.SkipAtomicValidation || p.IPColocationFactorWeight != 0 {
		// check the IP collocation factor
		if p.IPColocationFactorWeight > 0 || isInvalidNumber(p.IPColocationFactorWeight) {
			return fmt.Errorf("invalid IPColocationFactorWeight; must be negative (or 0 to disable) and a valid number")
		}
		if p.IPColocationFactorWeight != 0 && p.IPColocationFactorThreshold < 1 {
			return fmt.Errorf("invalid IPColocationFactorThreshold; must be at least 1")
		}
	}

	// check the behaviour penalty
	if !p.SkipAtomicValidation || p.BehaviourPenaltyWeight != 0 || p.BehaviourPenaltyThreshold != 0 {
		if p.BehaviourPenaltyWeight > 0 || isInvalidNumber(p.BehaviourPenaltyWeight) {
			return fmt.Errorf("invalid BehaviourPenaltyWeight; must be negative (or 0 to disable) and a valid number")
		}
		if p.BehaviourPenaltyWeight != 0 && (p.BehaviourPenaltyDecay <= 0 || p.BehaviourPenaltyDecay >= 1 || isInvalidNumber(p.BehaviourPenaltyDecay)) {
			return fmt.Errorf("invalid BehaviourPenaltyDecay; must be between 0 and 1")
		}
		if p.BehaviourPenaltyThreshold < 0 || isInvalidNumber(p.BehaviourPenaltyThreshold) {
			return fmt.Errorf("invalid BehaviourPenaltyThreshold; must be >= 0 and a valid number")
		}
	}

	// check the decay parameters
	if !p.SkipAtomicValidation || p.DecayInterval != 0 || p.DecayToZero != 0 {
		if p.DecayInterval < time.Second {
			return fmt.Errorf("invalid DecayInterval; must be at least 1s")
		}
		if p.DecayToZero <= 0 || p.DecayToZero >= 1 || isInvalidNumber(p.DecayToZero) {
			return fmt.Errorf("invalid DecayToZero; must be between 0 and 1")
		}
	}

	// no need to check the score retention; a value of 0 means that we don't retain scores
	return nil
}

func (p *TopicScoreParams) validate() error {
	// make sure we have a sane topic weight
	if p.TopicWeight < 0 || isInvalidNumber(p.TopicWeight) {
		return fmt.Errorf("invalid topic weight; must be >= 0 and a valid number")
	}

	// check P1
	if err := p.validateTimeInMeshParams(); err != nil {
		return err
	}

	// check P2
	if err := p.validateMessageDeliveryParams(); err != nil {
		return err
	}
	// check P3
	if err := p.validateMeshMessageDeliveryParams(); err != nil {
		return err
	}

	// check P3b
	if err := p.validateMessageFailurePenaltyParams(); err != nil {
		return err
	}

	// check P4
	if err := p.validateInvalidMessageDeliveryParams(); err != nil {
		return err
	}

	return nil
}

func (p *TopicScoreParams) validateTimeInMeshParams() error {
	if p.SkipAtomicValidation {
		// in non-atomic mode, parameters at their zero values are dismissed from validation.
		if p.TimeInMeshWeight == 0 && p.TimeInMeshQuantum == 0 && p.TimeInMeshCap == 0 {
			return nil
		}
	}

	// either atomic validation mode, or some parameters have been set a value,
	// hence, proceed with normal validation of all related parameters in this context.

	if p.TimeInMeshQuantum == 0 {
		return fmt.Errorf("invalid TimeInMeshQuantum; must be non zero")
	}
	if p.TimeInMeshWeight < 0 || isInvalidNumber(p.TimeInMeshWeight) {
		return fmt.Errorf("invalid TimeInMeshWeight; must be positive (or 0 to disable) and a valid number")
	}
	if p.TimeInMeshWeight != 0 && p.TimeInMeshQuantum <= 0 {
		return fmt.Errorf("invalid TimeInMeshQuantum; must be positive")
	}
	if p.TimeInMeshWeight != 0 && (p.TimeInMeshCap <= 0 || isInvalidNumber(p.TimeInMeshCap)) {
		return fmt.Errorf("invalid TimeInMeshCap; must be positive and a valid number")
	}

	return nil
}

func (p *TopicScoreParams) validateMessageDeliveryParams() error {
	if p.SkipAtomicValidation {
		// in non-atomic mode, parameters at their zero values are dismissed from validation.
		if p.FirstMessageDeliveriesWeight == 0 && p.FirstMessageDeliveriesCap == 0 && p.FirstMessageDeliveriesDecay == 0 {
			return nil
		}
	}

	// either atomic validation mode, or some parameters have been set a value,
	// hence, proceed with normal validation of all related parameters in this context.

	if p.FirstMessageDeliveriesWeight < 0 || isInvalidNumber(p.FirstMessageDeliveriesWeight) {
		return fmt.Errorf("invallid FirstMessageDeliveriesWeight; must be positive (or 0 to disable) and a valid number")
	}
	if p.FirstMessageDeliveriesWeight != 0 && (p.FirstMessageDeliveriesDecay <= 0 || p.FirstMessageDeliveriesDecay >= 1 || isInvalidNumber(p.FirstMessageDeliveriesDecay)) {
		return fmt.Errorf("invalid FirstMessageDeliveriesDecay; must be between 0 and 1")
	}
	if p.FirstMessageDeliveriesWeight != 0 && (p.FirstMessageDeliveriesCap <= 0 || isInvalidNumber(p.FirstMessageDeliveriesCap)) {
		return fmt.Errorf("invalid FirstMessageDeliveriesCap; must be positive and a valid number")
	}

	return nil
}

func (p *TopicScoreParams) validateMeshMessageDeliveryParams() error {
	if p.SkipAtomicValidation {
		// in non-atomic mode, parameters at their zero values are dismissed from validation.
		if p.MeshMessageDeliveriesWeight == 0 &&
			p.MeshMessageDeliveriesCap == 0 &&
			p.MeshMessageDeliveriesDecay == 0 &&
			p.MeshMessageDeliveriesThreshold == 0 &&
			p.MeshMessageDeliveriesWindow == 0 &&
			p.MeshMessageDeliveriesActivation == 0 {
			return nil
		}
	}

	// either atomic validation mode, or some parameters have been set a value,
	// hence, proceed with normal validation of all related parameters in this context.

	if p.MeshMessageDeliveriesWeight > 0 || isInvalidNumber(p.MeshMessageDeliveriesWeight) {
		return fmt.Errorf("invalid MeshMessageDeliveriesWeight; must be negative (or 0 to disable) and a valid number")
	}
	if p.MeshMessageDeliveriesWeight != 0 && (p.MeshMessageDeliveriesDecay <= 0 || p.MeshMessageDeliveriesDecay >= 1 || isInvalidNumber(p.MeshMessageDeliveriesDecay)) {
		return fmt.Errorf("invalid MeshMessageDeliveriesDecay; must be between 0 and 1")
	}
	if p.MeshMessageDeliveriesWeight != 0 && (p.MeshMessageDeliveriesCap <= 0 || isInvalidNumber(p.MeshMessageDeliveriesCap)) {
		return fmt.Errorf("invalid MeshMessageDeliveriesCap; must be positive and a valid number")
	}
	if p.MeshMessageDeliveriesWeight != 0 && (p.MeshMessageDeliveriesThreshold <= 0 || isInvalidNumber(p.MeshMessageDeliveriesThreshold)) {
		return fmt.Errorf("invalid MeshMessageDeliveriesThreshold; must be positive and a valid number")
	}
	if p.MeshMessageDeliveriesWindow < 0 {
		return fmt.Errorf("invalid MeshMessageDeliveriesWindow; must be non-negative")
	}
	if p.MeshMessageDeliveriesWeight != 0 && p.MeshMessageDeliveriesActivation < time.Second {
		return fmt.Errorf("invalid MeshMessageDeliveriesActivation; must be at least 1s")
	}

	return nil
}

func (p *TopicScoreParams) validateMessageFailurePenaltyParams() error {
	if p.SkipAtomicValidation {
		// in selective mode, parameters at their zero values are dismissed from validation.
		if p.MeshFailurePenaltyDecay == 0 && p.MeshFailurePenaltyWeight == 0 {
			return nil
		}
	}

	// either atomic validation mode, or some parameters have been set a value,
	// hence, proceed with normal validation of all related parameters in this context.

	if p.MeshFailurePenaltyWeight > 0 || isInvalidNumber(p.MeshFailurePenaltyWeight) {
		return fmt.Errorf("invalid MeshFailurePenaltyWeight; must be negative (or 0 to disable) and a valid number")
	}
	if p.MeshFailurePenaltyWeight != 0 && (isInvalidNumber(p.MeshFailurePenaltyDecay) || p.MeshFailurePenaltyDecay <= 0 || p.MeshFailurePenaltyDecay >= 1) {
		return fmt.Errorf("invalid MeshFailurePenaltyDecay; must be between 0 and 1")
	}

	return nil
}

func (p *TopicScoreParams) validateInvalidMessageDeliveryParams() error {
	if p.SkipAtomicValidation {
		// in selective mode, parameters at their zero values are dismissed from validation.
		if p.InvalidMessageDeliveriesDecay == 0 && p.InvalidMessageDeliveriesWeight == 0 {
			return nil
		}
	}

	// either atomic validation mode, or some parameters have been set a value,
	// hence, proceed with normal validation of all related parameters in this context.

	if p.InvalidMessageDeliveriesWeight > 0 || isInvalidNumber(p.InvalidMessageDeliveriesWeight) {
		return fmt.Errorf("invalid InvalidMessageDeliveriesWeight; must be negative (or 0 to disable) and a valid number")
	}
	if p.InvalidMessageDeliveriesDecay <= 0 || p.InvalidMessageDeliveriesDecay >= 1 || isInvalidNumber(p.InvalidMessageDeliveriesDecay) {
		return fmt.Errorf("invalid InvalidMessageDeliveriesDecay; must be between 0 and 1")
	}

	return nil
}

const (
	DefaultDecayInterval = time.Second
	DefaultDecayToZero   = 0.01
)

// ScoreParameterDecay computes the decay factor for a parameter, assuming the DecayInterval is 1s
// and that the value decays to zero if it drops below 0.01
func ScoreParameterDecay(decay time.Duration) float64 {
	return ScoreParameterDecayWithBase(decay, DefaultDecayInterval, DefaultDecayToZero)
}

// ScoreParameterDecayWithBase computes the decay factor for a parameter using base as the DecayInterval
func ScoreParameterDecayWithBase(decay time.Duration, base time.Duration, decayToZero float64) float64 {
	// the decay is linear, so after n ticks the value is factor^n
	// so factor^n = decayToZero => factor = decayToZero^(1/n)
	ticks := float64(decay / base)
	return math.Pow(decayToZero, 1/ticks)
}

// checks whether the provided floating-point number is `Not a Number`
// or an infinite number.
func isInvalidNumber(num float64) bool {
	return math.IsNaN(num) || math.IsInf(num, 0)
}
