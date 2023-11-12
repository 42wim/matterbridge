package communities

import (
	"golang.org/x/exp/slices"

	"github.com/status-im/status-go/protocol/protobuf"
)

var adminAuthorizedEventTypes = []protobuf.CommunityEvent_EventType{
	protobuf.CommunityEvent_COMMUNITY_EDIT,
	protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE,
	protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE,
	protobuf.CommunityEvent_COMMUNITY_CATEGORY_CREATE,
	protobuf.CommunityEvent_COMMUNITY_CATEGORY_DELETE,
	protobuf.CommunityEvent_COMMUNITY_CATEGORY_EDIT,
	protobuf.CommunityEvent_COMMUNITY_CHANNEL_CREATE,
	protobuf.CommunityEvent_COMMUNITY_CHANNEL_DELETE,
	protobuf.CommunityEvent_COMMUNITY_CHANNEL_EDIT,
	protobuf.CommunityEvent_COMMUNITY_CATEGORY_REORDER,
	protobuf.CommunityEvent_COMMUNITY_CHANNEL_REORDER,
	protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT,
	protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT,
	protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK,
	protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN,
	protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN,
}

var tokenMasterAuthorizedEventTypes = append(adminAuthorizedEventTypes, []protobuf.CommunityEvent_EventType{
	protobuf.CommunityEvent_COMMUNITY_TOKEN_ADD,
}...)

var ownerAuthorizedEventTypes = tokenMasterAuthorizedEventTypes

var rolesToAuthorizedEventTypes = map[protobuf.CommunityMember_Roles][]protobuf.CommunityEvent_EventType{
	protobuf.CommunityMember_ROLE_NONE:         []protobuf.CommunityEvent_EventType{},
	protobuf.CommunityMember_ROLE_OWNER:        ownerAuthorizedEventTypes,
	protobuf.CommunityMember_ROLE_ADMIN:        adminAuthorizedEventTypes,
	protobuf.CommunityMember_ROLE_TOKEN_MASTER: tokenMasterAuthorizedEventTypes,
}

var adminAuthorizedPermissionTypes = []protobuf.CommunityTokenPermission_Type{
	protobuf.CommunityTokenPermission_BECOME_MEMBER,
	protobuf.CommunityTokenPermission_CAN_VIEW_CHANNEL,
	protobuf.CommunityTokenPermission_CAN_VIEW_AND_POST_CHANNEL,
}

var tokenMasterAuthorizedPermissionTypes = append(adminAuthorizedPermissionTypes, []protobuf.CommunityTokenPermission_Type{}...)

var ownerAuthorizedPermissionTypes = append(tokenMasterAuthorizedPermissionTypes, []protobuf.CommunityTokenPermission_Type{
	protobuf.CommunityTokenPermission_BECOME_ADMIN,
	protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER,
}...)

var rolesToAuthorizedPermissionTypes = map[protobuf.CommunityMember_Roles][]protobuf.CommunityTokenPermission_Type{
	protobuf.CommunityMember_ROLE_NONE:         []protobuf.CommunityTokenPermission_Type{},
	protobuf.CommunityMember_ROLE_OWNER:        ownerAuthorizedPermissionTypes,
	protobuf.CommunityMember_ROLE_ADMIN:        adminAuthorizedPermissionTypes,
	protobuf.CommunityMember_ROLE_TOKEN_MASTER: tokenMasterAuthorizedPermissionTypes,
}

func canRolesPerformEvent(roles []protobuf.CommunityMember_Roles, eventType protobuf.CommunityEvent_EventType) bool {
	for _, role := range roles {
		if slices.Contains(rolesToAuthorizedEventTypes[role], eventType) {
			return true
		}
	}
	return false
}

func canRolesModifyPermission(roles []protobuf.CommunityMember_Roles, permissionType protobuf.CommunityTokenPermission_Type) bool {
	for _, role := range roles {
		if slices.Contains(rolesToAuthorizedPermissionTypes[role], permissionType) {
			return true
		}
	}
	return false
}

func canRolesKickOrBanMember(senderRoles []protobuf.CommunityMember_Roles, memberRoles []protobuf.CommunityMember_Roles) bool {
	// Owner can kick everyone
	if slices.Contains(senderRoles, protobuf.CommunityMember_ROLE_OWNER) {
		return true
	}

	// TokenMaster can kick normal members and admins
	if (slices.Contains(senderRoles, protobuf.CommunityMember_ROLE_TOKEN_MASTER)) &&
		!(slices.Contains(memberRoles, protobuf.CommunityMember_ROLE_TOKEN_MASTER) ||
			slices.Contains(memberRoles, protobuf.CommunityMember_ROLE_OWNER)) {
		return true
	}

	// Admins can kick normal members
	if (slices.Contains(senderRoles, protobuf.CommunityMember_ROLE_ADMIN)) &&
		!(slices.Contains(memberRoles, protobuf.CommunityMember_ROLE_ADMIN) ||
			slices.Contains(memberRoles, protobuf.CommunityMember_ROLE_TOKEN_MASTER) ||
			slices.Contains(memberRoles, protobuf.CommunityMember_ROLE_OWNER)) {
		return true
	}

	// Normal members can't kick anyone
	return false
}

func RolesAuthorizedToPerformEvent(senderRoles []protobuf.CommunityMember_Roles, memberRoles []protobuf.CommunityMember_Roles, event *CommunityEvent) bool {
	if !canRolesPerformEvent(senderRoles, event.Type) {
		return false
	}

	if event.Type == protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE ||
		event.Type == protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE {
		return canRolesModifyPermission(senderRoles, event.TokenPermission.Type)
	}

	if event.Type == protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN ||
		event.Type == protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK {
		return canRolesKickOrBanMember(senderRoles, memberRoles)
	}

	return true
}
