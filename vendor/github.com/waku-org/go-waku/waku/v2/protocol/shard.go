package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/waku-org/go-waku/waku/v2/hash"
)

const MaxShardIndex = uint16(1023)

// ClusterIndex is the clusterID used in sharding space.
// For shardIDs allocation and other magic numbers refer to RFC 51
const ClusterIndex = 1

// GenerationZeroShardsCount is number of shards supported in generation-0
const GenerationZeroShardsCount = 8

var (
	ErrTooManyShards     = errors.New("too many shards")
	ErrInvalidShard      = errors.New("invalid shard")
	ErrInvalidShardCount = errors.New("invalid shard count")
	ErrExpected130Bytes  = errors.New("invalid data: expected 130 bytes")
)

type RelayShards struct {
	ClusterID uint16   `json:"clusterID"`
	ShardIDs  []uint16 `json:"shardIDs"`
}

func NewRelayShards(clusterID uint16, shardIDs ...uint16) (RelayShards, error) {
	if len(shardIDs) > math.MaxUint8 {
		return RelayShards{}, ErrTooManyShards
	}

	shardIDSet := make(map[uint16]struct{})
	for _, index := range shardIDs {
		if index > MaxShardIndex {
			return RelayShards{}, ErrInvalidShard
		}
		shardIDSet[index] = struct{}{} // dedup
	}

	if len(shardIDSet) == 0 {
		return RelayShards{}, ErrInvalidShardCount
	}

	shardIDs = []uint16{}
	for index := range shardIDSet {
		shardIDs = append(shardIDs, index)
	}

	return RelayShards{ClusterID: clusterID, ShardIDs: shardIDs}, nil
}

func (rs RelayShards) Topics() []WakuPubSubTopic {
	var result []WakuPubSubTopic
	for _, i := range rs.ShardIDs {
		result = append(result, NewStaticShardingPubsubTopic(rs.ClusterID, i))
	}
	return result
}

func (rs RelayShards) Contains(cluster uint16, index uint16) bool {
	if rs.ClusterID != cluster {
		return false
	}

	found := false
	for _, idx := range rs.ShardIDs {
		if idx == index {
			found = true
		}
	}

	return found
}

func (rs RelayShards) ContainsShardPubsubTopic(topic WakuPubSubTopic) bool {
	if shardedTopic, err := ToShardPubsubTopic(topic); err != nil {
		return false
	} else {
		return rs.Contains(shardedTopic.Cluster(), shardedTopic.Shard())
	}
}

func TopicsToRelayShards(topic ...string) ([]RelayShards, error) {
	result := make([]RelayShards, 0)
	dict := make(map[uint16]map[uint16]struct{})
	for _, t := range topic {
		if !strings.HasPrefix(t, StaticShardingPubsubTopicPrefix) {
			continue
		}

		var ps StaticShardingPubsubTopic
		err := ps.Parse(t)
		if err != nil {
			return nil, err
		}

		shardIDs, ok := dict[ps.clusterID]
		if !ok {
			shardIDs = make(map[uint16]struct{})
		}

		shardIDs[ps.shardID] = struct{}{}
		dict[ps.clusterID] = shardIDs
	}

	for clusterID, shardIDs := range dict {
		idx := make([]uint16, 0, len(shardIDs))
		for shardID := range shardIDs {
			idx = append(idx, shardID)
		}

		rs, err := NewRelayShards(clusterID, idx...)
		if err != nil {
			return nil, err
		}

		result = append(result, rs)
	}

	return result, nil
}

func (rs RelayShards) ContainsTopic(topic string) bool {
	wTopic, err := ToWakuPubsubTopic(topic)
	if err != nil {
		return false
	}
	return rs.ContainsShardPubsubTopic(wTopic)
}

func (rs RelayShards) ShardList() ([]byte, error) {
	if len(rs.ShardIDs) > math.MaxUint8 {
		return nil, ErrTooManyShards
	}

	var result []byte

	result = binary.BigEndian.AppendUint16(result, rs.ClusterID)
	result = append(result, uint8(len(rs.ShardIDs)))
	for _, index := range rs.ShardIDs {
		result = binary.BigEndian.AppendUint16(result, index)
	}

	return result, nil
}

