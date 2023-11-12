package token

import (
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/services/wallet/bigint"
)

type DeployState uint8

const (
	Failed DeployState = iota
	InProgress
	Deployed
)

type PrivilegesLevel uint8

const (
	OwnerLevel PrivilegesLevel = iota
	MasterLevel
	CommunityLevel
)

type CommunityToken struct {
	TokenType          protobuf.CommunityTokenType `json:"tokenType"`
	CommunityID        string                      `json:"communityId"`
	Address            string                      `json:"address"`
	Name               string                      `json:"name"`
	Symbol             string                      `json:"symbol"`
	Description        string                      `json:"description"`
	Supply             *bigint.BigInt              `json:"supply"`
	InfiniteSupply     bool                        `json:"infiniteSupply"`
	Transferable       bool                        `json:"transferable"`
	RemoteSelfDestruct bool                        `json:"remoteSelfDestruct"`
	ChainID            int                         `json:"chainId"`
	DeployState        DeployState                 `json:"deployState"`
	Base64Image        string                      `json:"image"`
	Decimals           int                         `json:"decimals"`
	Deployer           string                      `json:"deployer"`
	PrivilegesLevel    PrivilegesLevel             `json:"privilegesLevel"`
}
