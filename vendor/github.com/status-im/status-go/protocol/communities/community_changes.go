package communities

import (
	"crypto/ecdsa"

	"github.com/status-im/status-go/protocol/protobuf"
)

type CommunityChatChanges struct {
	ChatModified                  *protobuf.CommunityChat
	MembersAdded                  map[string]*protobuf.CommunityMember
	MembersRemoved                map[string]*protobuf.CommunityMember
	CategoryModified              string
	PositionModified              int
	FirstMessageTimestampModified uint32
}

type CommunityChanges struct {
	Community *Community `json:"community"`

	ControlNodeChanged *ecdsa.PublicKey `json:"controlNodeChanged"`

	MembersAdded   map[string]*protobuf.CommunityMember `json:"membersAdded"`
	MembersRemoved map[string]*protobuf.CommunityMember `json:"membersRemoved"`

	TokenPermissionsAdded    map[string]*CommunityTokenPermission `json:"tokenPermissionsAdded"`
	TokenPermissionsModified map[string]*CommunityTokenPermission `json:"tokenPermissionsModified"`
	TokenPermissionsRemoved  map[string]*CommunityTokenPermission `json:"tokenPermissionsRemoved"`

	ChatsRemoved  map[string]*protobuf.CommunityChat `json:"chatsRemoved"`
	ChatsAdded    map[string]*protobuf.CommunityChat `json:"chatsAdded"`
	ChatsModified map[string]*CommunityChatChanges   `json:"chatsModified"`

	CategoriesRemoved  []string                               `json:"categoriesRemoved"`
	CategoriesAdded    map[string]*protobuf.CommunityCategory `json:"categoriesAdded"`
	CategoriesModified map[string]*protobuf.CommunityCategory `json:"categoriesModified"`

	MemberWalletsRemoved []string                               `json:"memberWalletsRemoved"`
	MemberWalletsAdded   map[string][]*protobuf.RevealedAccount `json:"memberWalletsAdded"`

	// ShouldMemberJoin indicates whether the user should join this community
	// automatically
	ShouldMemberJoin bool `json:"memberAdded"`

	// MemberKicked indicates whether the user has been kicked out
	MemberKicked bool `json:"memberRemoved"`
}

func EmptyCommunityChanges() *CommunityChanges {
	return &CommunityChanges{
		MembersAdded:   make(map[string]*protobuf.CommunityMember),
		MembersRemoved: make(map[string]*protobuf.CommunityMember),

		TokenPermissionsAdded:    make(map[string]*CommunityTokenPermission),
		TokenPermissionsModified: make(map[string]*CommunityTokenPermission),
		TokenPermissionsRemoved:  make(map[string]*CommunityTokenPermission),

		ChatsRemoved:  make(map[string]*protobuf.CommunityChat),
		ChatsAdded:    make(map[string]*protobuf.CommunityChat),
		ChatsModified: make(map[string]*CommunityChatChanges),

		CategoriesRemoved:  []string{},
		CategoriesAdded:    make(map[string]*protobuf.CommunityCategory),
		CategoriesModified: make(map[string]*protobuf.CommunityCategory),

		MemberWalletsRemoved: []string{},
		MemberWalletsAdded:   make(map[string][]*protobuf.RevealedAccount),
	}
}

func (c *CommunityChanges) HasNewMember(identity string) bool {
	if len(c.MembersAdded) == 0 {
		return false
	}
	_, ok := c.MembersAdded[identity]
	return ok
}

func (c *CommunityChanges) HasMemberLeft(identity string) bool {
	if len(c.MembersRemoved) == 0 {
		return false
	}
	_, ok := c.MembersRemoved[identity]
	return ok
}

func EvaluateCommunityChanges(origin, modified *Community) *CommunityChanges {
	changes := evaluateCommunityChangesByDescription(origin.Description(), modified.Description())

	if origin.ControlNode() != nil && !modified.ControlNode().Equal(origin.ControlNode()) {
		changes.ControlNodeChanged = modified.ControlNode()
	}

	originTokenPermissions := origin.tokenPermissions()
	modifiedTokenPermissions := modified.tokenPermissions()

	// Check for modified or removed token permissions
	for id, originPermission := range originTokenPermissions {
		if modifiedPermission := modifiedTokenPermissions[id]; modifiedPermission != nil {
			if !modifiedPermission.Equals(originPermission) {
				changes.TokenPermissionsModified[id] = modifiedPermission
			}
		} else {
			changes.TokenPermissionsRemoved[id] = originPermission
		}
	}

	// Check for added token permissions
	for id, permission := range modifiedTokenPermissions {
		if _, ok := originTokenPermissions[id]; !ok {
			changes.TokenPermissionsAdded[id] = permission
		}
	}

	changes.Community = modified
	return changes
}

