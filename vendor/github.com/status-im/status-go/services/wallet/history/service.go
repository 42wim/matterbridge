package history

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"math/big"
	"reflect"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"

	statustypes "github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
	statusrpc "github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/rpc/network"

	"github.com/status-im/status-go/services/wallet/balance"
	"github.com/status-im/status-go/services/wallet/market"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

const minPointsForGraph = 14 // for minimal time frame - 7 days, twice a day

// EventBalanceHistoryUpdateStarted and EventBalanceHistoryUpdateDone are used to notify the UI that balance history is being updated
const (
	EventBalanceHistoryUpdateStarted           walletevent.EventType = "wallet-balance-history-update-started"
	EventBalanceHistoryUpdateFinished          walletevent.EventType = "wallet-balance-history-update-finished"
	EventBalanceHistoryUpdateFinishedWithError walletevent.EventType = "wallet-balance-history-update-finished-with-error"
)

type ValuePoint struct {
	Value     float64 `json:"value"`
	Timestamp uint64  `json:"time"`
}

type Service struct {
	balance         *Balance
	db              *sql.DB
	accountsDB      *accounts.Database
	eventFeed       *event.Feed
	rpcClient       *statusrpc.Client
	networkManager  *network.Manager
	tokenManager    *token.Manager
	serviceContext  context.Context
	cancelFn        context.CancelFunc
	transferWatcher *Watcher
	exchange        *Exchange
	balanceCache    balance.CacheIface
}

func NewService(db *sql.DB, accountsDB *accounts.Database, eventFeed *event.Feed, rpcClient *statusrpc.Client, tokenManager *token.Manager, marketManager *market.Manager, balanceCache balance.CacheIface) *Service {
	return &Service{
		balance:        NewBalance(NewBalanceDB(db)),
		db:             db,
		accountsDB:     accountsDB,
		eventFeed:      eventFeed,
		rpcClient:      rpcClient,
		networkManager: rpcClient.NetworkManager,
		tokenManager:   tokenManager,
		exchange:       NewExchange(marketManager),
		balanceCache:   balanceCache,
	}
}

func (s *Service) Stop() {
	if s.cancelFn != nil {
		s.cancelFn()
	}

	s.stopTransfersWatcher()
}

func (s *Service) triggerEvent(eventType walletevent.EventType, account statustypes.Address, message string) {
	s.eventFeed.Send(walletevent.Event{
		Type: eventType,
		Accounts: []common.Address{
			common.Address(account),
		},
		Message: message,
	})
}

func (s *Service) Start() {
	log.Debug("Starting balance history service")

	s.startTransfersWatcher()

	go func() {
		s.serviceContext, s.cancelFn = context.WithCancel(context.Background())

		err := s.updateBalanceHistory(s.serviceContext)
		if s.serviceContext.Err() != nil {
			s.triggerEvent(EventBalanceHistoryUpdateFinished, statustypes.Address{}, "Service canceled")
		}
		if err != nil {
			s.triggerEvent(EventBalanceHistoryUpdateFinishedWithError, statustypes.Address{}, err.Error())
		}
	}()
}

func (s *Service) mergeChainsBalances(chainIDs []uint64, addresses []common.Address, tokenSymbol string, fromTimestamp uint64, data map[uint64][]*entry) ([]*DataPoint, error) {
	log.Debug("Merging balances", "address", addresses, "tokenSymbol", tokenSymbol, "fromTimestamp", fromTimestamp, "len(data)", len(data))

	toTimestamp := uint64(time.Now().UTC().Unix())
	allData := make([]*entry, 0)

	// Add edge points per chain
	// Iterate over chainIDs param, not data keys, because data may not contain all the chains, but we need edge points for all of them
	for _, chainID := range chainIDs {
		// edge points are needed to properly calculate total balance, as they contain the balance for the first and last timestamp
		chainData, err := s.balance.addEdgePoints(chainID, tokenSymbol, addresses, fromTimestamp, toTimestamp, data[chainID])
		if err != nil {
			return nil, err
		}
		allData = append(allData, chainData...)
	}

	// Sort by timestamp
	sort.Slice(allData, func(i, j int) bool {
		return allData[i].timestamp < allData[j].timestamp
	})

	log.Debug("Sorted balances", "len", len(allData))
	for _, entry := range allData {
		log.Debug("Sorted balances", "entry", entry)
	}

	// Add padding points to make chart look nice
	if len(allData) < minPointsForGraph {
		allData, _ = addPaddingPoints(tokenSymbol, addresses, toTimestamp, allData, minPointsForGraph)
	}

	return entriesToDataPoints(allData)
}

