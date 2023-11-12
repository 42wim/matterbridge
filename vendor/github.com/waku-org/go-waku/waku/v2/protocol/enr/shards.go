package enr

import (
	"errors"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/waku-org/go-waku/waku/v2/protocol"
)

func deleteShardingENREntries(localnode *enode.LocalNode) {
	localnode.Delete(enr.WithEntry(ShardingBitVectorEnrField, struct{}{}))
	localnode.Delete(enr.WithEntry(ShardingIndicesListEnrField, struct{}{}))
}

func WithWakuRelayShardList(rs protocol.RelayShards) ENROption {
	return func(localnode *enode.LocalNode) error {
		value, err := rs.ShardList()
		if err != nil {
			return err
		}
		deleteShardingENREntries(localnode)
		localnode.Set(enr.WithEntry(ShardingIndicesListEnrField, value))
		return nil
	}
}

func WithWakuRelayShardingBitVector(rs protocol.RelayShards) ENROption {
	return func(localnode *enode.LocalNode) error {
		deleteShardingENREntries(localnode)
		localnode.Set(enr.WithEntry(ShardingBitVectorEnrField, rs.BitVector()))
		return nil
	}
}

func WithWakuRelaySharding(rs protocol.RelayShards) ENROption {
	return func(localnode *enode.LocalNode) error {
		if len(rs.ShardIDs) >= 64 {
			return WithWakuRelayShardingBitVector(rs)(localnode)
		}

		return WithWakuRelayShardList(rs)(localnode)
	}
}

func WithWakuRelayShardingTopics(topics ...string) ENROption {
	return func(localnode *enode.LocalNode) error {
		rs, err := protocol.TopicsToRelayShards(topics...)
		if err != nil {
			return err
		}

		if len(rs) != 1 {
			return errors.New("expected a single RelayShards")
		}

		return WithWakuRelaySharding(rs[0])(localnode)
	}
}

// ENR record accessors

func RelayShardList(record *enr.Record) (*protocol.RelayShards, error) {
	var field []byte
	if err := record.Load(enr.WithEntry(ShardingIndicesListEnrField, &field)); err != nil {
		if enr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	res, err := protocol.FromShardList(field)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func RelayShardingBitVector(record *enr.Record) (*protocol.RelayShards, error) {
	var field []byte
	if err := record.Load(enr.WithEntry(ShardingBitVectorEnrField, &field)); err != nil {
		if enr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	res, err := protocol.FromBitVector(field)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func RelaySharding(record *enr.Record) (*protocol.RelayShards, error) {
	res, err := RelayShardList(record)
	if err != nil {
		return nil, err
	}

	if res != nil {
		return res, nil
	}

	return RelayShardingBitVector(record)
}

// Utils

func ContainsShard(record *enr.Record, cluster uint16, index uint16) bool {
	if index > protocol.MaxShardIndex {
		return false
	}

	rs, err := RelaySharding(record)
	if err != nil {
		return false
	}

	return rs.Contains(cluster, index)
}

func ContainsShardWithWakuTopic(record *enr.Record, topic protocol.WakuPubSubTopic) bool {
	if shardTopic, err := protocol.ToShardPubsubTopic(topic); err != nil {
		return false
	} else {
		return ContainsShard(record, shardTopic.Cluster(), shardTopic.Shard())
	}
}

func ContainsRelayShard(record *enr.Record, topic protocol.StaticShardingPubsubTopic) bool {
	return ContainsShardWithWakuTopic(record, topic)
}

func ContainsShardTopic(record *enr.Record, topic string) bool {
	shardTopic, err := protocol.ToWakuPubsubTopic(topic)
	if err != nil {
		return false
	}
	return ContainsShardWithWakuTopic(record, shardTopic)
}
