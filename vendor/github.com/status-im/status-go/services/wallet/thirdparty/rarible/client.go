package rarible

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	walletCommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/connection"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

const ownedNFTLimit = 100
const collectionOwnershipLimit = 50
const nftMetadataBatchLimit = 50

func (o *Client) ID() string {
	return RaribleID
}

func (o *Client) IsChainSupported(chainID walletCommon.ChainID) bool {
	_, err := getBaseURL(chainID)
	return err == nil
}

func (o *Client) IsConnected() bool {
	return o.connectionStatus.IsConnected()
}

func getBaseURL(chainID walletCommon.ChainID) (string, error) {
	switch uint64(chainID) {
	case walletCommon.EthereumMainnet, walletCommon.ArbitrumMainnet:
		return "https://api.rarible.org", nil
	case walletCommon.EthereumGoerli, walletCommon.ArbitrumSepolia:
		return "https://testnet-api.rarible.org", nil
	}

	return "", thirdparty.ErrChainIDNotSupported
}

func getItemBaseURL(chainID walletCommon.ChainID) (string, error) {
	baseURL, err := getBaseURL(chainID)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/v0.1/items", baseURL), nil
}

func getOwnershipBaseURL(chainID walletCommon.ChainID) (string, error) {
	baseURL, err := getBaseURL(chainID)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/v0.1/ownerships", baseURL), nil
}

func getCollectionBaseURL(chainID walletCommon.ChainID) (string, error) {
	baseURL, err := getBaseURL(chainID)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/v0.1/collections", baseURL), nil
}

type Client struct {
	thirdparty.CollectibleContractOwnershipProvider
	client           *http.Client
	mainnetAPIKey    string
	testnetAPIKey    string
	connectionStatus *connection.Status
}

func NewClient(mainnetAPIKey string, testnetAPIKey string) *Client {
	if mainnetAPIKey == "" {
		log.Warn("Rarible API key not available for Mainnet")
	}

	if testnetAPIKey == "" {
		log.Warn("Rarible API key not available for Testnet")
	}

	return &Client{
		client:           &http.Client{Timeout: time.Minute},
		mainnetAPIKey:    mainnetAPIKey,
		testnetAPIKey:    testnetAPIKey,
		connectionStatus: connection.NewStatus(),
	}
}

func (o *Client) getAPIKey(chainID walletCommon.ChainID) string {
	if chainID.IsMainnet() {
		return o.mainnetAPIKey
	}
	return o.testnetAPIKey
}

func (o *Client) doQuery(ctx context.Context, url string, apiKey string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")

	return o.doWithRetries(req, apiKey)
}

func (o *Client) doPostWithJSON(ctx context.Context, url string, payload any, apiKey string) (*http.Response, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	payloadString := string(payloadJSON)
	payloadReader := strings.NewReader(payloadString)

	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	return o.doWithRetries(req, apiKey)
}

func (o *Client) doWithRetries(req *http.Request, apiKey string) (*http.Response, error) {
	b := backoff.ExponentialBackOff{
		InitialInterval:     time.Millisecond * 1000,
		RandomizationFactor: 0.1,
		Multiplier:          1.5,
		MaxInterval:         time.Second * 32,
		MaxElapsedTime:      time.Second * 128,
		Clock:               backoff.SystemClock,
	}
	b.Reset()

	req.Header.Set("X-API-KEY", apiKey)

	op := func() (*http.Response, error) {
		resp, err := o.client.Do(req)
		if err != nil {
			return nil, backoff.Permanent(err)
		}

		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		err = fmt.Errorf("unsuccessful request: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, err
		}
		return nil, backoff.Permanent(err)
	}

	return backoff.RetryWithData(op, &b)
}

func (o *Client) FetchCollectibleOwnersByContractAddress(ctx context.Context, chainID walletCommon.ChainID, contractAddress common.Address) (*thirdparty.CollectibleContractOwnership, error) {
	ownership := thirdparty.CollectibleContractOwnership{
		ContractAddress: contractAddress,
		Owners:          make([]thirdparty.CollectibleOwner, 0),
	}

	queryParams := url.Values{
		"collection": {fmt.Sprintf("%s:%s", chainIDToChainString(chainID), contractAddress.String())},
		"size":       {strconv.Itoa(collectionOwnershipLimit)},
	}

	baseURL, err := getOwnershipBaseURL(chainID)

	if err != nil {
		return nil, err
	}

	for {
		url := fmt.Sprintf("%s/byCollection?%s", baseURL, queryParams.Encode())

		resp, err := o.doQuery(ctx, url, o.getAPIKey(chainID))
		if err != nil {
			o.connectionStatus.SetIsConnected(false)
			return nil, err
		}
		o.connectionStatus.SetIsConnected(true)

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var raribleOwnership ContractOwnershipContainer
		err = json.Unmarshal(body, &raribleOwnership)
		if err != nil {
			return nil, err
		}

		ownership.Owners = append(ownership.Owners, raribleContractOwnershipsToCommon(raribleOwnership.Ownerships)...)

		if raribleOwnership.Continuation == "" {
			break
		}

		queryParams["continuation"] = []string{raribleOwnership.Continuation}
	}

	return &ownership, nil
}

