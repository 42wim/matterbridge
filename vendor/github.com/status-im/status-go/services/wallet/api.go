package wallet

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc/network"
	"github.com/status-im/status-go/services/wallet/activity"
	"github.com/status-im/status-go/services/wallet/bridge"
	"github.com/status-im/status-go/services/wallet/collectibles"
	wcommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/currency"
	"github.com/status-im/status-go/services/wallet/history"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletconnect"
	wc "github.com/status-im/status-go/services/wallet/walletconnect"
	"github.com/status-im/status-go/services/wallet/walletevent"
	"github.com/status-im/status-go/transactions"
)

func NewAPI(s *Service) *API {
	router := NewRouter(s)
	return &API{s, s.reader, router}
}

// API is class with methods available over RPC.
type API struct {
	s      *Service
	reader *Reader
	router *Router
}

func (api *API) StartWallet(ctx context.Context) error {
	return api.reader.Start()
}

func (api *API) StopWallet(ctx context.Context) error {
	return api.s.Stop()
}

func (api *API) GetPairingsJSONFileContent() ([]byte, error) {
	return api.s.keycardPairings.GetPairingsJSONFileContent()
}

func (api *API) SetPairingsJSONFileContent(content []byte) error {
	return api.s.keycardPairings.SetPairingsJSONFileContent(content)
}

func (api *API) GetWalletToken(ctx context.Context, addresses []common.Address) (map[common.Address][]Token, error) {
	return api.reader.GetWalletToken(ctx, addresses)
}

// GetBalancesByChain return a map with key as chain id and value as map of account address and map of token address and balance
// [chainID][account][token]balance
func (api *API) GetBalancesByChain(ctx context.Context, chainIDs []uint64, addresses, tokens []common.Address) (map[uint64]map[common.Address]map[common.Address]*hexutil.Big, error) {
	clients, err := api.s.rpcClient.EthClients(chainIDs)
	if err != nil {
		return nil, err
	}

	return api.s.tokenManager.GetBalancesByChain(ctx, clients, addresses, tokens)
}

func (api *API) GetWalletTokenBalances(ctx context.Context, addresses []common.Address) (map[common.Address][]Token, error) {
	return api.reader.GetWalletTokenBalances(ctx, addresses)
}

func (api *API) FetchOrGetCachedWalletBalances(ctx context.Context, addresses []common.Address) (map[common.Address][]Token, error) {
	return api.reader.FetchOrGetCachedWalletBalances(ctx, addresses)
}

func (api *API) GetCachedWalletTokensWithoutMarketData(ctx context.Context) (map[common.Address][]Token, error) {
	return api.reader.GetCachedWalletTokensWithoutMarketData()
}

type DerivedAddress struct {
	Address        common.Address `json:"address"`
	PublicKey      types.HexBytes `json:"public-key,omitempty"`
	Path           string         `json:"path"`
	HasActivity    bool           `json:"hasActivity"`
	AlreadyCreated bool           `json:"alreadyCreated"`
}

// @deprecated
func (api *API) CheckRecentHistory(ctx context.Context, addresses []common.Address) error {
	return api.s.transferController.CheckRecentHistory([]uint64{api.s.rpcClient.UpstreamChainID}, addresses)
}

// @deprecated
func (api *API) CheckRecentHistoryForChainIDs(ctx context.Context, chainIDs []uint64, addresses []common.Address) error {
	return api.s.transferController.CheckRecentHistory(chainIDs, addresses)
}

func hexBigToBN(hexBig *hexutil.Big) *big.Int {
	var bN *big.Int
	if hexBig != nil {
		bN = hexBig.ToInt()
	}
	return bN
}

// @deprecated
// GetTransfersByAddress returns transfers for a single address
func (api *API) GetTransfersByAddress(ctx context.Context, address common.Address, toBlock, limit *hexutil.Big, fetchMore bool) ([]transfer.View, error) {
	log.Debug("[WalletAPI:: GetTransfersByAddress] get transfers for an address", "address", address)
	var intLimit = int64(1)
	if limit != nil {
		intLimit = limit.ToInt().Int64()
	}
	return api.s.transferController.GetTransfersByAddress(ctx, api.s.rpcClient.UpstreamChainID, address, hexBigToBN(toBlock), intLimit, fetchMore)
}

