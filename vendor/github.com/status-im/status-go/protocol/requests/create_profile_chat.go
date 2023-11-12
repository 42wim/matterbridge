package requests

import (
	"errors"

	"github.com/status-im/status-go/deprecation"
)

// Deprecated: errCreateProfileChatInvalidID shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
var errCreateProfileChatInvalidID = errors.New("create-public-chat: invalid id")

// Deprecated: CreateProfileChat shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
type CreateProfileChat struct {
	ID string `json:"id"`
}

// Deprecated: Validate shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
func (c *CreateProfileChat) Validate() error {
	// Return error to prevent usafe of deprecated function
	if deprecation.ChatProfileDeprecated {
		return errors.New("profile chats are deprecated")
	}

	if len(c.ID) == 0 {
		return errCreateProfileChatInvalidID
	}

	return nil
}
