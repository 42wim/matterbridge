package requests

import (
	"errors"

	"github.com/status-im/status-go/api/multiformat"
)

var ErrSendContactRequestInvalidID = errors.New("send-contact-request: invalid id")
var ErrSendContactRequestInvalidMessage = errors.New("send-contact-request: invalid message")

const legacyKeyLength = 132

type SendContactRequest struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (a *SendContactRequest) Validate() error {
	if len(a.ID) == 0 {
		return ErrSendContactRequestInvalidID
	}

	if len(a.Message) == 0 {
		return ErrSendContactRequestInvalidMessage
	}

	return nil
}

func ConvertCompressedToLegacyKey(k string) (string, error) {
	if len(k) == legacyKeyLength {
		return k, nil
	}
	return multiformat.DeserializeCompressedKey(k)
}

func (a *SendContactRequest) HexID() (string, error) {
	return ConvertCompressedToLegacyKey(a.ID)
}