// @deprecated
// LoadTransferByHash loads transfer to the database
// Only used by status-mobile
func (api *API) LoadTransferByHash(ctx context.Context, address common.Address, hash common.Hash) error {
	log.Debug("[WalletAPI:: LoadTransferByHash] get transfer by hash", "address", address, "hash", hash)
	return api.s.transferController.LoadTransferByHash(ctx, api.s.rpcClient, address, hash)
}

// @deprecated
func (api *API) GetTransfersByAddressAndChainID(ctx context.Context, chainID uint64, address common.Address, toBlock, limit *hexutil.Big, fetchMore bool) ([]transfer.View, error) {
	log.Debug("[WalletAPI:: GetTransfersByAddressAndChainIDs] get transfers for an address", "address", address)
	return api.s.transferController.GetTransfersByAddress(ctx, chainID, address, hexBigToBN(toBlock), limit.ToInt().Int64(), fetchMore)
}

// @deprecated
func (api *API) GetTransfersForIdentities(ctx context.Context, identities []transfer.TransactionIdentity) ([]transfer.View, error) {
	log.Debug("wallet.api.GetTransfersForIdentities", "identities.len", len(identities))

	return api.s.transferController.GetTransfersForIdentities(ctx, identities)
}

func (api *API) FetchDecodedTxData(ctx context.Context, data string) (*thirdparty.DataParsed, error) {
	log.Debug("[Wallet: FetchDecodedTxData]")

	return api.s.decoder.Decode(data)
}

// GetBalanceHistory retrieves token balance history for token identity on multiple chains
func (api *API) GetBalanceHistory(ctx context.Context, chainIDs []uint64, addresses []common.Address, tokenSymbol string, currencySymbol string, timeInterval history.TimeInterval) ([]*history.ValuePoint, error) {
	log.Debug("wallet.api.GetBalanceHistory", "chainIDs", chainIDs, "address", addresses, "tokenSymbol", tokenSymbol, "currencySymbol", currencySymbol, "timeInterval", timeInterval)

	var fromTimestamp uint64
	now := uint64(time.Now().UTC().Unix())
	switch timeInterval {
	case history.BalanceHistoryAllTime:
		fromTimestamp = 0
	case history.BalanceHistory1Year:
		fallthrough
	case history.BalanceHistory6Months:
		fallthrough
	case history.BalanceHistory1Month:
		fallthrough
	case history.BalanceHistory7Days:
		fromTimestamp = now - history.TimeIntervalDurationSecs(timeInterval)
	default:
		return nil, fmt.Errorf("unknown time interval: %v", timeInterval)
	}

	return api.GetBalanceHistoryRange(ctx, chainIDs, addresses, tokenSymbol, currencySymbol, fromTimestamp, now)
}

// GetBalanceHistoryRange retrieves token balance history for token identity on multiple chains for a time range
// 'toTimestamp' is ignored for now, but will be used in the future to limit the range of the history
func (api *API) GetBalanceHistoryRange(ctx context.Context, chainIDs []uint64, addresses []common.Address, tokenSymbol string, currencySymbol string, fromTimestamp uint64, _ uint64) ([]*history.ValuePoint, error) {
	log.Debug("wallet.api.GetBalanceHistoryRange", "chainIDs", chainIDs, "address", addresses, "tokenSymbol", tokenSymbol, "currencySymbol", currencySymbol, "fromTimestamp", fromTimestamp)
	return api.s.history.GetBalanceHistory(ctx, chainIDs, addresses, tokenSymbol, currencySymbol, fromTimestamp)
}

func (api *API) GetTokenList(ctx context.Context) (*token.ListWrapper, error) {
	log.Debug("call to get token list")
	rst := api.s.tokenManager.GetList()
	log.Debug("result from token list", "len", len(rst.Data))
	return rst, nil
}

// @deprecated
func (api *API) GetTokens(ctx context.Context, chainID uint64) ([]*token.Token, error) {
	log.Debug("call to get tokens")
	rst, err := api.s.tokenManager.GetTokens(chainID)
	log.Debug("result from token store", "len", len(rst))
	return rst, err
}

