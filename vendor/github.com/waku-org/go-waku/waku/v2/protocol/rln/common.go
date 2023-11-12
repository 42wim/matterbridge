package rln

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	rlnpb "github.com/waku-org/go-waku/waku/v2/protocol/rln/pb"
	"github.com/waku-org/go-zerokit-rln/rln"
	"google.golang.org/protobuf/proto"
)

type messageValidationResult int

const (
	validationError messageValidationResult = iota
	validMessage
	invalidMessage
	spamMessage
)

// the maximum clock difference between peers in seconds
const maxClockGapSeconds = 20

// maximum allowed gap between the epochs of messages' RateLimitProofs
const maxEpochGap = int64(maxClockGapSeconds / rln.EPOCH_UNIT_SECONDS)

// acceptable roots for merkle root validation of incoming messages
const acceptableRootWindowSize = 5

type RegistrationHandler = func(tx *types.Transaction)

type SpamHandler = func(msg *pb.WakuMessage, topic string) error

func toRLNSignal(wakuMessage *pb.WakuMessage) []byte {
	if wakuMessage == nil {
		return []byte{}
	}

	contentTopicBytes := []byte(wakuMessage.ContentTopic)
	return append(wakuMessage.Payload, contentTopicBytes...)
}

// Bytres2RateLimitProof converts a slice of bytes into a RateLimitProof instance
func BytesToRateLimitProof(data []byte) (*rln.RateLimitProof, error) {
	if data == nil {
		return nil, nil
	}

	rateLimitProof := &rlnpb.RateLimitProof{}
	err := proto.Unmarshal(data, rateLimitProof)
	if err != nil {
		return nil, err
	}

	result := &rln.RateLimitProof{
		Proof:         rln.ZKSNARK(rln.Bytes128(rateLimitProof.Proof)),
		MerkleRoot:    rln.MerkleNode(rln.Bytes32(rateLimitProof.MerkleRoot)),
		Epoch:         rln.Epoch(rln.Bytes32(rateLimitProof.Epoch)),
		ShareX:        rln.MerkleNode(rln.Bytes32(rateLimitProof.ShareX)),
		ShareY:        rln.MerkleNode(rln.Bytes32(rateLimitProof.ShareY)),
		Nullifier:     rln.Nullifier(rln.Bytes32(rateLimitProof.Nullifier)),
		RLNIdentifier: rln.RLNIdentifier(rln.Bytes32(rateLimitProof.RlnIdentifier)),
	}

	return result, nil
}
