package shard

import (
	wakuproto "github.com/waku-org/go-waku/waku/v2/protocol"

	"github.com/status-im/status-go/protocol/protobuf"
)

type Shard struct {
	Cluster uint16 `json:"cluster"`
	Index   uint16 `json:"index"`
}

func FromProtobuff(p *protobuf.Shard) *Shard {
	if p == nil {
		return nil
	}

	return &Shard{
		Cluster: uint16(p.Cluster),
		Index:   uint16(p.Index),
	}
}

func (s *Shard) Protobuffer() *protobuf.Shard {
	if s == nil {
		return nil
	}

	return &protobuf.Shard{
		Cluster: int32(s.Cluster),
		Index:   int32(s.Index),
	}
}
func (s *Shard) PubsubTopic() string {
	if s != nil {
		return wakuproto.NewStaticShardingPubsubTopic(s.Cluster, s.Index).String()
	}
	return ""
}

func DefaultNonProtectedPubsubTopic() string {
	return (&Shard{
		Cluster: MainStatusShardCluster,
		Index:   NonProtectedShardIndex,
	}).PubsubTopic()
}

const MainStatusShardCluster = 16
const DefaultShardIndex = 32
const NonProtectedShardIndex = 64
const UndefinedShardValue = 0

func DefaultShardPubsubTopic() string {
	return wakuproto.NewStaticShardingPubsubTopic(MainStatusShardCluster, DefaultShardIndex).String()
}

func DefaultShard() *Shard {
	return &Shard{
		Cluster: MainStatusShardCluster,
		Index:   NonProtectedShardIndex,
	}
}
