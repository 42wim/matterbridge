package protocol

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

const (
	BackupContactsPerBatch = 20
)

// backupTickerInterval is how often we should check for backups
var backupTickerInterval = 120 * time.Second

// backupIntervalSeconds is the amount of seconds we should allow between
// backups
var backupIntervalSeconds uint64 = 28800

func (m *Messenger) backupEnabled() (bool, error) {
	return m.settings.BackupEnabled()
}

func (m *Messenger) lastBackup() (uint64, error) {
	return m.settings.LastBackup()
}

func (m *Messenger) startBackupLoop() {
	ticker := time.NewTicker(backupTickerInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if !m.Online() {
					continue
				}

				enabled, err := m.backupEnabled()
				if err != nil {
					m.logger.Error("failed to fetch backup enabled")
					continue
				}
				if !enabled {
					m.logger.Debug("backup not enabled, skipping")
					continue
				}

				lastBackup, err := m.lastBackup()
				if err != nil {
					m.logger.Error("failed to fetch last backup time")
					continue
				}

				now := time.Now().Unix()
				if uint64(now) <= backupIntervalSeconds+lastBackup {
					m.logger.Debug("not backing up")
					continue
				}
				m.logger.Debug("backing up data")

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				_, err = m.BackupData(ctx)
				if err != nil {
					m.logger.Error("failed to backup data", zap.Error(err))
				}
			case <-m.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *Messenger) BackupData(ctx context.Context) (uint64, error) {
	clock, chat := m.getLastClockWithRelatedChat()
	contactsToBackup := m.backupContacts(ctx)
	communitiesToBackup, err := m.backupCommunities(ctx, clock)
	if err != nil {
		return 0, err
	}
	chatsToBackup := m.backupChats(ctx, clock)
	if err != nil {
		return 0, err
	}
	profileToBackup, err := m.backupProfile(ctx, clock)
	if err != nil {
		return 0, err
	}
	_, settings, errors := m.prepareSyncSettingsMessages(clock, true)
	if len(errors) != 0 {
		// return just the first error, the others have been logged
		return 0, errors[0]
	}

	keypairsToBackup, err := m.backupKeypairs()
	if err != nil {
		return 0, err
	}

	woAccountsToBackup, err := m.backupWatchOnlyAccounts()
	if err != nil {
		return 0, err
	}

	backupDetailsOnly := func() *protobuf.Backup {
		return &protobuf.Backup{
			Clock: clock,
			ChatsDetails: &protobuf.FetchingBackedUpDataDetails{
				DataNumber:  uint32(0),
				TotalNumber: uint32(len(chatsToBackup)),
			},
			ContactsDetails: &protobuf.FetchingBackedUpDataDetails{
				DataNumber:  uint32(0),
				TotalNumber: uint32(len(contactsToBackup)),
			},
			CommunitiesDetails: &protobuf.FetchingBackedUpDataDetails{
				DataNumber:  uint32(0),
				TotalNumber: uint32(len(communitiesToBackup)),
			},
			ProfileDetails: &protobuf.FetchingBackedUpDataDetails{
				DataNumber:  uint32(0),
				TotalNumber: uint32(len(profileToBackup)),
			},
			SettingsDetails: &protobuf.FetchingBackedUpDataDetails{
				DataNumber:  uint32(0),
				TotalNumber: uint32(len(settings)),
			},
			KeypairDetails: &protobuf.FetchingBackedUpDataDetails{
				DataNumber:  uint32(0),
				TotalNumber: uint32(len(keypairsToBackup)),
			},
			WatchOnlyAccountDetails: &protobuf.FetchingBackedUpDataDetails{
				DataNumber:  uint32(0),
				TotalNumber: uint32(len(woAccountsToBackup)),
			},
		}
	}

	// Update contacts messages encode and dispatch
	for i, d := range contactsToBackup {
		pb := backupDetailsOnly()
		pb.ContactsDetails.DataNumber = uint32(i + 1)
		pb.Contacts = d.Contacts
		err = m.encodeAndDispatchBackupMessage(ctx, pb, chat.ID)
		if err != nil {
			return 0, err
		}
	}

	// Update communities messages encode and dispatch
	for i, d := range communitiesToBackup {
		pb := backupDetailsOnly()
		pb.CommunitiesDetails.DataNumber = uint32(i + 1)
		pb.Communities = d.Communities
		err = m.encodeAndDispatchBackupMessage(ctx, pb, chat.ID)
		if err != nil {
			return 0, err
		}
	}

	// Update profile messages encode and dispatch
	for i, d := range profileToBackup {
		pb := backupDetailsOnly()
		pb.ProfileDetails.DataNumber = uint32(i + 1)
		pb.Profile = d.Profile
		err = m.encodeAndDispatchBackupMessage(ctx, pb, chat.ID)
		if err != nil {
			return 0, err
		}
	}

	// Update chats encode and dispatch
	for i, d := range chatsToBackup {
		pb := backupDetailsOnly()
		pb.ChatsDetails.DataNumber = uint32(i + 1)
		pb.Chats = d.Chats
		err = m.encodeAndDispatchBackupMessage(ctx, pb, chat.ID)
		if err != nil {
			return 0, err
		}
	}

	// Update settings messages encode and dispatch
	for i, d := range settings {
		pb := backupDetailsOnly()
		pb.SettingsDetails.DataNumber = uint32(i + 1)
		pb.Setting = d
		err = m.encodeAndDispatchBackupMessage(ctx, pb, chat.ID)
		if err != nil {
			return 0, err
		}
	}

	// Update keypairs messages encode and dispatch
	for i, d := range keypairsToBackup {
		pb := backupDetailsOnly()
		pb.KeypairDetails.DataNumber = uint32(i + 1)
		pb.Keypair = d.Keypair
		err = m.encodeAndDispatchBackupMessage(ctx, pb, chat.ID)
		if err != nil {
			return 0, err
		}
	}

	// Update watch only messages encode and dispatch
	for i, d := range woAccountsToBackup {
		pb := backupDetailsOnly()
		pb.WatchOnlyAccountDetails.DataNumber = uint32(i + 1)
		pb.WatchOnlyAccount = d.WatchOnlyAccount
		err = m.encodeAndDispatchBackupMessage(ctx, pb, chat.ID)
		if err != nil {
			return 0, err
		}
	}

	chat.LastClockValue = clock
	err = m.saveChat(chat)
	if err != nil {
		return 0, err
	}

	clockInSeconds := clock / 1000
	err = m.settings.SetLastBackup(clockInSeconds)
	if err != nil {
		return 0, err
	}
	if m.config.messengerSignalsHandler != nil {
		m.config.messengerSignalsHandler.BackupPerformed(clockInSeconds)
	}

	return clockInSeconds, nil
}

func (m *Messenger) encodeAndDispatchBackupMessage(ctx context.Context, message *protobuf.Backup, chatID string) error {
	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chatID,
		Payload:             encodedMessage,
		SkipEncryptionLayer: true,
		SendOnPersonalTopic: true,
		MessageType:         protobuf.ApplicationMetadataMessage_BACKUP,
	})

	return err
}