// @deprecated
func (api *API) GetCustomTokens(ctx context.Context) ([]*token.Token, error) {
	log.Debug("call to get custom tokens")
	rst, err := api.s.tokenManager.GetCustoms(true)
	log.Debug("result from database for custom tokens", "len", len(rst))
	return rst, err
}

func (api *API) DiscoverToken(ctx context.Context, chainID uint64, address common.Address) (*token.Token, error) {
	log.Debug("call to get discover token")
	token, err := api.s.tokenManager.DiscoverToken(ctx, chainID, address)
	return token, err
}

func (api *API) AddCustomToken(ctx context.Context, token token.Token) error {
	log.Debug("call to create or edit custom token")
	if token.ChainID == 0 {
		token.ChainID = api.s.rpcClient.UpstreamChainID
	}
	err := api.s.tokenManager.UpsertCustom(token)
	log.Debug("result from database for create or edit custom token", "err", err)
	return err
}

// @deprecated
func (api *API) DeleteCustomToken(ctx context.Context, address common.Address) error {
	log.Debug("call to remove custom token")
	err := api.s.tokenManager.DeleteCustom(api.s.rpcClient.UpstreamChainID, address)
	log.Debug("result from database for remove custom token", "err", err)
	return err
}

func (api *API) DeleteCustomTokenByChainID(ctx context.Context, chainID uint64, address common.Address) error {
	log.Debug("call to remove custom token")
	err := api.s.tokenManager.DeleteCustom(chainID, address)
	log.Debug("result from database for remove custom token", "err", err)
	return err
}

// @deprecated
// Not used by status-desktop anymore
func (api *API) GetPendingTransactions(ctx context.Context) ([]*transactions.PendingTransaction, error) {
	log.Debug("wallet.api.GetPendingTransactions")
	rst, err := api.s.pendingTxManager.GetAllPending()
	log.Debug("wallet.api.GetPendingTransactions RESULT", "len", len(rst))
	return rst, err
}

// @deprecated
// Not used by status-desktop anymore
func (api *API) GetPendingTransactionsForIdentities(ctx context.Context, identities []transfer.TransactionIdentity) (
	result []*transactions.PendingTransaction, err error) {

	log.Debug("wallet.api.GetPendingTransactionsForIdentities")

	result = make([]*transactions.PendingTransaction, 0, len(identities))
	var pt *transactions.PendingTransaction
	for _, identity := range identities {
		pt, err = api.s.pendingTxManager.GetPendingEntry(identity.ChainID, identity.Hash)
		result = append(result, pt)
	}

	log.Debug("wallet.api.GetPendingTransactionsForIdentities RES", "len", len(result))
	return
}

// @deprecated
// TODO - #11861: Remove this and replace with EventPendingTransactionStatusChanged event and Delete to confirm the transaction where it is needed
func (api *API) WatchTransactionByChainID(ctx context.Context, chainID uint64, transactionHash common.Hash) (err error) {
	log.Debug("wallet.api.WatchTransactionByChainID", "chainID", chainID, "transactionHash", transactionHash)
	var status *transactions.TxStatus
	defer func() {
		log.Debug("wallet.api.WatchTransactionByChainID return", "err", err, "chainID", chainID, "transactionHash", transactionHash)
	}()

	// Workaround to keep the blocking call until the clients use the PendingTxTracker APIs
	eventChan := make(chan walletevent.Event, 2)
	sub := api.s.feed.Subscribe(eventChan)
	defer sub.Unsubscribe()

	status, err = api.s.pendingTxManager.Watch(ctx, wcommon.ChainID(chainID), transactionHash)
	if err == nil && *status != transactions.Pending {
		return nil
	}

	for {
		select {
		case we := <-eventChan:
			if transactions.EventPendingTransactionStatusChanged == we.Type {
				var p transactions.StatusChangedPayload
				err = json.Unmarshal([]byte(we.Message), &p)
				if err != nil {
					return err
				}
				if p.ChainID == wcommon.ChainID(chainID) && p.Hash == transactionHash {
					return nil
				}
			}
		case <-time.After(10 * time.Minute):
			return errors.New("timeout watching for pending transaction")
		}
	}
}

