package communities

import (
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
)

func validateCommunityChat(desc *protobuf.CommunityDescription, chat *protobuf.CommunityChat) error {
	if chat == nil {
		return ErrInvalidCommunityDescription
	}
	if chat.Permissions == nil {
		return ErrInvalidCommunityDescriptionNoChatPermissions
	}
	if chat.Permissions.Access == protobuf.CommunityPermissions_UNKNOWN_ACCESS {
		return ErrInvalidCommunityDescriptionUnknownChatAccess
	}

	if len(chat.CategoryId) != 0 {
		if _, exists := desc.Categories[chat.CategoryId]; !exists {
			return ErrInvalidCommunityDescriptionUnknownChatCategory
		}
	}

	if chat.Identity == nil {
		return ErrInvalidCommunityDescriptionChatIdentity
	}

	for pk := range chat.Members {
		if desc.Members == nil {
			return ErrInvalidCommunityDescriptionMemberInChatButNotInOrg
		}
		// Check member is in the org as well
		if _, ok := desc.Members[pk]; !ok {
			return ErrInvalidCommunityDescriptionMemberInChatButNotInOrg
		}
	}

	return nil
}

func validateCommunityCategory(category *protobuf.CommunityCategory) error {
	if len(category.CategoryId) == 0 {
		return ErrInvalidCommunityDescriptionCategoryNoID
	}

	if len(category.Name) == 0 {
		return ErrInvalidCommunityDescriptionCategoryNoName
	}

	return nil
}

func ValidateCommunityDescription(desc *protobuf.CommunityDescription) error {
	if desc == nil {
		return ErrInvalidCommunityDescription
	}
	if desc.Permissions == nil {
		return ErrInvalidCommunityDescriptionNoOrgPermissions
	}
	if desc.Permissions.Access == protobuf.CommunityPermissions_UNKNOWN_ACCESS {
		return ErrInvalidCommunityDescriptionUnknownOrgAccess
	}

	valid := requests.ValidateTags(desc.Tags)
	if !valid {
		return ErrInvalidCommunityTags
	}

	for _, category := range desc.Categories {
		if err := validateCommunityCategory(category); err != nil {
			return err
		}
	}

	for _, chat := range desc.Chats {
		if err := validateCommunityChat(desc, chat); err != nil {
			return err
		}
	}

	return nil
}
