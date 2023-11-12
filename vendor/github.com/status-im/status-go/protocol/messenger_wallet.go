package protocol

import (
	"context"
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/accounts"
	walletsettings "github.com/status-im/status-go/multiaccounts/settings_wallet"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/encryption/multidevice"
	"github.com/status-im/status-go/protocol/protobuf"
)

var (
	checkBalancesInterval = time.Minute * 10

	ErrCannotChangeKeypairName = errors.New("cannot change profile keypair name")
)

func (m *Messenger) retrieveWalletBalances() error {
	if m.walletAPI == nil {
		m.logger.Warn("wallet api not enabled")
	}
	accounts, err := m.settings.GetActiveAccounts()
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		m.logger.Info("no accounts to sync wallet balance")
	}

	var ethAccounts []ethcommon.Address

	for _, acc := range accounts {
		m.logger.Info("syncing wallet address", zap.String("account", acc.Address.Hex()))
		ethAccounts = append(ethAccounts, ethcommon.BytesToAddress(acc.Address.Bytes()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	// TODO: publish tokens as a signal
	_, err = m.walletAPI.FetchOrGetCachedWalletBalances(ctx, ethAccounts)
	if err != nil {
		return err
	}

	return nil
}

func (m *Messenger) watchWalletBalances() {
	m.logger.Info("watching wallet balances")

	if m.walletAPI == nil {
		m.logger.Warn("wallet service not enabled")
		return
	}
	go func() {
		for {
			select {
			case <-time.After(checkBalancesInterval):

				err := m.retrieveWalletBalances()
				if err != nil {
					m.logger.Error("failed to retrieve wallet balances", zap.Error(err))
				}
			case <-m.quit:
				return
			}
		}
	}()
}

func (m *Messenger) UpdateKeypairName(keyUID string, name string) error {
	if keyUID == m.account.KeyUID && name != m.account.Name {
		// profile keypair name must always follow profile display name
		return ErrCannotChangeKeypairName
	}
	clock, _ := m.getLastClockWithRelatedChat()
	err := m.settings.UpdateKeypairName(keyUID, name, clock, keyUID == m.account.KeyUID)
	if err != nil {
		return err
	}

	return m.resolveAndSyncKeypairOrJustWalletAccount(keyUID, types.Address{}, clock, m.dispatchMessage)
}

func (m *Messenger) MoveWalletAccount(fromPosition int64, toPosition int64) error {
	clock, _ := m.getLastClockWithRelatedChat()

	err := m.settings.MoveWalletAccount(fromPosition, toPosition, clock)
	if err != nil {
		return err
	}

	return m.syncAccountsPositions(m.dispatchMessage)
}

func (m *Messenger) resolveAndSetAccountPropsMaintainedByBackend(acc *accounts.Account) error {
	// Account position is fully maintained by the backend, no need client to set it explicitly.
	// To support DragAndDrop feature for accounts there is exposed `MoveWalletAccount` which
	// moves an account to the passed position.
	//
	// Account operability is fully maintained by the backend, for new accounts created on this device
	// it is always set to fully operable, while for accounts received by syncing process or fetched from waku
	// is set by logic placed in `resolveAccountOperability` function.
	//
	// TODO: making not or partially operable accounts fully operable will be added later, but for sure it will
	// be handled by the backend only, no need client to set it explicitly.

	dbAccount, err := m.settings.GetAccountByAddress(acc.Address)
	if err != nil && err != accounts.ErrDbAccountNotFound {
		return err
	}
	if dbAccount != nil {
		acc.Position = dbAccount.Position
		acc.Operable = dbAccount.Operable
	} else {
		pos, err := m.settings.GetPositionForNextNewAccount()
		if err != nil {
			return err
		}
		acc.Position = pos
		acc.Operable = accounts.AccountFullyOperable
	}
	return nil
}

func (m *Messenger) SaveOrUpdateKeypair(keypair *accounts.Keypair) error {
	if keypair.KeyUID == m.account.KeyUID && keypair.Name != m.account.Name {
		// profile keypair name must always follow profile display name
		return ErrCannotChangeKeypairName
	}
	clock, _ := m.getLastClockWithRelatedChat()
	keypair.Clock = clock

	for _, acc := range keypair.Accounts {
		acc.Clock = clock
		err := m.resolveAndSetAccountPropsMaintainedByBackend(acc)
		if err != nil {
			return err
		}
	}

	err := m.settings.SaveOrUpdateKeypair(keypair)
	if err != nil {
		return err
	}

	return m.resolveAndSyncKeypairOrJustWalletAccount(keypair.KeyUID, types.Address{}, keypair.Clock, m.dispatchMessage)
}

func (m *Messenger) SaveOrUpdateAccount(acc *accounts.Account) error {
	clock, _ := m.getLastClockWithRelatedChat()
	acc.Clock = clock

	err := m.resolveAndSetAccountPropsMaintainedByBackend(acc)
	if err != nil {
		return err
	}

	err = m.settings.SaveOrUpdateAccounts([]*accounts.Account{acc}, true)
	if err != nil {
		return err
	}

	err = m.resolveAndSyncKeypairOrJustWalletAccount(acc.KeyUID, acc.Address, acc.Clock, m.dispatchMessage)
	if err != nil {
		return err
	}

	return m.UpdateProfileShowcaseWalletAccount(acc)
}

func (m *Messenger) MarkKeypairFullyOperable(keyUID string) error {
	clock, _ := m.getLastClockWithRelatedChat()

	err := m.settings.MarkKeypairFullyOperable(keyUID, clock, true)
	if err != nil {
		return err
	}

	return m.resolveAndSyncKeypairOrJustWalletAccount(keyUID, types.Address{}, clock, m.dispatchMessage)
}

func (m *Messenger) deleteKeystoreFileForAddress(address types.Address) error {
	acc, err := m.settings.GetAccountByAddress(address)
	if err != nil {
		return err
	}

	if acc.Operable == accounts.AccountNonOperable || acc.Operable == accounts.AccountPartiallyOperable {
		return nil
	}

	if acc.Type != accounts.AccountTypeWatch {
		kp, err := m.settings.GetKeypairByKeyUID(acc.KeyUID)
		if err != nil {
			return err
		}

		if !kp.MigratedToKeycard() {
			err = m.accountsManager.DeleteAccount(address)
			var e *account.ErrCannotLocateKeyFile
			if err != nil && !errors.As(err, &e) {
				return err
			}

			if acc.Type != accounts.AccountTypeKey {
				lastAcccountOfKeypairWithTheSameKey := len(kp.Accounts) == 1
				if lastAcccountOfKeypairWithTheSameKey {
					err = m.accountsManager.DeleteAccount(types.Address(ethcommon.HexToAddress(kp.DerivedFrom)))
					var e *account.ErrCannotLocateKeyFile
					if err != nil && !errors.As(err, &e) {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (m *Messenger) deleteKeystoreFilesForKeypair(keypair *accounts.Keypair) (err error) {
	if m.accountsManager == nil || keypair == nil || keypair.MigratedToKeycard() {
		return
	}

	anyAccountFullyOrPartiallyOperable := false
	for _, acc := range keypair.Accounts {
		if acc.Removed || acc.Operable == accounts.AccountNonOperable {
			continue
		}
		if !anyAccountFullyOrPartiallyOperable {
			anyAccountFullyOrPartiallyOperable = true
		}
		if acc.Operable == accounts.AccountPartiallyOperable {
			continue
		}
		err = m.accountsManager.DeleteAccount(acc.Address)
		var e *account.ErrCannotLocateKeyFile
		if err != nil && !errors.As(err, &e) {
			return err
		}
	}

	if anyAccountFullyOrPartiallyOperable && keypair.Type != accounts.KeypairTypeKey {
		err = m.accountsManager.DeleteAccount(types.Address(ethcommon.HexToAddress(keypair.DerivedFrom)))
		var e *account.ErrCannotLocateKeyFile
		if err != nil && !errors.As(err, &e) {
			return err
		}
	}

	return
}

func (m *Messenger) DeleteAccount(address types.Address) error {
	acc, err := m.settings.GetAccountByAddress(address)
	if err != nil {
		return err
	}

	if acc.Chat {
		return accounts.ErrCannotRemoveProfileAccount
	}

	if acc.Wallet {
		return accounts.ErrCannotRemoveDefaultWalletAccount
	}

	err = m.deleteKeystoreFileForAddress(address)
	if err != nil {
		return err
	}

	clock, _ := m.getLastClockWithRelatedChat()

	err = m.settings.RemoveAccount(address, clock)
	if err != nil {
		return err
	}

	err = m.resolveAndSyncKeypairOrJustWalletAccount(acc.KeyUID, acc.Address, clock, m.dispatchMessage)
	if err != nil {
		return err
	}

	// In case when user deletes an account, we need to send sync message after an account gets deleted,
	// and then (after that) update the positions of other accoutns. That's needed to handle properly
	// accounts order on the paired devices.
	err = m.settings.ResolveAccountsPositions(clock)
	if err != nil {
		return err
	}
	// Since some keypairs may be received out of expected order, we're aligning that by sending accounts position sync msg.
	err = m.syncAccountsPositions(m.dispatchMessage)
	if err != nil {
		return err
	}

	return m.DeleteProfileShowcaseWalletAccount(acc)
}

func (m *Messenger) DeleteKeypair(keyUID string) error {
	kp, err := m.settings.GetKeypairByKeyUID(keyUID)
	if err != nil {
		return err
	}

	if kp.Type == accounts.KeypairTypeProfile {
		return accounts.ErrCannotRemoveProfileKeypair
	}

	err = m.deleteKeystoreFilesForKeypair(kp)
	if err != nil {
		return err
	}

	clock, _ := m.getLastClockWithRelatedChat()

	err = m.settings.RemoveKeypair(keyUID, clock)
	if err != nil {
		return err
	}

	err = m.resolveAndSyncKeypairOrJustWalletAccount(kp.KeyUID, types.Address{}, clock, m.dispatchMessage)
	if err != nil {
		return err
	}

	// In case when user deletes entire keypair, we need to send sync message after a keypair gets deleted,
	// and then (after that) update the positions of other accoutns. That's needed to handle properly
	// accounts order on the paired devices.
	err = m.settings.ResolveAccountsPositions(clock)
	if err != nil {
		return err
	}
	// Since some keypairs may be received out of expected order, we're aligning that by sending accounts position sync msg.
	return m.syncAccountsPositions(m.dispatchMessage)
}

func (m *Messenger) prepareSyncAccountMessage(acc *accounts.Account) *protobuf.SyncAccount {
	return &protobuf.SyncAccount{
		Clock:                 acc.Clock,
		Address:               acc.Address.Bytes(),
		KeyUid:                acc.KeyUID,
		PublicKey:             acc.PublicKey,
		Path:                  acc.Path,
		Name:                  acc.Name,
		ColorId:               string(acc.ColorID),
		Emoji:                 acc.Emoji,
		Wallet:                acc.Wallet,
		Chat:                  acc.Chat,
		Hidden:                acc.Hidden,
		Removed:               acc.Removed,
		Operable:              acc.Operable.String(),
		Position:              acc.Position,
		ProdPreferredChainIDs: acc.ProdPreferredChainIDs,
		TestPreferredChainIDs: acc.TestPreferredChainIDs,
	}
}

func (m *Messenger) getMyInstallationMetadata() (*multidevice.InstallationMetadata, error) {
	installation, ok := m.allInstallations.Load(m.installationID)
	if !ok {
		return nil, errors.New("no installation found")
	}

	if installation.InstallationMetadata == nil {
		return nil, errors.New("no installation metadata")
	}

	return installation.InstallationMetadata, nil
}

func (m *Messenger) prepareSyncKeypairMessage(kp *accounts.Keypair) (*protobuf.SyncKeypair, error) {
	message := &protobuf.SyncKeypair{
		Clock:                   kp.Clock,
		KeyUid:                  kp.KeyUID,
		Name:                    kp.Name,
		Type:                    kp.Type.String(),
		DerivedFrom:             kp.DerivedFrom,
		LastUsedDerivationIndex: kp.LastUsedDerivationIndex,
		SyncedFrom:              kp.SyncedFrom,
		Removed:                 kp.Removed,
	}

	if kp.SyncedFrom == "" {
		installationMetadata, err := m.getMyInstallationMetadata()
		if err != nil {
			return nil, err
		}
		message.SyncedFrom = installationMetadata.Name
	}

	for _, acc := range kp.Accounts {
		sAcc := m.prepareSyncAccountMessage(acc)
		if sAcc == nil {
			continue
		}

		message.Accounts = append(message.Accounts, sAcc)
	}

	syncKcMsgs, err := m.prepareSyncKeycardsMessage(kp.KeyUID)
	if err != nil {
		return nil, err
	}
	message.Keycards = syncKcMsgs

	if m.walletAPI != nil {
		message.KeycardPairings, err = m.walletAPI.GetPairingsJSONFileContent()
		if err != nil {
			return nil, err
		}
	}

	return message, nil
}

func (m *Messenger) UpdateTokenPreferences(preferences []walletsettings.TokenPreferences) error {
	clock, _ := m.getLastClockWithRelatedChat()
	testNetworksEnabled, err := m.settings.GetTestNetworksEnabled()
	if err != nil {
		return err
	}

	groupByCommunity, err := m.settings.GetTokenGroupByCommunity()
	if err != nil {
		return err
	}

	err = m.settings.UpdateTokenPreferences(preferences, groupByCommunity, testNetworksEnabled, clock)
	if err != nil {
		return err
	}

	return m.syncTokenPreferences(m.dispatchMessage)
}

func (m *Messenger) GetTokenPreferences() ([]walletsettings.TokenPreferences, error) {
	testNetworksEnabled, err := m.settings.GetTestNetworksEnabled()
	if err != nil {
		return nil, err
	}

	list, err := m.settings.GetTokenPreferences(testNetworksEnabled)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (m *Messenger) prepareTokenPreferencesMessage(pref walletsettings.TokenPreferences) *protobuf.TokenPreferences {
	return &protobuf.TokenPreferences{
		Key:           pref.Key,
		Position:      int64(pref.Position),
		GroupPosition: int64(pref.GroupPosition),
		Visible:       pref.Visible,
		CommunityId:   pref.CommunityID,
	}
}

func (m *Messenger) syncTokenPreferences(rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, chat := m.getLastClockWithRelatedChat()

	lastUpdate, err := m.settings.GetClockOfLastTokenPreferencesChange()
	if err != nil {
		return err
	}

	testNetworksEnabled, err := m.settings.GetTestNetworksEnabled()
	if err != nil {
		return err
	}

	preferences, err := m.GetTokenPreferences()
	if err != nil {
		return err
	}

	message := &protobuf.SyncTokenPreferences{
		Clock:   lastUpdate,
		Testnet: testNetworksEnabled,
	}

	for _, pref := range preferences {
		message.Preferences = append(message.Preferences, m.prepareTokenPreferencesMessage(pref))
	}

	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_TOKEN_PREFERENCES,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	return err
}

func (m *Messenger) UpdateCollectiblePreferences(preferences []walletsettings.CollectiblePreferences) error {
	clock, _ := m.getLastClockWithRelatedChat()
	testNetworksEnabled, err := m.settings.GetTestNetworksEnabled()
	if err != nil {
		return err
	}

	groupByCommunity, err := m.settings.GetCollectibleGroupByCommunity()
	if err != nil {
		return err
	}

	groupByCollection, err := m.settings.GetCollectibleGroupByCollection()
	if err != nil {
		return err
	}

	err = m.settings.UpdateCollectiblePreferences(preferences, groupByCommunity, groupByCollection, testNetworksEnabled, clock)
	if err != nil {
		return err
	}

	return m.syncCollectiblePreferences(m.dispatchMessage)
}

func (m *Messenger) GetCollectiblePreferences() ([]walletsettings.CollectiblePreferences, error) {
	testNetworksEnabled, err := m.settings.GetTestNetworksEnabled()
	if err != nil {
		return nil, err
	}

	list, err := m.settings.GetCollectiblePreferences(testNetworksEnabled)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (m *Messenger) prepareCollectiblePreferencesMessage(pref walletsettings.CollectiblePreferences) *protobuf.CollectiblePreferences {
	return &protobuf.CollectiblePreferences{
		Type:     int64(pref.Type),
		Key:      pref.Key,
		Position: int64(pref.Position),
		Visible:  pref.Visible,
	}
}

func (m *Messenger) syncCollectiblePreferences(rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, chat := m.getLastClockWithRelatedChat()

	lastUpdate, err := m.settings.GetClockOfLastCollectiblePreferencesChange()
	if err != nil {
		return err
	}

	testNetworksEnabled, err := m.settings.GetTestNetworksEnabled()
	if err != nil {
		return err
	}

	preferences, err := m.GetCollectiblePreferences()
	if err != nil {
		return err
	}

	message := &protobuf.SyncCollectiblePreferences{
		Clock:   lastUpdate,
		Testnet: testNetworksEnabled,
	}

	for _, pref := range preferences {
		message.Preferences = append(message.Preferences, m.prepareCollectiblePreferencesMessage(pref))
	}

	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_COLLECTIBLE_PREFERENCES,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	return err
}

func (m *Messenger) syncAccountsPositions(rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, chat := m.getLastClockWithRelatedChat()

	allDbAccounts, err := m.settings.GetActiveAccounts()
	if err != nil {
		return err
	}

	lastUpdate, err := m.settings.GetClockOfLastAccountsPositionChange()
	if err != nil {
		return err
	}

	message := &protobuf.SyncAccountsPositions{
		Clock: lastUpdate,
	}

	for _, acc := range allDbAccounts {
		if acc.Chat {
			continue
		}
		message.Accounts = append(message.Accounts, m.prepareSyncAccountMessage(acc))
	}

	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACCOUNTS_POSITIONS,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	return err
}

func (m *Messenger) syncWalletAccount(acc *accounts.Account, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, chat := m.getLastClockWithRelatedChat()

	message := m.prepareSyncAccountMessage(acc)

	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACCOUNT,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	return err
}

func (m *Messenger) syncKeypair(keypair *accounts.Keypair, rawMessageHandler RawMessageHandler) (err error) {
	if !m.hasPairedDevices() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, chat := m.getLastClockWithRelatedChat()
	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		ResendAutomatically: true,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_KEYPAIR,
	}

	message, err := m.prepareSyncKeypairMessage(keypair)
	if err != nil {
		return err
	}

	rawMessage.Payload, err = proto.Marshal(message)
	if err != nil {
		return err
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	return err
}

// This function resolves which protobuf message needs to be sent.
//
// If `KeyUID` is empty (means it's a watch only account) we send `protobuf.SyncAccount` message
// otherwise means the account belong to a keypai, hence we send `protobuf.SyncKeypair` message
func (m *Messenger) resolveAndSyncKeypairOrJustWalletAccount(keyUID string, address types.Address, clock uint64, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	if keyUID == "" {
		var dbAccount *accounts.Account
		allDbAccounts, err := m.settings.GetAllAccounts() // removed accounts included
		if err != nil {
			return err
		}

		for _, acc := range allDbAccounts {
			if acc.Address == address {
				dbAccount = acc
				break
			}
		}

		if dbAccount == nil {
			return accounts.ErrDbAccountNotFound
		}

		err = m.syncWalletAccount(dbAccount, rawMessageHandler)
		if err != nil {
			return err
		}
	} else {
		var dbKeypair *accounts.Keypair
		allDbKeypairs, err := m.settings.GetAllKeypairs() // removed keypairs included
		if err != nil {
			return err
		}

		for _, kp := range allDbKeypairs {
			if kp.KeyUID == keyUID {
				dbKeypair = kp
				break
			}
		}

		if dbKeypair == nil {
			return accounts.ErrDbKeypairNotFound
		}

		err = m.syncKeypair(dbKeypair, rawMessageHandler)
		if err != nil {
			return err
		}
	}

	_, chat := m.getLastClockWithRelatedChat()
	chat.LastClockValue = clock
	return m.saveChat(chat)
}
