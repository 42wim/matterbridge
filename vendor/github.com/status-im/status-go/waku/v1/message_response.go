package v1

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/status-im/status-go/waku/common"
)

// MultiVersionResponse allows to decode response into chosen version.
type MultiVersionResponse struct {
	Version  uint
	Response rlp.RawValue
}

// DecodeResponse1 decodes response into first version of the messages response.
func (m MultiVersionResponse) DecodeResponse1() (resp common.MessagesResponse, err error) {
	return resp, rlp.DecodeBytes(m.Response, &resp)
}

// Version1MessageResponse first version of the message response.
type Version1MessageResponse struct {
	Version  uint
	Response common.MessagesResponse
}

// NewMessagesResponse returns instance of the version messages response.
func NewMessagesResponse(batch gethcommon.Hash, errors []common.EnvelopeError) Version1MessageResponse {
	return Version1MessageResponse{
		Version: 1,
		Response: common.MessagesResponse{
			Hash:   batch,
			Errors: errors,
		},
	}
}