func (m *Messenger) backupContacts(ctx context.Context) []*protobuf.Backup {
	var contacts []*protobuf.SyncInstallationContactV2
	m.allContacts.Range(func(contactID string, contact *Contact) (shouldContinue bool) {
		syncContact := m.buildSyncContactMessage(contact)
		if syncContact != nil {
			contacts = append(contacts, syncContact)
		}
		return true
	})

	var backupMessages []*protobuf.Backup
	for i := 0; i < len(contacts); i += BackupContactsPerBatch {
		j := i + BackupContactsPerBatch
		if j > len(contacts) {
			j = len(contacts)
		}

		contactsToAdd := contacts[i:j]

		backupMessage := &protobuf.Backup{
			Contacts: contactsToAdd,
		}
		backupMessages = append(backupMessages, backupMessage)
	}

	return backupMessages
}

func (m *Messenger) backupCommunities(ctx context.Context, clock uint64) ([]*protobuf.Backup, error) {
	joinedCs, err := m.communitiesManager.JoinedAndPendingCommunitiesWithRequests()
	if err != nil {
		return nil, err
	}

	deletedCs, err := m.communitiesManager.DeletedCommunities()
	if err != nil {
		return nil, err
	}

	var backupMessages []*protobuf.Backup
	cs := append(joinedCs, deletedCs...)
	for _, c := range cs {
		_, beingImported := m.importingCommunities[c.IDString()]
		if !beingImported {
			settings, err := m.communitiesManager.GetCommunitySettingsByID(c.ID())
			if err != nil {
				return nil, err
			}

			syncControlNode, err := m.communitiesManager.GetSyncControlNode(c.ID())
			if err != nil {
				return nil, err
			}

			syncMessage, err := c.ToSyncInstallationCommunityProtobuf(clock, settings, syncControlNode)
			if err != nil {
				return nil, err
			}

			encodedKeys, err := m.encryptor.GetAllHREncodedKeys(c.ID())
			if err != nil {
				return nil, err
			}
			syncMessage.EncryptionKeys = encodedKeys

			backupMessage := &protobuf.Backup{
				Communities: []*protobuf.SyncInstallationCommunity{syncMessage},
			}

			backupMessages = append(backupMessages, backupMessage)
		}
	}

	return backupMessages, nil
}

