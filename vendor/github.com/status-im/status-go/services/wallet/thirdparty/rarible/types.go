package rarible

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/status-go/services/wallet/bigint"
	walletCommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const RaribleID = "rarible"

const (
	ethereumString = "ETHEREUM"
	arbitrumString = "ARBITRUM"
)

func chainStringToChainID(chainString string, isMainnet bool) walletCommon.ChainID {
	chainID := walletCommon.UnknownChainID
	switch chainString {
	case ethereumString:
		if isMainnet {
			chainID = walletCommon.EthereumMainnet
		} else {
			chainID = walletCommon.EthereumGoerli
		}
	case arbitrumString:
		if isMainnet {
			chainID = walletCommon.ArbitrumMainnet
		} else {
			chainID = walletCommon.ArbitrumSepolia
		}
	}
	return walletCommon.ChainID(chainID)
}

func chainIDToChainString(chainID walletCommon.ChainID) string {
	chainString := ""
	switch uint64(chainID) {
	case walletCommon.EthereumMainnet, walletCommon.EthereumGoerli:
		chainString = ethereumString
	case walletCommon.ArbitrumMainnet, walletCommon.ArbitrumSepolia:
		chainString = arbitrumString
	}
	return chainString
}

func raribleToContractType(contractType string) walletCommon.ContractType {
	switch contractType {
	case "CRYPTO_PUNKS", "ERC721":
		return walletCommon.ContractTypeERC721
	case "ERC1155":
		return walletCommon.ContractTypeERC1155
	default:
		return walletCommon.ContractTypeUnknown
	}
}

func raribleContractIDToUniqueID(contractID string, isMainnet bool) (thirdparty.ContractID, error) {
	ret := thirdparty.ContractID{}

	parts := strings.Split(contractID, ":")
	if len(parts) != 2 {
		return ret, fmt.Errorf("invalid rarible contract id string %s", contractID)
	}

	ret.ChainID = chainStringToChainID(parts[0], isMainnet)
	if uint64(ret.ChainID) == walletCommon.UnknownChainID {
		return ret, fmt.Errorf("unknown rarible chainID in contract id string %s", contractID)
	}
	ret.Address = common.HexToAddress(parts[1])

	return ret, nil
}

func raribleCollectibleIDToUniqueID(collectibleID string, isMainnet bool) (thirdparty.CollectibleUniqueID, error) {
	ret := thirdparty.CollectibleUniqueID{}

	parts := strings.Split(collectibleID, ":")
	if len(parts) != 3 {
		return ret, fmt.Errorf("invalid rarible collectible id string %s", collectibleID)
	}

	ret.ContractID.ChainID = chainStringToChainID(parts[0], isMainnet)
	if uint64(ret.ContractID.ChainID) == walletCommon.UnknownChainID {
		return ret, fmt.Errorf("unknown rarible chainID in collectible id string %s", collectibleID)
	}
	ret.ContractID.Address = common.HexToAddress(parts[1])
	tokenID, ok := big.NewInt(0).SetString(parts[2], 10)
	if !ok {
		return ret, fmt.Errorf("invalid rarible tokenID %s", collectibleID)
	}
	ret.TokenID = &bigint.BigInt{
		Int: tokenID,
	}

	return ret, nil
}

type BatchTokenIDs struct {
	IDs []string `json:"ids"`
}

type CollectiblesContainer struct {
	Continuation string        `json:"continuation"`
	Collectibles []Collectible `json:"items"`
}

type Collectible struct {
	ID         string              `json:"id"`
	Blockchain string              `json:"blockchain"`
	Collection string              `json:"collection"`
	Contract   string              `json:"contract"`
	TokenID    *bigint.BigInt      `json:"tokenId"`
	Metadata   CollectibleMetadata `json:"meta"`
}

type CollectibleMetadata struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	ExternalURI     string      `json:"externalUri"`
	OriginalMetaURI string      `json:"originalMetaUri"`
	Attributes      []Attribute `json:"attributes"`
	Contents        []Content   `json:"content"`
}

type Attribute struct {
	Key   string         `json:"key"`
	Value AttributeValue `json:"value"`
}

type AttributeValue string

func (st *AttributeValue) UnmarshalJSON(b []byte) error {
	var item interface{}
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}

	switch v := item.(type) {
	case float64:
		*st = AttributeValue(strconv.FormatFloat(v, 'f', 2, 64))
	case int:
		*st = AttributeValue(strconv.Itoa(v))
	case string:
		*st = AttributeValue(v)
	}
	return nil
}

type Collection struct {
	ID           string             `json:"id"`
	Blockchain   string             `json:"blockchain"`
	ContractType string             `json:"type"`
	Name         string             `json:"name"`
	Metadata     CollectionMetadata `json:"meta"`
}

type CollectionMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Contents    []Content `json:"content"`
}

type Content struct {
	Type           string `json:"@type"`
	URL            string `json:"url"`
	Representation string `json:"representation"`
	Available      bool   `json:"available"`
}

type ContractOwnershipContainer struct {
	Continuation string              `json:"continuation"`
	Ownerships   []ContractOwnership `json:"ownerships"`
}

type ContractOwnership struct {
	ID         string         `json:"id"`
	Blockchain string         `json:"blockchain"`
	ItemID     string         `json:"itemId"`
	Contract   string         `json:"contract"`
	Collection string         `json:"collection"`
	TokenID    *bigint.BigInt `json:"tokenId"`
	Owner      string         `json:"owner"`
	Value      *bigint.BigInt `json:"value"`
}

