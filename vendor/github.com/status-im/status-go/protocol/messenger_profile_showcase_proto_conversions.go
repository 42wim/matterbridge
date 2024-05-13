package protocol

import (
	"github.com/status-im/status-go/protocol/identity"
	"github.com/status-im/status-go/protocol/protobuf"
)

func FromProfileShowcaseCommunityPreferenceProto(p *protobuf.ProfileShowcaseCommunityPreference) *identity.ProfileShowcaseCommunityPreference {
	return &identity.ProfileShowcaseCommunityPreference{
		CommunityID:        p.GetCommunityId(),
		ShowcaseVisibility: identity.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              int(p.Order),
	}
}

func FromProfileShowcaseCommunitiesPreferencesProto(preferences []*protobuf.ProfileShowcaseCommunityPreference) []*identity.ProfileShowcaseCommunityPreference {
	out := make([]*identity.ProfileShowcaseCommunityPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, FromProfileShowcaseCommunityPreferenceProto(p))
	}
	return out
}

func ToProfileShowcaseCommunityPreferenceProto(p *identity.ProfileShowcaseCommunityPreference) *protobuf.ProfileShowcaseCommunityPreference {
	return &protobuf.ProfileShowcaseCommunityPreference{
		CommunityId:        p.CommunityID,
		ShowcaseVisibility: protobuf.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              uint32(p.Order),
	}
}

func ToProfileShowcaseCommunitiesPreferencesProto(preferences []*identity.ProfileShowcaseCommunityPreference) []*protobuf.ProfileShowcaseCommunityPreference {
	out := make([]*protobuf.ProfileShowcaseCommunityPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, ToProfileShowcaseCommunityPreferenceProto(p))
	}
	return out
}

func FromProfileShowcaseAccountPreferenceProto(p *protobuf.ProfileShowcaseAccountPreference) *identity.ProfileShowcaseAccountPreference {
	return &identity.ProfileShowcaseAccountPreference{
		Address:            p.GetAddress(),
		ShowcaseVisibility: identity.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              int(p.Order),
	}
}

func FromProfileShowcaseAccountsPreferencesProto(preferences []*protobuf.ProfileShowcaseAccountPreference) []*identity.ProfileShowcaseAccountPreference {
	out := make([]*identity.ProfileShowcaseAccountPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, FromProfileShowcaseAccountPreferenceProto(p))
	}
	return out
}

func ToProfileShowcaseAccountPreferenceProto(p *identity.ProfileShowcaseAccountPreference) *protobuf.ProfileShowcaseAccountPreference {
	return &protobuf.ProfileShowcaseAccountPreference{
		Address:            p.Address,
		ShowcaseVisibility: protobuf.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              uint32(p.Order),
	}
}

func ToProfileShowcaseAccountsPreferenceProto(preferences []*identity.ProfileShowcaseAccountPreference) []*protobuf.ProfileShowcaseAccountPreference {
	out := make([]*protobuf.ProfileShowcaseAccountPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, ToProfileShowcaseAccountPreferenceProto(p))
	}
	return out
}

func FromProfileShowcaseCollectiblePreferenceProto(p *protobuf.ProfileShowcaseCollectiblePreference) *identity.ProfileShowcaseCollectiblePreference {
	return &identity.ProfileShowcaseCollectiblePreference{
		ContractAddress:    p.GetContractAddress(),
		ChainID:            p.GetChainId(),
		TokenID:            p.GetTokenId(),
		ShowcaseVisibility: identity.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              int(p.Order),
	}
}

func FromProfileShowcaseCollectiblesPreferencesProto(preferences []*protobuf.ProfileShowcaseCollectiblePreference) []*identity.ProfileShowcaseCollectiblePreference {
	out := make([]*identity.ProfileShowcaseCollectiblePreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, FromProfileShowcaseCollectiblePreferenceProto(p))
	}
	return out
}

func ToProfileShowcaseCollectiblePreferenceProto(p *identity.ProfileShowcaseCollectiblePreference) *protobuf.ProfileShowcaseCollectiblePreference {
	return &protobuf.ProfileShowcaseCollectiblePreference{
		ContractAddress:    p.ContractAddress,
		ChainId:            p.ChainID,
		TokenId:            p.TokenID,
		ShowcaseVisibility: protobuf.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              uint32(p.Order),
	}
}

func ToProfileShowcaseCollectiblesPreferenceProto(preferences []*identity.ProfileShowcaseCollectiblePreference) []*protobuf.ProfileShowcaseCollectiblePreference {
	out := make([]*protobuf.ProfileShowcaseCollectiblePreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, ToProfileShowcaseCollectiblePreferenceProto(p))
	}
	return out
}

func FromProfileShowcaseVerifiedTokenPreferenceProto(p *protobuf.ProfileShowcaseVerifiedTokenPreference) *identity.ProfileShowcaseVerifiedTokenPreference {
	return &identity.ProfileShowcaseVerifiedTokenPreference{
		Symbol:             p.GetSymbol(),
		ShowcaseVisibility: identity.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              int(p.Order),
	}
}

func FromProfileShowcaseVerifiedTokensPreferencesProto(preferences []*protobuf.ProfileShowcaseVerifiedTokenPreference) []*identity.ProfileShowcaseVerifiedTokenPreference {
	out := make([]*identity.ProfileShowcaseVerifiedTokenPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, FromProfileShowcaseVerifiedTokenPreferenceProto(p))
	}
	return out
}

func ToProfileShowcaseVerifiedTokenPreferenceProto(p *identity.ProfileShowcaseVerifiedTokenPreference) *protobuf.ProfileShowcaseVerifiedTokenPreference {
	return &protobuf.ProfileShowcaseVerifiedTokenPreference{
		Symbol:             p.Symbol,
		ShowcaseVisibility: protobuf.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              uint32(p.Order),
	}

}

