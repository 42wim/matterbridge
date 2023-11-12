package token

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/contracts"
	"github.com/status-im/status-go/contracts/community-tokens/assets"
	"github.com/status-im/status-go/contracts/ethscan"
	"github.com/status-im/status-go/contracts/ierc20"
	eth_node_types "github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/rpc/network"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/services/communitytokens"
	"github.com/status-im/status-go/services/utils"
	"github.com/status-im/status-go/services/wallet/async"
	"github.com/status-im/status-go/services/wallet/community"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

const (
	EventCommunityTokenReceived walletevent.EventType = "wallet-community-token-received"
)

var requestTimeout = 20 * time.Second
var nativeChainAddress = common.HexToAddress("0x")

type Token struct {
	Address common.Address `json:"address"`
	Name    string         `json:"name"`
	Symbol  string         `json:"symbol"`
	// Decimals defines how divisible the token is. For example, 0 would be
	// indivisible, whereas 18 would allow very small amounts of the token
	// to be traded.
	Decimals uint   `json:"decimals"`
	ChainID  uint64 `json:"chainId"`
	// PegSymbol indicates that the token is pegged to some fiat currency, using the
	// ISO 4217 alphabetic code. For example, an empty string means it is not
	// pegged, while "USD" means it's pegged to the United States Dollar.
	PegSymbol string `json:"pegSymbol"`
	Image     string `json:"image,omitempty"`

	CommunityData *community.Data `json:"community_data,omitempty"`
	Verified      bool            `json:"verified"`
	TokenListID   string          `json:"tokenListId"`
}

type ReceivedToken struct {
	Address       common.Address  `json:"address"`
	Name          string          `json:"name"`
	Symbol        string          `json:"symbol"`
	Image         string          `json:"image,omitempty"`
	ChainID       uint64          `json:"chainId"`
	CommunityData *community.Data `json:"community_data,omitempty"`
	Balance       *big.Int        `json:"balance"`
	TxHash        common.Hash     `json:"txHash"`
}

func (t *Token) IsNative() bool {
	return t.Address == nativeChainAddress
}

type List struct {
	Name    string   `json:"name"`
	Tokens  []*Token `json:"tokens"`
	Source  string   `json:"source"`
	Version string   `json:"version"`
}

type ListWrapper struct {
	UpdatedAt int64   `json:"updatedAt"`
	Data      []*List `json:"data"`
}

type addressTokenMap = map[common.Address]*Token
type storeMap = map[uint64]addressTokenMap

type ManagerInterface interface {
	LookupTokenIdentity(chainID uint64, address common.Address, native bool) *Token
	LookupToken(chainID *uint64, tokenSymbol string) (token *Token, isNative bool)
}

// Manager is used for accessing token store. It changes the token store based on overridden tokens
type Manager struct {
	db                *sql.DB
	RPCClient         *rpc.Client
	ContractMaker     *contracts.ContractMaker
	networkManager    *network.Manager
	stores            []store // Set on init, not changed afterwards
	communityTokensDB *communitytokens.Database
	communityManager  *community.Manager
	mediaServer       *server.MediaServer
	walletFeed        *event.Feed

	tokens []*Token

	tokenLock sync.RWMutex
}

func mergeTokens(sliceLists [][]*Token) []*Token {
	allKeys := make(map[string]bool)
	res := []*Token{}
	for _, list := range sliceLists {
		for _, token := range list {
			key := strconv.FormatUint(token.ChainID, 10) + token.Address.String()
			if _, value := allKeys[key]; !value {
				allKeys[key] = true
				res = append(res, token)
			}
		}
	}
	return res
}