func (api *API) GetCryptoOnRamps(ctx context.Context) ([]CryptoOnRamp, error) {
	return api.s.cryptoOnRampManager.Get()
}

/*
   Collectibles API Start
*/

func (api *API) FetchBalancesByOwnerAndContractAddress(ctx context.Context, chainID wcommon.ChainID, ownerAddress common.Address, contractAddresses []common.Address) (thirdparty.TokenBalancesPerContractAddress, error) {
	log.Debug("call to FetchBalancesByOwnerAndContractAddress")

	return api.s.collectiblesManager.FetchBalancesByOwnerAndContractAddress(ctx, chainID, ownerAddress, contractAddresses)
}

func (api *API) GetCollectibleOwnership(id thirdparty.CollectibleUniqueID) ([]thirdparty.AccountBalance, error) {
	return api.s.collectiblesManager.GetCollectibleOwnership(id)
}

func (api *API) RefetchOwnedCollectibles() error {
	log.Debug("wallet.api.RefetchOwnedCollectibles")

	api.s.collectibles.RefetchOwnedCollectibles()
	return nil
}

func (api *API) GetOwnedCollectiblesAsync(requestID int32, chainIDs []wcommon.ChainID, addresses []common.Address, filter collectibles.Filter, offset int, limit int, dataType collectibles.CollectibleDataType, fetchCriteria collectibles.FetchCriteria) error {
	log.Debug("wallet.api.GetOwnedCollectiblesAsync", "requestID", requestID, "chainIDs.count", len(chainIDs), "addr.count", len(addresses), "offset", offset, "limit", limit, "dataType", dataType, "fetchCriteria", fetchCriteria)

	api.s.collectibles.GetOwnedCollectiblesAsync(requestID, chainIDs, addresses, filter, offset, limit, dataType, fetchCriteria)
	return nil
}

func (api *API) GetCollectiblesByUniqueIDAsync(requestID int32, uniqueIDs []thirdparty.CollectibleUniqueID, dataType collectibles.CollectibleDataType) error {
	log.Debug("wallet.api.GetCollectiblesByUniqueIDAsync", "requestID", requestID, "uniqueIDs.count", len(uniqueIDs), "dataType", dataType)

	api.s.collectibles.GetCollectiblesByUniqueIDAsync(requestID, uniqueIDs, dataType)
	return nil
}

func (api *API) GetCollectibleOwnersByContractAddress(ctx context.Context, chainID wcommon.ChainID, contractAddress common.Address) (*thirdparty.CollectibleContractOwnership, error) {
	log.Debug("call to GetCollectibleOwnersByContractAddress")
	return api.s.collectiblesManager.FetchCollectibleOwnersByContractAddress(ctx, chainID, contractAddress)
}

func (api *API) SearchCollectibles(ctx context.Context, chainID wcommon.ChainID, text string, cursor string, limit int, providerID string) (*thirdparty.FullCollectibleDataContainer, error) {
	log.Debug("call to SearchCollectibles")
	return api.s.collectiblesManager.SearchCollectibles(ctx, chainID, text, cursor, limit, providerID)
}

func (api *API) SearchCollections(ctx context.Context, chainID wcommon.ChainID, text string, cursor string, limit int, providerID string) (*thirdparty.CollectionDataContainer, error) {
	log.Debug("call to SearchCollections")
	return api.s.collectiblesManager.SearchCollections(ctx, chainID, text, cursor, limit, providerID)
}

/*
   Collectibles API End
*/

func (api *API) AddEthereumChain(ctx context.Context, network params.Network) error {
	log.Debug("call to AddEthereumChain")
	return api.s.rpcClient.NetworkManager.Upsert(&network)
}

func (api *API) DeleteEthereumChain(ctx context.Context, chainID uint64) error {
	log.Debug("call to DeleteEthereumChain")
	return api.s.rpcClient.NetworkManager.Delete(chainID)
}

func (api *API) GetEthereumChains(ctx context.Context) ([]*network.CombinedNetwork, error) {
	log.Debug("call to GetEthereumChains")
	return api.s.rpcClient.NetworkManager.GetCombinedNetworks()
}

