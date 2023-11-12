package communitytokens

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/status-im/status-go/contracts/community-tokens/assets"
	"github.com/status-im/status-go/contracts/community-tokens/collectibles"
	"github.com/status-im/status-go/contracts/community-tokens/mastertoken"
	"github.com/status-im/status-go/contracts/community-tokens/ownertoken"
	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/services/wallet/bigint"
)

type TokenInstance interface {
	RemoteBurn(*bind.TransactOpts, []*big.Int) (*types.Transaction, error)
	Mint(*bind.TransactOpts, []string, *bigint.BigInt) (*types.Transaction, error)
	SetMaxSupply(*bind.TransactOpts, *big.Int) (*types.Transaction, error)
	PackMethod(ctx context.Context, methodName string, args ...interface{}) ([]byte, error)
}

// Owner Token
type OwnerTokenInstance struct {
	TokenInstance
	instance *ownertoken.OwnerToken
}

func (t OwnerTokenInstance) RemoteBurn(transactOpts *bind.TransactOpts, tokenIds []*big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("remote destruction for owner token not implemented")
}

func (t OwnerTokenInstance) Mint(transactOpts *bind.TransactOpts, walletAddresses []string, amount *bigint.BigInt) (*types.Transaction, error) {
	return nil, fmt.Errorf("minting for owner token not implemented")
}

func (t OwnerTokenInstance) SetMaxSupply(transactOpts *bind.TransactOpts, maxSupply *big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("setting max supply for owner token not implemented")
}

func (t OwnerTokenInstance) PackMethod(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	ownerTokenABI, err := abi.JSON(strings.NewReader(ownertoken.OwnerTokenABI))
	if err != nil {
		return []byte{}, err
	}
	return ownerTokenABI.Pack(methodName, args...)
}

// Master Token
type MasterTokenInstance struct {
	TokenInstance
	instance *mastertoken.MasterToken
	api      *API
}

func (t MasterTokenInstance) RemoteBurn(transactOpts *bind.TransactOpts, tokenIds []*big.Int) (*types.Transaction, error) {
	return t.instance.RemoteBurn(transactOpts, tokenIds)
}

func (t MasterTokenInstance) Mint(transactOpts *bind.TransactOpts, walletAddresses []string, amount *bigint.BigInt) (*types.Transaction, error) {
	usersAddresses := t.api.PrepareMintCollectiblesData(walletAddresses, amount)
	return t.instance.MintTo(transactOpts, usersAddresses)
}

func (t MasterTokenInstance) SetMaxSupply(transactOpts *bind.TransactOpts, maxSupply *big.Int) (*types.Transaction, error) {
	return t.instance.SetMaxSupply(transactOpts, maxSupply)
}

func (t MasterTokenInstance) PackMethod(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	masterTokenABI, err := abi.JSON(strings.NewReader(mastertoken.MasterTokenABI))
	if err != nil {
		return []byte{}, err
	}
	return masterTokenABI.Pack(methodName, args...)
}

// Collectible
type CollectibleInstance struct {
	TokenInstance
	instance *collectibles.Collectibles
	api      *API
}

func (t CollectibleInstance) RemoteBurn(transactOpts *bind.TransactOpts, tokenIds []*big.Int) (*types.Transaction, error) {
	return t.instance.RemoteBurn(transactOpts, tokenIds)
}

func (t CollectibleInstance) Mint(transactOpts *bind.TransactOpts, walletAddresses []string, amount *bigint.BigInt) (*types.Transaction, error) {
	usersAddresses := t.api.PrepareMintCollectiblesData(walletAddresses, amount)
	return t.instance.MintTo(transactOpts, usersAddresses)
}

func (t CollectibleInstance) SetMaxSupply(transactOpts *bind.TransactOpts, maxSupply *big.Int) (*types.Transaction, error) {
	return t.instance.SetMaxSupply(transactOpts, maxSupply)
}

func (t CollectibleInstance) PackMethod(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	collectiblesABI, err := abi.JSON(strings.NewReader(collectibles.CollectiblesABI))
	if err != nil {
		return []byte{}, err
	}
	return collectiblesABI.Pack(methodName, args...)
}

// Asset
type AssetInstance struct {
	TokenInstance
	instance *assets.Assets
	api      *API
}

func (t AssetInstance) RemoteBurn(transactOpts *bind.TransactOpts, tokenIds []*big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("remote destruction for assets not implemented")
}

// The amount should be in smallest denomination of the asset (like wei) with decimal = 18, eg.
// if we want to mint 2.34 of the token, then amount should be 234{16 zeros}.
func (t AssetInstance) Mint(transactOpts *bind.TransactOpts, walletAddresses []string, amount *bigint.BigInt) (*types.Transaction, error) {
	usersAddresses, amountsList := t.api.PrepareMintAssetsData(walletAddresses, amount)
	return t.instance.MintTo(transactOpts, usersAddresses, amountsList)
}

func (t AssetInstance) SetMaxSupply(transactOpts *bind.TransactOpts, maxSupply *big.Int) (*types.Transaction, error) {
	return t.instance.SetMaxSupply(transactOpts, maxSupply)
}

func (t AssetInstance) PackMethod(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	assetsABI, err := abi.JSON(strings.NewReader(assets.AssetsABI))
	if err != nil {
		return []byte{}, err
	}
	return assetsABI.Pack(methodName, args...)
}

// creator

func NewTokenInstance(api *API, chainID uint64, contractAddress string) (TokenInstance, error) {
	tokenType, err := api.s.db.GetTokenType(chainID, contractAddress)
	if err != nil {
		return nil, err
	}
	privLevel, err := api.s.db.GetTokenPrivilegesLevel(chainID, contractAddress)
	if err != nil {
		return nil, err
	}
	switch {
	case privLevel == token.OwnerLevel:
		contractInst, err := api.NewOwnerTokenInstance(chainID, contractAddress)
		if err != nil {
			return nil, err
		}
		return &OwnerTokenInstance{instance: contractInst}, nil
	case privLevel == token.MasterLevel:
		contractInst, err := api.NewMasterTokenInstance(chainID, contractAddress)
		if err != nil {
			return nil, err
		}
		return &MasterTokenInstance{instance: contractInst}, nil
	case tokenType == protobuf.CommunityTokenType_ERC721:
		contractInst, err := api.NewCollectiblesInstance(chainID, contractAddress)
		if err != nil {
			return nil, err
		}
		return &CollectibleInstance{instance: contractInst}, nil
	case tokenType == protobuf.CommunityTokenType_ERC20:
		contractInst, err := api.NewAssetsInstance(chainID, contractAddress)
		if err != nil {
			return nil, err
		}
		return &AssetInstance{instance: contractInst}, nil
	}

	return nil, fmt.Errorf("unknown type of contract: chain=%v, address=%v", chainID, contractAddress)
}
