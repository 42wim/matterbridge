package store

import (
	"context"

	"go.mau.fi/libsignal/groups/state/record"
	"go.mau.fi/libsignal/protocol"
)

type SenderKey interface {
	StoreSenderKey(ctx context.Context, senderKeyName *protocol.SenderKeyName, keyRecord *record.SenderKey) error
	LoadSenderKey(ctx context.Context, senderKeyName *protocol.SenderKeyName) (*record.SenderKey, error)
}
