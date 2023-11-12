package collectibles

import (
	"github.com/status-im/status-go/protocol/communities/token"
	w_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

// Combined Collection+Collectible info, used to display a detailed view of a collectible
type Collectible struct {
	DataType        CollectibleDataType            `json:"data_type"`
	ID              thirdparty.CollectibleUniqueID `json:"id"`
	ContractType    w_common.ContractType          `json:"contract_type"`
	CollectibleData *CollectibleData               `json:"collectible_data,omitempty"`
	CollectionData  *CollectionData                `json:"collection_data,omitempty"`
	CommunityData   *CommunityData                 `json:"community_data,omitempty"`
	Ownership       []thirdparty.AccountBalance    `json:"ownership,omitempty"`
}

type CollectibleData struct {
	Name               string                         `json:"name"`
	Description        *string                        `json:"description,omitempty"`
	ImageURL           *string                        `json:"image_url,omitempty"`
	AnimationURL       *string                        `json:"animation_url,omitempty"`
	AnimationMediaType *string                        `json:"animation_media_type,omitempty"`
	Traits             *[]thirdparty.CollectibleTrait `json:"traits,omitempty"`
	BackgroundColor    *string                        `json:"background_color,omitempty"`
}

type CollectionData struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	ImageURL string `json:"image_url"`
}

type CommunityData struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Color           string                `json:"color"`
	PrivilegesLevel token.PrivilegesLevel `json:"privileges_level"`
	ImageURL        *string               `json:"image_url,omitempty"`
}

func idToCollectible(id thirdparty.CollectibleUniqueID) Collectible {
	ret := Collectible{
		DataType: CollectibleDataTypeUniqueID,
		ID:       id,
	}
	return ret
}

func idsToCollectibles(ids []thirdparty.CollectibleUniqueID) []Collectible {
	res := make([]Collectible, 0, len(ids))

	for _, id := range ids {
		c := idToCollectible(id)
		res = append(res, c)
	}

	return res
}

func getContractType(c thirdparty.FullCollectibleData) w_common.ContractType {
	if c.CollectibleData.ContractType != w_common.ContractTypeUnknown {
		return c.CollectibleData.ContractType
	}
	if c.CollectionData != nil && c.CollectionData.ContractType != w_common.ContractTypeUnknown {
		return c.CollectionData.ContractType
	}
	return w_common.ContractTypeUnknown
}

func fullCollectibleDataToHeader(c thirdparty.FullCollectibleData) Collectible {
	ret := Collectible{
		DataType:     CollectibleDataTypeHeader,
		ID:           c.CollectibleData.ID,
		ContractType: getContractType(c),
		CollectibleData: &CollectibleData{
			Name:               c.CollectibleData.Name,
			ImageURL:           &c.CollectibleData.ImageURL,
			AnimationURL:       &c.CollectibleData.AnimationURL,
			AnimationMediaType: &c.CollectibleData.AnimationMediaType,
			BackgroundColor:    &c.CollectibleData.BackgroundColor,
		},
	}
	if c.CollectionData != nil {
		ret.CollectionData = &CollectionData{
			Name:     c.CollectionData.Name,
			Slug:     c.CollectionData.Slug,
			ImageURL: c.CollectionData.ImageURL,
		}
	}
	if c.CollectibleData.CommunityID != "" {
		communityData := communityInfoToData(c.CollectibleData.CommunityID, c.CommunityInfo, c.CollectibleCommunityInfo)
		ret.CommunityData = &communityData
	}
	ret.Ownership = c.Ownership
	return ret
}

func fullCollectiblesDataToHeaders(data []thirdparty.FullCollectibleData) []Collectible {
	res := make([]Collectible, 0, len(data))

	for _, c := range data {
		header := fullCollectibleDataToHeader(c)
		res = append(res, header)
	}

	return res
}

func fullCollectibleDataToDetails(c thirdparty.FullCollectibleData) Collectible {
	ret := Collectible{
		DataType:     CollectibleDataTypeDetails,
		ID:           c.CollectibleData.ID,
		ContractType: getContractType(c),
		CollectibleData: &CollectibleData{
			Name:               c.CollectibleData.Name,
			Description:        &c.CollectibleData.Description,
			ImageURL:           &c.CollectibleData.ImageURL,
			AnimationURL:       &c.CollectibleData.AnimationURL,
			AnimationMediaType: &c.CollectibleData.AnimationMediaType,
			BackgroundColor:    &c.CollectibleData.BackgroundColor,
			Traits:             &c.CollectibleData.Traits,
		},
	}
	if c.CollectionData != nil {
		ret.CollectionData = &CollectionData{
			Name:     c.CollectionData.Name,
			Slug:     c.CollectionData.Slug,
			ImageURL: c.CollectionData.ImageURL,
		}
	}
	if c.CollectibleData.CommunityID != "" {
		communityData := communityInfoToData(c.CollectibleData.CommunityID, c.CommunityInfo, c.CollectibleCommunityInfo)
		ret.CommunityData = &communityData
	}
	ret.Ownership = c.Ownership
	return ret
}

func fullCollectiblesDataToDetails(data []thirdparty.FullCollectibleData) []Collectible {
	res := make([]Collectible, 0, len(data))

	for _, c := range data {
		details := fullCollectibleDataToDetails(c)
		res = append(res, details)
	}

	return res
}

func fullCollectiblesDataToCommunityHeader(data []thirdparty.FullCollectibleData) []Collectible {
	res := make([]Collectible, 0, len(data))

	for _, c := range data {
		collectibleID := c.CollectibleData.ID
		communityID := c.CollectibleData.CommunityID

		if communityID == "" {
			continue
		}

		communityData := communityInfoToData(communityID, c.CommunityInfo, c.CollectibleCommunityInfo)

		header := Collectible{
			ID:           collectibleID,
			ContractType: getContractType(c),
			CollectibleData: &CollectibleData{
				Name: c.CollectibleData.Name,
			},
			CommunityData: &communityData,
			Ownership:     c.Ownership,
		}

		res = append(res, header)
	}

	return res
}

func communityInfoToData(communityID string, community *thirdparty.CommunityInfo, communityCollectible *thirdparty.CollectibleCommunityInfo) CommunityData {
	ret := CommunityData{
		ID: communityID,
	}

	if community != nil {
		ret.Name = community.CommunityName
		ret.Color = community.CommunityColor
		ret.ImageURL = &community.CommunityImage
	}

	if communityCollectible != nil {
		ret.PrivilegesLevel = communityCollectible.PrivilegesLevel
	}

	return ret
}
