package common

import (
	"fmt"
	"net/url"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/protobuf"
)

type MakeMediaServerURLType func(msgID string, previewURL string, imageID MediaServerImageID) string
type MakeMediaServerURLMessageWrapperType func(previewURL string, imageID MediaServerImageID) string

type LinkPreviewThumbnail struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
	// Non-empty when the thumbnail is available via the media server, i.e. after
	// the chat message is sent.
	URL string `json:"url,omitempty"`
	// Non-empty when the thumbnail payload needs to be shared with the client,
	// but before it has been persisted.
	DataURI string `json:"dataUri,omitempty"`
}

type LinkPreview struct {
	Type        protobuf.UnfurledLink_LinkType `json:"type"`
	URL         string                         `json:"url"`
	Hostname    string                         `json:"hostname"`
	Title       string                         `json:"title,omitempty"`
	Description string                         `json:"description,omitempty"`
	Favicon     LinkPreviewThumbnail           `json:"favicon,omitempty"`
	Thumbnail   LinkPreviewThumbnail           `json:"thumbnail,omitempty"`
}

type StatusContactLinkPreview struct {
	// PublicKey is: "0x" + hex-encoded decompressed public key.
	// We keep it a string here for correct json marshalling.
	PublicKey   string               `json:"publicKey"`
	DisplayName string               `json:"displayName"`
	Description string               `json:"description"`
	Icon        LinkPreviewThumbnail `json:"icon,omitempty"`
}

type StatusCommunityLinkPreview struct {
	CommunityID  string               `json:"communityId"`
	DisplayName  string               `json:"displayName"`
	Description  string               `json:"description"`
	MembersCount uint32               `json:"membersCount"`
	Color        string               `json:"color"`
	Icon         LinkPreviewThumbnail `json:"icon,omitempty"`
	Banner       LinkPreviewThumbnail `json:"banner,omitempty"`
}

type StatusCommunityChannelLinkPreview struct {
	ChannelUUID string                      `json:"channelUuid"`
	Emoji       string                      `json:"emoji"`
	DisplayName string                      `json:"displayName"`
	Description string                      `json:"description"`
	Color       string                      `json:"color"`
	Community   *StatusCommunityLinkPreview `json:"community"`
}

type StatusLinkPreview struct {
	URL       string                             `json:"url,omitempty"`
	Contact   *StatusContactLinkPreview          `json:"contact,omitempty"`
	Community *StatusCommunityLinkPreview        `json:"community,omitempty"`
	Channel   *StatusCommunityChannelLinkPreview `json:"channel,omitempty"`
}

func (thumbnail *LinkPreviewThumbnail) IsEmpty() bool {
	return thumbnail.Width == 0 &&
		thumbnail.Height == 0 &&
		thumbnail.URL == "" &&
		thumbnail.DataURI == ""
}

func (thumbnail *LinkPreviewThumbnail) clear() {
	thumbnail.Width = 0
	thumbnail.Height = 0
	thumbnail.URL = ""
	thumbnail.DataURI = ""
}

func (thumbnail *LinkPreviewThumbnail) validateForProto() error {
	if thumbnail.DataURI == "" {
		if thumbnail.Width == 0 && thumbnail.Height == 0 {
			return nil
		}
		return fmt.Errorf("dataUri is empty, but width/height are not zero")
	}

	if thumbnail.Width == 0 || thumbnail.Height == 0 {
		return fmt.Errorf("dataUri is not empty, but width/heigth are zero")
	}

	return nil
}

func (thumbnail *LinkPreviewThumbnail) convertToProto() (*protobuf.UnfurledLinkThumbnail, error) {
	var payload []byte
	var err error
	if thumbnail.DataURI != "" {
		payload, err = images.GetPayloadFromURI(thumbnail.DataURI)
		if err != nil {
			return nil, fmt.Errorf("could not get data URI payload, url='%s': %w", thumbnail.URL, err)
		}
	}

	return &protobuf.UnfurledLinkThumbnail{
		Width:   uint32(thumbnail.Width),
		Height:  uint32(thumbnail.Height),
		Payload: payload,
	}, nil
}

func (thumbnail *LinkPreviewThumbnail) loadFromProto(
	input *protobuf.UnfurledLinkThumbnail,
	URL string,
	imageID MediaServerImageID,
	makeMediaServerURL MakeMediaServerURLMessageWrapperType) {

	thumbnail.clear()
	thumbnail.Width = int(input.Width)
	thumbnail.Height = int(input.Height)

	if len(input.Payload) > 0 {
		thumbnail.URL = makeMediaServerURL(URL, imageID)
	}
}

