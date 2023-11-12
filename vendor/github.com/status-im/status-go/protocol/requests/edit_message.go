package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

var ErrEditMessageInvalidID = errors.New("edit-message: invalid id")
var ErrEditMessageInvalidText = errors.New("edit-message: invalid text")

type EditMessage struct {
	ID                 types.HexBytes                   `json:"id"`
	Text               string                           `json:"text"`
	ContentType        protobuf.ChatMessage_ContentType `json:"content-type"`
	LinkPreviews       []common.LinkPreview             `json:"linkPreviews"`
	StatusLinkPreviews []common.StatusLinkPreview       `json:"statusLinkPreviews"`
}

func (e *EditMessage) Validate() error {
	if len(e.ID) == 0 {
		return ErrEditMessageInvalidID
	}

	if len(e.Text) == 0 {
		return ErrEditMessageInvalidText
	}

	return nil
}