// Expects sorted data
func entriesToDataPoints(data []*entry) ([]*DataPoint, error) {
	var resSlice []*DataPoint
	var groupedEntries []*entry // Entries with the same timestamp

	type AddressKey struct {
		Address common.Address
		ChainID uint64
	}

	sumBalances := func(balanceMap map[AddressKey]*big.Int) *big.Int {
		// Sum balances of all accounts and chains in current timestamp
		sum := big.NewInt(0)
		for _, balance := range balanceMap {
			sum.Add(sum, balance)
		}
		return sum
	}

	updateBalanceMap := func(balanceMap map[AddressKey]*big.Int, entries []*entry) map[AddressKey]*big.Int {
		// Update balance map for this timestamp
		for _, entry := range entries {
			if entry.chainID == 0 {
				continue
			}
			key := AddressKey{
				Address: entry.address,
				ChainID: entry.chainID,
			}
			balanceMap[key] = entry.balance
		}
		return balanceMap
	}

	// Balance map always contains current balance for each address in specific timestamp
	// It is required to sum up balances from previous timestamp from accounts not present in current timestamp
	balanceMap := make(map[AddressKey]*big.Int)

	for _, entry := range data {
		if len(groupedEntries) > 0 {
			if entry.timestamp == groupedEntries[0].timestamp {
				groupedEntries = append(groupedEntries, entry)
				continue
			} else {
				// Split grouped entries into addresses
				balanceMap = updateBalanceMap(balanceMap, groupedEntries)
				// Calculate balance for all the addresses
				cumulativeBalance := sumBalances(balanceMap)
				// Points in slice contain balances for all chains
				resSlice = appendPointToSlice(resSlice, &DataPoint{
					Timestamp: uint64(groupedEntries[0].timestamp),
					Balance:   (*hexutil.Big)(cumulativeBalance),
				})

				// Reset grouped entries
				groupedEntries = nil
				groupedEntries = append(groupedEntries, entry)
			}
		} else {
			groupedEntries = append(groupedEntries, entry)
		}
	}

	// If only edge points are present, groupedEntries will be non-empty
	if len(groupedEntries) > 0 {
		// Split grouped entries into addresses
		balanceMap = updateBalanceMap(balanceMap, groupedEntries)
		// Calculate balance for all the addresses
		cumulativeBalance := sumBalances(balanceMap)
		resSlice = appendPointToSlice(resSlice, &DataPoint{
			Timestamp: uint64(groupedEntries[0].timestamp),
			Balance:   (*hexutil.Big)(cumulativeBalance),
		})
	}

	return resSlice, nil
}

func appendPointToSlice(slice []*DataPoint, point *DataPoint) []*DataPoint {
	// Replace the last point in slice if it has the same timestamp or add a new one if different
	if len(slice) > 0 {
		if slice[len(slice)-1].Timestamp != point.Timestamp {
			// Timestamps are different, appending to slice
			slice = append(slice, point)
		} else {
			// Replace last item in slice because timestamps are the same
			slice[len(slice)-1] = point
		}
	} else {
		slice = append(slice, point)
	}

	return slice
}

