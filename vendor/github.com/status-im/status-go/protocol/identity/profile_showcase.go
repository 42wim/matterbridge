package identity

import "errors"

var ErrorNoAccountProvidedWithTokenOrCollectible = errors.New("no account provided with tokens or collectible")

var ErrorExceedMaxProfileShowcaseCommunitiesLimit = errors.New("exeed maximum profile showcase communities limit")
var ErrorExceedMaxProfileShowcaseAccountsLimit = errors.New("exeed maximum profile showcase accounts limit")
var ErrorExceedMaxProfileShowcaseCollectiblesLimit = errors.New("exeed maximum profile showcase collectibles limit")
var ErrorExceedMaxProfileShowcaseVerifiedTokensLimit = errors.New("exeed maximum profile showcase verified tokens limit")
var ErrorExceedMaxProfileShowcaseUnverifiedTokensLimit = errors.New("exeed maximum profile showcase unverified tokens limit")
var ErrorExceedMaxProfileShowcaseSocialLinksLimit = errors.New("exeed maximum profile showcase communities limit")

const MaxProfileShowcaseSocialLinksLimit = 20
const MaxProfileShowcaseEntriesLimit = 100

type ProfileShowcaseVisibility int

const (
	ProfileShowcaseVisibilityNoOne ProfileShowcaseVisibility = iota
	ProfileShowcaseVisibilityIDVerifiedContacts
	ProfileShowcaseVisibilityContacts
	ProfileShowcaseVisibilityEveryone
)

type ProfileShowcaseMembershipStatus int

const (
	ProfileShowcaseMembershipStatusUnproven ProfileShowcaseMembershipStatus = iota
	ProfileShowcaseMembershipStatusProvenMember
	ProfileShowcaseMembershipStatusNotAMember
)

// Profile showcase preferences

type ProfileShowcaseCommunityPreference struct {
	CommunityID        string                    `json:"communityId"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseAccountPreference struct {
	Address            string                    `json:"address"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseCollectiblePreference struct {
	ContractAddress    string                    `json:"contractAddress"`
	ChainID            uint64                    `json:"chainId"`
	TokenID            string                    `json:"tokenId"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseVerifiedTokenPreference struct {
	Symbol             string                    `json:"symbol"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseUnverifiedTokenPreference struct {
	ContractAddress    string                    `json:"contractAddress"`
	ChainID            uint64                    `json:"chainId"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseSocialLinkPreference struct {
	URL                string                    `json:"url"`
	Text               string                    `json:"text"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcasePreferences struct {
	Clock            uint64                                      `json:"clock"`
	Communities      []*ProfileShowcaseCommunityPreference       `json:"communities"`
	Accounts         []*ProfileShowcaseAccountPreference         `json:"accounts"`
	Collectibles     []*ProfileShowcaseCollectiblePreference     `json:"collectibles"`
	VerifiedTokens   []*ProfileShowcaseVerifiedTokenPreference   `json:"verifiedTokens"`
	UnverifiedTokens []*ProfileShowcaseUnverifiedTokenPreference `json:"unverifiedTokens"`
	SocialLinks      []*ProfileShowcaseSocialLinkPreference      `json:"socialLinks"`
}

// Profile showcase for a contact

type ProfileShowcaseCommunity struct {
	CommunityID      string                          `json:"communityId"`
	Order            int                             `json:"order"`
	MembershipStatus ProfileShowcaseMembershipStatus `json:"membershipStatus"`
	Grant            []byte                          `json:"grant,omitempty"`
}

type ProfileShowcaseAccount struct {
	ContactID string `json:"contactId"`
	Address   string `json:"address"`
	Name      string `json:"name"`
	ColorID   string `json:"colorId"`
	Emoji     string `json:"emoji"`
	Order     int    `json:"order"`
}

type ProfileShowcaseCollectible struct {
	ContractAddress string `json:"contractAddress"`
	ChainID         uint64 `json:"chainId"`
	TokenID         string `json:"tokenId"`
	Order           int    `json:"order"`
}

type ProfileShowcaseVerifiedToken struct {
	Symbol string `json:"symbol"`
	Order  int    `json:"order"`
}

type ProfileShowcaseUnverifiedToken struct {
	ContractAddress string `json:"contractAddress"`
	ChainID         uint64 `json:"chainId"`
	Order           int    `json:"order"`
}

type ProfileShowcaseSocialLink struct {
	URL   string `json:"url"`
	Text  string `json:"text"`
	Order int    `json:"order"`
}

type ProfileShowcase struct {
	ContactID        string                            `json:"contactId"`
	Communities      []*ProfileShowcaseCommunity       `json:"communities"`
	Accounts         []*ProfileShowcaseAccount         `json:"accounts"`
	Collectibles     []*ProfileShowcaseCollectible     `json:"collectibles"`
	VerifiedTokens   []*ProfileShowcaseVerifiedToken   `json:"verifiedTokens"`
	UnverifiedTokens []*ProfileShowcaseUnverifiedToken `json:"unverifiedTokens"`
	SocialLinks      []*ProfileShowcaseSocialLink      `json:"socialLinks"`
}

func Validate(preferences *ProfileShowcasePreferences) error {
	if len(preferences.Communities) > MaxProfileShowcaseEntriesLimit {
		return ErrorExceedMaxProfileShowcaseCommunitiesLimit
	}
	if len(preferences.Accounts) > MaxProfileShowcaseEntriesLimit {
		return ErrorExceedMaxProfileShowcaseAccountsLimit
	}
	if len(preferences.Collectibles) > MaxProfileShowcaseEntriesLimit {
		return ErrorExceedMaxProfileShowcaseCollectiblesLimit
	}
	if len(preferences.VerifiedTokens) > MaxProfileShowcaseEntriesLimit {
		return ErrorExceedMaxProfileShowcaseVerifiedTokensLimit
	}
	if len(preferences.UnverifiedTokens) > MaxProfileShowcaseEntriesLimit {
		return ErrorExceedMaxProfileShowcaseUnverifiedTokensLimit
	}
	if len(preferences.SocialLinks) > MaxProfileShowcaseSocialLinksLimit {
		return ErrorExceedMaxProfileShowcaseSocialLinksLimit
	}

	if (len(preferences.VerifiedTokens) > 0 || len(preferences.UnverifiedTokens) > 0 || len(preferences.Collectibles) > 0) &&
		len(preferences.Accounts) == 0 {
		return ErrorNoAccountProvidedWithTokenOrCollectible
	}

	return nil
}
