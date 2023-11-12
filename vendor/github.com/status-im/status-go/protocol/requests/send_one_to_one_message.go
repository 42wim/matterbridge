package requests

import (
	"errors"
)

var ErrSendOneToOneMessageInvalidID = errors.New("send-one-to-one-message: invalid id")
var ErrSendOneToOneMessageInvalidMessage = errors.New("send-one-to-one-message: invalid message")

type SendOneToOneMessage struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (a *SendOneToOneMessage) Validate() error {
	if len(a.ID) == 0 {
		return ErrSendOneToOneMessageInvalidID
	}

	if len(a.Message) == 0 {
		return ErrSendOneToOneMessageInvalidMessage
	}

	return nil
}

func (a *SendOneToOneMessage) HexID() (string, error) {
	return ConvertCompressedToLegacyKey(a.ID)
}
