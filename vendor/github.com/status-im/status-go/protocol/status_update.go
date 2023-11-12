package protocol

import "github.com/status-im/status-go/protocol/protobuf"

type UserStatus struct {
	PublicKey  string `json:"publicKey,omitempty"`
	StatusType int    `json:"statusType"`
	Clock      uint64 `json:"clock"`
	CustomText string `json:"text"`
}

func ToUserStatus(msg *protobuf.StatusUpdate) UserStatus {
	return UserStatus{
		StatusType: int(msg.StatusType),
		Clock:      msg.Clock,
		CustomText: msg.CustomText,
	}
}
