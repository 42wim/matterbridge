package rln

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/rln/group_manager"
	rlnpb "github.com/waku-org/go-waku/waku/v2/protocol/rln/pb"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"github.com/waku-org/go-zerokit-rln/rln"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type WakuRLNRelay struct {
	timesource timesource.Timesource
	metrics    Metrics

	group_manager.Details

	nullifierLog *NullifierLog

	log *zap.Logger
}

const rlnDefaultTreePath = "./rln_tree.db"

func GetRLNInstanceAndRootTracker(treePath string) (*rln.RLN, *group_manager.MerkleRootTracker, error) {
	if treePath == "" {
		treePath = rlnDefaultTreePath
	}

	rlnInstance, err := rln.NewWithConfig(rln.DefaultTreeDepth, &rln.TreeConfig{
		CacheCapacity: 15000,
		Mode:          rln.HighThroughput,
		Compression:   false,
		FlushInterval: 500 * time.Millisecond,
		Path:          treePath,
	})
	if err != nil {
		return nil, nil, err
	}

	rootTracker := group_manager.NewMerkleRootTracker(acceptableRootWindowSize, rlnInstance)

	return rlnInstance, rootTracker, nil
}
func New(
	Details group_manager.Details,
	timesource timesource.Timesource,
	reg prometheus.Registerer,
	log *zap.Logger) *WakuRLNRelay {

	// create the WakuRLNRelay
	rlnPeer := &WakuRLNRelay{
		Details:    Details,
		metrics:    newMetrics(reg),
		log:        log,
		timesource: timesource,
	}

	return rlnPeer
}

func (rlnRelay *WakuRLNRelay) Start(ctx context.Context) error {
	rlnRelay.nullifierLog = NewNullifierLog(ctx, rlnRelay.log)

	err := rlnRelay.GroupManager.Start(ctx)
	if err != nil {
		return err
	}

	log.Info("rln relay topic validator mounted")

	return nil
}

// Stop will stop any operation or goroutine started while using WakuRLNRelay
func (rlnRelay *WakuRLNRelay) Stop() error {
	return rlnRelay.GroupManager.Stop()
}

// ValidateMessage validates the supplied message based on the waku-rln-relay routing protocol i.e.,
// the message's epoch is within `maxEpochGap` of the current epoch
// the message's has valid rate limit proof
// the message's does not violate the rate limit
// if `optionalTime` is supplied, then the current epoch is calculated based on that, otherwise the current time will be used
func (rlnRelay *WakuRLNRelay) ValidateMessage(msg *pb.WakuMessage, optionalTime *time.Time) (messageValidationResult, error) {
	if msg == nil {
		return validationError, errors.New("nil message")
	}

	//  checks if the `msg`'s epoch is far from the current epoch
	// it corresponds to the validation of rln external nullifier
	var epoch rln.Epoch
	if optionalTime != nil {
		epoch = rln.CalcEpoch(*optionalTime)
	} else {
		// get current rln epoch
		epoch = rln.CalcEpoch(rlnRelay.timesource.Now())
	}

	msgProof, err := BytesToRateLimitProof(msg.RateLimitProof)
	if err != nil {
		rlnRelay.log.Debug("invalid message: could not extract proof", zap.Error(err))
		rlnRelay.metrics.RecordInvalidMessage(proofExtractionErr)
	}

	if msgProof == nil {
		// message does not contain a proof
		rlnRelay.log.Debug("invalid message: message does not contain a proof")
		rlnRelay.metrics.RecordInvalidMessage(invalidNoProof)
		return invalidMessage, nil
	}

	proofMD, err := rlnRelay.RLN.ExtractMetadata(*msgProof)
	if err != nil {
		rlnRelay.log.Debug("could not extract metadata", zap.Error(err))
		rlnRelay.metrics.RecordError(proofMetadataExtractionErr)
		return invalidMessage, nil
	}

	// calculate the gaps and validate the epoch
	gap := rln.Diff(epoch, msgProof.Epoch)
	if int64(math.Abs(float64(gap))) > maxEpochGap {
		// message's epoch is too old or too ahead
		// accept messages whose epoch is within +-MAX_EPOCH_GAP from the current epoch
		rlnRelay.log.Debug("invalid message: epoch gap exceeds a threshold", zap.Int64("gap", gap))
		rlnRelay.metrics.RecordInvalidMessage(invalidEpoch)

		return invalidMessage, nil
	}

	if !(rlnRelay.RootTracker.ContainsRoot(msgProof.MerkleRoot)) {
		rlnRelay.log.Debug("invalid message: unexpected root", logging.HexBytes("msgRoot", msgProof.MerkleRoot[:]))
		rlnRelay.metrics.RecordInvalidMessage(invalidRoot)
		return invalidMessage, nil
	}

	start := time.Now()
	valid, err := rlnRelay.verifyProof(msg, msgProof)
	if err != nil {
		rlnRelay.log.Debug("could not verify proof", zap.Error(err))
		rlnRelay.metrics.RecordError(proofVerificationErr)
		return invalidMessage, nil
	}
	rlnRelay.metrics.RecordProofVerification(time.Since(start))

	if !valid {
		// invalid proof
		rlnRelay.log.Debug("Invalid proof")
		rlnRelay.metrics.RecordInvalidMessage(invalidProof)
		return invalidMessage, nil
	}

	// check if double messaging has happened
	hasDup, err := rlnRelay.nullifierLog.HasDuplicate(proofMD)
	if err != nil {
		rlnRelay.log.Debug("validation error", zap.Error(err))
		rlnRelay.metrics.RecordError(duplicateCheckErr)
		return validationError, err
	}

	if hasDup {
		rlnRelay.log.Debug("spam received")
		return spamMessage, nil
	}

	err = rlnRelay.nullifierLog.Insert(proofMD)
	if err != nil {
		rlnRelay.log.Debug("could not insert proof into log")
		rlnRelay.metrics.RecordError(logInsertionErr)
		return validationError, err
	}

	rlnRelay.log.Debug("message is valid")

	rootIndex := rlnRelay.RootTracker.IndexOf(msgProof.MerkleRoot)
	rlnRelay.metrics.RecordValidMessages(rootIndex)

	return validMessage, nil
}

