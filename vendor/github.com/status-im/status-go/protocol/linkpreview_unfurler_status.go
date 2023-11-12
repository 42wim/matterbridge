package protocol

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/status-im/status-go/api/multiformat"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/protocol/communities"
)

type StatusUnfurler struct {
	m      *Messenger
	logger *zap.Logger
	url    string
}

func NewStatusUnfurler(URL string, messenger *Messenger, logger *zap.Logger) *StatusUnfurler {
	return &StatusUnfurler{
		m:      messenger,
		logger: logger.With(zap.String("url", URL)),
		url:    URL,
	}
}

func updateThumbnail(image *images.IdentityImage, thumbnail *common.LinkPreviewThumbnail) error {
	if image.IsEmpty() {
		return nil
	}

	width, height, err := images.GetImageDimensions(image.Payload)
	if err != nil {
		return fmt.Errorf("failed to get image dimensions: %w", err)
	}

	dataURI, err := image.GetDataURI()
	if err != nil {
		return fmt.Errorf("failed to get data uri: %w", err)
	}

	thumbnail.Width = width
	thumbnail.Height = height
	thumbnail.DataURI = dataURI

	return nil
}

func (u *StatusUnfurler) buildContactData(publicKey string) (*common.StatusContactLinkPreview, error) {
	// contactID == "0x" + secp251k1 compressed public key as hex-encoded string
	contactID, err := multiformat.DeserializeCompressedKey(publicKey)
	if err != nil {
		return nil, err
	}

	contact := u.m.GetContactByID(contactID)

	// If no contact found locally, fetch it from waku
	if contact == nil {
		if contact, err = u.m.FetchContact(contactID, true); err != nil {
			return nil, fmt.Errorf("failed to request contact info from mailserver for public key '%s': %w", publicKey, err)
		}
	}

	c := &common.StatusContactLinkPreview{
		PublicKey:   contactID,
		DisplayName: contact.DisplayName,
		Description: contact.Bio,
	}

	if image, ok := contact.Images[images.SmallDimName]; ok {
		if err = updateThumbnail(&image, &c.Icon); err != nil {
			u.logger.Warn("unfurling status link: failed to set contact thumbnail", zap.Error(err))
		}
	}

	return c, nil
}

func (u *StatusUnfurler) buildCommunityData(communityID string, shard *shard.Shard) (*communities.Community, *common.StatusCommunityLinkPreview, error) {
	// This automatically checks the database
	community, err := u.m.FetchCommunity(&FetchCommunityRequest{
		CommunityKey:    communityID,
		Shard:           shard,
		TryDatabase:     true,
		WaitForResponse: true,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get community info for communityID '%s': %w", communityID, err)
	}

	if community == nil {
		return community, nil, fmt.Errorf("community info fetched, but it is empty")
	}

	statusCommunityLinkPreviews, err := community.ToStatusLinkPreview()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get status community link preview for communityID '%s': %w", communityID, err)
	}

	return community, statusCommunityLinkPreviews, nil
}

func (u *StatusUnfurler) buildChannelData(channelUUID string, communityID string, communityShard *shard.Shard) (*common.StatusCommunityChannelLinkPreview, error) {
	community, communityData, err := u.buildCommunityData(communityID, communityShard)
	if err != nil {
		return nil, fmt.Errorf("failed to build channel community data: %w", err)
	}

	channel, ok := community.Chats()[channelUUID]
	if !ok {
		return nil, fmt.Errorf("channel with channelID '%s' not found in community '%s'", channelUUID, communityID)
	}

	return &common.StatusCommunityChannelLinkPreview{
		ChannelUUID: channelUUID,
		Emoji:       channel.Identity.Emoji,
		DisplayName: channel.Identity.DisplayName,
		Description: channel.Identity.Description,
		Color:       channel.Identity.Color,
		Community:   communityData,
	}, nil
}

func (u *StatusUnfurler) Unfurl() (*common.StatusLinkPreview, error) {
	preview := new(common.StatusLinkPreview)
	preview.URL = u.url

	resp, err := ParseSharedURL(u.url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shared url: %w", err)
	}

	// If a URL has been successfully parsed,
	// any further errors should not be returned, only logged.

	if resp.Contact != nil {
		preview.Contact, err = u.buildContactData(resp.Contact.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("error when building contact data: %w", err)
		}
		return preview, nil
	}

	// NOTE: Currently channel data comes together with community data,
	//		 both `Community` and `Channel` fields will be present.
	//		 So we check for Channel first, then Community.

	if resp.Channel != nil {
		if resp.Community == nil {
			return preview, fmt.Errorf("channel community can't be empty")
		}
		preview.Channel, err = u.buildChannelData(resp.Channel.ChannelUUID, resp.Community.CommunityID, resp.Shard)
		if err != nil {
			return nil, fmt.Errorf("error when building channel data: %w", err)
		}
		return preview, nil
	}

	if resp.Community != nil {
		_, preview.Community, err = u.buildCommunityData(resp.Community.CommunityID, resp.Shard)
		if err != nil {
			return nil, fmt.Errorf("error when building community data: %w", err)
		}
		return preview, nil
	}

	return nil, fmt.Errorf("shared url does not contain contact, community or channel data")
}
