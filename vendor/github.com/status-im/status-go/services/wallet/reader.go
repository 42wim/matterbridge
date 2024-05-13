package wallet

import (
	"context"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/wallet/async"
	"github.com/status-im/status-go/services/wallet/community"
	"github.com/status-im/status-go/services/wallet/market"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

// WalletTickReload emitted every 15mn to reload the wallet balance and history
const EventWalletTickReload walletevent.EventType = "wallet-tick-reload"
const EventWalletTickCheckConnected walletevent.EventType = "wallet-tick-check-connected"

const (
	walletTickReloadPeriod      = 10 * time.Minute
	activityReloadDelay         = 30 // Wait this many seconds after activity is detected before triggering a wallet reload
	activityReloadMarginSeconds = 30 // Trigger a wallet reload if activity is detected this many seconds before the last reload
)

func getFixedCurrencies() []string {
	return []string{"USD"}
}

func belongsToMandatoryTokens(symbol string) bool {
	var mandatoryTokens = []string{"ETH", "DAI", "SNT", "STT"}
	for _, t := range mandatoryTokens {
		if t == symbol {
			return true
		}
	}
	return false
}

func NewReader(rpcClient *rpc.Client, tokenManager *token.Manager, marketManager *market.Manager, communityManager *community.Manager, accountsDB *accounts.Database, persistence *Persistence, walletFeed *event.Feed) *Reader {
	return &Reader{
		rpcClient:           rpcClient,
		tokenManager:        tokenManager,
		marketManager:       marketManager,
		communityManager:    communityManager,
		accountsDB:          accountsDB,
		persistence:         persistence,
		walletFeed:          walletFeed,
		refreshBalanceCache: true,
	}
}

type Reader struct {
	rpcClient                      *rpc.Client
	tokenManager                   *token.Manager
	marketManager                  *market.Manager
	communityManager               *community.Manager
	accountsDB                     *accounts.Database
	persistence                    *Persistence
	walletFeed                     *event.Feed
	cancel                         context.CancelFunc
	walletEventsWatcher            *walletevent.Watcher
	lastWalletTokenUpdateTimestamp sync.Map
	reloadDelayTimer               *time.Timer
	refreshBalanceCache            bool
	rw                             sync.RWMutex
}

type TokenMarketValues struct {
	MarketCap       float64 `json:"marketCap"`
	HighDay         float64 `json:"highDay"`
	LowDay          float64 `json:"lowDay"`
	ChangePctHour   float64 `json:"changePctHour"`
	ChangePctDay    float64 `json:"changePctDay"`
	ChangePct24hour float64 `json:"changePct24hour"`
	Change24hour    float64 `json:"change24hour"`
	Price           float64 `json:"price"`
	HasError        bool    `json:"hasError"`
}

type ChainBalance struct {
	RawBalance     string         `json:"rawBalance"`
	Balance        *big.Float     `json:"balance"`
	Balance1DayAgo string         `json:"balance1DayAgo"`
	Address        common.Address `json:"address"`
	ChainID        uint64         `json:"chainId"`
	HasError       bool           `json:"hasError"`
}

type Token struct {
	Name                    string                       `json:"name"`
	Symbol                  string                       `json:"symbol"`
	Decimals                uint                         `json:"decimals"`
	BalancesPerChain        map[uint64]ChainBalance      `json:"balancesPerChain"`
	Description             string                       `json:"description"`
	AssetWebsiteURL         string                       `json:"assetWebsiteUrl"`
	BuiltOn                 string                       `json:"builtOn"`
	MarketValuesPerCurrency map[string]TokenMarketValues `json:"marketValuesPerCurrency"`
	PegSymbol               string                       `json:"pegSymbol"`
	Verified                bool                         `json:"verified"`
	Image                   string                       `json:"image,omitempty"`
	CommunityData           *community.Data              `json:"community_data,omitempty"`
}

func splitVerifiedTokens(tokens []*token.Token) ([]*token.Token, []*token.Token) {
	verified := make([]*token.Token, 0)
	unverified := make([]*token.Token, 0)

	for _, t := range tokens {
		if t.Verified {
			verified = append(verified, t)
		} else {
			unverified = append(unverified, t)
		}
	}

	return verified, unverified
}

func getTokenBySymbols(tokens []*token.Token) map[string][]*token.Token {
	res := make(map[string][]*token.Token)

	for _, t := range tokens {
		if _, ok := res[t.Symbol]; !ok {
			res[t.Symbol] = make([]*token.Token, 0)
		}

		res[t.Symbol] = append(res[t.Symbol], t)
	}

	return res
}

func getTokenAddresses(tokens []*token.Token) []common.Address {
	set := make(map[common.Address]bool)
	for _, token := range tokens {
		set[token.Address] = true
	}
	res := make([]common.Address, 0)
	for address := range set {
		res = append(res, address)
	}
	return res
}

func (r *Reader) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	r.startWalletEventsWatcher()

	go func() {
		ticker := time.NewTicker(walletTickReloadPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.triggerWalletReload()
			}
		}
	}()
	return nil
}