// @deprecated
func (api *API) FetchPrices(ctx context.Context, symbols []string, currencies []string) (map[string]map[string]float64, error) {
	log.Debug("call to FetchPrices")
	return api.s.marketManager.FetchPrices(symbols, currencies)
}

// @deprecated
func (api *API) FetchMarketValues(ctx context.Context, symbols []string, currency string) (map[string]thirdparty.TokenMarketValues, error) {
	log.Debug("call to FetchMarketValues")
	return api.s.marketManager.FetchTokenMarketValues(symbols, currency)
}

func (api *API) GetHourlyMarketValues(ctx context.Context, symbol string, currency string, limit int, aggregate int) ([]thirdparty.HistoricalPrice, error) {
	log.Debug("call to GetHourlyMarketValues")
	return api.s.marketManager.FetchHistoricalHourlyPrices(symbol, currency, limit, aggregate)
}

func (api *API) GetDailyMarketValues(ctx context.Context, symbol string, currency string, limit int, allData bool, aggregate int) ([]thirdparty.HistoricalPrice, error) {
	log.Debug("call to GetDailyMarketValues")
	return api.s.marketManager.FetchHistoricalDailyPrices(symbol, currency, limit, allData, aggregate)
}

// @deprecated
func (api *API) FetchTokenDetails(ctx context.Context, symbols []string) (map[string]thirdparty.TokenDetails, error) {
	log.Debug("call to FetchTokenDetails")
	return api.s.marketManager.FetchTokenDetails(symbols)
}

func (api *API) GetSuggestedFees(ctx context.Context, chainID uint64) (*SuggestedFees, error) {
	log.Debug("call to GetSuggestedFees")
	return api.s.feesManager.suggestedFees(ctx, chainID)
}

func (api *API) GetEstimatedLatestBlockNumber(ctx context.Context, chainID uint64) (uint64, error) {
	log.Debug("call to GetEstimatedLatestBlockNumber, chainID:", chainID)
	return api.s.blockChainState.GetEstimatedLatestBlockNumber(ctx, chainID)
}

// @deprecated
func (api *API) GetTransactionEstimatedTime(ctx context.Context, chainID uint64, maxFeePerGas *big.Float) (TransactionEstimation, error) {
	log.Debug("call to getTransactionEstimatedTime")
	return api.s.feesManager.transactionEstimatedTime(ctx, chainID, maxFeePerGas), nil
}

func (api *API) GetSuggestedRoutes(
	ctx context.Context,
	sendType SendType,
	addrFrom common.Address,
	addrTo common.Address,
	amountIn *hexutil.Big,
	tokenID string,
	disabledFromChainIDs,
	disabledToChaindIDs,
	preferedChainIDs []uint64,
	gasFeeMode GasFeeMode,
	fromLockedAmount map[uint64]*hexutil.Big,
) (*SuggestedRoutes, error) {
	log.Debug("call to GetSuggestedRoutes")
	return api.router.suggestedRoutes(ctx, sendType, addrFrom, addrTo, amountIn.ToInt(), tokenID, disabledFromChainIDs,
		disabledToChaindIDs, preferedChainIDs, gasFeeMode, fromLockedAmount)
}

// Generates addresses for the provided paths, response doesn't include `HasActivity` value (if you need it check `GetAddressDetails` function)
func (api *API) GetDerivedAddresses(ctx context.Context, password string, derivedFrom string, paths []string) ([]*DerivedAddress, error) {
	info, err := api.s.gethManager.AccountsGenerator().LoadAccount(derivedFrom, password)
	if err != nil {
		return nil, err
	}

	return api.getDerivedAddresses(info.ID, paths)
}

// Generates addresses for the provided paths derived from the provided mnemonic, response doesn't include `HasActivity` value (if you need it check `GetAddressDetails` function)
func (api *API) GetDerivedAddressesForMnemonic(ctx context.Context, mnemonic string, paths []string) ([]*DerivedAddress, error) {
	mnemonicNoExtraSpaces := strings.Join(strings.Fields(mnemonic), " ")

	info, err := api.s.gethManager.AccountsGenerator().ImportMnemonic(mnemonicNoExtraSpaces, "")
	if err != nil {
		return nil, err
	}

	return api.getDerivedAddresses(info.ID, paths)
}

