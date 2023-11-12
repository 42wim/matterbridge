package protocol

import (
	"github.com/pkg/errors"
)

var (
	ErrChatIDEmpty      = errors.New("chat ID is empty")
	ErrChatNotFound     = errors.New("can't find chat")
	ErrNotImplemented   = errors.New("not implemented")
	ErrContactNotFound  = errors.New("contact not found")
	ErrCommunityIDEmpty = errors.New("community ID is empty")
	ErrUserNotMember    = errors.New("user not a member")
)
