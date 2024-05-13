package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/status-im/status-go/contracts"
	gaspriceoracle "github.com/status-im/status-go/contracts/gas-price-oracle"
	"github.com/status-im/status-go/contracts/ierc20"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/wallet/async"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/bridge"
	walletCommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/transactions"
)

const EstimateUsername = "RandomUsername"
const EstimatePubKey = "0x04bb2024ce5d72e45d4a4f8589ae657ef9745855006996115a23a1af88d536cf02c0524a585fce7bfa79d6a9669af735eda6205d6c7e5b3cdc2b8ff7b2fa1f0b56"
const ERC721TransferString = "ERC721Transfer"
const ERC1155TransferString = "ERC1155Transfer"

type SendType int

const (
	Transfer SendType = iota
	ENSRegister
	ENSRelease
	ENSSetPubKey
	StickersBuy
	Bridge
	ERC721Transfer
	ERC1155Transfer
)

func (s SendType) IsCollectiblesTransfer() bool {
	return s == ERC721Transfer || s == ERC1155Transfer
}

func (s SendType) FetchPrices(service *Service, tokenID string) (map[string]float64, error) {
	symbols := []string{tokenID, "ETH"}
	if s.IsCollectiblesTransfer() {
		symbols = []string{"ETH"}
	}

	pricesMap, err := service.marketManager.FetchPrices(symbols, []string{"USD"})
	if err != nil {
		return nil, err
	}
	prices := make(map[string]float64, 0)
	for symbol, pricePerCurrency := range pricesMap {
		prices[symbol] = pricePerCurrency["USD"]
	}
	if s.IsCollectiblesTransfer() {
		prices[tokenID] = 0
	}
	return prices, nil
}

func (s SendType) FindToken(service *Service, account common.Address, network *params.Network, tokenID string) *token.Token {
	if !s.IsCollectiblesTransfer() {
		return service.tokenManager.FindToken(network, tokenID)
	}

	parts := strings.Split(tokenID, ":")
	contractAddress := common.HexToAddress(parts[0])
	collectibleTokenID, success := new(big.Int).SetString(parts[1], 10)
	if !success {
		return nil
	}
	uniqueID, err := service.collectibles.GetOwnedCollectible(walletCommon.ChainID(network.ChainID), account, contractAddress, collectibleTokenID)
	if err != nil || uniqueID == nil {
		return nil
	}

	return &token.Token{
		Address:  contractAddress,
		Symbol:   collectibleTokenID.String(),
		Decimals: 0,
		ChainID:  network.ChainID,
	}
}

func (s SendType) isTransfer() bool {
	return s == Transfer || s.IsCollectiblesTransfer()
}

func (s SendType) isAvailableBetween(from, to *params.Network) bool {
	if s.IsCollectiblesTransfer() {
		return from.ChainID == to.ChainID
	}

	if s == Bridge {
		return from.ChainID != to.ChainID
	}

	return true
}

func (s SendType) canUseBridge(b bridge.Bridge) bool {
	if s == ERC721Transfer && b.Name() != ERC721TransferString {
		return false
	}

	if s != ERC721Transfer && b.Name() == ERC721TransferString {
		return false
	}

	if s == ERC1155Transfer && b.Name() != ERC1155TransferString {
		return false
	}

	if s != ERC1155Transfer && b.Name() == ERC1155TransferString {
		return false
	}

	return true
}

func (s SendType) isAvailableFor(network *params.Network) bool {
	if s == Transfer || s == Bridge || s.IsCollectiblesTransfer() {
		return true
	}

	if network.ChainID == 1 || network.ChainID == 5 || network.ChainID == 11155111 {
		return true
	}

	return false
}

