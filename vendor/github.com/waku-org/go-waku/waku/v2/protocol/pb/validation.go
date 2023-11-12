package pb

import (
	"errors"
)

const MaxMetaAttrLength = 64

var (
	ErrMissingPayload      = errors.New("missing Payload field")
	ErrMissingContentTopic = errors.New("missing ContentTopic field")
	ErrInvalidMetaLength   = errors.New("invalid length for Meta field")
)

func (msg *WakuMessage) Validate() error {
	if len(msg.Payload) == 0 {
		return ErrMissingPayload
	}

	if msg.ContentTopic == "" {
		return ErrMissingContentTopic
	}

	if len(msg.Meta) > MaxMetaAttrLength {
		return ErrInvalidMetaLength
	}

	return nil
}
