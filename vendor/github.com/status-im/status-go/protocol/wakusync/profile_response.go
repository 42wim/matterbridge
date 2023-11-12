package wakusync

import (
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/identity"
	"github.com/status-im/status-go/services/ens"
)

type BackedUpProfile struct {
	DisplayName        string                 `json:"displayName,omitempty"`
	Images             []images.IdentityImage `json:"images,omitempty"`
	SocialLinks        []*identity.SocialLink `json:"socialLinks,omitempty"`
	EnsUsernameDetails []*ens.UsernameDetail  `json:"ensUsernameDetails,omitempty"`
}

func (sfwr *WakuBackedUpDataResponse) SetDisplayName(displayName string) {
	sfwr.Profile.DisplayName = displayName
}

func (sfwr *WakuBackedUpDataResponse) SetImages(images []images.IdentityImage) {
	sfwr.Profile.Images = images
}

func (sfwr *WakuBackedUpDataResponse) SetSocialLinks(socialLinks []*identity.SocialLink) {
	sfwr.Profile.SocialLinks = socialLinks
}

func (sfwr *WakuBackedUpDataResponse) SetEnsUsernameDetails(ensUsernameDetails []*ens.UsernameDetail) {
	sfwr.Profile.EnsUsernameDetails = ensUsernameDetails
}