func (preview *LinkPreview) validateForProto() error {
	switch preview.Type {
	case protobuf.UnfurledLink_IMAGE:
		if preview.URL == "" {
			return fmt.Errorf("empty url")
		}
		if err := preview.Thumbnail.validateForProto(); err != nil {
			return fmt.Errorf("thumbnail is not valid for proto: %w", err)
		}
		return nil
	default: // Validate as a link type by default.
		if preview.Title == "" {
			return fmt.Errorf("title is empty")
		}
		if preview.URL == "" {
			return fmt.Errorf("url is empty")
		}
		if err := preview.Thumbnail.validateForProto(); err != nil {
			return fmt.Errorf("thumbnail is not valid for proto: %w", err)
		}
		return nil
	}
}

func (preview *StatusLinkPreview) validateForProto() error {
	if preview.URL == "" {
		return fmt.Errorf("url can't be empty")
	}

	// At least and only one of Contact/Community/Channel should be present in the preview
	if preview.Contact != nil && preview.Community != nil {
		return fmt.Errorf("both contact and community are set at the same time")
	}
	if preview.Community != nil && preview.Channel != nil {
		return fmt.Errorf("both community and channel are set at the same time")
	}
	if preview.Channel != nil && preview.Contact != nil {
		return fmt.Errorf("both contact and channel are set at the same time")
	}
	if preview.Contact == nil && preview.Community == nil && preview.Channel == nil {
		return fmt.Errorf("none of contact/community/channel are set")
	}

	if preview.Contact != nil {
		if preview.Contact.PublicKey == "" {
			return fmt.Errorf("contact publicKey is empty")
		}
		if err := preview.Contact.Icon.validateForProto(); err != nil {
			return fmt.Errorf("contact icon invalid: %w", err)
		}
		return nil
	}

	if preview.Community != nil {
		return preview.Community.validateForProto()
	}

	if preview.Channel != nil {
		if preview.Channel.ChannelUUID == "" {
			return fmt.Errorf("channelUuid is empty")
		}
		if preview.Channel.Community == nil {
			return fmt.Errorf("channel community is nil")
		}
		if err := preview.Channel.Community.validateForProto(); err != nil {
			return fmt.Errorf("channel community is not valid: %w", err)
		}
		return nil
	}
	return nil
}

func (preview *StatusCommunityLinkPreview) validateForProto() error {
	if preview == nil {
		return fmt.Errorf("community preview is empty")
	}
	if preview.CommunityID == "" {
		return fmt.Errorf("communityId is empty")
	}
	if err := preview.Icon.validateForProto(); err != nil {
		return fmt.Errorf("community icon is invalid: %w", err)
	}
	if err := preview.Banner.validateForProto(); err != nil {
		return fmt.Errorf("community banner is invalid: %w", err)
	}
	return nil
}

func (preview *StatusCommunityLinkPreview) convertToProto() (*protobuf.UnfurledStatusCommunityLink, error) {
	if preview == nil {
		return nil, nil
	}

	icon, err := preview.Icon.convertToProto()
	if err != nil {
		return nil, err
	}

	banner, err := preview.Banner.convertToProto()
	if err != nil {
		return nil, err
	}

	communityID, err := types.DecodeHex(preview.CommunityID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode community id: %w", err)
	}

	community := &protobuf.UnfurledStatusCommunityLink{
		CommunityId:  communityID,
		DisplayName:  preview.DisplayName,
		Description:  preview.Description,
		MembersCount: preview.MembersCount,
		Color:        preview.Color,
		Icon:         icon,
		Banner:       banner,
	}

	return community, nil
}

func (preview *StatusCommunityLinkPreview) loadFromProto(c *protobuf.UnfurledStatusCommunityLink,
	URL string, thumbnailPrefix MediaServerImageIDPrefix,
	makeMediaServerURL MakeMediaServerURLMessageWrapperType) {

	preview.CommunityID = types.EncodeHex(c.CommunityId)
	preview.DisplayName = c.DisplayName
	preview.Description = c.Description
	preview.MembersCount = c.MembersCount
	preview.Color = c.Color
	preview.Icon.clear()
	preview.Banner.clear()

	if icon := c.GetIcon(); icon != nil {
		preview.Icon.loadFromProto(icon, URL, CreateImageID(thumbnailPrefix, MediaServerIconPostfix), makeMediaServerURL)
	}
	if banner := c.GetBanner(); banner != nil {
		preview.Banner.loadFromProto(banner, URL, CreateImageID(thumbnailPrefix, MediaServerBannerPostfix), makeMediaServerURL)
	}
}

