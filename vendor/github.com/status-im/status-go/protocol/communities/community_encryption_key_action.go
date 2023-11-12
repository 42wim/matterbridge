package communities

import (
	"github.com/status-im/status-go/protocol/protobuf"
)

type KeyDistributor interface {
	Generate(community *Community, keyActions *EncryptionKeyActions) error
	Distribute(community *Community, keyActions *EncryptionKeyActions) error
}

type EncryptionKeyActionType int

const (
	EncryptionKeyNone EncryptionKeyActionType = iota
	EncryptionKeyAdd
	EncryptionKeyRemove
	EncryptionKeyRekey
	EncryptionKeySendToMembers
)

type EncryptionKeyAction struct {
	ActionType     EncryptionKeyActionType
	Members        map[string]*protobuf.CommunityMember
	RemovedMembers map[string]*protobuf.CommunityMember
}

type EncryptionKeyActions struct {
	// community-level encryption key action
	CommunityKeyAction EncryptionKeyAction

	// channel-level encryption key actions
	ChannelKeysActions map[string]EncryptionKeyAction // key is: chatID
}

func EvaluateCommunityEncryptionKeyActions(origin, modified *Community) *EncryptionKeyActions {
	if origin == nil {
		// `modified` is a new community, create empty `origin` community
		origin = &Community{
			config: &Config{
				ID: modified.config.ID,
				CommunityDescription: &protobuf.CommunityDescription{
					Members:                 map[string]*protobuf.CommunityMember{},
					Permissions:             &protobuf.CommunityPermissions{},
					Identity:                &protobuf.ChatIdentity{},
					Chats:                   map[string]*protobuf.CommunityChat{},
					Categories:              map[string]*protobuf.CommunityCategory{},
					AdminSettings:           &protobuf.CommunityAdminSettings{},
					TokenPermissions:        map[string]*protobuf.CommunityTokenPermission{},
					CommunityTokensMetadata: []*protobuf.CommunityTokenMetadata{},
				},
			},
		}
	}

	changes := EvaluateCommunityChanges(origin, modified)

	result := &EncryptionKeyActions{
		CommunityKeyAction: *evaluateCommunityLevelEncryptionKeyAction(origin, modified, changes),
		ChannelKeysActions: *evaluateChannelLevelEncryptionKeyActions(origin, modified, changes),
	}
	return result
}

func evaluateCommunityLevelEncryptionKeyAction(origin, modified *Community, changes *CommunityChanges) *EncryptionKeyAction {
	return evaluateEncryptionKeyAction(
		origin.Encrypted(),
		modified.Encrypted(),
		changes.ControlNodeChanged != nil,
		modified.config.CommunityDescription.Members,
		changes.MembersAdded,
		changes.MembersRemoved,
	)
}

func evaluateChannelLevelEncryptionKeyActions(origin, modified *Community, changes *CommunityChanges) *map[string]EncryptionKeyAction {
	result := make(map[string]EncryptionKeyAction)

	for channelID := range modified.config.CommunityDescription.Chats {
		membersAdded := make(map[string]*protobuf.CommunityMember)
		membersRemoved := make(map[string]*protobuf.CommunityMember)

		chatChanges, ok := changes.ChatsModified[channelID]
		if ok {
			membersAdded = chatChanges.MembersAdded
			membersRemoved = chatChanges.MembersRemoved
		}

		result[channelID] = *evaluateEncryptionKeyAction(
			origin.ChannelEncrypted(channelID),
			modified.ChannelEncrypted(channelID),
			changes.ControlNodeChanged != nil,
			modified.config.CommunityDescription.Chats[channelID].Members,
			membersAdded,
			membersRemoved,
		)
	}

	return &result
}

func evaluateEncryptionKeyAction(originEncrypted, modifiedEncrypted, controlNodeChanged bool,
	allMembers, membersAdded, membersRemoved map[string]*protobuf.CommunityMember) *EncryptionKeyAction {
	result := &EncryptionKeyAction{
		ActionType: EncryptionKeyNone,
		Members:    map[string]*protobuf.CommunityMember{},
	}

	copyMap := func(source map[string]*protobuf.CommunityMember) map[string]*protobuf.CommunityMember {
		to := make(map[string]*protobuf.CommunityMember)
		for pubKey, member := range source {
			to[pubKey] = member
		}
		return to
	}

	// control node changed on closed community/channel
	if controlNodeChanged && modifiedEncrypted {
		result.ActionType = EncryptionKeyRekey
		result.Members = copyMap(allMembers)
		return result
	}

	// encryption was just added
	if modifiedEncrypted && !originEncrypted {
		result.ActionType = EncryptionKeyAdd
		result.Members = copyMap(allMembers)
		return result
	}

	// encryption was just removed
	if !modifiedEncrypted && originEncrypted {
		result.ActionType = EncryptionKeyRemove
		result.Members = copyMap(allMembers)
		return result
	}

	// open community/channel does not require any actions
	if !modifiedEncrypted {
		return result
	}

	if len(membersRemoved) > 0 {
		result.ActionType = EncryptionKeyRekey
		result.Members = copyMap(allMembers)
		result.RemovedMembers = copyMap(membersRemoved)
	} else if len(membersAdded) > 0 {
		result.ActionType = EncryptionKeySendToMembers
		result.Members = copyMap(membersAdded)
	}

	return result
}