func NewTokenManager(
	db *sql.DB,
	RPCClient *rpc.Client,
	communityManager *community.Manager,
	networkManager *network.Manager,
	appDB *sql.DB,
	mediaServer *server.MediaServer,
	walletFeed *event.Feed,
) *Manager {
	maker, _ := contracts.NewContractMaker(RPCClient)
	stores := []store{newUniswapStore(), newDefaultStore()}
	tokens := make([]*Token, 0)

	networks, err := networkManager.GetAll()
	if err != nil {
		return nil
	}

	for _, store := range stores {
		validTokens := make([]*Token, 0)
		for _, token := range store.GetTokens() {
			token.Verified = true

			for _, network := range networks {
				if network.ChainID == token.ChainID {
					validTokens = append(validTokens, token)
					break
				}
			}
		}

		tokens = mergeTokens([][]*Token{tokens, validTokens})
	}

	return &Manager{
		db:                db,
		RPCClient:         RPCClient,
		ContractMaker:     maker,
		networkManager:    networkManager,
		communityManager:  communityManager,
		stores:            stores,
		communityTokensDB: communitytokens.NewCommunityTokensDatabase(appDB),
		tokens:            tokens,
		mediaServer:       mediaServer,
		walletFeed:        walletFeed,
	}
}

// overrideTokensInPlace overrides tokens in the store with the ones from the networks
// BEWARE: overridden tokens will have their original address removed and replaced by the one in networks
func overrideTokensInPlace(networks []params.Network, tokens []*Token) {
	for _, network := range networks {
		if len(network.TokenOverrides) == 0 {
			continue
		}

		for _, overrideToken := range network.TokenOverrides {
			for _, token := range tokens {
				if token.Symbol == overrideToken.Symbol {
					token.Address = overrideToken.Address
				}
			}
		}
	}
}

func (tm *Manager) getTokens() []*Token {
	tm.tokenLock.RLock()
	defer tm.tokenLock.RUnlock()

	return tm.tokens
}

func (tm *Manager) SetTokens(tokens []*Token) {
	tm.tokenLock.Lock()
	defer tm.tokenLock.Unlock()
	tm.tokens = tokens
}

func (tm *Manager) FindToken(network *params.Network, tokenSymbol string) *Token {
	if tokenSymbol == network.NativeCurrencySymbol {
		return tm.ToToken(network)
	}

	return tm.GetToken(network.ChainID, tokenSymbol)
}

func (tm *Manager) LookupToken(chainID *uint64, tokenSymbol string) (token *Token, isNative bool) {
	if chainID == nil {
		networks, err := tm.networkManager.Get(false)
		if err != nil {
			return nil, false
		}

		for _, network := range networks {
			if tokenSymbol == network.NativeCurrencySymbol {
				return tm.ToToken(network), true
			}
			token := tm.GetToken(network.ChainID, tokenSymbol)
			if token != nil {
				return token, false
			}
		}
	} else {
		network := tm.networkManager.Find(*chainID)
		if network != nil && tokenSymbol == network.NativeCurrencySymbol {
			return tm.ToToken(network), true
		}
		return tm.GetToken(*chainID, tokenSymbol), false
	}
	return nil, false
}

// GetToken returns token by chainID and tokenSymbol. Use ToToken for native token
func (tm *Manager) GetToken(chainID uint64, tokenSymbol string) *Token {
	allTokens, err := tm.GetTokens(chainID)
	if err != nil {
		return nil
	}
	for _, token := range allTokens {
		if token.Symbol == tokenSymbol {
			return token
		}
	}
	return nil
}

func (tm *Manager) LookupTokenIdentity(chainID uint64, address common.Address, native bool) *Token {
	network := tm.networkManager.Find(chainID)
	if native {
		return tm.ToToken(network)
	}

	return tm.FindTokenByAddress(chainID, address)
}

func (tm *Manager) FindTokenByAddress(chainID uint64, address common.Address) *Token {
	allTokens, err := tm.GetTokens(chainID)
	if err != nil {
		return nil
	}
	for _, token := range allTokens {
		if token.Address == address {
			return token
		}
	}

	return nil
}

func (tm *Manager) FindOrCreateTokenByAddress(ctx context.Context, chainID uint64, address common.Address) *Token {
	// If token comes datasource, simply returns it
	for _, token := range tm.getTokens() {
		if token.ChainID != chainID {
			continue
		}
		if token.Address == address {
			return token
		}
	}

	// Create custom token if not known or try to link with a community
	customTokens, err := tm.GetCustoms(false)
	if err != nil {
		return nil
	}

	for _, token := range customTokens {
		if token.Address == address {
			tm.discoverTokenCommunityID(ctx, token, address)
			return token
		}
	}

	token, err := tm.DiscoverToken(ctx, chainID, address)
	if err != nil {
		return nil
	}

	err = tm.UpsertCustom(*token)
	if err != nil {
		return nil
	}

	tm.discoverTokenCommunityID(ctx, token, address)
	return token
}

