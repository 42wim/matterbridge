package opensea

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	walletCommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/connection"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

const assetLimitV2 = 50

func getV2BaseURL(chainID walletCommon.ChainID) (string, error) {
	switch uint64(chainID) {
	case walletCommon.EthereumMainnet, walletCommon.ArbitrumMainnet, walletCommon.OptimismMainnet:
		return "https://api.opensea.io/v2", nil
	case walletCommon.EthereumSepolia, walletCommon.ArbitrumSepolia, walletCommon.OptimismSepolia:
		return "https://testnets-api.opensea.io/v2", nil
	}

	return "", thirdparty.ErrChainIDNotSupported
}

func (o *ClientV2) ID() string {
	return OpenseaV2ID
}

func (o *ClientV2) IsChainSupported(chainID walletCommon.ChainID) bool {
	_, err := getV2BaseURL(chainID)
	return err == nil
}

func (o *ClientV2) IsConnected() bool {
	return o.connectionStatus.IsConnected()
}

func getV2URL(chainID walletCommon.ChainID, path string) (string, error) {
	baseURL, err := getV2BaseURL(chainID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", baseURL, path), nil
}

type ClientV2 struct {
	client           *HTTPClient
	apiKey           string
	connectionStatus *connection.Status
	urlGetter        urlGetter
}

// new opensea v2 client.
func NewClientV2(apiKey string, httpClient *HTTPClient) *ClientV2 {
	if apiKey == "" {
		log.Warn("OpenseaV2 API key not available")
	}

	return &ClientV2{
		client:           httpClient,
		apiKey:           apiKey,
		connectionStatus: connection.NewStatus(),
		urlGetter:        getV2URL,
	}
}

func (o *ClientV2) FetchAllAssetsByOwnerAndContractAddress(ctx context.Context, chainID walletCommon.ChainID, owner common.Address, contractAddresses []common.Address, cursor string, limit int) (*thirdparty.FullCollectibleDataContainer, error) {
	// No dedicated endpoint to filter owned assets by contract address.
	// Will probably be available at some point, for now do the filtering ourselves.
	assets := new(thirdparty.FullCollectibleDataContainer)

	// Build map for more efficient contract address check
	contractHashMap := make(map[string]bool)
	for _, contractAddress := range contractAddresses {
		contractID := thirdparty.ContractID{
			ChainID: chainID,
			Address: contractAddress,
		}
		contractHashMap[contractID.HashKey()] = true
	}

	assets.PreviousCursor = cursor
	assets.NextCursor = cursor
	assets.Provider = o.ID()

	for {
		assetsPage, err := o.FetchAllAssetsByOwner(ctx, chainID, owner, assets.NextCursor, assetLimitV2)
		if err != nil {
			return nil, err
		}

		for _, asset := range assetsPage.Items {
			if contractHashMap[asset.CollectibleData.ID.ContractID.HashKey()] {
				assets.Items = append(assets.Items, asset)
			}
		}

		assets.NextCursor = assetsPage.NextCursor

		if assets.NextCursor == "" {
			break
		}

		if limit > thirdparty.FetchNoLimit && len(assets.Items) >= limit {
			break
		}
	}

	return assets, nil
}

func (o *ClientV2) FetchAllAssetsByOwner(ctx context.Context, chainID walletCommon.ChainID, owner common.Address, cursor string, limit int) (*thirdparty.FullCollectibleDataContainer, error) {
	pathParams := []string{
		"chain", chainIDToChainString(chainID),
		"account", owner.String(),
		"nfts",
	}

	queryParams := url.Values{}

	return o.fetchAssets(ctx, chainID, pathParams, queryParams, limit, cursor)
}

func (o *ClientV2) FetchAssetsByCollectibleUniqueID(ctx context.Context, uniqueIDs []thirdparty.CollectibleUniqueID) ([]thirdparty.FullCollectibleData, error) {
	return o.fetchDetailedAssets(ctx, uniqueIDs)
}

func (o *ClientV2) fetchAssets(ctx context.Context, chainID walletCommon.ChainID, pathParams []string, queryParams url.Values, limit int, cursor string) (*thirdparty.FullCollectibleDataContainer, error) {
	assets := new(thirdparty.FullCollectibleDataContainer)

	tmpLimit := assetLimitV2
	if limit > thirdparty.FetchNoLimit && limit < tmpLimit {
		tmpLimit = limit
	}
	queryParams["limit"] = []string{strconv.Itoa(tmpLimit)}

	assets.PreviousCursor = cursor
	if cursor != "" {
		queryParams["next"] = []string{cursor}
	}
	assets.Provider = o.ID()

	for {
		path := fmt.Sprintf("%s?%s", strings.Join(pathParams, "/"), queryParams.Encode())
		url, err := o.urlGetter(chainID, path)
		if err != nil {
			return nil, err
		}

		body, err := o.client.doGetRequest(ctx, url, o.apiKey)
		if err != nil {
			o.connectionStatus.SetIsConnected(false)
			return nil, err
		}
		o.connectionStatus.SetIsConnected(true)

		// If body is empty, it means the account has no collectibles for this chain.
		// (Workaround implemented in http_client.go)
		if body == nil {
			assets.NextCursor = ""
			break
		}

		// if Json is not returned there must be an error
		if !json.Valid(body) {
			return nil, fmt.Errorf("invalid json: %s", string(body))
		}

		container := NFTContainer{}
		err = json.Unmarshal(body, &container)
		if err != nil {
			return nil, err
		}

		for _, asset := range container.NFTs {
			assets.Items = append(assets.Items, asset.toCommon(chainID))
		}
		assets.NextCursor = container.NextCursor

		if assets.NextCursor == "" {
			break
		}

		queryParams["next"] = []string{assets.NextCursor}

		if limit > thirdparty.FetchNoLimit && len(assets.Items) >= limit {
			break
		}
	}

	return assets, nil
}

func (o *ClientV2) fetchDetailedAssets(ctx context.Context, uniqueIDs []thirdparty.CollectibleUniqueID) ([]thirdparty.FullCollectibleData, error) {
	assets := make([]thirdparty.FullCollectibleData, 0, len(uniqueIDs))

	for _, id := range uniqueIDs {
		path := fmt.Sprintf("chain/%s/contract/%s/nfts/%s", chainIDToChainString(id.ContractID.ChainID), id.ContractID.Address.String(), id.TokenID.String())
		url, err := o.urlGetter(id.ContractID.ChainID, path)
		if err != nil {
			return nil, err
		}

		body, err := o.client.doGetRequest(ctx, url, o.apiKey)
		if err != nil {
			if ctx.Err() == nil {
				o.connectionStatus.SetIsConnected(false)
			}
			return nil, err
		}
		o.connectionStatus.SetIsConnected(true)

		// if Json is not returned there must be an error
		if !json.Valid(body) {
			return nil, fmt.Errorf("invalid json: %s", string(body))
		}

		nftContainer := DetailedNFTContainer{}
		err = json.Unmarshal(body, &nftContainer)
		if err != nil {
			return nil, err
		}

		assets = append(assets, nftContainer.NFT.toCommon(id.ContractID.ChainID))
	}

	return assets, nil
}

func (o *ClientV2) fetchContractDataByContractID(ctx context.Context, id thirdparty.ContractID) (*ContractData, error) {
	path := fmt.Sprintf("chain/%s/contract/%s", chainIDToChainString(id.ChainID), id.Address.String())
	url, err := o.urlGetter(id.ChainID, path)
	if err != nil {
		return nil, err
	}

	body, err := o.client.doGetRequest(ctx, url, o.apiKey)
	if err != nil {
		if ctx.Err() == nil {
			o.connectionStatus.SetIsConnected(false)
		}
		return nil, err
	}
	o.connectionStatus.SetIsConnected(true)

	// if Json is not returned there must be an error
	if !json.Valid(body) {
		return nil, fmt.Errorf("invalid json: %s", string(body))
	}

	contract := ContractData{}
	err = json.Unmarshal(body, &contract)
	if err != nil {
		return nil, err
	}

	return &contract, nil
}

func (o *ClientV2) fetchCollectionDataBySlug(ctx context.Context, chainID walletCommon.ChainID, slug string) (*CollectionData, error) {
	path := fmt.Sprintf("collections/%s", slug)
	url, err := o.urlGetter(chainID, path)
	if err != nil {
		return nil, err
	}

	body, err := o.client.doGetRequest(ctx, url, o.apiKey)
	if err != nil {
		o.connectionStatus.SetIsConnected(false)
		return nil, err
	}
	o.connectionStatus.SetIsConnected(true)

	// if Json is not returned there must be an error
	if !json.Valid(body) {
		return nil, fmt.Errorf("invalid json: %s", string(body))
	}

	collection := CollectionData{}
	err = json.Unmarshal(body, &collection)
	if err != nil {
		return nil, err
	}

	return &collection, nil
}

func (o *ClientV2) FetchCollectionsDataByContractID(ctx context.Context, contractIDs []thirdparty.ContractID) ([]thirdparty.CollectionData, error) {
	ret := make([]thirdparty.CollectionData, 0, len(contractIDs))

	for _, id := range contractIDs {
		contractData, err := o.fetchContractDataByContractID(ctx, id)
		if err != nil {
			return nil, err
		}

		if contractData == nil || contractData.Collection == "" {
			continue
		}

		collectionData, err := o.fetchCollectionDataBySlug(ctx, id.ChainID, contractData.Collection)
		if err != nil {
			return nil, err
		}

		ret = append(ret, collectionData.toCommon(id, contractData.ContractStandard))
	}

	return ret, nil
}
