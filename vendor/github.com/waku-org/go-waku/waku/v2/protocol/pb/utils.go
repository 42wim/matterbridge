package pb

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/waku-org/go-waku/waku/v2/hash"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Hash calculates the hash of a waku message
func (msg *WakuMessage) Hash(pubsubTopic string) []byte {
	return hash.SHA256([]byte(pubsubTopic), msg.Payload, []byte(msg.ContentTopic), msg.Meta, toBytes(msg.GetTimestamp()))
}

func toBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

func (msg *WakuMessage) LogFields(pubsubTopic string) []zapcore.Field {
	return []zapcore.Field{
		zap.String("hash", hexutil.Encode(msg.Hash(pubsubTopic))),
		zap.String("pubsubTopic", pubsubTopic),
		zap.String("contentTopic", msg.ContentTopic),
		zap.Int64("timestamp", msg.GetTimestamp()),
	}
}

func (msg *WakuMessage) Logger(logger *zap.Logger, pubsubTopic string) *zap.Logger {
	return logger.With(msg.LogFields(pubsubTopic)...)
}
