package opensea

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/status-go/services/wallet/bigint"
	walletCommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	OpenseaV2ID           = "openseaV2"
	ethereumMainnetString = "ethereum"
	arbitrumMainnetString = "arbitrum"
	optimismMainnetString = "optimism"
	ethereumSepoliaString = "sepolia"
	arbitrumSepoliaString = "arbitrum_sepolia"
	optimismSepoliaString = "optimism_sepolia"
)

type urlGetter func(walletCommon.ChainID, string) (string, error)

func chainIDToChainString(chainID walletCommon.ChainID) string {
	chainString := ""
	switch uint64(chainID) {
	case walletCommon.EthereumMainnet:
		chainString = ethereumMainnetString
	case walletCommon.ArbitrumMainnet:
		chainString = arbitrumMainnetString
	case walletCommon.OptimismMainnet:
		chainString = optimismMainnetString
	case walletCommon.EthereumSepolia:
		chainString = ethereumSepoliaString
	case walletCommon.ArbitrumSepolia:
		chainString = arbitrumSepoliaString
	case walletCommon.OptimismSepolia:
		chainString = optimismSepoliaString
	}
	return chainString
}

func openseaToContractType(contractType string) walletCommon.ContractType {
	switch contractType {
	case "cryptopunks", "erc721":
		return walletCommon.ContractTypeERC721
	case "erc1155":
		return walletCommon.ContractTypeERC1155
	default:
		return walletCommon.ContractTypeUnknown
	}
}

type NFTContainer struct {
	NFTs       []NFT  `json:"nfts"`
	NextCursor string `json:"next"`
}

type NFT struct {
	TokenID       *bigint.BigInt `json:"identifier"`
	Collection    string         `json:"collection"`
	Contract      common.Address `json:"contract"`
	TokenStandard string         `json:"token_standard"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ImageURL      string         `json:"image_url"`
	MetadataURL   string         `json:"metadata_url"`
}

type DetailedNFTContainer struct {
	NFT DetailedNFT `json:"nft"`
}

type DetailedNFT struct {
	TokenID       *bigint.BigInt `json:"identifier"`
	Collection    string         `json:"collection"`
	Contract      common.Address `json:"contract"`
	TokenStandard string         `json:"token_standard"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ImageURL      string         `json:"image_url"`
	AnimationURL  string         `json:"animation_url"`
	MetadataURL   string         `json:"metadata_url"`
	Owners        []OwnerV2      `json:"owners"`
	Traits        []TraitV2      `json:"traits"`
}

type OwnerV2 struct {
	Address  common.Address `json:"address"`
	Quantity *bigint.BigInt `json:"quantity"`
}

type TraitValue string

func (st *TraitValue) UnmarshalJSON(b []byte) error {
	var item interface{}
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}

	switch v := item.(type) {
	case float64:
		*st = TraitValue(strconv.FormatFloat(v, 'f', 2, 64))
	case int:
		*st = TraitValue(strconv.Itoa(v))
	case string:
		*st = TraitValue(v)

	}
	return nil
}

type TraitV2 struct {
	TraitType   string     `json:"trait_type"`
	DisplayType string     `json:"display_type"`
	MaxValue    string     `json:"max_value"`
	TraitCount  int        `json:"trait_count"`
	Order       string     `json:"order"`
	Value       TraitValue `json:"value"`
}

type ContractData struct {
	Address          common.Address `json:"address"`
	Chain            string         `json:"chain"`
	Collection       string         `json:"collection"`
	ContractStandard string         `json:"contract_standard"`
	Name             string         `json:"name"`
}

type ContractID struct {
	Address common.Address `json:"address"`
	Chain   string         `json:"chain"`
}

type CollectionData struct {
	Collection  string         `json:"collection"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Owner       common.Address `json:"owner"`
	ImageURL    string         `json:"image_url"`
	Contracts   []ContractID   `json:"contracts"`
}

func (c *NFT) id(chainID walletCommon.ChainID) thirdparty.CollectibleUniqueID {
	return thirdparty.CollectibleUniqueID{
		ContractID: thirdparty.ContractID{
			ChainID: chainID,
			Address: c.Contract,
		},
		TokenID: c.TokenID,
	}
}

func (c *NFT) toCollectiblesData(chainID walletCommon.ChainID) thirdparty.CollectibleData {
	return thirdparty.CollectibleData{
		ID:           c.id(chainID),
		ContractType: openseaToContractType(c.TokenStandard),
		Provider:     OpenseaV2ID,
		Name:         c.Name,
		Description:  c.Description,
		ImageURL:     c.ImageURL,
		AnimationURL: c.ImageURL,
		Traits:       make([]thirdparty.CollectibleTrait, 0),
		TokenURI:     c.MetadataURL,
	}
}

func (c *NFT) toCommon(chainID walletCommon.ChainID) thirdparty.FullCollectibleData {
	return thirdparty.FullCollectibleData{
		CollectibleData: c.toCollectiblesData(chainID),
		CollectionData:  nil,
	}
}

func openseaV2ToCollectibleTraits(traits []TraitV2) []thirdparty.CollectibleTrait {
	ret := make([]thirdparty.CollectibleTrait, 0, len(traits))
	caser := cases.Title(language.Und, cases.NoLower)
	for _, orig := range traits {
		dest := thirdparty.CollectibleTrait{
			TraitType:   strings.Replace(orig.TraitType, "_", " ", 1),
			Value:       caser.String(string(orig.Value)),
			DisplayType: orig.DisplayType,
			MaxValue:    orig.MaxValue,
		}

		ret = append(ret, dest)
	}
	return ret
}

func (c *DetailedNFT) id(chainID walletCommon.ChainID) thirdparty.CollectibleUniqueID {
	return thirdparty.CollectibleUniqueID{
		ContractID: thirdparty.ContractID{
			ChainID: chainID,
			Address: c.Contract,
		},
		TokenID: c.TokenID,
	}
}

func (c *DetailedNFT) toCollectiblesData(chainID walletCommon.ChainID) thirdparty.CollectibleData {
	return thirdparty.CollectibleData{
		ID:           c.id(chainID),
		ContractType: openseaToContractType(c.TokenStandard),
		Provider:     OpenseaV2ID,
		Name:         c.Name,
		Description:  c.Description,
		ImageURL:     c.ImageURL,
		AnimationURL: c.AnimationURL,
		Traits:       openseaV2ToCollectibleTraits(c.Traits),
		TokenURI:     c.MetadataURL,
	}
}

func (c *DetailedNFT) toCommon(chainID walletCommon.ChainID) thirdparty.FullCollectibleData {
	return thirdparty.FullCollectibleData{
		CollectibleData: c.toCollectiblesData(chainID),
		CollectionData:  nil,
	}
}

func (c *CollectionData) toCommon(id thirdparty.ContractID, tokenStandard string) thirdparty.CollectionData {
	ret := thirdparty.CollectionData{
		ID:           id,
		ContractType: openseaToContractType(tokenStandard),
		Provider:     OpenseaV2ID,
		Name:         c.Name,
		Slug:         c.Collection,
		ImageURL:     c.ImageURL,
	}
	return ret
}
