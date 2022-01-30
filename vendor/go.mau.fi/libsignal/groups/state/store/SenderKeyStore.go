package store

import (
	"go.mau.fi/libsignal/groups/state/record"
	"go.mau.fi/libsignal/protocol"
)

type SenderKey interface {
	StoreSenderKey(senderKeyName *protocol.SenderKeyName, keyRecord *record.SenderKey)
	LoadSenderKey(senderKeyName *protocol.SenderKeyName) *record.SenderKey
}