// ConvertLinkPreviewsToProto expects previews to be correctly sent by the
// client because we can't attempt to re-unfurl URLs at this point (it's
// actually undesirable). We run a basic validation as an additional safety net.
func (m *Message) ConvertLinkPreviewsToProto() ([]*protobuf.UnfurledLink, error) {
	if len(m.LinkPreviews) == 0 {
		return nil, nil
	}

	unfurledLinks := make([]*protobuf.UnfurledLink, 0, len(m.LinkPreviews))

	for _, preview := range m.LinkPreviews {
		// Do not process subsequent previews because we do expect all previews to
		// be valid at this stage.
		if err := preview.validateForProto(); err != nil {
			return nil, fmt.Errorf("invalid link preview, url='%s': %w", preview.URL, err)
		}

		var thumbnailPayload []byte
		var faviconPayload []byte
		var err error
		if preview.Thumbnail.DataURI != "" {
			thumbnailPayload, err = images.GetPayloadFromURI(preview.Thumbnail.DataURI)
			if err != nil {
				return nil, fmt.Errorf("could not get data URI payload for link preview thumbnail, url='%s': %w", preview.URL, err)
			}
		}
		if preview.Favicon.DataURI != "" {
			faviconPayload, err = images.GetPayloadFromURI(preview.Favicon.DataURI)
			if err != nil {
				return nil, fmt.Errorf("could not get data URI payload for link preview favicon, url='%s': %w", preview.URL, err)
			}
		}

		ul := &protobuf.UnfurledLink{
			Type:             preview.Type,
			Url:              preview.URL,
			Title:            preview.Title,
			Description:      preview.Description,
			ThumbnailWidth:   uint32(preview.Thumbnail.Width),
			ThumbnailHeight:  uint32(preview.Thumbnail.Height),
			ThumbnailPayload: thumbnailPayload,
			FaviconPayload:   faviconPayload,
		}
		unfurledLinks = append(unfurledLinks, ul)
	}

	return unfurledLinks, nil
}

func (m *Message) ConvertFromProtoToLinkPreviews(makeThumbnailMediaServerURL func(msgID string, previewURL string) string,
	makeFaviconMediaServerURL func(msgID string, previewURL string) string) []LinkPreview {
	var links []*protobuf.UnfurledLink

	if links = m.GetUnfurledLinks(); links == nil {
		return nil
	}

	previews := make([]LinkPreview, 0, len(links))
	for _, link := range links {
		parsedURL, err := url.Parse(link.Url)
		var hostname string
		// URL parsing in Go can fail with URLs that weren't correctly URL encoded.
		// This shouldn't happen in general, but if an error happens we just reuse
		// the full URL.
		if err != nil {
			hostname = link.Url
		} else {
			hostname = parsedURL.Hostname()
		}
		lp := LinkPreview{
			Description: link.Description,
			Hostname:    hostname,
			Title:       link.Title,
			Type:        link.Type,
			URL:         link.Url,
		}
		mediaURL := ""
		if len(link.ThumbnailPayload) > 0 {
			mediaURL = makeThumbnailMediaServerURL(m.ID, link.Url)
		}
		if link.GetThumbnailPayload() != nil {
			lp.Thumbnail.Width = int(link.ThumbnailWidth)
			lp.Thumbnail.Height = int(link.ThumbnailHeight)
			lp.Thumbnail.URL = mediaURL
		}
		faviconMediaURL := ""
		if len(link.FaviconPayload) > 0 {
			faviconMediaURL = makeFaviconMediaServerURL(m.ID, link.Url)
		}
		if link.GetFaviconPayload() != nil {
			lp.Favicon.URL = faviconMediaURL
		}
		previews = append(previews, lp)
	}

	return previews
}

