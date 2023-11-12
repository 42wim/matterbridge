package thirdparty

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/services/wallet/bigint"
	w_common "github.com/status-im/status-go/services/wallet/common"
)

func generateContractType(seed int) w_common.ContractType {
	if seed%2 == 0 {
		return w_common.ContractTypeERC721
	}
	return w_common.ContractTypeERC1155
}

func GenerateTestCollectiblesData(count int) (result []CollectibleData) {
	base := rand.Intn(100) // nolint: gosec

	result = make([]CollectibleData, 0, count)
	for i := base; i < count+base; i++ {
		bigI := big.NewInt(int64(i))
		newCollectible := CollectibleData{
			ID: CollectibleUniqueID{
				ContractID: ContractID{
					ChainID: w_common.ChainID(i % 4),
					Address: common.BigToAddress(bigI),
				},
				TokenID: &bigint.BigInt{Int: bigI},
			},
			ContractType:       generateContractType(i),
			Provider:           fmt.Sprintf("provider-%d", i),
			Name:               fmt.Sprintf("name-%d", i),
			Description:        fmt.Sprintf("description-%d", i),
			Permalink:          fmt.Sprintf("permalink-%d", i),
			ImageURL:           fmt.Sprintf("imageurl-%d", i),
			ImagePayload:       []byte(fmt.Sprintf("imagepayload-%d", i)),
			AnimationURL:       fmt.Sprintf("animationurl-%d", i),
			AnimationMediaType: fmt.Sprintf("animationmediatype-%d", i),
			Traits: []CollectibleTrait{
				{
					TraitType:   fmt.Sprintf("traittype-%d", i),
					Value:       fmt.Sprintf("traitvalue-%d", i),
					DisplayType: fmt.Sprintf("displaytype-%d", i),
					MaxValue:    fmt.Sprintf("maxvalue-%d", i),
				},
				{
					TraitType:   fmt.Sprintf("traittype-%d", i),
					Value:       fmt.Sprintf("traitvalue-%d", i),
					DisplayType: fmt.Sprintf("displaytype-%d", i),
					MaxValue:    fmt.Sprintf("maxvalue-%d", i),
				},
				{
					TraitType:   fmt.Sprintf("traittype-%d", i),
					Value:       fmt.Sprintf("traitvalue-%d", i),
					DisplayType: fmt.Sprintf("displaytype-%d", i),
					MaxValue:    fmt.Sprintf("maxvalue-%d", i),
				},
			},
			BackgroundColor: fmt.Sprintf("backgroundcolor-%d", i),
			TokenURI:        fmt.Sprintf("tokenuri-%d", i),
			CommunityID:     fmt.Sprintf("communityid-%d", i%5),
		}
		result = append(result, newCollectible)
	}
	return result
}

func GenerateTestCollectiblesCommunityData(count int) []CollectibleCommunityInfo {
	base := rand.Intn(100) // nolint: gosec

	result := make([]CollectibleCommunityInfo, 0, count)
	for i := base; i < count+base; i++ {
		newCommunityInfo := CollectibleCommunityInfo{
			PrivilegesLevel: token.PrivilegesLevel(i) % (token.CommunityLevel + 1),
		}
		result = append(result, newCommunityInfo)
	}
	return result
}

func GenerateTestCollectiblesOwnership(count int) []AccountBalance {
	base := rand.Intn(100) // nolint: gosec

	ret := make([]AccountBalance, 0, count)
	for i := base; i < count+base; i++ {
		ret = append(ret, AccountBalance{
			Address: common.HexToAddress(fmt.Sprintf("0x%x", i)),
			Balance: &bigint.BigInt{Int: big.NewInt(int64(i))},
		})
	}
	return ret
}

func GenerateTestCollectionsData(count int) (result []CollectionData) {
	base := rand.Intn(100) // nolint: gosec

	result = make([]CollectionData, 0, count)
	for i := base; i < count+base; i++ {
		bigI := big.NewInt(int64(count))
		traits := make(map[string]CollectionTrait)
		for j := 0; j < 3; j++ {
			traits[fmt.Sprintf("traittype-%d", j)] = CollectionTrait{
				Min: float64(i+j) / 2,
				Max: float64(i+j) * 2,
			}
		}

		newCollection := CollectionData{
			ID: ContractID{
				ChainID: w_common.ChainID(i),
				Address: common.BigToAddress(bigI),
			},
			ContractType: generateContractType(i),
			Provider:     fmt.Sprintf("provider-%d", i),
			Name:         fmt.Sprintf("name-%d", i),
			Slug:         fmt.Sprintf("slug-%d", i),
			ImageURL:     fmt.Sprintf("imageurl-%d", i),
			ImagePayload: []byte(fmt.Sprintf("imagepayload-%d", i)),
			Traits:       traits,
			CommunityID:  fmt.Sprintf("community-%d", i),
		}
		result = append(result, newCollection)
	}
	return result
}

func GenerateTestCommunityInfo(count int) map[string]CommunityInfo {
	base := rand.Intn(100) // nolint: gosec

	result := make(map[string]CommunityInfo)
	for i := base; i < count+base; i++ {
		communityID := fmt.Sprintf("communityid-%d", i)
		newCommunity := CommunityInfo{
			CommunityName:         fmt.Sprintf("communityname-%d", i),
			CommunityColor:        fmt.Sprintf("communitycolor-%d", i),
			CommunityImage:        fmt.Sprintf("communityimage-%d", i),
			CommunityImagePayload: []byte(fmt.Sprintf("communityimagepayload-%d", i)),
		}
		result[communityID] = newCommunity
	}

	return result
}

func GenerateTestFullCollectiblesData(count int) []FullCollectibleData {
	collectiblesData := GenerateTestCollectiblesData(count)
	collectionsData := GenerateTestCollectionsData(count)
	communityInfoMap := GenerateTestCommunityInfo(count)
	communityInfo := make([]CommunityInfo, 0, count)
	for _, info := range communityInfoMap {
		communityInfo = append(communityInfo, info)
	}
	communityData := GenerateTestCollectiblesCommunityData(count)

	ret := make([]FullCollectibleData, 0, count)
	for i := 0; i < count; i++ {
		// Ensure consistent ContracType
		collectionsData[i].ContractType = collectiblesData[i].ContractType

		ret = append(ret, FullCollectibleData{
			CollectibleData:          collectiblesData[i],
			CollectionData:           &collectionsData[i],
			CommunityInfo:            &communityInfo[i],
			CollectibleCommunityInfo: &communityData[i],
			Ownership:                GenerateTestCollectiblesOwnership(rand.Intn(5) + 1), // nolint: gosec
		})
	}
	return ret
}