func (s SendType) EstimateGas(service *Service, network *params.Network, from common.Address, tokenID string) uint64 {
	tx := transactions.SendTxArgs{
		From:  (types.Address)(from),
		Value: (*hexutil.Big)(zero),
	}
	if s == ENSRegister {
		estimate, err := service.ens.API().RegisterEstimate(context.Background(), network.ChainID, tx, EstimateUsername, EstimatePubKey)
		if err != nil {
			return 400000
		}
		return estimate
	}

	if s == ENSRelease {
		estimate, err := service.ens.API().ReleaseEstimate(context.Background(), network.ChainID, tx, EstimateUsername)
		if err != nil {
			return 200000
		}
		return estimate
	}

	if s == ENSSetPubKey {
		estimate, err := service.ens.API().SetPubKeyEstimate(context.Background(), network.ChainID, tx, fmt.Sprint(EstimateUsername, ".stateofus.eth"), EstimatePubKey)
		if err != nil {
			return 400000
		}
		return estimate
	}

	if s == StickersBuy {
		packID := &bigint.BigInt{Int: big.NewInt(2)}
		estimate, err := service.stickers.API().BuyEstimate(context.Background(), network.ChainID, (types.Address)(from), packID)
		if err != nil {
			return 400000
		}
		return estimate
	}

	return 0
}

var zero = big.NewInt(0)

type Path struct {
	BridgeName              string
	From                    *params.Network
	To                      *params.Network
	MaxAmountIn             *hexutil.Big
	AmountIn                *hexutil.Big
	AmountInLocked          bool
	AmountOut               *hexutil.Big
	GasAmount               uint64
	GasFees                 *SuggestedFees
	BonderFees              *hexutil.Big
	TokenFees               *big.Float
	Cost                    *big.Float
	EstimatedTime           TransactionEstimation
	ApprovalRequired        bool
	ApprovalGasFees         *big.Float
	ApprovalAmountRequired  *hexutil.Big
	ApprovalContractAddress *common.Address
}

func (p *Path) Equal(o *Path) bool {
	return p.From.ChainID == o.From.ChainID && p.To.ChainID == o.To.ChainID
}

type Graph = []*Node

type Node struct {
	Path     *Path
	Children Graph
}

func newNode(path *Path) *Node {
	return &Node{Path: path, Children: make(Graph, 0)}
}

func buildGraph(AmountIn *big.Int, routes []*Path, level int, sourceChainIDs []uint64) Graph {
	graph := make(Graph, 0)
	for _, route := range routes {
		found := false
		for _, chainID := range sourceChainIDs {
			if chainID == route.From.ChainID {
				found = true
				break
			}
		}
		if found {
			continue
		}
		node := newNode(route)

		newRoutes := make([]*Path, 0)
		for _, r := range routes {
			if route.Equal(r) {
				continue
			}
			newRoutes = append(newRoutes, r)
		}

		newAmountIn := new(big.Int).Sub(AmountIn, route.MaxAmountIn.ToInt())
		if newAmountIn.Sign() > 0 {
			newSourceChainIDs := make([]uint64, len(sourceChainIDs))
			copy(newSourceChainIDs, sourceChainIDs)
			newSourceChainIDs = append(newSourceChainIDs, route.From.ChainID)
			node.Children = buildGraph(newAmountIn, newRoutes, level+1, newSourceChainIDs)

			if len(node.Children) == 0 {
				continue
			}
		}

		graph = append(graph, node)
	}

	return graph
}

func (n Node) buildAllRoutes() [][]*Path {
	res := make([][]*Path, 0)

	if len(n.Children) == 0 && n.Path != nil {
		res = append(res, []*Path{n.Path})
	}

	for _, node := range n.Children {
		for _, route := range node.buildAllRoutes() {
			extendedRoute := route
			if n.Path != nil {
				extendedRoute = append([]*Path{n.Path}, route...)
			}
			res = append(res, extendedRoute)
		}
	}

	return res
}