// Generates addresses for the provided paths, response doesn't include `HasActivity` value (if you need it check `GetAddressDetails` function)
func (api *API) getDerivedAddresses(id string, paths []string) ([]*DerivedAddress, error) {
	addedAccounts, err := api.s.accountsDB.GetActiveAccounts()
	if err != nil {
		return nil, err
	}

	info, err := api.s.gethManager.AccountsGenerator().DeriveAddresses(id, paths)
	if err != nil {
		return nil, err
	}

	derivedAddresses := make([]*DerivedAddress, 0)
	for accPath, acc := range info {

		derivedAddress := &DerivedAddress{
			Address:   common.HexToAddress(acc.Address),
			PublicKey: types.Hex2Bytes(acc.PublicKey),
			Path:      accPath,
		}

		for _, account := range addedAccounts {
			if types.Address(derivedAddress.Address) == account.Address {
				derivedAddress.AlreadyCreated = true
				break
			}
		}

		derivedAddresses = append(derivedAddresses, derivedAddress)
	}

	return derivedAddresses, nil
}

func (api *API) AddressExists(ctx context.Context, address types.Address) (bool, error) {
	return api.s.accountsDB.AddressExists(address)
}

// Returns details for the passed address (response doesn't include derivation path)
func (api *API) GetAddressDetails(ctx context.Context, chainID uint64, address string) (*DerivedAddress, error) {
	result := &DerivedAddress{
		Address: common.HexToAddress(address),
	}
	addressExists, err := api.s.accountsDB.AddressExists(types.Address(result.Address))
	if err != nil {
		return result, err
	}

	result.AlreadyCreated = addressExists

	chainClient, err := api.s.rpcClient.EthClient(chainID)
	if err != nil {
		return result, err
	}

	balance, err := api.s.tokenManager.GetChainBalance(ctx, chainClient, result.Address)
	if err != nil {
		return result, err
	}

	result.HasActivity = balance.Cmp(big.NewInt(0)) != 0
	return result, nil
}

func (api *API) SignMessage(ctx context.Context, message types.HexBytes, address common.Address, password string) (string, error) {
	log.Debug("[WalletAPI::SignMessage]", "message", message, "address", address)
	return api.s.transactionManager.SignMessage(message, address, password)
}

func (api *API) BuildTransaction(ctx context.Context, chainID uint64, sendTxArgsJSON string) (response *transfer.TxResponse, err error) {
	log.Debug("[WalletAPI::BuildTransaction]", "chainID", chainID, "sendTxArgsJSON", sendTxArgsJSON)
	var params transactions.SendTxArgs
	err = json.Unmarshal([]byte(sendTxArgsJSON), &params)
	if err != nil {
		return nil, err
	}
	return api.s.transactionManager.BuildTransaction(chainID, params)
}

func (api *API) BuildRawTransaction(ctx context.Context, chainID uint64, sendTxArgsJSON string, signature string) (response *transfer.TxResponse, err error) {
	log.Debug("[WalletAPI::BuildRawTransaction]", "chainID", chainID, "sendTxArgsJSON", sendTxArgsJSON, "signature", signature)

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return nil, err
	}

	var params transactions.SendTxArgs
	err = json.Unmarshal([]byte(sendTxArgsJSON), &params)
	if err != nil {
		return nil, err
	}

	return api.s.transactionManager.BuildRawTransaction(chainID, params, sig)
}

func (api *API) SendTransactionWithSignature(ctx context.Context, chainID uint64, txType transactions.PendingTrxType,
	sendTxArgsJSON string, signature string) (hash types.Hash, err error) {
	log.Debug("[WalletAPI::SendTransactionWithSignature]", "chainID", chainID, "txType", txType, "sendTxArgsJSON", sendTxArgsJSON, "signature", signature)
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return hash, err
	}

	var params transactions.SendTxArgs
	err = json.Unmarshal([]byte(sendTxArgsJSON), &params)
	if err != nil {
		return hash, err
	}
	return api.s.transactionManager.SendTransactionWithSignature(chainID, txType, params, sig)
}