func FromShardList(buf []byte) (RelayShards, error) {
	if len(buf) < 3 {
		return RelayShards{}, fmt.Errorf("insufficient data: expected at least 3 bytes, got %d bytes", len(buf))
	}

	cluster := binary.BigEndian.Uint16(buf[0:2])
	length := int(buf[2])

	if len(buf) != 3+2*length {
		return RelayShards{}, fmt.Errorf("invalid data: `length` field is %d but %d bytes were provided", length, len(buf))
	}

	shardIDs := make([]uint16, length)
	for i := 0; i < length; i++ {
		shardIDs[i] = binary.BigEndian.Uint16(buf[3+2*i : 5+2*i])
	}

	return NewRelayShards(cluster, shardIDs...)
}

func setBit(n byte, pos uint) byte {
	n |= (1 << pos)
	return n
}

func hasBit(n byte, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func (rs RelayShards) BitVector() []byte {
	// The value is comprised of a two-byte shard cluster index in network byte
	// order concatenated with a 128-byte wide bit vector. The bit vector
	// indicates which shards of the respective shard cluster the node is part
	// of. The right-most bit in the bit vector represents shard 0, the left-most
	// bit represents shard 1023.
	var result []byte
	result = binary.BigEndian.AppendUint16(result, rs.ClusterID)

	vec := make([]byte, 128)
	for _, index := range rs.ShardIDs {
		n := vec[index/8]
		vec[index/8] = byte(setBit(n, uint(index%8)))
	}

	return append(result, vec...)
}

// Generate a RelayShards from a byte slice
func FromBitVector(buf []byte) (RelayShards, error) {
	if len(buf) != 130 {
		return RelayShards{}, ErrExpected130Bytes
	}

	cluster := binary.BigEndian.Uint16(buf[0:2])
	var shardIDs []uint16

	for i := uint16(0); i < 128; i++ {
		for j := uint(0); j < 8; j++ {
			if !hasBit(buf[2+i], j) {
				continue
			}

			shardIDs = append(shardIDs, uint16(j)+8*i)
		}
	}

	return RelayShards{ClusterID: cluster, ShardIDs: shardIDs}, nil
}

// GetShardFromContentTopic runs Autosharding logic and returns a pubSubTopic
// This is based on Autosharding algorithm defined in RFC 51
func GetShardFromContentTopic(topic ContentTopic, shardCount int) StaticShardingPubsubTopic {
	bytes := []byte(topic.ApplicationName)
	bytes = append(bytes, []byte(topic.ApplicationVersion)...)

	hash := hash.SHA256(bytes)
	//We only use the last 64 bits of the hash as having more shards is unlikely.
	hashValue := binary.BigEndian.Uint64(hash[24:])

	shard := hashValue % uint64(shardCount)

	return NewStaticShardingPubsubTopic(ClusterIndex, uint16(shard))
}

func GetPubSubTopicFromContentTopic(cTopicString string) (string, error) {
	cTopic, err := StringToContentTopic(cTopicString)
	if err != nil {
		return "", fmt.Errorf("%s : %s", err.Error(), cTopicString)
	}
	pTopic := GetShardFromContentTopic(cTopic, GenerationZeroShardsCount)

	return pTopic.String(), nil
}

func GeneratePubsubToContentTopicMap(pubsubTopic string, contentTopics []string) (map[string][]string, error) {

	pubSubTopicMap := make(map[string][]string, 0)

	if pubsubTopic == "" {
		//Should we derive pubsub topic from contentTopic so that peer selection and discovery can be done accordingly?
		for _, cTopic := range contentTopics {
			pTopic, err := GetPubSubTopicFromContentTopic(cTopic)
			if err != nil {
				return nil, err
			}
			_, ok := pubSubTopicMap[pTopic]
			if !ok {
				pubSubTopicMap[pTopic] = []string{}
			}
			pubSubTopicMap[pTopic] = append(pubSubTopicMap[pTopic], cTopic)
		}
	} else {
		pubSubTopicMap[pubsubTopic] = append(pubSubTopicMap[pubsubTopic], contentTopics...)
	}
	return pubSubTopicMap, nil
}