func filterRoutes(routes [][]*Path, amountIn *big.Int, fromLockedAmount map[uint64]*hexutil.Big) [][]*Path {
	if len(fromLockedAmount) == 0 {
		return routes
	}

	filteredRoutesLevel1 := make([][]*Path, 0)
	for _, route := range routes {
		routeOk := true
		fromIncluded := make(map[uint64]bool)
		fromExcluded := make(map[uint64]bool)
		for chainID, amount := range fromLockedAmount {
			if amount.ToInt().Cmp(zero) == 0 {
				fromExcluded[chainID] = false
			} else {
				fromIncluded[chainID] = false
			}

		}
		for _, path := range route {
			if _, ok := fromExcluded[path.From.ChainID]; ok {
				routeOk = false
				break
			}
			if _, ok := fromIncluded[path.From.ChainID]; ok {
				fromIncluded[path.From.ChainID] = true
			}
		}
		for _, value := range fromIncluded {
			if !value {
				routeOk = false
				break
			}
		}

		if routeOk {
			filteredRoutesLevel1 = append(filteredRoutesLevel1, route)
		}
	}

	filteredRoutesLevel2 := make([][]*Path, 0)
	for _, route := range filteredRoutesLevel1 {
		routeOk := true
		for _, path := range route {
			if amount, ok := fromLockedAmount[path.From.ChainID]; ok {
				requiredAmountIn := new(big.Int).Sub(amountIn, amount.ToInt())
				restAmountIn := big.NewInt(0)

				for _, otherPath := range route {
					if path.Equal(otherPath) {
						continue
					}
					restAmountIn = new(big.Int).Add(otherPath.MaxAmountIn.ToInt(), restAmountIn)
				}
				if restAmountIn.Cmp(requiredAmountIn) >= 0 {
					path.AmountIn = amount
					path.AmountInLocked = true
				} else {
					routeOk = false
					break
				}
			}
		}
		if routeOk {
			filteredRoutesLevel2 = append(filteredRoutesLevel2, route)
		}
	}

	return filteredRoutesLevel2
}

func findBest(routes [][]*Path) []*Path {
	var best []*Path
	bestCost := big.NewFloat(math.Inf(1))
	for _, route := range routes {
		currentCost := big.NewFloat(0)
		for _, path := range route {
			currentCost = new(big.Float).Add(currentCost, path.Cost)
		}

		if currentCost.Cmp(bestCost) == -1 {
			best = route
			bestCost = currentCost
		}
	}

	return best
}

type SuggestedRoutes struct {
	Best                  []*Path
	Candidates            []*Path
	TokenPrice            float64
	NativeChainTokenPrice float64
}

func newSuggestedRoutes(
	amountIn *big.Int,
	candidates []*Path,
	fromLockedAmount map[uint64]*hexutil.Big,
) *SuggestedRoutes {
	if len(candidates) == 0 {
		return &SuggestedRoutes{
			Candidates: candidates,
			Best:       candidates,
		}
	}

	node := &Node{
		Path:     nil,
		Children: buildGraph(amountIn, candidates, 0, []uint64{}),
	}
	routes := node.buildAllRoutes()
	routes = filterRoutes(routes, amountIn, fromLockedAmount)
	best := findBest(routes)

	if len(best) > 0 {
		sort.Slice(best, func(i, j int) bool {
			return best[i].AmountInLocked
		})
		rest := new(big.Int).Set(amountIn)
		for _, path := range best {
			diff := new(big.Int).Sub(rest, path.MaxAmountIn.ToInt())
			if diff.Cmp(zero) >= 0 {
				path.AmountIn = (*hexutil.Big)(path.MaxAmountIn.ToInt())
			} else {
				path.AmountIn = (*hexutil.Big)(new(big.Int).Set(rest))
			}
			rest.Sub(rest, path.AmountIn.ToInt())
		}
	}

	return &SuggestedRoutes{
		Candidates: candidates,
		Best:       best,
	}
}

func NewRouter(s *Service) *Router {
	bridges := make(map[string]bridge.Bridge)
	transfer := bridge.NewTransferBridge(s.rpcClient, s.transactor)
	erc721Transfer := bridge.NewERC721TransferBridge(s.rpcClient, s.transactor)
	erc1155Transfer := bridge.NewERC1155TransferBridge(s.rpcClient, s.transactor)
	cbridge := bridge.NewCbridge(s.rpcClient, s.transactor, s.tokenManager)
	hop := bridge.NewHopBridge(s.rpcClient, s.transactor, s.tokenManager)
	bridges[transfer.Name()] = transfer
	bridges[erc721Transfer.Name()] = erc721Transfer
	bridges[hop.Name()] = hop
	bridges[cbridge.Name()] = cbridge
	bridges[erc1155Transfer.Name()] = erc1155Transfer

	return &Router{s, bridges, s.rpcClient}
}