func (m *Messenger) backupChats(ctx context.Context, clock uint64) []*protobuf.Backup {
	var oneToOneAndGroupChats []*protobuf.SyncChat
	m.allChats.Range(func(chatID string, chat *Chat) bool {
		if !chat.OneToOne() && !chat.PrivateGroupChat() {
			return true
		}
		syncChat := protobuf.SyncChat{
			Clock:    clock,
			Id:       chatID,
			ChatType: uint32(chat.ChatType),
			Active:   chat.Active,
		}
		chatMuteTill, _ := time.Parse(time.RFC3339, chat.MuteTill.Format(time.RFC3339))
		if chat.Muted && chatMuteTill.Equal(time.Time{}) {
			// Only set Muted if it is "permanently" muted
			syncChat.Muted = true
		}
		if chat.PrivateGroupChat() {
			syncChat.Name = chat.Name // The Name is only useful in the case of a group chat

			syncChat.MembershipUpdateEvents = make([]*protobuf.MembershipUpdateEvents, len(chat.MembershipUpdates))
			for i, membershipUpdate := range chat.MembershipUpdates {
				syncChat.MembershipUpdateEvents[i] = &protobuf.MembershipUpdateEvents{
					Clock:      membershipUpdate.ClockValue,
					Type:       uint32(membershipUpdate.Type),
					Members:    membershipUpdate.Members,
					Name:       membershipUpdate.Name,
					Signature:  membershipUpdate.Signature,
					ChatId:     membershipUpdate.ChatID,
					From:       membershipUpdate.From,
					RawPayload: membershipUpdate.RawPayload,
					Color:      membershipUpdate.Color,
				}
			}
		}
		oneToOneAndGroupChats = append(oneToOneAndGroupChats, &syncChat)
		return true
	})

	var backupMessages []*protobuf.Backup
	backupMessage := &protobuf.Backup{
		Chats: oneToOneAndGroupChats,
	}
	backupMessages = append(backupMessages, backupMessage)
	return backupMessages
}

func (m *Messenger) buildSyncContactMessage(contact *Contact) *protobuf.SyncInstallationContactV2 {
	var ensName string
	if contact.ENSVerified {
		ensName = contact.EnsName
	}

	oneToOneChat, ok := m.allChats.Load(contact.ID)
	muted := false
	if ok {
		muted = oneToOneChat.Muted
	}

	return &protobuf.SyncInstallationContactV2{
		LastUpdatedLocally:        contact.LastUpdatedLocally,
		LastUpdated:               contact.LastUpdated,
		Id:                        contact.ID,
		DisplayName:               contact.DisplayName,
		EnsName:                   ensName,
		LocalNickname:             contact.LocalNickname,
		Added:                     contact.added(),
		Blocked:                   contact.Blocked,
		Muted:                     muted,
		HasAddedUs:                contact.hasAddedUs(),
		Removed:                   contact.Removed,
		ContactRequestLocalState:  int64(contact.ContactRequestLocalState),
		ContactRequestRemoteState: int64(contact.ContactRequestRemoteState),
		ContactRequestRemoteClock: int64(contact.ContactRequestRemoteClock),
		ContactRequestLocalClock:  int64(contact.ContactRequestLocalClock),
		VerificationStatus:        int64(contact.VerificationStatus),
		TrustStatus:               int64(contact.TrustStatus),
	}
}

