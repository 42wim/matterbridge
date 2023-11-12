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

package waku

import (
	"github.com/status-im/status-go/waku/common"
)

// Config represents the configuration state of a waku node.
type Config struct {
	MaxMessageSize           uint32   `toml:",omitempty"`
	MinimumAcceptedPoW       float64  `toml:",omitempty"`
	BloomFilterMode          bool     `toml:",omitempty"` // when true, we only match against bloom filter
	LightClient              bool     `toml:",omitempty"` // when true, it does not forward messages
	FullNode                 bool     `toml:",omitempty"` // when true, it forwards all messages
	RestrictLightClientsConn bool     `toml:",omitempty"` // when true, do not accept light client as peers if it is a light client itself
	EnableConfirmations      bool     `toml:",omitempty"` // when true, sends message confirmations
	SoftBlacklistedPeerIDs   []string `toml:",omitempty"`
}

var DefaultConfig = Config{
	MaxMessageSize:           common.DefaultMaxMessageSize,
	MinimumAcceptedPoW:       common.DefaultMinimumPoW,
	RestrictLightClientsConn: true,
}
