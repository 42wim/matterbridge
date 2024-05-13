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

package wakuv2

import (
	"errors"

	"go.uber.org/zap"

	"github.com/waku-org/go-waku/waku/v2/protocol/relay"

	"github.com/status-im/status-go/protocol/common/shard"

	ethdisc "github.com/ethereum/go-ethereum/p2p/dnsdisc"

	"github.com/status-im/status-go/wakuv2/common"
)

var (
	ErrBadLightClientConfig = errors.New("either peer exchange server or discv5 must be disabled, and the peer exchange client must be enabled")
	ErrBadFullNodeConfig    = errors.New("peer exchange server and discv5 must be enabled, and the peer exchange client must be disabled")
)

// Config represents the configuration state of a waku node.
type Config struct {
	MaxMessageSize           uint32           `toml:",omitempty"` // Maximal message length allowed by the waku node
	Host                     string           `toml:",omitempty"`
	Port                     int              `toml:",omitempty"`
	EnablePeerExchangeServer bool             `toml:",omitempty"` // PeerExchange server makes sense only when discv5 is running locally as it will have a cache of peers that it can respond to in case a PeerExchange request comes from the PeerExchangeClient
	EnablePeerExchangeClient bool             `toml:",omitempty"`
	KeepAliveInterval        int              `toml:",omitempty"`
	MinPeersForRelay         int              `toml:",omitempty"` // Indicates the minimum number of peers required for using Relay Protocol
	MinPeersForFilter        int              `toml:",omitempty"` // Indicates the minimum number of peers required for using Filter Protocol
	LightClient              bool             `toml:",omitempty"` // Indicates if the node is a light client
	WakuNodes                []string         `toml:",omitempty"`
	Rendezvous               bool             `toml:",omitempty"`
	DiscV5BootstrapNodes     []string         `toml:",omitempty"`
	Nameserver               string           `toml:",omitempty"` // Optional nameserver to use for dns discovery
	Resolver                 ethdisc.Resolver `toml:",omitempty"` // Optional resolver to use for dns discovery
	EnableDiscV5             bool             `toml:",omitempty"` // Indicates whether discv5 is enabled or not
	DiscoveryLimit           int              `toml:",omitempty"` // Indicates the number of nodes to discover with peer exchange client
	AutoUpdate               bool             `toml:",omitempty"`
	UDPPort                  int              `toml:",omitempty"`
	EnableStore              bool             `toml:",omitempty"`
	StoreCapacity            int              `toml:",omitempty"`
	StoreSeconds             int              `toml:",omitempty"`
	TelemetryServerURL       string           `toml:",omitempty"`
	DefaultShardPubsubTopic  string           `toml:",omitempty"` // Pubsub topic to be used by default for messages that do not have a topic assigned (depending whether sharding is used or not)
	UseShardAsDefaultTopic   bool             `toml:",omitempty"`
	ClusterID                uint16           `toml:",omitempty"`
	EnableConfirmations      bool             `toml:",omitempty"` // Enable sending message confirmations
	SkipPublishToTopic       bool             `toml:",omitempty"` // Used in testing
}

func (c *Config) Validate(logger *zap.Logger) error {
	if c.LightClient && (c.EnablePeerExchangeServer || c.EnableDiscV5 || !c.EnablePeerExchangeClient) {
		logger.Warn("bad configuration for a light client", zap.Error(ErrBadLightClientConfig))
		return nil
	}
	if !c.LightClient && (!c.EnablePeerExchangeServer || !c.EnableDiscV5 || c.EnablePeerExchangeClient) {
		logger.Warn("bad configuration for a full node", zap.Error(ErrBadFullNodeConfig))
		return nil
	}
	return nil
}

var DefaultConfig = Config{
	MaxMessageSize:    common.DefaultMaxMessageSize,
	Host:              "0.0.0.0",
	Port:              0,
	KeepAliveInterval: 10, // second
	DiscoveryLimit:    20,
	MinPeersForRelay:  1, // TODO: determine correct value with Vac team
	MinPeersForFilter: 2, // TODO: determine correct value with Vac team and via testing
	AutoUpdate:        false,
}

func setDefaults(cfg *Config) *Config {
	if cfg == nil {
		cfg = new(Config)
	}

	if cfg.MaxMessageSize == 0 {
		cfg.MaxMessageSize = DefaultConfig.MaxMessageSize
	}

	if cfg.Host == "" {
		cfg.Host = DefaultConfig.Host
	}

	if cfg.KeepAliveInterval == 0 {
		cfg.KeepAliveInterval = DefaultConfig.KeepAliveInterval
	}

	if cfg.DiscoveryLimit == 0 {
		cfg.DiscoveryLimit = DefaultConfig.DiscoveryLimit
	}

	if cfg.MinPeersForRelay == 0 {
		cfg.MinPeersForRelay = DefaultConfig.MinPeersForRelay
	}

	if cfg.MinPeersForFilter == 0 {
		cfg.MinPeersForFilter = DefaultConfig.MinPeersForFilter
	}

	if cfg.DefaultShardPubsubTopic == "" {
		if cfg.UseShardAsDefaultTopic {
			cfg.DefaultShardPubsubTopic = shard.DefaultShardPubsubTopic()
		} else {
			cfg.DefaultShardPubsubTopic = relay.DefaultWakuTopic
		}
	}

	return cfg
}