// GetBalanceHistory returns token count balance
func (s *Service) GetBalanceHistory(ctx context.Context, chainIDs []uint64, addresses []common.Address, tokenSymbol string, currencySymbol string, fromTimestamp uint64) ([]*ValuePoint, error) {
	log.Debug("GetBalanceHistory", "chainIDs", chainIDs, "address", addresses, "tokenSymbol", tokenSymbol, "currencySymbol", currencySymbol, "fromTimestamp", fromTimestamp)

	chainDataMap := make(map[uint64][]*entry)
	for _, chainID := range chainIDs {
		chainData, err := s.balance.get(ctx, chainID, tokenSymbol, addresses, fromTimestamp) // TODO Make chainID a slice?
		if err != nil {
			return nil, err
		}

		if len(chainData) == 0 {
			continue
		}

		chainDataMap[chainID] = chainData
	}

	// Need to get balance for all the chains for the first timestamp, otherwise total values will be incorrect
	data, err := s.mergeChainsBalances(chainIDs, addresses, tokenSymbol, fromTimestamp, chainDataMap)

	if err != nil {
		return nil, err
	} else if len(data) == 0 {
		return make([]*ValuePoint, 0), nil
	}

	return s.dataPointsToValuePoints(chainIDs, tokenSymbol, currencySymbol, data)
}

func (s *Service) dataPointsToValuePoints(chainIDs []uint64, tokenSymbol string, currencySymbol string, data []*DataPoint) ([]*ValuePoint, error) {
	if len(data) == 0 {
		return make([]*ValuePoint, 0), nil
	}

	// Check if historical exchange rate for data point is present and fetch remaining if not
	lastDayTime := time.Unix(int64(data[len(data)-1].Timestamp), 0).UTC()
	currentTime := time.Now().UTC()
	currentDayStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	if lastDayTime.After(currentDayStart) {
		// No chance to have today, use the previous day value for the last data point
		lastDayTime = lastDayTime.AddDate(0, 0, -1)
	}

	lastDayValue, err := s.exchange.GetExchangeRateForDay(tokenSymbol, currencySymbol, lastDayTime)
	if err != nil {
		err := s.exchange.FetchAndCacheMissingRates(tokenSymbol, currencySymbol)
		if err != nil {
			log.Error("Error fetching exchange rates", "tokenSymbol", tokenSymbol, "currencySymbol", currencySymbol, "err", err)
			return nil, err
		}

		lastDayValue, err = s.exchange.GetExchangeRateForDay(tokenSymbol, currencySymbol, lastDayTime)
		if err != nil {
			log.Error("Exchange rate missing for", "tokenSymbol", tokenSymbol, "currencySymbol", currencySymbol, "lastDayTime", lastDayTime, "err", err)
			return nil, err
		}
	}

	decimals, err := s.decimalsForToken(tokenSymbol, chainIDs[0])
	if err != nil {
		return nil, err
	}
	weisInOneMain := big.NewFloat(math.Pow(10, float64(decimals)))

	var res []*ValuePoint
	for _, d := range data {
		var dayValue float32
		dayTime := time.Unix(int64(d.Timestamp), 0).UTC()
		if dayTime.After(currentDayStart) {
			// No chance to have today, use the previous day value for the last data point
			if lastDayValue > 0 {
				dayValue = lastDayValue
			} else {
				log.Warn("Exchange rate missing for", "dayTime", dayTime, "err", err)
				continue
			}
		} else {
			dayValue, err = s.exchange.GetExchangeRateForDay(tokenSymbol, currencySymbol, dayTime)
			if err != nil {
				log.Warn("Exchange rate missing for", "dayTime", dayTime, "err", err)
				continue
			}
		}

		// The big.Int values are discarded, hence copy the original values
		res = append(res, &ValuePoint{
			Timestamp: d.Timestamp,
			Value:     tokenToValue((*big.Int)(d.Balance), dayValue, weisInOneMain),
		})
	}

	return res, nil
}