func evaluateCommunityChangesByDescription(origin, modified *protobuf.CommunityDescription) *CommunityChanges {
	changes := EmptyCommunityChanges()

	// Check for new members at the org level
	for pk, member := range modified.Members {
		if _, ok := origin.Members[pk]; !ok {
			if changes.MembersAdded == nil {
				changes.MembersAdded = make(map[string]*protobuf.CommunityMember)
			}
			changes.MembersAdded[pk] = member
		}
	}

	// Check for removed members at the org level
	for pk, member := range origin.Members {
		if _, ok := modified.Members[pk]; !ok {
			if changes.MembersRemoved == nil {
				changes.MembersRemoved = make(map[string]*protobuf.CommunityMember)
			}
			changes.MembersRemoved[pk] = member
		}
	}

	// check for removed chats
	for chatID, chat := range origin.Chats {
		if modified.Chats == nil {
			modified.Chats = make(map[string]*protobuf.CommunityChat)
		}
		if _, ok := modified.Chats[chatID]; !ok {
			if changes.ChatsRemoved == nil {
				changes.ChatsRemoved = make(map[string]*protobuf.CommunityChat)
			}

			changes.ChatsRemoved[chatID] = chat
		}
	}

	for chatID, chat := range modified.Chats {
		if origin.Chats == nil {
			origin.Chats = make(map[string]*protobuf.CommunityChat)
		}

		if _, ok := origin.Chats[chatID]; !ok {
			if changes.ChatsAdded == nil {
				changes.ChatsAdded = make(map[string]*protobuf.CommunityChat)
			}

			changes.ChatsAdded[chatID] = chat
		} else {
			// Check for members added
			for pk, member := range modified.Chats[chatID].Members {
				if _, ok := origin.Chats[chatID].Members[pk]; !ok {
					if changes.ChatsModified[chatID] == nil {
						changes.ChatsModified[chatID] = &CommunityChatChanges{
							MembersAdded:   make(map[string]*protobuf.CommunityMember),
							MembersRemoved: make(map[string]*protobuf.CommunityMember),
						}
					}

					changes.ChatsModified[chatID].MembersAdded[pk] = member
				}
			}

			// check for members removed
			for pk, member := range origin.Chats[chatID].Members {
				if _, ok := modified.Chats[chatID].Members[pk]; !ok {
					if changes.ChatsModified[chatID] == nil {
						changes.ChatsModified[chatID] = &CommunityChatChanges{
							MembersAdded:   make(map[string]*protobuf.CommunityMember),
							MembersRemoved: make(map[string]*protobuf.CommunityMember),
						}
					}

					changes.ChatsModified[chatID].MembersRemoved[pk] = member
				}
			}

			// check if first message timestamp was modified
			if origin.Chats[chatID].Identity.FirstMessageTimestamp !=
				modified.Chats[chatID].Identity.FirstMessageTimestamp {
				if changes.ChatsModified[chatID] == nil {
					changes.ChatsModified[chatID] = &CommunityChatChanges{
						MembersAdded:   make(map[string]*protobuf.CommunityMember),
						MembersRemoved: make(map[string]*protobuf.CommunityMember),
					}
				}
				changes.ChatsModified[chatID].FirstMessageTimestampModified = modified.Chats[chatID].Identity.FirstMessageTimestamp
			}
		}
	}

	// Check for categories that were removed
	for categoryID := range origin.Categories {
		if modified.Categories == nil {
			modified.Categories = make(map[string]*protobuf.CommunityCategory)
		}

		if modified.Chats == nil {
			modified.Chats = make(map[string]*protobuf.CommunityChat)
		}

		if _, ok := modified.Categories[categoryID]; !ok {
			changes.CategoriesRemoved = append(changes.CategoriesRemoved, categoryID)
		}

		if origin.Chats == nil {
			origin.Chats = make(map[string]*protobuf.CommunityChat)
		}
	}

	// Check for categories that were added
	for categoryID, category := range modified.Categories {
		if origin.Categories == nil {
			origin.Categories = make(map[string]*protobuf.CommunityCategory)
		}
		if _, ok := origin.Categories[categoryID]; !ok {
			if changes.CategoriesAdded == nil {
				changes.CategoriesAdded = make(map[string]*protobuf.CommunityCategory)
			}

			changes.CategoriesAdded[categoryID] = category
		} else {
			if origin.Categories[categoryID].Name != category.Name || origin.Categories[categoryID].Position != category.Position {
				changes.CategoriesModified[categoryID] = category
			}
		}
	}

	// Check for chat categories that were modified
	for chatID, chat := range modified.Chats {
		if origin.Chats == nil {
			origin.Chats = make(map[string]*protobuf.CommunityChat)
		}

		if _, ok := origin.Chats[chatID]; !ok {
			continue // It's a new chat
		}

		if origin.Chats[chatID].CategoryId != chat.CategoryId {
			if changes.ChatsModified[chatID] == nil {
				changes.ChatsModified[chatID] = &CommunityChatChanges{
					MembersAdded:   make(map[string]*protobuf.CommunityMember),
					MembersRemoved: make(map[string]*protobuf.CommunityMember),
				}
			}

			changes.ChatsModified[chatID].CategoryModified = chat.CategoryId
		}
	}

	return changes
}