func (api *API) CreateMultiTransaction(ctx context.Context, multiTransactionCommand *transfer.MultiTransactionCommand, data []*bridge.TransactionBridge, password string) (*transfer.MultiTransactionCommandResult, error) {
	log.Debug("[WalletAPI:: CreateMultiTransaction] create multi transaction")
	return api.s.transactionManager.CreateMultiTransactionFromCommand(ctx, multiTransactionCommand, data, api.router.bridges, password)
}

func (api *API) ProceedWithTransactionsSignatures(ctx context.Context, signatures map[string]transfer.SignatureDetails) (*transfer.MultiTransactionCommandResult, error) {
	log.Debug("[WalletAPI:: ProceedWithTransactionsSignatures] sign with signatures and send multi transaction")
	return api.s.transactionManager.ProceedWithTransactionsSignatures(ctx, signatures)
}

func (api *API) GetMultiTransactions(ctx context.Context, transactionIDs []wcommon.MultiTransactionIDType) ([]*transfer.MultiTransaction, error) {
	log.Debug("wallet.api.GetMultiTransactions", "IDs.len", len(transactionIDs))
	return api.s.transactionManager.GetMultiTransactions(ctx, transactionIDs)
}

func (api *API) GetCachedCurrencyFormats() (currency.FormatPerSymbol, error) {
	log.Debug("call to GetCachedCurrencyFormats")
	return api.s.currency.GetCachedCurrencyFormats()
}

func (api *API) FetchAllCurrencyFormats() (currency.FormatPerSymbol, error) {
	log.Debug("call to FetchAllCurrencyFormats")
	return api.s.currency.FetchAllCurrencyFormats()
}

// @deprecated replaced by session APIs; see #12120
func (api *API) FilterActivityAsync(requestID int32, addresses []common.Address, allAddresses bool, chainIDs []wcommon.ChainID, filter activity.Filter, offset int, limit int) error {
	log.Debug("wallet.api.FilterActivityAsync", "requestID", requestID, "addr.count", len(addresses), "allAddresses", allAddresses, "chainIDs.count", len(chainIDs), "offset", offset, "limit", limit)

	api.s.activity.FilterActivityAsync(requestID, addresses, allAddresses, chainIDs, filter, offset, limit)
	return nil
}

// @deprecated replaced by session APIs; see #12120
func (api *API) CancelActivityFilterTask(requestID int32) error {
	log.Debug("wallet.api.CancelActivityFilterTask", "requestID", requestID)

	api.s.activity.CancelFilterTask(requestID)
	return nil
}

func (api *API) StartActivityFilterSession(addresses []common.Address, allAddresses bool, chainIDs []wcommon.ChainID, filter activity.Filter, firstPageCount int) (activity.SessionID, error) {
	log.Debug("wallet.api.StartActivityFilterSession", "addr.count", len(addresses), "allAddresses", allAddresses, "chainIDs.count", len(chainIDs), "firstPageCount", firstPageCount)

	return api.s.activity.StartFilterSession(addresses, allAddresses, chainIDs, filter, firstPageCount), nil
}

func (api *API) UpdateActivityFilterForSession(sessionID activity.SessionID, filter activity.Filter, firstPageCount int) error {
	log.Debug("wallet.api.UpdateActivityFilterForSession", "sessionID", sessionID, "firstPageCount", firstPageCount)

	return api.s.activity.UpdateFilterForSession(sessionID, filter, firstPageCount)
}

func (api *API) ResetActivityFilterSession(id activity.SessionID, firstPageCount int) error {
	log.Debug("wallet.api.ResetActivityFilterSession", "id", id, "firstPageCount", firstPageCount)

	return api.s.activity.ResetFilterSession(id, firstPageCount)
}

func (api *API) GetMoreForActivityFilterSession(id activity.SessionID, pageCount int) error {
	log.Debug("wallet.api.GetMoreForActivityFilterSession", "id", id, "pageCount", pageCount)

	return api.s.activity.GetMoreForFilterSession(id, pageCount)
}

func (api *API) StopActivityFilterSession(id activity.SessionID) {
	log.Debug("wallet.api.StopActivityFilterSession", "id", id)

	api.s.activity.StopFilterSession(id)
}

func (api *API) GetMultiTxDetails(ctx context.Context, multiTxID int) (*activity.EntryDetails, error) {
	log.Debug("wallet.api.GetMultiTxDetails", "multiTxID", multiTxID)

	return api.s.activity.GetMultiTxDetails(ctx, multiTxID)
}

