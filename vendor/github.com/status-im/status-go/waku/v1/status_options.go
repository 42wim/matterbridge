package v1

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/status-im/status-go/waku/common"
)

// statusOptionKey is a current type used in StatusOptions as a key.
type statusOptionKey uint64

var (
	defaultMinPoW = math.Float64bits(0.001)
	idxFieldKey   = make(map[int]statusOptionKey)
	keyFieldIdx   = make(map[statusOptionKey]int)
)

// StatusOptions defines additional information shared between peers
// during the handshake.
// There might be more options provided then fields in StatusOptions
// and they should be ignored during deserialization to stay forward compatible.
// In the case of RLP, options should be serialized to an array of tuples
// where the first item is a field name and the second is a RLP-serialized value.
type StatusOptions struct {
	PoWRequirement       *uint64            `rlp:"key=0"` // RLP does not support float64 natively
	BloomFilter          []byte             `rlp:"key=1"`
	LightNodeEnabled     *bool              `rlp:"key=2"`
	ConfirmationsEnabled *bool              `rlp:"key=3"`
	PacketRateLimits     *common.RateLimits `rlp:"key=4"`
	TopicInterest        []common.TopicType `rlp:"key=5"`
	BytesRateLimits      *common.RateLimits `rlp:"key=6"`
}

func StatusOptionsFromHost(host common.WakuHost) StatusOptions {
	opts := StatusOptions{}

	packetRateLimits := host.PacketRateLimits()
	opts.PacketRateLimits = &packetRateLimits

	bytesRateLimits := host.BytesRateLimits()
	opts.BytesRateLimits = &bytesRateLimits

	lightNode := host.LightClientMode()
	opts.LightNodeEnabled = &lightNode

	minPoW := host.MinPow()
	opts.SetPoWRequirementFromF(minPoW)

	confirmationsEnabled := host.ConfirmationsEnabled()
	opts.ConfirmationsEnabled = &confirmationsEnabled

	bloomFilterMode := host.BloomFilterMode()
	if bloomFilterMode {
		opts.BloomFilter = host.BloomFilter()
	} else {
		opts.TopicInterest = host.TopicInterest()
	}

	return opts
}

// initFLPKeyFields initialises the values of `idxFieldKey` and `keyFieldIdx`
func initRLPKeyFields() {
	o := StatusOptions{}
	v := reflect.ValueOf(o)

	for i := 0; i < v.NumField(); i++ {
		// skip unexported fields
		if !v.Field(i).CanInterface() {
			continue
		}
		rlpTag := v.Type().Field(i).Tag.Get("rlp")

		// skip fields without rlp field tag
		if rlpTag == "" {
			continue
		}

		keys := strings.Split(rlpTag, "=")

		if len(keys) != 2 || keys[0] != "key" {
			panic("invalid value of \"rlp\" tag, expected \"key=N\" where N is uint")
		}
		key, err := strconv.ParseUint(keys[1], 10, 64)
		if err != nil {
			panic("could not parse \"rlp\" key")
		}

		// typecast key to be of statusOptionKey type
		keyFieldIdx[statusOptionKey(key)] = i
		idxFieldKey[i] = statusOptionKey(key)
	}
}

// WithDefaults adds the default values for a given peer.
// This are not the host default values, but the default values that ought to
// be used when receiving from an update from a peer.
func (o StatusOptions) WithDefaults() StatusOptions {
	if o.PoWRequirement == nil {
		o.PoWRequirement = &defaultMinPoW
	}

	if o.LightNodeEnabled == nil {
		lightNodeEnabled := false
		o.LightNodeEnabled = &lightNodeEnabled
	}

	if o.ConfirmationsEnabled == nil {
		confirmationsEnabled := false
		o.ConfirmationsEnabled = &confirmationsEnabled
	}

	if o.PacketRateLimits == nil {
		o.PacketRateLimits = &common.RateLimits{}
	}

	if o.BytesRateLimits == nil {
		o.BytesRateLimits = &common.RateLimits{}
	}

	if o.BloomFilter == nil {
		o.BloomFilter = common.MakeFullNodeBloom()
	}

	return o
}

func (o StatusOptions) PoWRequirementF() *float64 {
	if o.PoWRequirement == nil {
		return nil
	}
	result := math.Float64frombits(*o.PoWRequirement)
	return &result
}

func (o *StatusOptions) SetPoWRequirementFromF(val float64) {
	requirement := math.Float64bits(val)
	o.PoWRequirement = &requirement
}

func (o StatusOptions) EncodeRLP(w io.Writer) error {
	v := reflect.ValueOf(o)
	var optionsList []interface{}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.IsNil() {
			value := field.Interface()
			key, ok := idxFieldKey[i]
			if !ok {
				continue
			}
			if value != nil {
				optionsList = append(optionsList, []interface{}{key, value})
			}
		}
	}
	return rlp.Encode(w, optionsList)
}

func (o *StatusOptions) DecodeRLP(s *rlp.Stream) error {
	_, err := s.List()
	if err != nil {
		return fmt.Errorf("expected an outer list: %v", err)
	}

	v := reflect.ValueOf(o)

loop:
	for {
		_, err := s.List()
		switch err {
		case nil:
			// continue to decode a key
		case rlp.EOL:
			break loop
		default:
			return fmt.Errorf("expected an inner list: %v", err)
		}
		var key statusOptionKey
		if err := s.Decode(&key); err != nil {
			return fmt.Errorf("invalid key: %v", err)
		}
		// Skip processing if a key does not exist.
		// It might happen when there is a new peer
		// which supports a new option with
		// a higher index.
		idx, ok := keyFieldIdx[key]
		if !ok {
			// Read the rest of the list items and dump peer.
			_, err := s.Raw()
			if err != nil {
				return fmt.Errorf("failed to read the value of key %d: %v", key, err)
			}
			continue
		}
		if err := s.Decode(v.Elem().Field(idx).Addr().Interface()); err != nil {
			return fmt.Errorf("failed to decode an option %d: %v", key, err)
		}
		if err := s.ListEnd(); err != nil {
			return err
		}
	}

	return s.ListEnd()
}

func (o StatusOptions) Validate() error {
	if len(o.TopicInterest) > 10000 {
		return errors.New("topic interest is limited by 1000 items")
	}
	return nil
}
