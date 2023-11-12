// Copyright 2019 The Waku Library Authors.
//
// The Waku library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Waku library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty off
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Waku library. If not, see <http://www.gnu.org/licenses/>.
//
// This software uses the go-ethereum library, which is licensed
// under the GNU Lesser General Public Library, version 3 or any later.

package common

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TopicType represents a cryptographically secure, probabilistic partial
// classifications of a message, determined as the first (leftmost) 4 bytes of the
// SHA3 hash of some arbitrary data given by the original author of the message.
type TopicType [TopicLength]byte

type TopicSet map[TopicType]struct{}

func NewTopicSet(topics []TopicType) TopicSet {
	s := make(TopicSet, len(topics))
	for _, t := range topics {
		s[t] = struct{}{}
	}
	return s
}

func NewTopicSetFromBytes(byteArrays [][]byte) TopicSet {
	topics := make([]TopicType, len(byteArrays))
	for i, byteArr := range byteArrays {
		topics[i] = BytesToTopic(byteArr)
	}
	return NewTopicSet(topics)
}

// BytesToTopic converts from the byte array representation of a topic
// into the TopicType type.
func BytesToTopic(b []byte) (t TopicType) {
	sz := TopicLength
	if x := len(b); x < TopicLength {
		sz = x
	}
	for i := 0; i < sz; i++ {
		t[i] = b[i]
	}
	return t
}

func StringToTopic(s string) (t TopicType) {
	str, _ := hexutil.Decode(s)
	return BytesToTopic(str)
}

// String converts a topic byte array to a string representation.
func (t *TopicType) String() string {
	return hexutil.Encode(t[:])
}

// MarshalText returns the hex representation of t.
func (t TopicType) MarshalText() ([]byte, error) {
	return hexutil.Bytes(t[:]).MarshalText()
}

// UnmarshalText parses a hex representation to a topic.
func (t *TopicType) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Topic", input, t[:])
}

// Converts a topic to its 23/WAKU2-TOPICS representation
func (t TopicType) ContentTopic() string {
	enc := hexutil.Encode(t[:])
	return "/waku/1/" + enc + "/rfc26"
}

func ExtractTopicFromContentTopic(s string) (TopicType, error) {
	p := strings.Split(s, "/")

	if len(p) != 5 || p[1] != "waku" || p[2] != "1" || p[4] != "rfc26" {
		return TopicType{}, errors.New("invalid content topic format")
	}

	str, err := hexutil.Decode(p[3])
	if err != nil {
		return TopicType{}, err
	}

	result := BytesToTopic(str)
	return result, nil
}
