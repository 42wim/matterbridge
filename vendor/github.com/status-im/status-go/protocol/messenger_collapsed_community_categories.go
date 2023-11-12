package protocol

import (
	"github.com/status-im/status-go/protocol/requests"
)

type CollapsedCommunityCategory struct {
	CommunityID string `json:"communityId"`
	CategoryID  string `json:"categoryId"`
	Collapsed   bool   `json:"collapsed"`
}

func (m *Messenger) ToggleCollapsedCommunityCategory(request *requests.ToggleCollapsedCommunityCategory) error {
	if err := request.Validate(); err != nil {
		return err
	}

	collapsedCategory := CollapsedCommunityCategory{
		CommunityID: request.CommunityID,
		CategoryID:  request.CategoryID,
		Collapsed:   request.Collapsed,
	}

	return m.persistence.UpsertCollapsedCommunityCategory(collapsedCategory)
}

func (m *Messenger) CollapsedCommunityCategories() ([]CollapsedCommunityCategory, error) {
	return m.persistence.CollapsedCommunityCategories()
}
