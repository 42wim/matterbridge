package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type WakuPubSubTopic interface {
	String() string
}

const defaultPubsubTopic = "/waku/2/default-waku/proto"

type DefaultPubsubTopic struct{}

func (DefaultPubsubTopic) String() string {
	return defaultPubsubTopic
}

// StaticShardingPubsubTopicPrefix is the expected prefix to be used for static sharding pubsub topics
const StaticShardingPubsubTopicPrefix = "/waku/2/rs"

// waku pubsub topic errors
var ErrNotWakuPubsubTopic = errors.New("not a waku pubsub topic")

// shard pubsub topic errors
var ErrNotShardPubsubTopic = errors.New("not a shard pubsub topic")
var ErrInvalidStructure = errors.New("invalid topic structure")
var ErrInvalidShardedTopicPrefix = errors.New("must start with " + StaticShardingPubsubTopicPrefix)
var ErrMissingClusterIndex = errors.New("missing shard_cluster_index")
var ErrMissingShardNumber = errors.New("missing shard_number")

// ErrInvalidNumberFormat indicates that a number exceeds the allowed range
var ErrInvalidNumberFormat = errors.New("only 2^16 numbers are allowed")

// StaticShardingPubsubTopic describes a pubSub topic as per StaticSharding
type StaticShardingPubsubTopic struct {
	clusterID uint16
	shardID   uint16
}

// NewStaticShardingPubsubTopic creates a new pubSub topic
func NewStaticShardingPubsubTopic(cluster uint16, shard uint16) StaticShardingPubsubTopic {
	return StaticShardingPubsubTopic{
		clusterID: cluster,
		shardID:   shard,
	}
}

// Cluster returns the sharded cluster index
func (s StaticShardingPubsubTopic) Cluster() uint16 {
	return s.clusterID
}

// Shard returns the shard number
func (s StaticShardingPubsubTopic) Shard() uint16 {
	return s.shardID
}

// Equal compares StaticShardingPubsubTopic
func (s StaticShardingPubsubTopic) Equal(t2 StaticShardingPubsubTopic) bool {
	return s.String() == t2.String()
}

// String formats StaticShardingPubsubTopic to RFC 23 specific string format for pubsub topic.
func (s StaticShardingPubsubTopic) String() string {
	return fmt.Sprintf("%s/%d/%d", StaticShardingPubsubTopicPrefix, s.clusterID, s.shardID)
}

// Parse parses a topic string into a StaticShardingPubsubTopic
func (s *StaticShardingPubsubTopic) Parse(topic string) error {
	if !strings.HasPrefix(topic, StaticShardingPubsubTopicPrefix) {
		return ErrInvalidShardedTopicPrefix
	}

	parts := strings.Split(topic[11:], "/")
	if len(parts) != 2 {
		return ErrInvalidStructure
	}

	clusterPart := parts[0]
	if len(clusterPart) == 0 {
		return ErrMissingClusterIndex
	}

	clusterInt, err := strconv.ParseUint(clusterPart, 10, 16)
	if err != nil {
		return ErrInvalidNumberFormat
	}

	shardPart := parts[1]
	if len(shardPart) == 0 {
		return ErrMissingShardNumber
	}

	shardInt, err := strconv.ParseUint(shardPart, 10, 16)
	if err != nil {
		return ErrInvalidNumberFormat
	}

	s.shardID = uint16(shardInt)
	s.clusterID = uint16(clusterInt)

	return nil
}

func ToShardPubsubTopic(topic WakuPubSubTopic) (StaticShardingPubsubTopic, error) {
	result, ok := topic.(StaticShardingPubsubTopic)
	if !ok {
		return StaticShardingPubsubTopic{}, ErrNotShardPubsubTopic
	}
	return result, nil
}

// ToWakuPubsubTopic takes a pubSub topic string and creates a WakuPubsubTopic object.
func ToWakuPubsubTopic(topic string) (WakuPubSubTopic, error) {
	if topic == defaultPubsubTopic {
		return DefaultPubsubTopic{}, nil
	}
	if strings.HasPrefix(topic, StaticShardingPubsubTopicPrefix) {
		s := StaticShardingPubsubTopic{}
		err := s.Parse(topic)
		if err != nil {
			return s, err
		}
		return s, nil
	}
	return nil, ErrNotWakuPubsubTopic
}