func (m *Message) ConvertStatusLinkPreviewsToProto() (*protobuf.UnfurledStatusLinks, error) {
	if len(m.StatusLinkPreviews) == 0 {
		return nil, nil
	}

	unfurledLinks := make([]*protobuf.UnfurledStatusLink, 0, len(m.StatusLinkPreviews))

	for _, preview := range m.StatusLinkPreviews {
		// We expect all previews to be valid at this stage
		if err := preview.validateForProto(); err != nil {
			return nil, fmt.Errorf("invalid status link preview, url='%s': %w", preview.URL, err)
		}

		ul := &protobuf.UnfurledStatusLink{
			Url: preview.URL,
		}

		if preview.Contact != nil {
			decompressedPublicKey, err := types.DecodeHex(preview.Contact.PublicKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decode contact public key: %w", err)
			}

			publicKey, err := crypto.UnmarshalPubkey(decompressedPublicKey)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal decompressed public key: %w", err)
			}

			compressedPublicKey := crypto.CompressPubkey(publicKey)

			icon, err := preview.Contact.Icon.convertToProto()
			if err != nil {
				return nil, err
			}

			ul.Payload = &protobuf.UnfurledStatusLink_Contact{
				Contact: &protobuf.UnfurledStatusContactLink{
					PublicKey:   compressedPublicKey,
					DisplayName: preview.Contact.DisplayName,
					Description: preview.Contact.Description,
					Icon:        icon,
				},
			}
		}

		if preview.Community != nil {
			communityPreview, err := preview.Community.convertToProto()
			if err != nil {
				return nil, err
			}
			ul.Payload = &protobuf.UnfurledStatusLink_Community{
				Community: communityPreview,
			}
		}

		if preview.Channel != nil {
			communityPreview, err := preview.Channel.Community.convertToProto()
			if err != nil {
				return nil, err
			}

			ul.Payload = &protobuf.UnfurledStatusLink_Channel{
				Channel: &protobuf.UnfurledStatusChannelLink{
					ChannelUuid: preview.Channel.ChannelUUID,
					Emoji:       preview.Channel.Emoji,
					DisplayName: preview.Channel.DisplayName,
					Description: preview.Channel.Description,
					Color:       preview.Channel.Color,
					Community:   communityPreview,
				},
			}

		}

		unfurledLinks = append(unfurledLinks, ul)
	}

	return &protobuf.UnfurledStatusLinks{UnfurledStatusLinks: unfurledLinks}, nil
}

func (m *Message) ConvertFromProtoToStatusLinkPreviews(makeMediaServerURL func(msgID string, previewURL string, imageID MediaServerImageID) string) []StatusLinkPreview {
	if m.GetUnfurledStatusLinks() == nil {
		return nil
	}

	links := m.UnfurledStatusLinks.GetUnfurledStatusLinks()

	if links == nil {
		return nil
	}

	// This wrapper adds the messageID to the callback
	makeMediaServerURLMessageWrapper := func(previewURL string, imageID MediaServerImageID) string {
		return makeMediaServerURL(m.ID, previewURL, imageID)
	}

	previews := make([]StatusLinkPreview, 0, len(links))

	for _, link := range links {
		lp := StatusLinkPreview{
			URL: link.Url,
		}

		if c := link.GetContact(); c != nil {
			publicKey, err := crypto.DecompressPubkey(c.PublicKey)
			if err != nil {
				continue
			}

			lp.Contact = &StatusContactLinkPreview{
				PublicKey:   types.EncodeHex(gethcrypto.FromECDSAPub(publicKey)),
				DisplayName: c.DisplayName,
				Description: c.Description,
			}
			if icon := c.GetIcon(); icon != nil {
				lp.Contact.Icon.loadFromProto(icon, link.Url, MediaServerContactIcon, makeMediaServerURLMessageWrapper)
			}
		}

		if c := link.GetCommunity(); c != nil {
			lp.Community = new(StatusCommunityLinkPreview)
			lp.Community.loadFromProto(c, link.Url, MediaServerCommunityPrefix, makeMediaServerURLMessageWrapper)
		}

		if c := link.GetChannel(); c != nil {
			lp.Channel = &StatusCommunityChannelLinkPreview{
				ChannelUUID: c.ChannelUuid,
				Emoji:       c.Emoji,
				DisplayName: c.DisplayName,
				Description: c.Description,
				Color:       c.Color,
			}
			if c.Community != nil {
				lp.Channel.Community = new(StatusCommunityLinkPreview)
				lp.Channel.Community.loadFromProto(c.Community, link.Url, MediaServerChannelCommunityPrefix, makeMediaServerURLMessageWrapper)
			}
		}

		previews = append(previews, lp)
	}

	return previews
}