func (api *API) GetTxDetails(ctx context.Context, id string) (*activity.EntryDetails, error) {
	log.Debug("wallet.api.GetTxDetails", "id", id)

	return api.s.activity.GetTxDetails(ctx, id)
}

func (api *API) GetRecipientsAsync(requestID int32, chainIDs []wcommon.ChainID, addresses []common.Address, offset int, limit int) (ignored bool, err error) {
	log.Debug("wallet.api.GetRecipientsAsync", "addresses.len", len(addresses), "chainIDs.len", len(chainIDs), "offset", offset, "limit", limit)

	ignored = api.s.activity.GetRecipientsAsync(requestID, chainIDs, addresses, offset, limit)
	return ignored, err
}

func (api *API) GetOldestActivityTimestampAsync(requestID int32, addresses []common.Address) error {
	log.Debug("wallet.api.GetOldestActivityTimestamp", "addresses.len", len(addresses))

	api.s.activity.GetOldestTimestampAsync(requestID, addresses)
	return nil
}

func (api *API) GetActivityCollectiblesAsync(requestID int32, chainIDs []wcommon.ChainID, addresses []common.Address, offset int, limit int) error {
	log.Debug("wallet.api.GetActivityCollectiblesAsync", "addresses.len", len(addresses), "chainIDs.len", len(chainIDs), "offset", offset, "limit", limit)

	api.s.activity.GetActivityCollectiblesAsync(requestID, chainIDs, addresses, offset, limit)

	return nil
}

func (api *API) FetchChainIDForURL(ctx context.Context, rpcURL string) (*big.Int, error) {
	log.Debug("wallet.api.VerifyURL", "rpcURL", rpcURL)

	rpcClient, err := gethrpc.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial upstream server: %s", err)
	}
	client := ethclient.NewClient(rpcClient)
	return client.ChainID(ctx)
}

// WCPairSessionProposal responds to "session_proposal" event
func (api *API) WCPairSessionProposal(ctx context.Context, sessionProposalJSON string) (*wc.PairSessionResponse, error) {
	log.Debug("wallet.api.wc.PairSessionProposal", "proposal.len", len(sessionProposalJSON))

	var data wc.SessionProposal
	err := json.Unmarshal([]byte(sessionProposalJSON), &data)
	if err != nil {
		return nil, err
	}

	return api.s.walletConnect.PairSessionProposal(data)
}

// WCSaveOrUpdateSession records a session established between Status app and dapp
func (api *API) WCSaveOrUpdateSession(ctx context.Context, sessionProposalJSON string) error {
	log.Debug("wallet.api.wc.WCSaveOrUpdateSession", "proposal.len", len(sessionProposalJSON))

	var data wc.Session
	err := json.Unmarshal([]byte(sessionProposalJSON), &data)
	if err != nil {
		return err
	}

	return api.s.walletConnect.SaveOrUpdateSession(data)
}

// WCChangeSessionState changes the active state of a session
func (api *API) WCChangeSessionState(ctx context.Context, topic walletconnect.Topic, active bool) error {
	log.Debug("wallet.api.wc.WCChangeSessionState", "topic", topic, "active", active)

	return api.s.walletConnect.ChangeSessionState(topic, active)
}

// WCSessionRequest responds to "session_request" event
func (api *API) WCSessionRequest(ctx context.Context, sessionRequestJSON string) (*transfer.TxResponse, error) {
	log.Debug("wallet.api.wc.SessionRequest", "request.len", len(sessionRequestJSON))

	var request wc.SessionRequest
	err := json.Unmarshal([]byte(sessionRequestJSON), &request)
	if err != nil {
		return nil, err
	}

	return api.s.walletConnect.SessionRequest(request)
}

// WCAuthRequest responds to "auth_request" event
func (api *API) WCAuthRequest(ctx context.Context, address common.Address, authMessage string) (*transfer.TxResponse, error) {
	log.Debug("wallet.api.wc.AuthRequest", "address", address, "authMessage", authMessage)

	return api.s.walletConnect.AuthRequest(address, authMessage)
}
