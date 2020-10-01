package gumble

import (
	"github.com/golang/protobuf/proto"
	"layeh.com/gumble/gumble/MumbleProto"
)

// ACL contains a list of ACLGroups and ACLRules linked to a channel.
type ACL struct {
	// The channel to which the ACL belongs.
	Channel *Channel
	// The ACL's groups.
	Groups []*ACLGroup
	// The ACL's rules.
	Rules []*ACLRule
	// Does the ACL inherits the parent channel's ACLs?
	Inherits bool
}

func (a *ACL) writeMessage(client *Client) error {
	packet := MumbleProto.ACL{
		ChannelId:   &a.Channel.ID,
		Groups:      make([]*MumbleProto.ACL_ChanGroup, len(a.Groups)),
		Acls:        make([]*MumbleProto.ACL_ChanACL, len(a.Rules)),
		InheritAcls: &a.Inherits,
		Query:       proto.Bool(false),
	}

	for i, group := range a.Groups {
		packet.Groups[i] = &MumbleProto.ACL_ChanGroup{
			Name:        &group.Name,
			Inherit:     &group.InheritUsers,
			Inheritable: &group.Inheritable,
			Add:         make([]uint32, 0, len(group.UsersAdd)),
			Remove:      make([]uint32, 0, len(group.UsersRemove)),
		}
		for _, user := range group.UsersAdd {
			packet.Groups[i].Add = append(packet.Groups[i].Add, user.UserID)
		}
		for _, user := range group.UsersRemove {
			packet.Groups[i].Remove = append(packet.Groups[i].Remove, user.UserID)
		}
	}

	for i, rule := range a.Rules {
		packet.Acls[i] = &MumbleProto.ACL_ChanACL{
			ApplyHere: &rule.AppliesCurrent,
			ApplySubs: &rule.AppliesChildren,
			Grant:     proto.Uint32(uint32(rule.Granted)),
			Deny:      proto.Uint32(uint32(rule.Denied)),
		}
		if rule.User != nil {
			packet.Acls[i].UserId = &rule.User.UserID
		}
		if rule.Group != nil {
			packet.Acls[i].Group = &rule.Group.Name
		}
	}

	return client.Conn.WriteProto(&packet)
}

// ACLUser is a registered user who is part of or can be part of an ACL group
// or rule.
type ACLUser struct {
	// The user ID of the user.
	UserID uint32
	// The name of the user.
	Name string
}

// ACLGroup is a named group of registered users which can be used in an
// ACLRule.
type ACLGroup struct {
	// The ACL group name.
	Name string
	// Is the group inherited from the parent channel's ACL?
	Inherited bool
	// Are group members are inherited from the parent channel's ACL?
	InheritUsers bool
	// Can the group be inherited by child channels?
	Inheritable bool
	// The users who are explicitly added to, explicitly removed from, and
	// inherited into the group.
	UsersAdd, UsersRemove, UsersInherited map[uint32]*ACLUser
}

// ACL group names that are built-in.
const (
	ACLGroupEveryone       = "all"
	ACLGroupAuthenticated  = "auth"
	ACLGroupInsideChannel  = "in"
	ACLGroupOutsideChannel = "out"
)

// ACLRule is a set of granted and denied permissions given to an ACLUser or
// ACLGroup.
type ACLRule struct {
	// Does the rule apply to the channel in which the rule is defined?
	AppliesCurrent bool
	// Does the rule apply to the children of the channel in which the rule is
	// defined?
	AppliesChildren bool
	// Is the rule inherited from the parent channel's ACL?
	Inherited bool

	// The permissions granted by the rule.
	Granted Permission
	// The permissions denied by the rule.
	Denied Permission

	// The ACL user the rule applies to. Can be nil.
	User *ACLUser
	// The ACL group the rule applies to. Can be nil.
	Group *ACLGroup
}