func (r *Reader) Stop() {
	if r.cancel != nil {
		r.cancel()
	}

	r.stopWalletEventsWatcher()

	r.cancelDelayedWalletReload()

	r.lastWalletTokenUpdateTimestamp = sync.Map{}
}

func (r *Reader) triggerWalletReload() {
	r.cancelDelayedWalletReload()

	r.walletFeed.Send(walletevent.Event{
		Type: EventWalletTickReload,
	})
}

func (r *Reader) triggerDelayedWalletReload() {
	r.cancelDelayedWalletReload()

	r.reloadDelayTimer = time.AfterFunc(time.Duration(activityReloadDelay)*time.Second, r.triggerWalletReload)
}

func (r *Reader) cancelDelayedWalletReload() {

	if r.reloadDelayTimer != nil {
		r.reloadDelayTimer.Stop()
		r.reloadDelayTimer = nil
	}
}

func (r *Reader) startWalletEventsWatcher() {
	if r.walletEventsWatcher != nil {
		return
	}

	// Respond to ETH/Token transfers
	walletEventCb := func(event walletevent.Event) {
		if event.Type != transfer.EventInternalETHTransferDetected &&
			event.Type != transfer.EventInternalERC20TransferDetected {
			return
		}

		for _, address := range event.Accounts {
			timestamp, ok := r.lastWalletTokenUpdateTimestamp.Load(address)
			timecheck := int64(0)
			if ok {
				timecheck = timestamp.(int64) - activityReloadMarginSeconds
			}

			if !ok || event.At > timecheck {
				r.triggerDelayedWalletReload()
				r.invalidateBalanceCache()
				break
			}
		}
	}

	r.walletEventsWatcher = walletevent.NewWatcher(r.walletFeed, walletEventCb)

	r.walletEventsWatcher.Start()
}

func (r *Reader) stopWalletEventsWatcher() {
	if r.walletEventsWatcher != nil {
		r.walletEventsWatcher.Stop()
		r.walletEventsWatcher = nil
	}
}

func (r *Reader) tokensCachedForAddresses(addresses []common.Address) bool {
	for _, address := range addresses {
		cachedTokens, err := r.GetCachedWalletTokensWithoutMarketData()
		if err != nil {
			return false
		}

		_, ok := cachedTokens[address]
		if !ok {
			return false
		}
	}

	return true
}

func (r *Reader) isCacheTimestampValidForAddress(address common.Address) bool {
	_, ok := r.lastWalletTokenUpdateTimestamp.Load(address)
	return ok
}

func (r *Reader) areCacheTimestampsValid(addresses []common.Address) bool {
	for _, address := range addresses {
		if !r.isCacheTimestampValidForAddress(address) {
			return false
		}
	}

	return true
}

func (r *Reader) isBalanceCacheValid(addresses []common.Address) bool {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return !r.refreshBalanceCache && r.tokensCachedForAddresses(addresses) && r.areCacheTimestampsValid(addresses)
}

func (r *Reader) balanceRefreshed() {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.refreshBalanceCache = false
}

func (r *Reader) invalidateBalanceCache() {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.refreshBalanceCache = true
}

func (r *Reader) FetchOrGetCachedWalletBalances(ctx context.Context, addresses []common.Address) (map[common.Address][]Token, error) {
	if !r.isBalanceCacheValid(addresses) {
		balances, err := r.GetWalletTokenBalances(ctx, addresses)
		if err != nil {
			return nil, err
		}
		r.balanceRefreshed()

		return balances, nil
	}

	tokens, err := r.getWalletTokenBalances(ctx, addresses, false)
	return tokens, err
}