func (s *Service) decimalsForToken(tokenSymbol string, chainID uint64) (int, error) {
	network := s.networkManager.Find(chainID)
	if network == nil {
		return 0, errors.New("network not found")
	}
	token := s.tokenManager.FindToken(network, tokenSymbol)
	if token == nil {
		return 0, errors.New("token not found")
	}
	return int(token.Decimals), nil
}

func tokenToValue(tokenCount *big.Int, mainDenominationValue float32, weisInOneMain *big.Float) float64 {
	weis := new(big.Float).SetInt(tokenCount)
	mainTokens := new(big.Float).Quo(weis, weisInOneMain)
	mainTokenValue := new(big.Float).SetFloat64(float64(mainDenominationValue))
	res, accuracy := new(big.Float).Mul(mainTokens, mainTokenValue).Float64()
	if res == 0 && accuracy == big.Below {
		return math.SmallestNonzeroFloat64
	} else if res == math.Inf(1) && accuracy == big.Above {
		return math.Inf(1)
	}

	return res
}

// updateBalanceHistory iterates over all networks depending on test/prod for the s.visibleTokenSymbol
// and updates the balance history for the given address
//
// expects ctx to have cancellation support and processing to be cancelled by the caller
func (s *Service) updateBalanceHistory(ctx context.Context) error {
	log.Debug("updateBalanceHistory started")

	addresses, err := s.accountsDB.GetWalletAddresses()
	if err != nil {
		return err
	}

	areTestNetworksEnabled, err := s.accountsDB.GetTestNetworksEnabled()
	if err != nil {
		return err
	}

	onlyEnabledNetworks := false
	networks, err := s.networkManager.Get(onlyEnabledNetworks)
	if err != nil {
		return err
	}

	for _, address := range addresses {
		s.triggerEvent(EventBalanceHistoryUpdateStarted, address, "")

		for _, network := range networks {
			if network.IsTest != areTestNetworksEnabled {
				continue
			}

			entries, err := s.balance.db.getEntriesWithoutBalances(network.ChainID, common.Address(address))
			if err != nil {
				log.Error("Error getting blocks without balances", "chainID", network.ChainID, "address", address.String(), "err", err)
				return err
			}

			log.Debug("Blocks without balances", "chainID", network.ChainID, "address", address.String(), "entries", entries)

			client, err := s.rpcClient.EthClient(network.ChainID)
			if err != nil {
				log.Error("Error getting client", "chainID", network.ChainID, "address", address.String(), "err", err)
				return err
			}

			err = s.addEntriesToDB(ctx, client, network, address, entries)
			if err != nil {
				return err
			}
		}
		s.triggerEvent(EventBalanceHistoryUpdateFinished, address, "")
	}

	log.Debug("updateBalanceHistory finished")
	return nil
}

func (s *Service) addEntriesToDB(ctx context.Context, client chain.ClientInterface, network *params.Network, address statustypes.Address, entries []*entry) (err error) {
	for _, entry := range entries {
		var balance *big.Int
		// tokenAddess is zero for native currency
		if (entry.tokenAddress == common.Address{}) {
			// Check in cache
			balance = s.balanceCache.GetBalance(common.Address(address), network.ChainID, entry.block)
			log.Debug("Balance from cache", "chainID", network.ChainID, "address", address.String(), "block", entry.block, "balance", balance)

			if balance == nil {
				balance, err = client.BalanceAt(ctx, common.Address(address), entry.block)
				if balance == nil {
					log.Error("Error getting balance", "chainID", network.ChainID, "address", address.String(), "err", err, "unwrapped", errors.Unwrap(err))
					return err
				}
				time.Sleep(50 * time.Millisecond) // TODO Remove this sleep after fixing exceeding rate limit
			}
			entry.tokenSymbol = network.NativeCurrencySymbol
		} else {
			// Check token first if it is supported
			token := s.tokenManager.FindTokenByAddress(network.ChainID, entry.tokenAddress)
			if token == nil {
				log.Warn("Token not found", "chainID", network.ChainID, "address", address.String(), "tokenAddress", entry.tokenAddress.String())
				// TODO Add "supported=false" flag to such tokens to avoid checking them again and again
				continue // Skip token that we don't have symbol for. For example we don't have tokens in store for goerli optimism
			} else {
				entry.tokenSymbol = token.Symbol
			}

			// Check balance for token
			balance, err = s.tokenManager.GetTokenBalanceAt(ctx, client, common.Address(address), entry.tokenAddress, entry.block)
			log.Debug("Balance from token manager", "chainID", network.ChainID, "address", address.String(), "block", entry.block, "balance", balance)

			if err != nil {
				log.Error("Error getting token balance", "chainID", network.ChainID, "address", address.String(), "tokenAddress", entry.tokenAddress.String(), "err", err)
				return err
			}
		}

		entry.balance = balance
		err = s.balance.db.add(entry)
		if err != nil {
			log.Error("Error adding balance", "chainID", network.ChainID, "address", address.String(), "err", err)
			return err
		}
	}

	return nil
}

