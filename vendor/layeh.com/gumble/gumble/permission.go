package gumble

// Permission is a bitmask of permissions given to a certain user.
type Permission int

// Permissions that can be applied in any channel.
const (
	PermissionWrite Permission = 1 << iota
	PermissionTraverse
	PermissionEnter
	PermissionSpeak
	PermissionMuteDeafen
	PermissionMove
	PermissionMakeChannel
	PermissionLinkChannel
	PermissionWhisper
	PermissionTextMessage
	PermissionMakeTemporaryChannel
)

// Permissions that can only be applied in the root channel.
const (
	PermissionKick Permission = 0x10000 << iota
	PermissionBan
	PermissionRegister
	PermissionRegisterSelf
)

// Has returns true if the Permission p contains Permission o has part of its
// bitmask.
func (p Permission) Has(o Permission) bool {
	return p&o == o
}