func ToProfileShowcaseVerifiedTokensPreferenceProto(preferences []*identity.ProfileShowcaseVerifiedTokenPreference) []*protobuf.ProfileShowcaseVerifiedTokenPreference {
	out := make([]*protobuf.ProfileShowcaseVerifiedTokenPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, ToProfileShowcaseVerifiedTokenPreferenceProto(p))
	}
	return out
}

func FromProfileShowcaseUnverifiedTokenPreferenceProto(p *protobuf.ProfileShowcaseUnverifiedTokenPreference) *identity.ProfileShowcaseUnverifiedTokenPreference {
	return &identity.ProfileShowcaseUnverifiedTokenPreference{
		ContractAddress:    p.GetContractAddress(),
		ChainID:            p.GetChainId(),
		ShowcaseVisibility: identity.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              int(p.Order),
	}
}

func FromProfileShowcaseUnverifiedTokensPreferencesProto(preferences []*protobuf.ProfileShowcaseUnverifiedTokenPreference) []*identity.ProfileShowcaseUnverifiedTokenPreference {
	out := make([]*identity.ProfileShowcaseUnverifiedTokenPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, FromProfileShowcaseUnverifiedTokenPreferenceProto(p))
	}
	return out
}

func ToProfileShowcaseUnverifiedTokenPreferenceProto(p *identity.ProfileShowcaseUnverifiedTokenPreference) *protobuf.ProfileShowcaseUnverifiedTokenPreference {
	return &protobuf.ProfileShowcaseUnverifiedTokenPreference{
		ContractAddress:    p.ContractAddress,
		ChainId:            p.ChainID,
		ShowcaseVisibility: protobuf.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              uint32(p.Order),
	}
}

func ToProfileShowcaseUnverifiedTokensPreferenceProto(preferences []*identity.ProfileShowcaseUnverifiedTokenPreference) []*protobuf.ProfileShowcaseUnverifiedTokenPreference {
	out := make([]*protobuf.ProfileShowcaseUnverifiedTokenPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, ToProfileShowcaseUnverifiedTokenPreferenceProto(p))
	}
	return out
}

func FromProfileShowcaseSocialLinkPreferenceProto(p *protobuf.ProfileShowcaseSocialLinkPreference) *identity.ProfileShowcaseSocialLinkPreference {
	return &identity.ProfileShowcaseSocialLinkPreference{
		Text:               p.GetText(),
		URL:                p.GetUrl(),
		ShowcaseVisibility: identity.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              int(p.Order),
	}
}

func FromProfileShowcaseSocialLinksPreferencesProto(preferences []*protobuf.ProfileShowcaseSocialLinkPreference) []*identity.ProfileShowcaseSocialLinkPreference {
	out := make([]*identity.ProfileShowcaseSocialLinkPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, FromProfileShowcaseSocialLinkPreferenceProto(p))
	}
	return out
}

func ToProfileShowcaseSocialLinkPreferenceProto(p *identity.ProfileShowcaseSocialLinkPreference) *protobuf.ProfileShowcaseSocialLinkPreference {
	return &protobuf.ProfileShowcaseSocialLinkPreference{
		Text:               p.Text,
		Url:                p.URL,
		ShowcaseVisibility: protobuf.ProfileShowcaseVisibility(p.ShowcaseVisibility),
		Order:              uint32(p.Order),
	}
}

func ToProfileShowcaseSocialLinksPreferenceProto(preferences []*identity.ProfileShowcaseSocialLinkPreference) []*protobuf.ProfileShowcaseSocialLinkPreference {
	out := make([]*protobuf.ProfileShowcaseSocialLinkPreference, 0, len(preferences))
	for _, p := range preferences {
		out = append(out, ToProfileShowcaseSocialLinkPreferenceProto(p))
	}
	return out
}

func FromProfileShowcasePreferencesProto(p *protobuf.SyncProfileShowcasePreferences) *identity.ProfileShowcasePreferences {
	return &identity.ProfileShowcasePreferences{
		Clock:            p.GetClock(),
		Communities:      FromProfileShowcaseCommunitiesPreferencesProto(p.Communities),
		Accounts:         FromProfileShowcaseAccountsPreferencesProto(p.Accounts),
		Collectibles:     FromProfileShowcaseCollectiblesPreferencesProto(p.Collectibles),
		VerifiedTokens:   FromProfileShowcaseVerifiedTokensPreferencesProto(p.VerifiedTokens),
		UnverifiedTokens: FromProfileShowcaseUnverifiedTokensPreferencesProto(p.UnverifiedTokens),
		SocialLinks:      FromProfileShowcaseSocialLinksPreferencesProto(p.SocialLinks),
	}
}

func ToProfileShowcasePreferencesProto(p *identity.ProfileShowcasePreferences) *protobuf.SyncProfileShowcasePreferences {
	return &protobuf.SyncProfileShowcasePreferences{
		Clock:            p.Clock,
		Communities:      ToProfileShowcaseCommunitiesPreferencesProto(p.Communities),
		Accounts:         ToProfileShowcaseAccountsPreferenceProto(p.Accounts),
		Collectibles:     ToProfileShowcaseCollectiblesPreferenceProto(p.Collectibles),
		VerifiedTokens:   ToProfileShowcaseVerifiedTokensPreferenceProto(p.VerifiedTokens),
		UnverifiedTokens: ToProfileShowcaseUnverifiedTokensPreferenceProto(p.UnverifiedTokens),
		SocialLinks:      ToProfileShowcaseSocialLinksPreferenceProto(p.SocialLinks),
	}
}
