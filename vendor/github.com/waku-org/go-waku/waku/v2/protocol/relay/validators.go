package relay

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/waku-org/go-waku/waku/v2/hash"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"go.uber.org/zap"
)

func msgHash(pubSubTopic string, msg *pb.WakuMessage) []byte {
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(msg.GetTimestamp()))

	var ephemeralByte byte
	if msg.GetEphemeral() {
		ephemeralByte = 1
	}

	return hash.SHA256(
		[]byte(pubSubTopic),
		msg.Payload,
		[]byte(msg.ContentTopic),
		timestampBytes,
		[]byte{ephemeralByte},
	)
}

type validatorFn = func(ctx context.Context, msg *pb.WakuMessage, topic string) bool

func (w *WakuRelay) RegisterDefaultValidator(fn validatorFn) {
	w.topicValidatorMutex.Lock()
	defer w.topicValidatorMutex.Unlock()
	w.defaultTopicValidators = append(w.defaultTopicValidators, fn)
}

func (w *WakuRelay) RegisterTopicValidator(topic string, fn validatorFn) {
	w.topicValidatorMutex.Lock()
	defer w.topicValidatorMutex.Unlock()

	w.topicValidators[topic] = append(w.topicValidators[topic], fn)
}

func (w *WakuRelay) RemoveTopicValidator(topic string) {
	w.topicValidatorMutex.Lock()
	defer w.topicValidatorMutex.Unlock()

	delete(w.topicValidators, topic)
}

func (w *WakuRelay) topicValidator(topic string) func(ctx context.Context, peerID peer.ID, message *pubsub.Message) bool {
	return func(ctx context.Context, peerID peer.ID, message *pubsub.Message) bool {
		msg, err := pb.Unmarshal(message.Data)
		if err != nil {
			return false
		}

		w.topicValidatorMutex.RLock()
		validators := w.topicValidators[topic]
		validators = append(validators, w.defaultTopicValidators...)
		w.topicValidatorMutex.RUnlock()
		exists := len(validators) > 0

		if exists {
			for _, v := range validators {
				if !v(ctx, msg, topic) {
					return false
				}
			}
		}

		return true
	}
}

// AddSignedTopicValidator registers a gossipsub validator for a topic which will check that messages Meta field contains a valid ECDSA signature for the specified pubsub topic. This is used as a DoS prevention mechanism
func (w *WakuRelay) AddSignedTopicValidator(topic string, publicKey *ecdsa.PublicKey) error {
	w.log.Info("adding validator to signed topic", zap.String("topic", topic), zap.String("publicKey", hex.EncodeToString(elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y))))

	fn := signedTopicBuilder(w.timesource, publicKey)

	w.RegisterTopicValidator(topic, fn)

	if !w.IsSubscribed(topic) {
		w.log.Warn("relay is not subscribed to signed topic", zap.String("topic", topic))
	}

	return nil
}

const messageWindowDuration = time.Minute * 5

func withinTimeWindow(t timesource.Timesource, msg *pb.WakuMessage) bool {
	if msg.GetTimestamp() == 0 {
		return false
	}

	now := t.Now()
	msgTime := time.Unix(0, msg.GetTimestamp())

	return now.Sub(msgTime).Abs() <= messageWindowDuration
}

func signedTopicBuilder(t timesource.Timesource, publicKey *ecdsa.PublicKey) validatorFn {
	publicKeyBytes := crypto.FromECDSAPub(publicKey)
	return func(ctx context.Context, msg *pb.WakuMessage, topic string) bool {
		if !withinTimeWindow(t, msg) {
			return false
		}

		msgHash := msgHash(topic, msg)
		signature := msg.Meta

		return secp256k1.VerifySignature(publicKeyBytes, msgHash, signature)
	}
}

// SignMessage adds an ECDSA signature to a WakuMessage as an opt-in mechanism for DoS prevention
func SignMessage(privKey *ecdsa.PrivateKey, msg *pb.WakuMessage, pubsubTopic string) error {
	msgHash := msgHash(pubsubTopic, msg)
	sign, err := secp256k1.Sign(msgHash, crypto.FromECDSA(privKey))
	if err != nil {
		return err
	}

	msg.Meta = sign[0:64] // Remove V
	return nil
}
