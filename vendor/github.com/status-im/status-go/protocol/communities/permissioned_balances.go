package communities

import (
	"context"
	"math/big"
	"strconv"

	"github.com/pkg/errors"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

type PermissionedBalance struct {
	Type     protobuf.CommunityTokenType `json:"type"`
	Symbol   string                      `json:"symbol"`
	Name     string                      `json:"name"`
	Amount   *bigint.BigInt              `json:"amount"`
	Decimals uint64                      `json:"decimals"`
}

func calculatePermissionedBalancesERC20(
	accountAddresses []gethcommon.Address,
	balances BalancesByChain,
	tokenPermissions []*CommunityTokenPermission,
) map[gethcommon.Address]map[string]*PermissionedBalance {
	res := make(map[gethcommon.Address]map[string]*PermissionedBalance)

	// Set with composite key (chain ID + wallet address + contract address) to
	// store if we already processed the balance.
	usedBalances := make(map[string]bool)

	for _, permission := range tokenPermissions {
		for _, criteria := range permission.TokenCriteria {
			if criteria.Type != protobuf.CommunityTokenType_ERC20 {
				continue
			}

			for _, accountAddress := range accountAddresses {
				for chainID, hexContractAddress := range criteria.ContractAddresses {
					usedKey := strconv.FormatUint(chainID, 10) + "-" + accountAddress.Hex() + "-" + hexContractAddress

					if _, ok := balances[chainID]; !ok {
						continue
					}
					if _, ok := balances[chainID][accountAddress]; !ok {
						continue
					}

					contractAddress := gethcommon.HexToAddress(hexContractAddress)
					value, ok := balances[chainID][accountAddress][contractAddress]
					if !ok {
						continue
					}

					// Skip the contract address if it has been used already in the sum.
					if _, ok := usedBalances[usedKey]; ok {
						continue
					}

					if _, ok := res[accountAddress]; !ok {
						res[accountAddress] = make(map[string]*PermissionedBalance, 0)
					}
					if _, ok := res[accountAddress][criteria.Symbol]; !ok {
						res[accountAddress][criteria.Symbol] = &PermissionedBalance{
							Type:     criteria.Type,
							Symbol:   criteria.Symbol,
							Name:     criteria.Name,
							Decimals: criteria.Decimals,
							Amount:   &bigint.BigInt{Int: big.NewInt(0)},
						}
					}

					res[accountAddress][criteria.Symbol].Amount.Add(
						res[accountAddress][criteria.Symbol].Amount.Int,
						value.ToInt(),
					)
					usedBalances[usedKey] = true
				}
			}
		}
	}

	return res
}

func isERC721CriteriaSatisfied(tokenBalances []thirdparty.TokenBalance, criteria *protobuf.TokenCriteria) bool {
	// No token IDs to compare against, so the criteria is satisfied.
	if len(criteria.TokenIds) == 0 {
		return true
	}

	for _, tokenID := range criteria.TokenIds {
		tokenIDBigInt := new(big.Int).SetUint64(tokenID)
		for _, asset := range tokenBalances {
			if asset.TokenID.Cmp(tokenIDBigInt) == 0 && asset.Balance.Sign() > 0 {
				return true
			}
		}
	}

	return false
}

func (m *Manager) calculatePermissionedBalancesERC721(
	accountAddresses []gethcommon.Address,
	balances CollectiblesByChain,
	tokenPermissions []*CommunityTokenPermission,
) map[gethcommon.Address]map[string]*PermissionedBalance {
	res := make(map[gethcommon.Address]map[string]*PermissionedBalance)

	// Set with composite key (chain ID + wallet address + contract address) to
	// store if we already processed the balance.
	usedBalances := make(map[string]bool)

	for _, permission := range tokenPermissions {
		for _, criteria := range permission.TokenCriteria {
			if criteria.Type != protobuf.CommunityTokenType_ERC721 {
				continue
			}

			for _, accountAddress := range accountAddresses {
				for chainID, hexContractAddress := range criteria.ContractAddresses {
					usedKey := strconv.FormatUint(chainID, 10) + "-" + accountAddress.Hex() + "-" + hexContractAddress

					if _, ok := balances[chainID]; !ok {
						continue
					}
					if _, ok := balances[chainID][accountAddress]; !ok {
						continue
					}

					contractAddress := gethcommon.HexToAddress(hexContractAddress)
					tokenBalances, ok := balances[chainID][accountAddress][contractAddress]
					if !ok || len(tokenBalances) == 0 {
						continue
					}

					// Skip the contract address if it has been used already in the sum.
					if _, ok := usedBalances[usedKey]; ok {
						continue
					}

					usedBalances[usedKey] = true

					if _, ok := res[accountAddress]; !ok {
						res[accountAddress] = make(map[string]*PermissionedBalance, 0)
					}
					if _, ok := res[accountAddress][criteria.Symbol]; !ok {
						res[accountAddress][criteria.Symbol] = &PermissionedBalance{
							Type:     criteria.Type,
							Symbol:   criteria.Symbol,
							Name:     criteria.Name,
							Decimals: criteria.Decimals,
							Amount:   &bigint.BigInt{Int: big.NewInt(0)},
						}
					}

					if isERC721CriteriaSatisfied(tokenBalances, criteria) {
						// We don't care about summing balances, thus setting as 1 is
						// sufficient.
						res[accountAddress][criteria.Symbol].Amount = &bigint.BigInt{Int: big.NewInt(1)}
					}
				}
			}
		}
	}

	return res
}

func (m *Manager) calculatePermissionedBalances(
	chainIDs []uint64,
	accountAddresses []gethcommon.Address,
	erc20Balances BalancesByChain,
	erc721Balances CollectiblesByChain,
	tokenPermissions []*CommunityTokenPermission,
) map[gethcommon.Address][]PermissionedBalance {
	res := make(map[gethcommon.Address][]PermissionedBalance, 0)

	aggregatedERC721Balances := m.calculatePermissionedBalancesERC721(accountAddresses, erc721Balances, tokenPermissions)
	for accountAddress, tokens := range aggregatedERC721Balances {
		for _, permissionedToken := range tokens {
			if permissionedToken.Amount.Sign() > 0 {
				res[accountAddress] = append(res[accountAddress], *permissionedToken)
			}
		}
	}

	aggregatedERC20Balances := calculatePermissionedBalancesERC20(accountAddresses, erc20Balances, tokenPermissions)
	for accountAddress, tokens := range aggregatedERC20Balances {
		for _, permissionedToken := range tokens {
			if permissionedToken.Amount.Sign() > 0 {
				res[accountAddress] = append(res[accountAddress], *permissionedToken)
			}
		}
	}

	return res
}

func keepRoleTokenPermissions(tokenPermissions map[string]*CommunityTokenPermission) []*CommunityTokenPermission {
	res := make([]*CommunityTokenPermission, 0)
	for _, p := range tokenPermissions {
		if p.Type == protobuf.CommunityTokenPermission_BECOME_MEMBER ||
			p.Type == protobuf.CommunityTokenPermission_BECOME_ADMIN ||
			p.Type == protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER ||
			p.Type == protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER {
			res = append(res, p)
		}
	}
	return res
}

// GetPermissionedBalances returns balances indexed by account address.
//
// It assumes balances in different chains with the same symbol can be summed.
// It also assumes the criteria's decimals field is the same across different
// criteria when they refer to the same asset (by symbol).
func (m *Manager) GetPermissionedBalances(
	ctx context.Context,
	communityID types.HexBytes,
	accountAddresses []gethcommon.Address,
) (map[gethcommon.Address][]PermissionedBalance, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, err
	}
	if community == nil {
		return nil, errors.Errorf("community does not exist ID='%s'", communityID)
	}

	tokenPermissions := keepRoleTokenPermissions(community.TokenPermissions())

	allChainIDs, err := m.tokenManager.GetAllChainIDs()
	if err != nil {
		return nil, err
	}
	accountsAndChainIDs := combineAddressesAndChainIDs(accountAddresses, allChainIDs)

	erc20TokenCriteriaByChain, erc721TokenCriteriaByChain, _ := ExtractTokenCriteria(tokenPermissions)

	accounts := make([]gethcommon.Address, 0, len(accountsAndChainIDs))
	for _, accountAndChainIDs := range accountsAndChainIDs {
		accounts = append(accounts, accountAndChainIDs.Address)
	}

	erc20ChainIDsSet := make(map[uint64]bool)
	erc20TokenAddresses := make([]gethcommon.Address, 0)
	for chainID, criterionByContractAddress := range erc20TokenCriteriaByChain {
		erc20ChainIDsSet[chainID] = true
		for contractAddress := range criterionByContractAddress {
			erc20TokenAddresses = append(erc20TokenAddresses, gethcommon.HexToAddress(contractAddress))
		}
	}

	erc721ChainIDsSet := make(map[uint64]bool)
	for chainID := range erc721TokenCriteriaByChain {
		erc721ChainIDsSet[chainID] = true
	}

	erc20ChainIDs := calculateChainIDsSet(accountsAndChainIDs, erc20ChainIDsSet)
	erc721ChainIDs := calculateChainIDsSet(accountsAndChainIDs, erc721ChainIDsSet)

	erc20Balances, err := m.tokenManager.GetBalancesByChain(ctx, accounts, erc20TokenAddresses, erc20ChainIDs)
	if err != nil {
		return nil, err
	}

	erc721Balances := make(CollectiblesByChain)
	if len(erc721ChainIDs) > 0 {
		balances, err := m.GetOwnedERC721Tokens(accounts, erc721TokenCriteriaByChain, erc721ChainIDs)
		if err != nil {
			return nil, err
		}

		erc721Balances = balances
	}

	return m.calculatePermissionedBalances(allChainIDs, accountAddresses, erc20Balances, erc721Balances, tokenPermissions), nil
}