func (s *Service) startTransfersWatcher() {
	if s.transferWatcher != nil {
		return
	}

	transferLoadedCb := func(chainID uint64, addresses []common.Address, block *big.Int) {
		log.Debug("Balance history watcher: transfer loaded:", "chainID", chainID, "addresses", addresses, "block", block.Uint64())

		client, err := s.rpcClient.EthClient(chainID)
		if err != nil {
			log.Error("Error getting client", "chainID", chainID, "err", err)
			return
		}

		transferDB := transfer.NewDB(s.db)

		for _, address := range addresses {
			network := s.networkManager.Find(chainID)

			transfers, err := transferDB.GetTransfersByAddressAndBlock(chainID, address, block, 1500) // 1500 is quite arbitrary and far from real, but should be enough to cover all transfers in a block
			if err != nil {
				log.Error("Error getting transfers", "chainID", chainID, "address", address.String(), "err", err)
				continue
			}

			if len(transfers) == 0 {
				log.Debug("No transfers found", "chainID", chainID, "address", address.String(), "block", block.Uint64())
				continue
			}

			entries := transfersToEntries(address, block, transfers) // TODO Remove address and block after testing that they match
			unique := removeDuplicates(entries)
			log.Debug("Entries after filtering", "entries", entries, "unique", unique)

			err = s.addEntriesToDB(s.serviceContext, client, network, statustypes.Address(address), unique)
			if err != nil {
				log.Error("Error adding entries to DB", "chainID", chainID, "address", address.String(), "err", err)
				continue
			}

			// No event triggering here, because noone cares about balance history updates yet
		}
	}

	s.transferWatcher = NewWatcher(s.eventFeed, transferLoadedCb)
	s.transferWatcher.Start()
}

func removeDuplicates(entries []*entry) []*entry {
	unique := make([]*entry, 0, len(entries))
	for _, entry := range entries {
		found := false
		for _, u := range unique {
			if reflect.DeepEqual(entry, u) {
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, entry)
		}
	}

	return unique
}

func transfersToEntries(address common.Address, block *big.Int, transfers []transfer.Transfer) []*entry {
	entries := make([]*entry, 0)

	for _, transfer := range transfers {
		if transfer.Address != address {
			panic("Address mismatch") // coding error
		}

		if transfer.BlockNumber.Cmp(block) != 0 {
			panic("Block number mismatch") // coding error
		}
		entry := &entry{
			chainID:      transfer.NetworkID,
			address:      transfer.Address,
			tokenAddress: transfer.Receipt.ContractAddress,
			block:        transfer.BlockNumber,
			timestamp:    (int64)(transfer.Timestamp),
		}

		entries = append(entries, entry)
	}

	return entries
}

func (s *Service) stopTransfersWatcher() {
	if s.transferWatcher != nil {
		s.transferWatcher.Stop()
		s.transferWatcher = nil
	}
}