func (rlnRelay *WakuRLNRelay) verifyProof(msg *pb.WakuMessage, proof *rln.RateLimitProof) (bool, error) {
	contentTopicBytes := []byte(msg.ContentTopic)
	input := append(msg.Payload, contentTopicBytes...)
	return rlnRelay.RLN.Verify(input, *proof, rlnRelay.RootTracker.Roots()...)
}

func (rlnRelay *WakuRLNRelay) AppendRLNProof(msg *pb.WakuMessage, senderEpochTime time.Time) error {
	// returns error if it could not create and append a `RateLimitProof` to the supplied `msg`
	// `senderEpochTime` indicates the number of seconds passed since Unix epoch. The fractional part holds sub-seconds.
	// The `epoch` field of `RateLimitProof` is derived from the provided `senderEpochTime` (using `calcEpoch()`)

	if msg == nil {
		return errors.New("nil message")
	}

	input := toRLNSignal(msg)

	start := time.Now()
	proof, err := rlnRelay.generateProof(input, rln.CalcEpoch(senderEpochTime))
	if err != nil {
		return err
	}
	rlnRelay.metrics.RecordProofGeneration(time.Since(start))

	b, err := proto.Marshal(proof)
	if err != nil {
		return err
	}

	msg.RateLimitProof = b
	//If msgTimeStamp is not set, then set it to timestamp of proof
	if msg.Timestamp == nil {
		msg.Timestamp = proto.Int64(senderEpochTime.Unix())
	}

	return nil
}

// Validator returns a validator for the waku messages.
// The message validation logic is according to https://rfc.vac.dev/spec/17/
func (rlnRelay *WakuRLNRelay) Validator(
	spamHandler SpamHandler) func(ctx context.Context, msg *pb.WakuMessage, topic string) bool {
	return func(ctx context.Context, msg *pb.WakuMessage, topic string) bool {

		hash := msg.Hash(topic)

		log := rlnRelay.log.With(
			logging.HexBytes("hash", hash),
			zap.String("pubsubTopic", topic),
			zap.String("contentTopic", msg.ContentTopic),
		)

		log.Debug("rln-relay topic validator called")

		rlnRelay.metrics.RecordMessage()

		// validate the message
		validationRes, err := rlnRelay.ValidateMessage(msg, nil)
		if err != nil {
			log.Debug("validating message", zap.Error(err))
			return false
		}

		switch validationRes {
		case validMessage:
			log.Debug("message verified")
			return true
		case invalidMessage:
			log.Debug("message could not be verified")
			return false
		case spamMessage:
			log.Debug("spam message found")

			rlnRelay.metrics.RecordSpam(msg.ContentTopic)

			if spamHandler != nil {
				if err := spamHandler(msg, topic); err != nil {
					log.Error("executing spam handler", zap.Error(err))
				}
			}

			return false
		default:
			log.Debug("unhandled validation result", zap.Int("validationResult", int(validationRes)))
			return false
		}
	}
}

func (rlnRelay *WakuRLNRelay) generateProof(input []byte, epoch rln.Epoch) (*rlnpb.RateLimitProof, error) {
	identityCredentials, err := rlnRelay.GroupManager.IdentityCredentials()
	if err != nil {
		return nil, err
	}

	membershipIndex := rlnRelay.GroupManager.MembershipIndex()

	proof, err := rlnRelay.RLN.GenerateProof(input, identityCredentials, membershipIndex, epoch)
	if err != nil {
		return nil, err
	}

	return &rlnpb.RateLimitProof{
		Proof:         proof.Proof[:],
		MerkleRoot:    proof.MerkleRoot[:],
		Epoch:         proof.Epoch[:],
		ShareX:        proof.ShareX[:],
		ShareY:        proof.ShareY[:],
		Nullifier:     proof.Nullifier[:],
		RlnIdentifier: proof.RLNIdentifier[:],
	}, nil
}

func (rlnRelay *WakuRLNRelay) IdentityCredential() (rln.IdentityCredential, error) {
	return rlnRelay.GroupManager.IdentityCredentials()
}

func (rlnRelay *WakuRLNRelay) MembershipIndex() uint {
	return rlnRelay.GroupManager.MembershipIndex()
}

// IsReady returns true if the RLN Relay protocol is ready to relay messages
func (rlnRelay *WakuRLNRelay) IsReady(ctx context.Context) (bool, error) {
	return rlnRelay.GroupManager.IsReady(ctx)
}