func containsNetworkChainID(network *params.Network, chainIDs []uint64) bool {
	for _, chainID := range chainIDs {
		if chainID == network.ChainID {
			return true
		}
	}

	return false
}

type Router struct {
	s         *Service
	bridges   map[string]bridge.Bridge
	rpcClient *rpc.Client
}

func (r *Router) requireApproval(ctx context.Context, sendType SendType, bridge bridge.Bridge, account common.Address, network *params.Network, token *token.Token, amountIn *big.Int) (
	bool, *big.Int, uint64, uint64, *common.Address, error) {
	if sendType.IsCollectiblesTransfer() {
		return false, nil, 0, 0, nil, nil
	}

	if token.IsNative() {
		return false, nil, 0, 0, nil, nil
	}
	contractMaker, err := contracts.NewContractMaker(r.rpcClient)
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	bridgeAddress := bridge.GetContractAddress(network, token)
	if bridgeAddress == nil {
		return false, nil, 0, 0, nil, nil
	}

	contract, err := contractMaker.NewERC20(network.ChainID, token.Address)
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	allowance, err := contract.Allowance(&bind.CallOpts{
		Context: ctx,
	}, account, *bridgeAddress)

	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	if allowance.Cmp(amountIn) >= 0 {
		return false, nil, 0, 0, nil, nil
	}

	ethClient, err := r.rpcClient.EthClient(network.ChainID)
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	erc20ABI, err := abi.JSON(strings.NewReader(ierc20.IERC20ABI))
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	data, err := erc20ABI.Pack("approve", bridgeAddress, amountIn)
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	estimate, err := ethClient.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  account,
		To:    &token.Address,
		Value: big.NewInt(0),
		Data:  data,
	})
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	// fetching l1 fee
	oracleContractAddress, err := gaspriceoracle.ContractAddress(network.ChainID)
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	oracleContract, err := gaspriceoracle.NewGaspriceoracleCaller(oracleContractAddress, ethClient)
	if err != nil {
		return false, nil, 0, 0, nil, err
	}

	callOpt := &bind.CallOpts{}

	l1Fee, _ := oracleContract.GetL1Fee(callOpt, data)

	return true, amountIn, estimate, l1Fee.Uint64(), bridgeAddress, nil
}

func (r *Router) getBalance(ctx context.Context, network *params.Network, token *token.Token, account common.Address) (*big.Int, error) {
	client, err := r.s.rpcClient.EthClient(network.ChainID)
	if err != nil {
		return nil, err
	}

	return r.s.tokenManager.GetBalance(ctx, client, account, token.Address)
}

func (r *Router) getERC1155Balance(ctx context.Context, network *params.Network, token *token.Token, account common.Address) (*big.Int, error) {
	tokenID, success := new(big.Int).SetString(token.Symbol, 10)
	if !success {
		return nil, errors.New("failed to convert token symbol to big.Int")
	}

	balances, err := r.s.collectiblesManager.FetchERC1155Balances(
		ctx,
		account,
		walletCommon.ChainID(network.ChainID),
		token.Address,
		[]*bigint.BigInt{&bigint.BigInt{Int: tokenID}},
	)
	if err != nil {
		return nil, err
	}

	if len(balances) != 1 || balances[0] == nil {
		return nil, errors.New("invalid ERC1155 balance fetch response")
	}

	return balances[0].Int, nil
}

