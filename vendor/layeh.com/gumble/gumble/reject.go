package gumble

import (
	"strconv"

	"layeh.com/gumble/gumble/MumbleProto"
)

// RejectType describes why a client connection was rejected by the server.
type RejectType int

// The possible reason why a client connection was rejected by the server.
const (
	RejectNone              RejectType = RejectType(MumbleProto.Reject_None)
	RejectVersion           RejectType = RejectType(MumbleProto.Reject_WrongVersion)
	RejectUserName          RejectType = RejectType(MumbleProto.Reject_InvalidUsername)
	RejectUserCredentials   RejectType = RejectType(MumbleProto.Reject_WrongUserPW)
	RejectServerPassword    RejectType = RejectType(MumbleProto.Reject_WrongServerPW)
	RejectUsernameInUse     RejectType = RejectType(MumbleProto.Reject_UsernameInUse)
	RejectServerFull        RejectType = RejectType(MumbleProto.Reject_ServerFull)
	RejectNoCertificate     RejectType = RejectType(MumbleProto.Reject_NoCertificate)
	RejectAuthenticatorFail RejectType = RejectType(MumbleProto.Reject_AuthenticatorFail)
)

// RejectError is returned by DialWithDialer when the server rejects the client
// connection.
type RejectError struct {
	Type   RejectType
	Reason string
}

// Error implements error.
func (e RejectError) Error() string {
	var msg string
	switch e.Type {
	case RejectNone:
		msg = "none"
	case RejectVersion:
		msg = "wrong client version"
	case RejectUserName:
		msg = "invalid username"
	case RejectUserCredentials:
		msg = "incorrect user credentials"
	case RejectServerPassword:
		msg = "incorrect server password"
	case RejectUsernameInUse:
		msg = "username in use"
	case RejectServerFull:
		msg = "server full"
	case RejectNoCertificate:
		msg = "no certificate"
	case RejectAuthenticatorFail:
		msg = "authenticator fail"
	default:
		msg = "unknown type " + strconv.Itoa(int(e.Type))
	}
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	return msg
}
