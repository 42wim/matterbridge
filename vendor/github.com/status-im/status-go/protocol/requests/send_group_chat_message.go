package requests

import (
	"errors"
)

var ErrSendGroupChatMessageInvalidID = errors.New("send-group-chat-message: invalid id")
var ErrSendGroupChatMessageInvalidMessage = errors.New("send-group-chat-message: invalid message")

type SendGroupChatMessage struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (a *SendGroupChatMessage) Validate() error {
	if len(a.ID) == 0 {
		return ErrSendGroupChatMessageInvalidID
	}

	if len(a.Message) == 0 {
		return ErrSendGroupChatMessageInvalidMessage
	}

	return nil
}