func (o *Client) FetchAllAssetsByOwner(ctx context.Context, chainID walletCommon.ChainID, owner common.Address, cursor string, limit int) (*thirdparty.FullCollectibleDataContainer, error) {
	assets := new(thirdparty.FullCollectibleDataContainer)

	queryParams := url.Values{
		"owner":       {fmt.Sprintf("%s:%s", ethereumString, owner.String())},
		"blockchains": {chainIDToChainString(chainID)},
	}

	tmpLimit := ownedNFTLimit
	if limit > thirdparty.FetchNoLimit && limit < tmpLimit {
		tmpLimit = limit
	}
	queryParams["size"] = []string{strconv.Itoa(tmpLimit)}

	if len(cursor) > 0 {
		queryParams["continuation"] = []string{cursor}
		assets.PreviousCursor = cursor
	}
	assets.Provider = o.ID()

	baseURL, err := getItemBaseURL(chainID)

	if err != nil {
		return nil, err
	}

	for {
		url := fmt.Sprintf("%s/byOwner?%s", baseURL, queryParams.Encode())

		resp, err := o.doQuery(ctx, url, o.getAPIKey(chainID))
		if err != nil {
			o.connectionStatus.SetIsConnected(false)
			return nil, err
		}
		o.connectionStatus.SetIsConnected(true)

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// if Json is not returned there must be an error
		if !json.Valid(body) {
			return nil, fmt.Errorf("invalid json: %s", string(body))
		}

		var container CollectiblesContainer
		err = json.Unmarshal(body, &container)
		if err != nil {
			return nil, err
		}

		assets.Items = append(assets.Items, raribleToCollectiblesData(container.Collectibles, chainID.IsMainnet())...)
		assets.NextCursor = container.Continuation

		if len(assets.NextCursor) == 0 {
			break
		}

		queryParams["continuation"] = []string{assets.NextCursor}

		if limit != thirdparty.FetchNoLimit && len(assets.Items) >= limit {
			break
		}
	}

	return assets, nil
}

func (o *Client) FetchAllAssetsByOwnerAndContractAddress(ctx context.Context, chainID walletCommon.ChainID, owner common.Address, contractAddresses []common.Address, cursor string, limit int) (*thirdparty.FullCollectibleDataContainer, error) {
	return nil, thirdparty.ErrEndpointNotSupported
}

func getCollectibleUniqueIDBatches(ids []thirdparty.CollectibleUniqueID) []BatchTokenIDs {
	batches := make([]BatchTokenIDs, 0)

	for startIdx := 0; startIdx < len(ids); startIdx += nftMetadataBatchLimit {
		endIdx := startIdx + nftMetadataBatchLimit
		if endIdx > len(ids) {
			endIdx = len(ids)
		}

		pageIDs := ids[startIdx:endIdx]

		batchIDs := BatchTokenIDs{
			IDs: make([]string, 0, len(pageIDs)),
		}
		for _, id := range pageIDs {
			batchID := fmt.Sprintf("%s:%s:%s", chainIDToChainString(id.ContractID.ChainID), id.ContractID.Address.String(), id.TokenID.String())
			batchIDs.IDs = append(batchIDs.IDs, batchID)
		}

		batches = append(batches, batchIDs)
	}

	return batches
}

func (o *Client) fetchAssetsByBatchTokenIDs(ctx context.Context, chainID walletCommon.ChainID, batchIDs BatchTokenIDs) ([]thirdparty.FullCollectibleData, error) {
	baseURL, err := getItemBaseURL(chainID)

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/byIds", baseURL)

	resp, err := o.doPostWithJSON(ctx, url, batchIDs, o.getAPIKey(chainID))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// if Json is not returned there must be an error
	if !json.Valid(body) {
		return nil, fmt.Errorf("invalid json: %s", string(body))
	}

	var assets CollectiblesContainer
	err = json.Unmarshal(body, &assets)
	if err != nil {
		return nil, err
	}

	ret := raribleToCollectiblesData(assets.Collectibles, chainID.IsMainnet())

	return ret, nil
}

func (o *Client) FetchAssetsByCollectibleUniqueID(ctx context.Context, uniqueIDs []thirdparty.CollectibleUniqueID) ([]thirdparty.FullCollectibleData, error) {
	ret := make([]thirdparty.FullCollectibleData, 0, len(uniqueIDs))

	idsPerChainID := thirdparty.GroupCollectibleUIDsByChainID(uniqueIDs)

	for chainID, ids := range idsPerChainID {
		batches := getCollectibleUniqueIDBatches(ids)
		for _, batch := range batches {
			assets, err := o.fetchAssetsByBatchTokenIDs(ctx, chainID, batch)
			if err != nil {
				return nil, err
			}

			ret = append(ret, assets...)
		}
	}

	return ret, nil
}

func (o *Client) FetchCollectionsDataByContractID(ctx context.Context, contractIDs []thirdparty.ContractID) ([]thirdparty.CollectionData, error) {
	ret := make([]thirdparty.CollectionData, 0, len(contractIDs))

	for _, contractID := range contractIDs {
		baseURL, err := getCollectionBaseURL(contractID.ChainID)

		if err != nil {
			return nil, err
		}

		url := fmt.Sprintf("%s/%s:%s", baseURL, chainIDToChainString(contractID.ChainID), contractID.Address.String())

		resp, err := o.doQuery(ctx, url, o.getAPIKey(contractID.ChainID))
		if err != nil {
			o.connectionStatus.SetIsConnected(false)
			return nil, err
		}
		o.connectionStatus.SetIsConnected(true)

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// if Json is not returned there must be an error
		if !json.Valid(body) {
			return nil, fmt.Errorf("invalid json: %s", string(body))
		}

		var collection Collection
		err = json.Unmarshal(body, &collection)
		if err != nil {
			return nil, err
		}

		ret = append(ret, collection.toCommon(contractID))
	}

	return ret, nil
}
