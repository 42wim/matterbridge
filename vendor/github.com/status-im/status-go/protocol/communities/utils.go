package communities

import (
	"fmt"
	"strings"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

func CalculateRequestID(publicKey string, communityID types.HexBytes) types.HexBytes {
	idString := fmt.Sprintf("%s-%s", publicKey, communityID)
	return crypto.Keccak256([]byte(idString))
}

func ExtractTokenCriteria(permissions []*CommunityTokenPermission) (erc20TokenCriteria map[uint64]map[string]*protobuf.TokenCriteria, erc721TokenCriteria map[uint64]map[string]*protobuf.TokenCriteria, ensTokenCriteria []string) {
	erc20TokenCriteria = make(map[uint64]map[string]*protobuf.TokenCriteria)
	erc721TokenCriteria = make(map[uint64]map[string]*protobuf.TokenCriteria)
	ensTokenCriteria = make([]string, 0)

	for _, tokenPermission := range permissions {
		for _, tokenRequirement := range tokenPermission.TokenCriteria {

			isERC721 := tokenRequirement.Type == protobuf.CommunityTokenType_ERC721
			isERC20 := tokenRequirement.Type == protobuf.CommunityTokenType_ERC20
			isENS := tokenRequirement.Type == protobuf.CommunityTokenType_ENS

			for chainID, contractAddress := range tokenRequirement.ContractAddresses {

				_, existsERC721 := erc721TokenCriteria[chainID]

				if isERC721 && !existsERC721 {
					erc721TokenCriteria[chainID] = make(map[string]*protobuf.TokenCriteria)
				}
				_, existsERC20 := erc20TokenCriteria[chainID]

				if isERC20 && !existsERC20 {
					erc20TokenCriteria[chainID] = make(map[string]*protobuf.TokenCriteria)
				}

				_, existsERC721 = erc721TokenCriteria[chainID][contractAddress]
				if isERC721 && !existsERC721 {
					erc721TokenCriteria[chainID][strings.ToLower(contractAddress)] = tokenRequirement
				}

				_, existsERC20 = erc20TokenCriteria[chainID][contractAddress]
				if isERC20 && !existsERC20 {
					erc20TokenCriteria[chainID][strings.ToLower(contractAddress)] = tokenRequirement
				}

				if isENS {
					ensTokenCriteria = append(ensTokenCriteria, tokenRequirement.EnsPattern)
				}
			}
		}
	}
	return
}
