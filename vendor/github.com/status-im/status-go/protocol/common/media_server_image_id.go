package common

type MediaServerImageIDPrefix string
type MediaServerImageIDPostfix string
type MediaServerImageID string

func CreateImageID(prefix MediaServerImageIDPrefix, postfix MediaServerImageIDPostfix) MediaServerImageID {
	return MediaServerImageID(string(prefix) + string(postfix))
}

const (
	MediaServerIconPostfix   MediaServerImageIDPostfix = "icon"
	MediaServerBannerPostfix MediaServerImageIDPostfix = "banner"
)

const (
	MediaServerContactPrefix          MediaServerImageIDPrefix = "contact-"
	MediaServerCommunityPrefix        MediaServerImageIDPrefix = "community-"
	MediaServerChannelCommunityPrefix MediaServerImageIDPrefix = "community-channel-"
)

const (
	MediaServerContactIcon            = MediaServerImageID(string(MediaServerContactPrefix) + string(MediaServerIconPostfix))
	MediaServerCommunityIcon          = MediaServerImageID(string(MediaServerCommunityPrefix) + string(MediaServerIconPostfix))
	MediaServerCommunityBanner        = MediaServerImageID(string(MediaServerCommunityPrefix) + string(MediaServerBannerPostfix))
	MediaServerChannelCommunityIcon   = MediaServerImageID(string(MediaServerChannelCommunityPrefix) + string(MediaServerIconPostfix))
	MediaServerChannelCommunityBanner = MediaServerImageID(string(MediaServerChannelCommunityPrefix) + string(MediaServerBannerPostfix))
)
