package protocol

import (
	"context"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/protocol/protobuf"
)

func (m *Messenger) prepareSyncKeycardsMessage(keyUID string) (message []*protobuf.SyncKeycard, err error) {
	keycards, err := m.settings.GetKeycardsWithSameKeyUID(keyUID)
	if err != nil {
		return
	}

	for _, kc := range keycards {
		syncKeycard := kc.ToSyncKeycard()
		message = append(message, syncKeycard)
	}

	return
}

func (m *Messenger) dispatchKeycardActivity(keyUID string, keycardUID string, newKeycardUID string, accountAddresses []types.Address,
	callback func(uint64) error) error {
	clock, _ := m.getLastClockWithRelatedChat()

	finalKeyUID := keyUID
	if finalKeyUID == "" {
		dbKeycard, err := m.settings.GetKeycardByKeycardUID(keycardUID)
		if err != nil {
			return err
		}
		finalKeyUID = dbKeycard.KeyUID
	}

	if err := callback(clock); err != nil {
		return err
	}

	return m.resolveAndSyncKeypairOrJustWalletAccount(finalKeyUID, types.Address{}, clock, m.dispatchMessage)
}

// This function stores keycard to db and notifies paired devices about that if keycard with `KeycardUID` is not already stored.
// Keycard position is fully maintained by the backend.
// If keycard is already stored, this function updates `KeycardName` and adds accounts which are not already added, in this case
// `KeycardLocked` and `Position` remains as they were, they won't be changed.
func (m *Messenger) SaveOrUpdateKeycard(ctx context.Context, keycard *accounts.Keycard) (err error) {
	dbKeycard, err := m.settings.GetKeycardByKeycardUID(keycard.KeycardUID)
	if err != nil && err != accounts.ErrNoKeycardForPassedKeycardUID {
		return err
	}

	if dbKeycard == nil {
		position, err := m.settings.GetPositionForNextNewKeycard()
		if err != nil {
			return err
		}
		keycard.Position = position
		keycard.KeycardLocked = false
	} else {
		keycard.Position = dbKeycard.Position
		keycard.KeycardLocked = dbKeycard.KeycardLocked
	}

	return m.dispatchKeycardActivity(keycard.KeyUID, "", "", []types.Address{}, func(clock uint64) error {
		return m.settings.SaveOrUpdateKeycard(*keycard, clock, true)
	})
}

func (m *Messenger) SetKeycardName(ctx context.Context, keycardUID string, kpName string) error {
	return m.dispatchKeycardActivity("", keycardUID, "", []types.Address{}, func(clock uint64) error {
		return m.settings.SetKeycardName(keycardUID, kpName, clock)
	})
}

func (m *Messenger) KeycardLocked(ctx context.Context, keycardUID string) error {
	return m.dispatchKeycardActivity("", keycardUID, "", []types.Address{}, func(clock uint64) error {
		return m.settings.KeycardLocked(keycardUID, clock)
	})
}

func (m *Messenger) KeycardUnlocked(ctx context.Context, keycardUID string) error {
	return m.dispatchKeycardActivity("", keycardUID, "", []types.Address{}, func(clock uint64) error {
		return m.settings.KeycardUnlocked(keycardUID, clock)
	})
}

func (m *Messenger) DeleteKeycardAccounts(ctx context.Context, keycardUID string, accountAddresses []types.Address) error {
	return m.dispatchKeycardActivity("", keycardUID, "", accountAddresses, func(clock uint64) error {
		return m.settings.DeleteKeycardAccounts(keycardUID, accountAddresses, clock)
	})
}

func (m *Messenger) DeleteKeycard(ctx context.Context, keycardUID string) error {
	return m.dispatchKeycardActivity("", keycardUID, "", []types.Address{}, func(clock uint64) error {
		return m.settings.DeleteKeycard(keycardUID, clock)
	})
}

func (m *Messenger) DeleteAllKeycardsWithKeyUID(ctx context.Context, keyUID string) error {
	return m.dispatchKeycardActivity(keyUID, "", "", []types.Address{}, func(clock uint64) error {
		return m.settings.DeleteAllKeycardsWithKeyUID(keyUID, clock)
	})
}

func (m *Messenger) UpdateKeycardUID(ctx context.Context, oldKeycardUID string, newKeycardUID string) error {
	return m.dispatchKeycardActivity("", oldKeycardUID, newKeycardUID, []types.Address{}, func(clock uint64) error {
		return m.settings.UpdateKeycardUID(oldKeycardUID, newKeycardUID, clock)
	})
}