func raribleContractOwnershipsToCommon(raribleOwnerships []ContractOwnership) []thirdparty.CollectibleOwner {
	balancesPerOwner := make(map[common.Address][]thirdparty.TokenBalance)
	for _, raribleOwnership := range raribleOwnerships {
		owner := common.HexToAddress(raribleOwnership.Owner)
		if _, ok := balancesPerOwner[owner]; !ok {
			balancesPerOwner[owner] = make([]thirdparty.TokenBalance, 0)
		}

		balance := thirdparty.TokenBalance{
			TokenID: raribleOwnership.TokenID,
			Balance: raribleOwnership.Value,
		}
		balancesPerOwner[owner] = append(balancesPerOwner[owner], balance)
	}

	ret := make([]thirdparty.CollectibleOwner, 0, len(balancesPerOwner))
	for owner, balances := range balancesPerOwner {
		ret = append(ret, thirdparty.CollectibleOwner{
			OwnerAddress:  owner,
			TokenBalances: balances,
		})
	}

	return ret
}

func raribleToCollectibleTraits(attributes []Attribute) []thirdparty.CollectibleTrait {
	ret := make([]thirdparty.CollectibleTrait, 0, len(attributes))
	caser := cases.Title(language.Und, cases.NoLower)
	for _, orig := range attributes {
		dest := thirdparty.CollectibleTrait{
			TraitType: orig.Key,
			Value:     caser.String(string(orig.Value)),
		}

		ret = append(ret, dest)
	}
	return ret
}

func raribleToCollectiblesData(l []Collectible, isMainnet bool) []thirdparty.FullCollectibleData {
	ret := make([]thirdparty.FullCollectibleData, 0, len(l))
	for _, c := range l {
		id, err := raribleCollectibleIDToUniqueID(c.ID, isMainnet)
		if err != nil {
			continue
		}
		item := c.toCommon(id)
		ret = append(ret, item)
	}
	return ret
}

func (c *Collection) toCommon(id thirdparty.ContractID) thirdparty.CollectionData {
	ret := thirdparty.CollectionData{
		ID:           id,
		ContractType: raribleToContractType(c.ContractType),
		Provider:     RaribleID,
		Name:         c.Metadata.Name,
		Slug:         "", /* Missing from the API for now */
		ImageURL:     getImageURL(c.Metadata.Contents),
		Traits:       make(map[string]thirdparty.CollectionTrait, 0), /* Missing from the API for now */
	}
	return ret
}

func contentTypeValue(contentType string, includeOriginal bool) int {
	ret := -1

	switch contentType {
	case "PREVIEW":
		ret = 1
	case "PORTRAIT":
		ret = 2
	case "BIG":
		ret = 3
	case "ORIGINAL":
		if includeOriginal {
			ret = 4
		}
	}

	return ret
}

func isNewContentBigger(current string, new string, includeOriginal bool) bool {
	currentValue := contentTypeValue(current, includeOriginal)
	newValue := contentTypeValue(new, includeOriginal)

	return newValue > currentValue
}

func getBiggestContentURL(contents []Content, contentType string, includeOriginal bool) string {
	ret := Content{
		Type:           "",
		URL:            "",
		Representation: "",
		Available:      false,
	}

	for _, content := range contents {
		if content.Type == contentType {
			if isNewContentBigger(ret.Representation, content.Representation, includeOriginal) {
				ret = content
			}
		}
	}

	return ret.URL
}

func getAnimationURL(contents []Content) string {
	// Try to get the biggest content of type "VIDEO"
	ret := getBiggestContentURL(contents, "VIDEO", true)

	// If empty, try to get the biggest content of type "IMAGE", including the "ORIGINAL" representation
	if ret == "" {
		ret = getBiggestContentURL(contents, "IMAGE", true)
	}

	return ret
}

func getImageURL(contents []Content) string {
	// Get the biggest content of type "IMAGE", excluding the "ORIGINAL" representation
	ret := getBiggestContentURL(contents, "IMAGE", false)

	// If empty, allow the "ORIGINAL" representation
	if ret == "" {
		ret = getBiggestContentURL(contents, "IMAGE", true)
	}

	return ret
}

func (c *Collectible) toCollectibleData(id thirdparty.CollectibleUniqueID) thirdparty.CollectibleData {
	imageURL := getImageURL(c.Metadata.Contents)
	animationURL := getAnimationURL(c.Metadata.Contents)

	if animationURL == "" {
		animationURL = imageURL
	}

	return thirdparty.CollectibleData{
		ID:           id,
		ContractType: walletCommon.ContractTypeUnknown, // Rarible doesn't provide the contract type with the collectible
		Provider:     RaribleID,
		Name:         c.Metadata.Name,
		Description:  c.Metadata.Description,
		Permalink:    c.Metadata.ExternalURI,
		ImageURL:     imageURL,
		AnimationURL: animationURL,
		Traits:       raribleToCollectibleTraits(c.Metadata.Attributes),
		TokenURI:     c.Metadata.OriginalMetaURI,
	}
}

func (c *Collectible) toCommon(id thirdparty.CollectibleUniqueID) thirdparty.FullCollectibleData {
	return thirdparty.FullCollectibleData{
		CollectibleData: c.toCollectibleData(id),
		CollectionData:  nil,
	}
}
