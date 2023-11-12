package thirdparty

// Community-related info used by the wallet, cached in the wallet db.
type CommunityInfo struct {
	CommunityName         string `json:"community_name"`
	CommunityColor        string `json:"community_color"`
	CommunityImage        string `json:"community_image"`
	CommunityImagePayload []byte
}

type CommunityInfoProvider interface {
	FetchCommunityInfo(communityID string) (*CommunityInfo, error)

	// Collectible-related methods
	GetCommunityID(tokenURI string) string
	FillCollectibleMetadata(collectible *FullCollectibleData) error
}