func (r *Reader) getWalletTokenBalances(ctx context.Context, addresses []common.Address, updateBalances bool) (map[common.Address][]Token, error) {
	areTestNetworksEnabled, err := r.accountsDB.GetTestNetworksEnabled()
	if err != nil {
		return nil, err
	}

	networks, err := r.rpcClient.NetworkManager.Get(false)
	if err != nil {
		return nil, err
	}
	availableNetworks := make([]*params.Network, 0)
	for _, network := range networks {
		if network.IsTest != areTestNetworksEnabled {
			continue
		}
		availableNetworks = append(availableNetworks, network)
	}

	cachedTokens, err := r.GetCachedWalletTokensWithoutMarketData()
	if err != nil {
		return nil, err
	}

	chainIDs := make([]uint64, 0)
	for _, network := range availableNetworks {
		chainIDs = append(chainIDs, network.ChainID)
	}

	allTokens, err := r.tokenManager.GetTokensByChainIDs(chainIDs)
	if err != nil {
		return nil, err
	}

	for _, network := range availableNetworks {
		allTokens = append(allTokens, r.tokenManager.ToToken(network))
	}

	tokenAddresses := getTokenAddresses(allTokens)

	clients, err := r.rpcClient.EthClients(chainIDs)
	if err != nil {
		return nil, err
	}

	verifiedTokens, unverifiedTokens := splitVerifiedTokens(allTokens)

	cachedBalancesPerChain := map[common.Address]map[common.Address]map[uint64]string{}
	updateAnyway := false
	if !updateBalances {
	cacheCheck:
		for _, address := range addresses {
			if res, ok := cachedTokens[address]; !ok || len(res) == 0 {
				updateAnyway = true
				break
			}

			networkFound := map[uint64]bool{}
			for _, token := range cachedTokens[address] {
				for _, chain := range chainIDs {
					if _, ok := token.BalancesPerChain[chain]; ok {
						networkFound[chain] = true
					}
				}
			}

			for _, chain := range chainIDs {
				if !networkFound[chain] {
					updateAnyway = true
					break cacheCheck
				}
			}
		}
	}

	if !updateBalances && !updateAnyway {
		for address, tokens := range cachedTokens {
			for _, token := range tokens {
				for _, balance := range token.BalancesPerChain {
					if _, ok := cachedBalancesPerChain[address]; !ok {
						cachedBalancesPerChain[address] = map[common.Address]map[uint64]string{}
					}
					if _, ok := cachedBalancesPerChain[address][balance.Address]; !ok {
						cachedBalancesPerChain[address][balance.Address] = map[uint64]string{}
					}
					cachedBalancesPerChain[address][balance.Address][balance.ChainID] = balance.RawBalance
				}
			}

		}
	}

	var latestBalances map[uint64]map[common.Address]map[common.Address]*hexutil.Big
	if updateBalances || updateAnyway {
		latestBalances, err = r.tokenManager.GetBalancesByChain(ctx, clients, addresses, tokenAddresses)
		if err != nil {
			for _, client := range clients {
				client.SetIsConnected(false)
			}
			log.Info("tokenManager.GetBalancesByChain error", "err", err)
			return nil, err
		}
	}

	result := make(map[common.Address][]Token)
	dayAgoTimestamp := time.Now().Add(-24 * time.Hour).Unix()

	for _, address := range addresses {
		for _, tokenList := range [][]*token.Token{verifiedTokens, unverifiedTokens} {
			for symbol, tokens := range getTokenBySymbols(tokenList) {
				balancesPerChain := make(map[uint64]ChainBalance)
				decimals := tokens[0].Decimals
				isVisible := false
				for _, token := range tokens {
					var balance *big.Float
					hexBalance := &big.Int{}
					if latestBalances != nil {
						hexBalance = latestBalances[token.ChainID][address][token.Address].ToInt()
					} else {
						if cachedRawBalance, ok := cachedBalancesPerChain[address][token.Address][token.ChainID]; ok {
							hexBalance, _ = new(big.Int).SetString(cachedRawBalance, 10)
						}
					}
					balance = big.NewFloat(0.0)
					if hexBalance != nil {
						balance = new(big.Float).Quo(
							new(big.Float).SetInt(hexBalance),
							big.NewFloat(math.Pow(10, float64(decimals))),
						)
					}

					hasError := false
					if client, ok := clients[token.ChainID]; ok {
						hasError = err != nil || !client.GetIsConnected()
					}
					if !isVisible {
						isVisible = balance.Cmp(big.NewFloat(0.0)) > 0 || r.isCachedToken(cachedTokens, address, token.Symbol, token.ChainID)
					}
					balancesPerChain[token.ChainID] = ChainBalance{
						RawBalance:     hexBalance.String(),
						Balance:        balance,
						Balance1DayAgo: "0",
						Address:        token.Address,
						ChainID:        token.ChainID,
						HasError:       hasError,
					}
				}

				if !isVisible && !belongsToMandatoryTokens(symbol) {
					continue
				}

				for _, balance := range balancesPerChain {
					balance1DayAgo, err := r.tokenManager.GetTokenHistoricalBalance(address, balance.ChainID, symbol, dayAgoTimestamp)
					if err != nil {
						return nil, err
					}
					if balance1DayAgo != nil {
						balance.Balance1DayAgo = balance1DayAgo.String()
						balancesPerChain[balance.ChainID] = balance
					}
				}

				walletToken := Token{
					Name:             tokens[0].Name,
					Symbol:           symbol,
					BalancesPerChain: balancesPerChain,
					Decimals:         decimals,
					PegSymbol:        token.GetTokenPegSymbol(symbol),
					Verified:         tokens[0].Verified,
					CommunityData:    tokens[0].CommunityData,
					Image:            tokens[0].Image,
				}

				result[address] = append(result[address], walletToken)
			}
		}
	}

	r.updateTokenUpdateTimestamp(addresses)

	return result, r.persistence.SaveTokens(result)
}