func (tm *Manager) MarkAsPreviouslyOwnedToken(token *Token, owner common.Address) error {
	if token == nil {
		return errors.New("token is nil")
	}
	if (owner == common.Address{}) {
		return errors.New("owner is nil")
	}
	count := 0
	err := tm.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM token_balances WHERE user_address = ? AND token_address = ? AND chain_id = ?)`, owner.Hex(), token.Address.Hex(), token.ChainID).Scan(&count)
	if err != nil || count > 0 {
		return err
	}
	_, err = tm.db.Exec(`INSERT INTO token_balances(user_address,token_name,token_symbol,token_address,token_decimals,chain_id,token_decimals,raw_balance,balance) VALUES (?,?,?,?,?,?,?,?,?)`, owner.Hex(), token.Name, token.Symbol, token.Address.Hex(), token.Decimals, token.ChainID, 0, "0", "0")
	return err
}

func (tm *Manager) discoverTokenCommunityID(ctx context.Context, token *Token, address common.Address) {
	if token == nil || token.CommunityData != nil {
		// Token is invalid or is alrady discovered. Nothing to do here.
		return
	}
	backend, err := tm.RPCClient.EthClient(token.ChainID)
	if err != nil {
		return
	}
	caller, err := assets.NewAssetsCaller(address, backend)
	if err != nil {
		return
	}
	uri, err := caller.BaseTokenURI(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return
	}

	update, err := tm.db.Prepare("UPDATE tokens SET community_id=? WHERE network_id=? AND address=?")
	if err != nil {
		log.Error("Cannot prepare token update query", err)
		return
	}

	if uri == "" {
		// Update token community ID to prevent further checks
		_, err := update.Exec("", token.ChainID, token.Address)
		if err != nil {
			log.Error("Cannot update community id", err)
		}
		return
	}

	uri = strings.TrimSuffix(uri, "/")
	communityIDHex, err := utils.DeserializePublicKey(uri)
	if err != nil {
		return
	}
	communityID := eth_node_types.EncodeHex(communityIDHex)

	token.CommunityData = &community.Data{
		ID: communityID,
	}

	_, err = update.Exec(communityID, token.ChainID, token.Address)
	if err != nil {
		log.Error("Cannot update community id", err)
	}
}

func (tm *Manager) FindSNT(chainID uint64) *Token {
	tokens, err := tm.GetTokens(chainID)
	if err != nil {
		return nil
	}

	for _, token := range tokens {
		if token.Symbol == "SNT" || token.Symbol == "STT" {
			return token
		}
	}

	return nil
}

func (tm *Manager) getNativeTokens() ([]*Token, error) {
	tokens := make([]*Token, 0)
	networks, err := tm.networkManager.Get(false)
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		tokens = append(tokens, tm.ToToken(network))
	}

	return tokens, nil
}

func (tm *Manager) GetAllTokens() ([]*Token, error) {
	allTokens, err := tm.GetCustoms(true)
	if err != nil {
		log.Error("can't fetch custom tokens", "error", err)
	}

	allTokens = append(tm.getTokens(), allTokens...)

	overrideTokensInPlace(tm.networkManager.GetConfiguredNetworks(), allTokens)

	native, err := tm.getNativeTokens()
	if err != nil {
		return nil, err
	}

	allTokens = append(allTokens, native...)

	return allTokens, nil
}

func (tm *Manager) GetTokens(chainID uint64) ([]*Token, error) {
	tokens, err := tm.GetAllTokens()
	if err != nil {
		return nil, err
	}

	res := make([]*Token, 0)

	for _, token := range tokens {
		if token.ChainID == chainID {
			res = append(res, token)
		}
	}

	return res, nil
}

func (tm *Manager) GetTokensByChainIDs(chainIDs []uint64) ([]*Token, error) {
	tokens, err := tm.GetAllTokens()
	if err != nil {
		return nil, err
	}

	res := make([]*Token, 0)

	for _, token := range tokens {
		for _, chainID := range chainIDs {
			if token.ChainID == chainID {
				res = append(res, token)
			}
		}
	}

	return res, nil
}

func (tm *Manager) GetList() *ListWrapper {
	data := make([]*List, 0)
	nativeTokens, err := tm.getNativeTokens()
	if err == nil {
		data = append(data, &List{
			Name:    "native",
			Tokens:  nativeTokens,
			Source:  "native",
			Version: "1.0.0",
		})
	}

	customTokens, err := tm.GetCustoms(true)
	if err == nil && len(customTokens) > 0 {
		data = append(data, &List{
			Name:    "custom",
			Tokens:  customTokens,
			Source:  "custom",
			Version: "1.0.0",
		})
	}

	updatedAt := time.Now().Unix()
	for _, store := range tm.stores {
		updatedAt = store.GetUpdatedAt()
		data = append(data, &List{
			Name:    store.GetName(),
			Tokens:  store.GetTokens(),
			Source:  store.GetSource(),
			Version: store.GetVersion(),
		})
	}
	return &ListWrapper{
		Data:      data,
		UpdatedAt: updatedAt,
	}
}

func (tm *Manager) DiscoverToken(ctx context.Context, chainID uint64, address common.Address) (*Token, error) {
	caller, err := tm.ContractMaker.NewERC20(chainID, address)
	if err != nil {
		return nil, err
	}

	name, err := caller.Name(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	symbol, err := caller.Symbol(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	decimal, err := caller.Decimals(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	return &Token{
		Address:  address,
		Name:     name,
		Symbol:   symbol,
		Decimals: uint(decimal),
		ChainID:  chainID,
	}, nil
}

func (tm *Manager) getTokensFromDB(query string, args ...any) ([]*Token, error) {
	communityTokens := []*token.CommunityToken{}
	if tm.communityTokensDB != nil {
		// Error is skipped because it's only returning optional metadata
		communityTokens, _ = tm.communityTokensDB.GetCommunityERC20Metadata()
	}

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rst []*Token
	for rows.Next() {
		token := &Token{}
		var communityIDDB sql.NullString
		err := rows.Scan(&token.Address, &token.Name, &token.Symbol, &token.Decimals, &token.ChainID, &communityIDDB)
		if err != nil {
			return nil, err
		}

		if communityIDDB.Valid {
			communityID := communityIDDB.String
			for _, communityToken := range communityTokens {
				if communityToken.CommunityID != communityID || uint64(communityToken.ChainID) != token.ChainID || communityToken.Symbol != token.Symbol {
					continue
				}
				token.Image = tm.mediaServer.MakeCommunityTokenImagesURL(communityID, token.ChainID, token.Symbol)
				break
			}

			token.CommunityData = &community.Data{
				ID: communityID,
			}
		}

		_ = tm.fillCommunityData(token)

		rst = append(rst, token)
	}

	return rst, nil
}

func (tm *Manager) GetCustoms(onlyCommunityCustoms bool) ([]*Token, error) {
	if onlyCommunityCustoms {
		return tm.getTokensFromDB("SELECT address, name, symbol, decimals, network_id, community_id FROM tokens WHERE community_id IS NOT NULL AND community_id != ''")
	}
	return tm.getTokensFromDB("SELECT address, name, symbol, decimals, network_id, community_id FROM tokens")
}

func (tm *Manager) ToToken(network *params.Network) *Token {
	return &Token{
		Address:  common.HexToAddress("0x"),
		Name:     network.NativeCurrencyName,
		Symbol:   network.NativeCurrencySymbol,
		Decimals: uint(network.NativeCurrencyDecimals),
		ChainID:  network.ChainID,
		Verified: true,
	}
}

func (tm *Manager) UpsertCustom(token Token) error {
	insert, err := tm.db.Prepare("INSERT OR REPLACE INTO TOKENS (network_id, address, name, symbol, decimals) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = insert.Exec(token.ChainID, token.Address, token.Name, token.Symbol, token.Decimals)
	return err
}

func (tm *Manager) DeleteCustom(chainID uint64, address common.Address) error {
	_, err := tm.db.Exec(`DELETE FROM TOKENS WHERE address = ? and network_id = ?`, address, chainID)
	return err
}

func (tm *Manager) GetTokenBalance(ctx context.Context, client chain.ClientInterface, account common.Address, token common.Address) (*big.Int, error) {
	caller, err := ierc20.NewIERC20Caller(token, client)
	if err != nil {
		return nil, err
	}

	return caller.BalanceOf(&bind.CallOpts{
		Context: ctx,
	}, account)
}

func (tm *Manager) GetTokenBalanceAt(ctx context.Context, client chain.ClientInterface, account common.Address, token common.Address, blockNumber *big.Int) (*big.Int, error) {
	caller, err := ierc20.NewIERC20Caller(token, client)
	if err != nil {
		return nil, err
	}

	balance, err := caller.BalanceOf(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: blockNumber,
	}, account)

	if err != nil {
		if err != bind.ErrNoCode {
			return nil, err
		}
		balance = big.NewInt(0)
	}

	return balance, nil
}

func (tm *Manager) GetChainBalance(ctx context.Context, client chain.ClientInterface, account common.Address) (*big.Int, error) {
	return client.BalanceAt(ctx, account, nil)
}

func (tm *Manager) GetBalance(ctx context.Context, client chain.ClientInterface, account common.Address, token common.Address) (*big.Int, error) {
	if token == nativeChainAddress {
		return tm.GetChainBalance(ctx, client, account)
	}

	return tm.GetTokenBalance(ctx, client, account, token)
}

func (tm *Manager) GetBalancesByChain(parent context.Context, clients map[uint64]chain.ClientInterface, accounts, tokens []common.Address) (map[uint64]map[common.Address]map[common.Address]*hexutil.Big, error) {
	return tm.GetBalancesAtByChain(parent, clients, accounts, tokens, nil)
}

func (tm *Manager) GetBalancesAtByChain(parent context.Context, clients map[uint64]chain.ClientInterface, accounts, tokens []common.Address, atBlocks map[uint64]*big.Int) (map[uint64]map[common.Address]map[common.Address]*hexutil.Big, error) {
	var (
		group    = async.NewAtomicGroup(parent)
		mu       sync.Mutex
		response = map[uint64]map[common.Address]map[common.Address]*hexutil.Big{}
	)

	updateBalance := func(chainID uint64, account common.Address, token common.Address, balance *big.Int) {
		mu.Lock()
		if _, ok := response[chainID]; !ok {
			response[chainID] = map[common.Address]map[common.Address]*hexutil.Big{}
		}

		if _, ok := response[chainID][account]; !ok {
			response[chainID][account] = map[common.Address]*hexutil.Big{}
		}

		if _, ok := response[chainID][account][token]; !ok {
			zeroHex := hexutil.Big(*big.NewInt(0))
			response[chainID][account][token] = &zeroHex
		}
		sum := big.NewInt(0).Add(response[chainID][account][token].ToInt(), balance)
		sumHex := hexutil.Big(*sum)
		response[chainID][account][token] = &sumHex
		mu.Unlock()
	}

	for clientIdx := range clients {
		// Keep the reference to the client. DO NOT USE A LOOP, the client will be overridden in the coroutine
		client := clients[clientIdx]

		ethScanContract, availableAtBlock, err := tm.ContractMaker.NewEthScan(client.NetworkID())
		if err != nil {
			log.Error("error scanning contract", "err", err)
			return nil, err
		}

		atBlock := atBlocks[client.NetworkID()]

		fetchChainBalance := false
		var tokenChunks [][]common.Address
		chunkSize := 500
		for i := 0; i < len(tokens); i += chunkSize {
			end := i + chunkSize
			if end > len(tokens) {
				end = len(tokens)
			}

			tokenChunks = append(tokenChunks, tokens[i:end])
		}

		for _, token := range tokens {
			if token == nativeChainAddress {
				fetchChainBalance = true
			}
		}
		if fetchChainBalance {
			group.Add(func(parent context.Context) error {
				ctx, cancel := context.WithTimeout(parent, requestTimeout)
				defer cancel()
				res, err := ethScanContract.EtherBalances(&bind.CallOpts{
					Context:     ctx,
					BlockNumber: atBlock,
				}, accounts)
				if err != nil {
					log.Error("can't fetch chain balance 5", err)
					return nil
				}
				for idx, account := range accounts {
					balance := new(big.Int)
					balance.SetBytes(res[idx].Data)
					updateBalance(client.NetworkID(), account, common.HexToAddress("0x"), balance)
				}

				return nil
			})
		}

		for accountIdx := range accounts {
			// Keep the reference to the account. DO NOT USE A LOOP, the account will be overridden in the coroutine
			account := accounts[accountIdx]
			for idx := range tokenChunks {
				// Keep the reference to the chunk. DO NOT USE A LOOP, the chunk will be overridden in the coroutine
				chunk := tokenChunks[idx]

				group.Add(func(parent context.Context) error {
					ctx, cancel := context.WithTimeout(parent, requestTimeout)
					defer cancel()
					var res []ethscan.BalanceScannerResult
					if atBlock == nil || big.NewInt(int64(availableAtBlock)).Cmp(atBlock) < 0 {
						res, err = ethScanContract.TokensBalance(&bind.CallOpts{
							Context:     ctx,
							BlockNumber: atBlock,
						}, account, chunk)
						if err != nil {
							log.Error("can't fetch erc20 token balance 6", "account", account, "error", err)
							return nil
						}

						if len(res) != len(chunk) {
							log.Error("can't fetch erc20 token balance 7", "account", account, "error response not complete")
							return nil
						}

						for idx, token := range chunk {
							if !res[idx].Success {
								continue
							}
							balance := new(big.Int)
							balance.SetBytes(res[idx].Data)
							updateBalance(client.NetworkID(), account, token, balance)
						}
						return nil
					}

					for _, token := range chunk {
						balance, err := tm.GetTokenBalanceAt(ctx, client, account, token, atBlock)
						if err != nil {
							if err != bind.ErrNoCode {
								log.Error("can't fetch erc20 token balance 8", "account", account, "token", token, "error on fetching token balance")

								return nil
							}
						}
						updateBalance(client.NetworkID(), account, token, balance)
					}

					return nil
				})
			}
		}

	}
	select {
	case <-group.WaitAsync():
	case <-parent.Done():
		return nil, parent.Err()
	}
	return response, group.Error()
}

func (tm *Manager) SignalCommunityTokenReceived(address common.Address, txHash common.Hash, value *big.Int, t *Token) {
	if tm.walletFeed == nil || t == nil || t.CommunityData == nil {
		return
	}

	if len(t.CommunityData.Name) == 0 {
		_ = tm.fillCommunityData(t)
	}
	if len(t.CommunityData.Name) == 0 && tm.communityManager != nil {
		communityData, _ := tm.communityManager.FetchCommunityMetadata(t.CommunityData.ID)
		if communityData != nil {
			t.CommunityData.Name = communityData.CommunityName
			t.CommunityData.Color = communityData.CommunityColor
			t.CommunityData.Image = tm.communityManager.GetCommunityImageURL(t.CommunityData.ID)
		}
	}

	receivedToken := ReceivedToken{
		Address:       t.Address,
		Name:          t.Name,
		Symbol:        t.Symbol,
		Image:         t.Image,
		ChainID:       t.ChainID,
		CommunityData: t.CommunityData,
		Balance:       value,
		TxHash:        txHash,
	}

	encodedMessage, err := json.Marshal(receivedToken)
	if err != nil {
		return
	}

	tm.walletFeed.Send(walletevent.Event{
		Type:    EventCommunityTokenReceived,
		ChainID: t.ChainID,
		Accounts: []common.Address{
			address,
		},
		Message: string(encodedMessage),
	})
}

func (tm *Manager) fillCommunityData(token *Token) error {
	if token == nil || token.CommunityData == nil || tm.communityManager == nil {
		return nil
	}

	communityInfo, _, err := tm.communityManager.GetCommunityInfo(token.CommunityData.ID)
	if err != nil {
		return err
	}
	if err == nil && communityInfo != nil {
		// Fetched data from cache. Cache is refreshed during every wallet token list call.
		token.CommunityData.Name = communityInfo.CommunityName
		token.CommunityData.Color = communityInfo.CommunityColor
		token.CommunityData.Image = communityInfo.CommunityImage
	}
	return nil
}