func (r *Router) suggestedRoutes(
	ctx context.Context,
	sendType SendType,
	addrFrom common.Address,
	addrTo common.Address,
	amountIn *big.Int,
	tokenID string,
	disabledFromChainIDs,
	disabledToChaindIDs,
	preferedChainIDs []uint64,
	gasFeeMode GasFeeMode,
	fromLockedAmount map[uint64]*hexutil.Big,
) (*SuggestedRoutes, error) {
	areTestNetworksEnabled, err := r.s.accountsDB.GetTestNetworksEnabled()
	if err != nil {
		return nil, err
	}

	networks, err := r.s.rpcClient.NetworkManager.Get(false)
	if err != nil {
		return nil, err
	}

	prices, err := sendType.FetchPrices(r.s, tokenID)
	if err != nil {
		return nil, err
	}
	var (
		group      = async.NewAtomicGroup(ctx)
		mu         sync.Mutex
		candidates = make([]*Path, 0)
	)
	for networkIdx := range networks {
		network := networks[networkIdx]
		if network.IsTest != areTestNetworksEnabled {
			continue
		}

		if containsNetworkChainID(network, disabledFromChainIDs) {
			continue
		}

		if !sendType.isAvailableFor(network) {
			continue
		}

		token := sendType.FindToken(r.s, addrFrom, network, tokenID)
		if token == nil {
			continue
		}

		nativeToken := r.s.tokenManager.FindToken(network, network.NativeCurrencySymbol)
		if nativeToken == nil {
			continue
		}

		group.Add(func(c context.Context) error {
			gasFees, err := r.s.feesManager.suggestedFees(ctx, network.ChainID)
			if err != nil {
				return err
			}

			// Default value is 1 as in case of erc721 as we built the token we are sure the account owns it
			balance := big.NewInt(1)
			if sendType == ERC1155Transfer {
				balance, err = r.getERC1155Balance(ctx, network, token, addrFrom)
				if err != nil {
					return err
				}
			} else if sendType != ERC721Transfer {
				balance, err = r.getBalance(ctx, network, token, addrFrom)
				if err != nil {
					return err
				}
			}

			maxAmountIn := (*hexutil.Big)(balance)
			if amount, ok := fromLockedAmount[network.ChainID]; ok {
				if amount.ToInt().Cmp(balance) == 1 {
					return errors.New("locked amount cannot be bigger than balance")
				}
				maxAmountIn = amount
			}

			nativeBalance, err := r.getBalance(ctx, network, nativeToken, addrFrom)
			if err != nil {
				return err
			}
			maxFees := gasFees.feeFor(gasFeeMode)

			estimatedTime := r.s.feesManager.transactionEstimatedTime(ctx, network.ChainID, maxFees)
			for _, bridge := range r.bridges {
				if !sendType.canUseBridge(bridge) {
					continue
				}

				for _, dest := range networks {
					if dest.IsTest != areTestNetworksEnabled {
						continue
					}

					if !sendType.isAvailableFor(network) {
						continue
					}

					if !sendType.isAvailableBetween(network, dest) {
						continue
					}

					if len(preferedChainIDs) > 0 && !containsNetworkChainID(dest, preferedChainIDs) {
						continue
					}
					if containsNetworkChainID(dest, disabledToChaindIDs) {
						continue
					}

					can, err := bridge.Can(network, dest, token, maxAmountIn.ToInt())
					if err != nil || !can {
						continue
					}
					bonderFees, tokenFees, err := bridge.CalculateFees(network, dest, token, amountIn, prices["ETH"], prices[tokenID], gasFees.GasPrice)
					if err != nil {
						continue
					}
					if bonderFees.Cmp(zero) != 0 {
						if maxAmountIn.ToInt().Cmp(amountIn) >= 0 {
							if bonderFees.Cmp(amountIn) >= 0 {
								continue
							}
						} else {
							if bonderFees.Cmp(maxAmountIn.ToInt()) >= 0 {
								continue
							}
						}
					}
					gasLimit := uint64(0)
					if sendType.isTransfer() {
						gasLimit, err = bridge.EstimateGas(network, dest, addrFrom, addrTo, token, amountIn)
						if err != nil {
							continue
						}
					} else {
						gasLimit = sendType.EstimateGas(r.s, network, addrFrom, tokenID)
					}

					approvalRequired, approvalAmountRequired, approvalGasLimit, l1ApprovalFee, approvalContractAddress, err := r.requireApproval(ctx, sendType, bridge, addrFrom, network, token, amountIn)
					if err != nil {
						continue
					}

					tx, err := bridge.BuildTx(network, dest, addrFrom, addrTo, token, amountIn, bonderFees)
					if err != nil {
						continue
					}

					l1GasFeeWei, _ := r.s.feesManager.getL1Fee(ctx, network.ChainID, tx)
					l1GasFeeWei += l1ApprovalFee
					gasFees.L1GasFee = weiToGwei(big.NewInt(int64(l1GasFeeWei)))

					requiredNativeBalance := new(big.Int).Mul(gweiToWei(maxFees), big.NewInt(int64(gasLimit)))
					requiredNativeBalance.Add(requiredNativeBalance, new(big.Int).Mul(gweiToWei(maxFees), big.NewInt(int64(approvalGasLimit))))
					requiredNativeBalance.Add(requiredNativeBalance, big.NewInt(int64(l1GasFeeWei))) // add l1Fee to requiredNativeBalance, in case of L1 chain l1Fee is 0

					if nativeBalance.Cmp(requiredNativeBalance) <= 0 {
						continue
					}

					// Removed the required fees from maxAMount in case of native token tx
					if token.IsNative() {
						maxAmountIn = (*hexutil.Big)(new(big.Int).Sub(maxAmountIn.ToInt(), requiredNativeBalance))
					}

					ethPrice := big.NewFloat(prices["ETH"])

					approvalGasFees := new(big.Float).Mul(gweiToEth(maxFees), big.NewFloat((float64(approvalGasLimit))))

					approvalGasCost := new(big.Float)
					approvalGasCost.Mul(approvalGasFees, ethPrice)

					l1GasCost := new(big.Float)
					l1GasCost.Mul(gasFees.L1GasFee, ethPrice)

					gasCost := new(big.Float)
					gasCost.Mul(new(big.Float).Mul(gweiToEth(maxFees), big.NewFloat(float64(gasLimit))), ethPrice)

					tokenFeesAsFloat := new(big.Float).Quo(
						new(big.Float).SetInt(tokenFees),
						big.NewFloat(math.Pow(10, float64(token.Decimals))),
					)
					tokenCost := new(big.Float)
					tokenCost.Mul(tokenFeesAsFloat, big.NewFloat(prices[tokenID]))

					cost := new(big.Float)
					cost.Add(tokenCost, gasCost)
					cost.Add(cost, approvalGasCost)
					cost.Add(cost, l1GasCost)
					mu.Lock()
					candidates = append(candidates, &Path{
						BridgeName:              bridge.Name(),
						From:                    network,
						To:                      dest,
						MaxAmountIn:             maxAmountIn,
						AmountIn:                (*hexutil.Big)(zero),
						AmountOut:               (*hexutil.Big)(zero),
						GasAmount:               gasLimit,
						GasFees:                 gasFees,
						BonderFees:              (*hexutil.Big)(bonderFees),
						TokenFees:               tokenFeesAsFloat,
						Cost:                    cost,
						EstimatedTime:           estimatedTime,
						ApprovalRequired:        approvalRequired,
						ApprovalGasFees:         approvalGasFees,
						ApprovalAmountRequired:  (*hexutil.Big)(approvalAmountRequired),
						ApprovalContractAddress: approvalContractAddress,
					})
					mu.Unlock()
				}
			}
			return nil
		})
	}

	group.Wait()

	suggestedRoutes := newSuggestedRoutes(amountIn, candidates, fromLockedAmount)
	suggestedRoutes.TokenPrice = prices[tokenID]
	suggestedRoutes.NativeChainTokenPrice = prices["ETH"]
	for _, path := range suggestedRoutes.Best {
		amountOut, err := r.bridges[path.BridgeName].CalculateAmountOut(path.From, path.To, (*big.Int)(path.AmountIn), tokenID)
		if err != nil {
			continue
		}
		path.AmountOut = (*hexutil.Big)(amountOut)
	}

	return suggestedRoutes, nil
}