func (r *Reader) GetWalletTokenBalances(ctx context.Context, addresses []common.Address) (map[common.Address][]Token, error) {
	return r.getWalletTokenBalances(ctx, addresses, true)
}

func (r *Reader) GetWalletToken(ctx context.Context, addresses []common.Address) (map[common.Address][]Token, error) {
	areTestNetworksEnabled, err := r.accountsDB.GetTestNetworksEnabled()
	if err != nil {
		return nil, err
	}

	networks, err := r.rpcClient.NetworkManager.Get(false)
	if err != nil {
		return nil, err
	}
	availableNetworks := make([]*params.Network, 0)
	for _, network := range networks {
		if network.IsTest != areTestNetworksEnabled {
			continue
		}
		availableNetworks = append(availableNetworks, network)
	}

	cachedTokens, err := r.GetCachedWalletTokensWithoutMarketData()
	if err != nil {
		return nil, err
	}

	chainIDs := make([]uint64, 0)
	for _, network := range availableNetworks {
		chainIDs = append(chainIDs, network.ChainID)
	}

	currencies := make([]string, 0)
	currency, err := r.accountsDB.GetCurrency()
	if err != nil {
		return nil, err
	}
	currencies = append(currencies, currency)
	currencies = append(currencies, getFixedCurrencies()...)
	allTokens, err := r.tokenManager.GetTokensByChainIDs(chainIDs)

	if err != nil {
		return nil, err
	}
	for _, network := range availableNetworks {
		allTokens = append(allTokens, r.tokenManager.ToToken(network))
	}

	tokenAddresses := getTokenAddresses(allTokens)

	clients, err := r.rpcClient.EthClients(chainIDs)
	if err != nil {
		return nil, err
	}

	balances, err := r.tokenManager.GetBalancesByChain(ctx, clients, addresses, tokenAddresses)
	if err != nil {
		for _, client := range clients {
			client.SetIsConnected(false)
		}
		log.Info("tokenManager.GetBalancesByChain error", "err", err)
		return nil, err
	}

	verifiedTokens, unverifiedTokens := splitVerifiedTokens(allTokens)
	tokenSymbols := make([]string, 0)
	result := make(map[common.Address][]Token)

	for _, address := range addresses {
		for _, tokenList := range [][]*token.Token{verifiedTokens, unverifiedTokens} {
			for symbol, tokens := range getTokenBySymbols(tokenList) {
				balancesPerChain := make(map[uint64]ChainBalance)
				decimals := tokens[0].Decimals
				isVisible := false
				for _, token := range tokens {
					hexBalance := balances[token.ChainID][address][token.Address]
					balance := big.NewFloat(0.0)
					if hexBalance != nil {
						balance = new(big.Float).Quo(
							new(big.Float).SetInt(hexBalance.ToInt()),
							big.NewFloat(math.Pow(10, float64(decimals))),
						)
					}
					hasError := false
					if client, ok := clients[token.ChainID]; ok {
						hasError = err != nil || !client.GetIsConnected()
					}
					if !isVisible {
						isVisible = balance.Cmp(big.NewFloat(0.0)) > 0 || r.isCachedToken(cachedTokens, address, token.Symbol, token.ChainID)
					}
					balancesPerChain[token.ChainID] = ChainBalance{
						RawBalance: hexBalance.ToInt().String(),
						Balance:    balance,
						Address:    token.Address,
						ChainID:    token.ChainID,
						HasError:   hasError,
					}
				}

				if !isVisible && !belongsToMandatoryTokens(symbol) {
					continue
				}

				walletToken := Token{
					Name:             tokens[0].Name,
					Symbol:           symbol,
					BalancesPerChain: balancesPerChain,
					Decimals:         decimals,
					PegSymbol:        token.GetTokenPegSymbol(symbol),
					Verified:         tokens[0].Verified,
					CommunityData:    tokens[0].CommunityData,
					Image:            tokens[0].Image,
				}

				tokenSymbols = append(tokenSymbols, symbol)
				result[address] = append(result[address], walletToken)
			}
		}
	}

	var (
		group             = async.NewAtomicGroup(ctx)
		prices            = map[string]map[string]float64{}
		tokenDetails      = map[string]thirdparty.TokenDetails{}
		tokenMarketValues = map[string]thirdparty.TokenMarketValues{}
	)

	group.Add(func(parent context.Context) error {
		prices, err = r.marketManager.FetchPrices(tokenSymbols, currencies)
		if err != nil {
			log.Info("marketManager.FetchPrices err", err)
		}
		return nil
	})

	group.Add(func(parent context.Context) error {
		tokenDetails, err = r.marketManager.FetchTokenDetails(tokenSymbols)
		if err != nil {
			log.Info("marketManager.FetchTokenDetails err", err)
		}
		return nil
	})

	group.Add(func(parent context.Context) error {
		tokenMarketValues, err = r.marketManager.FetchTokenMarketValues(tokenSymbols, currency)
		if err != nil {
			log.Info("marketManager.FetchTokenMarketValues err", err)
		}
		return nil
	})

	select {
	case <-group.WaitAsync():
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	err = group.Error()
	if err != nil {
		return nil, err
	}

	for address, tokens := range result {
		for index, token := range tokens {
			marketValuesPerCurrency := make(map[string]TokenMarketValues)
			for _, currency := range currencies {
				if _, ok := tokenMarketValues[token.Symbol]; !ok {
					continue
				}
				marketValuesPerCurrency[currency] = TokenMarketValues{
					MarketCap:       tokenMarketValues[token.Symbol].MKTCAP,
					HighDay:         tokenMarketValues[token.Symbol].HIGHDAY,
					LowDay:          tokenMarketValues[token.Symbol].LOWDAY,
					ChangePctHour:   tokenMarketValues[token.Symbol].CHANGEPCTHOUR,
					ChangePctDay:    tokenMarketValues[token.Symbol].CHANGEPCTDAY,
					ChangePct24hour: tokenMarketValues[token.Symbol].CHANGEPCT24HOUR,
					Change24hour:    tokenMarketValues[token.Symbol].CHANGE24HOUR,
					Price:           prices[token.Symbol][currency],
					HasError:        !r.marketManager.IsConnected,
				}
			}

			if _, ok := tokenDetails[token.Symbol]; !ok {
				continue
			}

			result[address][index].Description = tokenDetails[token.Symbol].Description
			result[address][index].AssetWebsiteURL = tokenDetails[token.Symbol].AssetWebsiteURL
			result[address][index].BuiltOn = tokenDetails[token.Symbol].BuiltOn
			result[address][index].MarketValuesPerCurrency = marketValuesPerCurrency
		}
	}

	r.updateTokenUpdateTimestamp(addresses)

	return result, r.persistence.SaveTokens(result)
}

func (r *Reader) isCachedToken(cachedTokens map[common.Address][]Token, address common.Address, symbol string, chainID uint64) bool {
	if tokens, ok := cachedTokens[address]; ok {
		for _, t := range tokens {
			if t.Symbol != symbol {
				continue
			}
			_, ok := t.BalancesPerChain[chainID]
			if ok {
				return true
			}
		}
	}
	return false
}

// GetCachedWalletTokensWithoutMarketData returns the latest fetched balances, minus
// price information
func (r *Reader) GetCachedWalletTokensWithoutMarketData() (map[common.Address][]Token, error) {
	return r.persistence.GetTokens()
}

func (r *Reader) updateTokenUpdateTimestamp(addresses []common.Address) {
	for _, address := range addresses {
		r.lastWalletTokenUpdateTimestamp.Store(address, time.Now().Unix())
	}
}