func (m *Messenger) backupProfile(ctx context.Context, clock uint64) ([]*protobuf.Backup, error) {
	displayName, err := m.settings.DisplayName()
	if err != nil {
		return nil, err
	}

	displayNameClock, err := m.settings.GetSettingLastSynced(settings.DisplayName)
	if err != nil {
		return nil, err
	}

	keyUID := m.account.KeyUID
	images, err := m.multiAccounts.GetIdentityImages(keyUID)
	if err != nil {
		return nil, err
	}

	pictureProtos := make([]*protobuf.SyncProfilePicture, len(images))
	for i, image := range images {
		p := &protobuf.SyncProfilePicture{}
		p.Name = image.Name
		p.Payload = image.Payload
		p.Width = uint32(image.Width)
		p.Height = uint32(image.Height)
		p.FileSize = uint32(image.FileSize)
		p.ResizeTarget = uint32(image.ResizeTarget)
		if image.Clock == 0 {
			p.Clock = clock
		} else {
			p.Clock = image.Clock
		}
		pictureProtos[i] = p
	}

	socialLinks, err := m.settings.GetSocialLinks()
	if err != nil {
		return nil, err
	}

	socialLinksClock, err := m.settings.GetSocialLinksClock()
	if err != nil {
		return nil, err
	}

	syncSocialLinks := &protobuf.SyncSocialLinks{
		SocialLinks: socialLinks.ToProtobuf(),
		Clock:       socialLinksClock,
	}

	ensUsernameDetails, err := m.getEnsUsernameDetails()
	if err != nil {
		return nil, err
	}
	ensUsernameDetailProtos := make([]*protobuf.SyncEnsUsernameDetail, len(ensUsernameDetails))
	for i, ensUsernameDetail := range ensUsernameDetails {
		ensUsernameDetailProtos[i] = &protobuf.SyncEnsUsernameDetail{
			Username: ensUsernameDetail.Username,
			Clock:    ensUsernameDetail.Clock,
			Removed:  ensUsernameDetail.Removed,
			ChainId:  ensUsernameDetail.ChainID,
		}
	}

	backupMessage := &protobuf.Backup{
		Profile: &protobuf.BackedUpProfile{
			KeyUid:             keyUID,
			DisplayName:        displayName,
			Pictures:           pictureProtos,
			DisplayNameClock:   displayNameClock,
			SocialLinks:        syncSocialLinks,
			EnsUsernameDetails: ensUsernameDetailProtos,
		},
	}

	backupMessages := []*protobuf.Backup{backupMessage}

	return backupMessages, nil
}

func (m *Messenger) backupKeypairs() ([]*protobuf.Backup, error) {
	keypairs, err := m.settings.GetActiveKeypairs()
	if err != nil {
		return nil, err
	}

	var backupMessages []*protobuf.Backup
	for _, kp := range keypairs {

		kp.SyncedFrom = accounts.SyncedFromBackup
		keypair, err := m.prepareSyncKeypairMessage(kp)
		if err != nil {
			return nil, err
		}

		backupMessage := &protobuf.Backup{
			Keypair: keypair,
		}

		backupMessages = append(backupMessages, backupMessage)
	}

	return backupMessages, nil
}

func (m *Messenger) backupWatchOnlyAccounts() ([]*protobuf.Backup, error) {
	accounts, err := m.settings.GetAllWatchOnlyAccounts()
	if err != nil {
		return nil, err
	}

	var backupMessages []*protobuf.Backup
	for _, acc := range accounts {

		backupMessage := &protobuf.Backup{}
		backupMessage.WatchOnlyAccount = m.prepareSyncAccountMessage(acc)

		backupMessages = append(backupMessages, backupMessage)
	}

	return backupMessages, nil
}
