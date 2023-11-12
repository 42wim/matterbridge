package protocol

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"

	gethcommon "github.com/ethereum/go-ethereum/common"
	multiAccCommon "github.com/status-im/status-go/multiaccounts/common"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	v1protocol "github.com/status-im/status-go/protocol/v1"
	"github.com/status-im/status-go/services/wallet"
)

func (m *Messenger) UpsertSavedAddress(ctx context.Context, sa wallet.SavedAddress) error {
	sa.UpdateClock, _ = m.getLastClockWithRelatedChat()
	err := m.savedAddressesManager.UpdateMetadataAndUpsertSavedAddress(sa)
	if err != nil {
		return err
	}
	return m.syncNewSavedAddress(ctx, &sa, sa.UpdateClock, m.dispatchMessage)
}

func (m *Messenger) DeleteSavedAddress(ctx context.Context, address gethcommon.Address, isTest bool) error {
	updateClock, _ := m.getLastClockWithRelatedChat()
	_, err := m.savedAddressesManager.DeleteSavedAddress(address, isTest, updateClock)
	if err != nil {
		return err
	}
	return m.syncDeletedSavedAddress(ctx, address, isTest, updateClock, m.dispatchMessage)
}

func (m *Messenger) GetSavedAddresses(ctx context.Context) ([]*wallet.SavedAddress, error) {
	return m.savedAddressesManager.GetSavedAddresses()
}

func (m *Messenger) garbageCollectRemovedSavedAddresses() error {
	return m.savedAddressesManager.DeleteSoftRemovedSavedAddresses(uint64(time.Now().AddDate(0, 0, -30).Unix()))
}

func (m *Messenger) dispatchSyncSavedAddress(ctx context.Context, syncMessage *protobuf.SyncSavedAddress, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	clock, chat := m.getLastClockWithRelatedChat()

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_SAVED_ADDRESS,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) syncNewSavedAddress(ctx context.Context, savedAddress *wallet.SavedAddress, updateClock uint64, rawMessageHandler RawMessageHandler) error {
	return m.dispatchSyncSavedAddress(ctx, &protobuf.SyncSavedAddress{
		Address:         savedAddress.Address.Bytes(),
		Name:            savedAddress.Name,
		Removed:         savedAddress.Removed,
		UpdateClock:     updateClock,
		ChainShortNames: savedAddress.ChainShortNames,
		Ens:             savedAddress.ENSName,
		IsTest:          savedAddress.IsTest,
		Color:           string(savedAddress.ColorID),
	}, rawMessageHandler)
}

func (m *Messenger) syncDeletedSavedAddress(ctx context.Context, address gethcommon.Address, isTest bool, updateClock uint64, rawMessageHandler RawMessageHandler) error {
	return m.dispatchSyncSavedAddress(ctx, &protobuf.SyncSavedAddress{
		Address:     address.Bytes(),
		UpdateClock: updateClock,
		Removed:     true,
		IsTest:      isTest,
	}, rawMessageHandler)
}

func (m *Messenger) syncSavedAddress(ctx context.Context, savedAddress *wallet.SavedAddress, rawMessageHandler RawMessageHandler) (err error) {
	if savedAddress.Removed {
		if err = m.syncDeletedSavedAddress(ctx, savedAddress.Address, savedAddress.IsTest, savedAddress.UpdateClock, rawMessageHandler); err != nil {
			return err
		}
	} else {
		if err = m.syncNewSavedAddress(ctx, savedAddress, savedAddress.UpdateClock, rawMessageHandler); err != nil {
			return err
		}
	}
	return
}

func (m *Messenger) HandleSyncSavedAddress(state *ReceivedMessageState, syncMessage *protobuf.SyncSavedAddress, statusMessage *v1protocol.StatusMessage) (err error) {
	address := gethcommon.BytesToAddress(syncMessage.Address)
	if syncMessage.Removed {
		deleted, err := m.savedAddressesManager.DeleteSavedAddress(
			address, syncMessage.IsTest, syncMessage.UpdateClock)
		if err != nil {
			return err
		}
		if deleted {
			state.Response.AddSavedAddress(&wallet.SavedAddress{Address: address, ENSName: syncMessage.Ens, IsTest: syncMessage.IsTest, Removed: true})
		}
	} else {
		sa := wallet.SavedAddress{
			Address:         address,
			Name:            syncMessage.Name,
			ChainShortNames: syncMessage.ChainShortNames,
			ENSName:         syncMessage.Ens,
			IsTest:          syncMessage.IsTest,
			ColorID:         multiAccCommon.CustomizationColor(syncMessage.Color),
		}
		sa.UpdateClock = syncMessage.UpdateClock

		added, err := m.savedAddressesManager.AddSavedAddressIfNewerUpdate(sa)
		if err != nil {
			return err
		}
		if added {
			state.Response.AddSavedAddress(&sa)
		}
	}
	return
}
