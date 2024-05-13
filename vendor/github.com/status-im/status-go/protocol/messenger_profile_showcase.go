package protocol

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"errors"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	eth_common "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/identity"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/services/wallet/bigint"
	w_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

var errorDecryptingPayloadEncryptionKey = errors.New("decrypting the payload encryption key resulted in no error and a nil key")
var errorConvertCollectibleTokenIDToInt = errors.New("failed to convert collectible token id to bigint")
var errorNoAccountPresentedForCollectible = errors.New("account holding the collectible is not presented in the profile showcase")
var errorDublicateAccountAddress = errors.New("duplicate account address")
var errorAccountVisibilityLowerThanCollectible = errors.New("account visibility lower than collectible")

func sortProfileEntyByOrder(slice interface{}, getOrder func(int) int) {
	sort.Slice(slice, func(i, j int) bool {
		return getOrder(j) > getOrder(i)
	})
}

func toCollectibleUniqueID(contractAddress string, tokenID string, chainID uint64) (thirdparty.CollectibleUniqueID, error) {
	tokenIDInt := new(big.Int)
	tokenIDInt, isTokenIDOk := tokenIDInt.SetString(tokenID, 10)
	if !isTokenIDOk {
		return thirdparty.CollectibleUniqueID{}, errorConvertCollectibleTokenIDToInt
	}

	return thirdparty.CollectibleUniqueID{
		ContractID: thirdparty.ContractID{
			ChainID: w_common.ChainID(chainID),
			Address: eth_common.HexToAddress(contractAddress),
		},
		TokenID: &bigint.BigInt{Int: tokenIDInt},
	}, nil
}

func (m *Messenger) fetchCollectibleOwner(contractAddress string, tokenID string, chainID uint64) ([]thirdparty.AccountBalance, error) {
	collectibleID, err := toCollectibleUniqueID(contractAddress, tokenID, chainID)
	if err != nil {
		return nil, err
	}

	balance, err := m.communitiesManager.GetCollectiblesManager().GetCollectibleOwnership(collectibleID)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (m *Messenger) validateCollectiblesOwnership(accounts []*identity.ProfileShowcaseAccountPreference,
	collectibles []*identity.ProfileShowcaseCollectiblePreference) error {
	accountsMap := make(map[string]identity.ProfileShowcaseVisibility)

	for _, accountProfile := range accounts {
		addressCapitalized := strings.ToUpper(accountProfile.Address)
		if _, ok := accountsMap[addressCapitalized]; ok {
			return errorDublicateAccountAddress
		}
		accountsMap[addressCapitalized] = accountProfile.ShowcaseVisibility
	}

	for _, collectibleProfile := range collectibles {
		balances, err := m.fetchCollectibleOwner(collectibleProfile.ContractAddress, collectibleProfile.TokenID,
			collectibleProfile.ChainID)
		if err != nil {
			return err
		}

		// NOTE: ERC721 tokens can have only a single holder
		// but ERC1155 which can be supported later can have more than one holder and balances > 1
		found := false
		for _, balance := range balances {
			addressCapitalized := strings.ToUpper(balance.Address.String())
			if accountShowcaseVisibility, ok := accountsMap[addressCapitalized]; ok {
				if accountShowcaseVisibility < collectibleProfile.ShowcaseVisibility {
					return errorAccountVisibilityLowerThanCollectible
				}
				found = true
				break
			}
		}
		if !found {
			return errorNoAccountPresentedForCollectible
		}
	}

	return nil
}

func (m *Messenger) validateCommunityMembershipEntry(
	entry *identity.ProfileShowcaseCommunity,
	community *communities.Community,
	contactPubKey *ecdsa.PublicKey) (identity.ProfileShowcaseMembershipStatus, error) {
	if community == nil {
		return identity.ProfileShowcaseMembershipStatusUnproven, nil
	}

	if community.Encrypted() {
		// NOTE: commentend for 0.177.x release, actual fix is here:
		// https://github.com/status-im/status-go/pull/5024
		return identity.ProfileShowcaseMembershipStatusProvenMember, nil
		// grant, err := community.VerifyGrantSignature(entry.Grant)
		// if err != nil {
		// 	m.logger.Warn("failed to verify grant signature ", zap.Error(err))
		// 	return identity.ProfileShowcaseMembershipStatusNotAMember, nil
		// }

		// if grant != nil && bytes.Equal(grant.MemberId, crypto.CompressPubkey(contactPubKey)) {
		// 	return identity.ProfileShowcaseMembershipStatusProvenMember, nil
		// }
		// // Show as not a member if membership can't be proven
		// return identity.ProfileShowcaseMembershipStatusNotAMember, nil
	}

	if community.HasMember(contactPubKey) {
		return identity.ProfileShowcaseMembershipStatusProvenMember, nil
	}

	return identity.ProfileShowcaseMembershipStatusNotAMember, nil
}

func (m *Messenger) validateCommunitiesMembership(communities []*identity.ProfileShowcaseCommunity, contactPubKey *ecdsa.PublicKey) ([]*identity.ProfileShowcaseCommunity, error) {
	validatedCommunities := []*identity.ProfileShowcaseCommunity{}

	for _, communityEntry := range communities {
		community, err := m.FetchCommunity(&FetchCommunityRequest{
			CommunityKey:    communityEntry.CommunityID,
			Shard:           nil,
			TryDatabase:     true,
			WaitForResponse: true,
		})
		if err != nil {
			m.logger.Warn("failed to fetch community for profile entry ", zap.Error(err))
			continue
		}

		communityEntry.MembershipStatus, err = m.validateCommunityMembershipEntry(communityEntry, community, contactPubKey)
		if err != nil {
			m.logger.Warn("failed to verify grant signature ", zap.Error(err))
		}
		validatedCommunities = append(validatedCommunities, communityEntry)
	}

	return validatedCommunities, nil
}

func (m *Messenger) toProfileShowcaseCommunityProto(preferences []*identity.ProfileShowcaseCommunityPreference, visibility identity.ProfileShowcaseVisibility) []*protobuf.ProfileShowcaseCommunity {
	entries := []*protobuf.ProfileShowcaseCommunity{}
	for _, preference := range preferences {
		if preference.ShowcaseVisibility != visibility {
			continue
		}

		entry := &protobuf.ProfileShowcaseCommunity{
			CommunityId: preference.CommunityID,
			Order:       uint32(preference.Order),
		}

		community, err := m.communitiesManager.GetByIDString(preference.CommunityID)
		if err != nil {
			m.logger.Warn("failed to get community for profile entry ", zap.Error(err))
		}

		if community != nil && community.Encrypted() {
			grant, _, err := m.communitiesManager.GetCommunityGrant(preference.CommunityID)
			if err != nil {
				m.logger.Warn("failed to get community for profile entry ", zap.Error(err))
			}

			entry.Grant = grant
		}

		entries = append(entries, entry)
	}
	return entries
}

func (m *Messenger) toProfileShowcaseAccountProto(preferences []*identity.ProfileShowcaseAccountPreference, visibility identity.ProfileShowcaseVisibility) []*protobuf.ProfileShowcaseAccount {
	entries := []*protobuf.ProfileShowcaseAccount{}
	for _, preference := range preferences {
		if preference.ShowcaseVisibility != visibility {
			continue
		}

		account, err := m.settings.GetAccountByAddress(types.HexToAddress(preference.Address))
		if err != nil {
			m.logger.Warn("failed to get account for profile entry ", zap.Error(err))
		}
		if account == nil {
			m.logger.Warn("can not find wallet account for profile entry ")
			continue
		}

		entries = append(entries, &protobuf.ProfileShowcaseAccount{
			Address: preference.Address,
			Name:    account.Name,
			ColorId: string(account.ColorID),
			Emoji:   account.Emoji,
			Order:   uint32(preference.Order),
		})
	}
	return entries
}

func (m *Messenger) toProfileShowcaseCollectibleProto(preferences []*identity.ProfileShowcaseCollectiblePreference, visibility identity.ProfileShowcaseVisibility) []*protobuf.ProfileShowcaseCollectible {
	entries := []*protobuf.ProfileShowcaseCollectible{}
	for _, preference := range preferences {
		if preference.ShowcaseVisibility != visibility {
			continue
		}

		entries = append(entries, &protobuf.ProfileShowcaseCollectible{
			ContractAddress: preference.ContractAddress,
			ChainId:         preference.ChainID,
			TokenId:         preference.TokenID,
			Order:           uint32(preference.Order),
		})
	}
	return entries
}

func (m *Messenger) toProfileShowcaseVerifiedTokensProto(preferences []*identity.ProfileShowcaseVerifiedTokenPreference, visibility identity.ProfileShowcaseVisibility) []*protobuf.ProfileShowcaseVerifiedToken {
	entries := []*protobuf.ProfileShowcaseVerifiedToken{}
	for _, preference := range preferences {
		if preference.ShowcaseVisibility != visibility {
			continue
		}

		entries = append(entries, &protobuf.ProfileShowcaseVerifiedToken{
			Symbol: preference.Symbol,
			Order:  uint32(preference.Order),
		})
	}
	return entries
}

func (m *Messenger) toProfileShowcaseUnverifiedTokensProto(preferences []*identity.ProfileShowcaseUnverifiedTokenPreference, visibility identity.ProfileShowcaseVisibility) []*protobuf.ProfileShowcaseUnverifiedToken {
	entries := []*protobuf.ProfileShowcaseUnverifiedToken{}
	for _, preference := range preferences {
		if preference.ShowcaseVisibility != visibility {
			continue
		}

		entries = append(entries, &protobuf.ProfileShowcaseUnverifiedToken{
			ContractAddress: preference.ContractAddress,
			ChainId:         preference.ChainID,
			Order:           uint32(preference.Order),
		})
	}
	return entries
}

func (m *Messenger) toProfileShowcaseSocialLinksProto(preferences []*identity.ProfileShowcaseSocialLinkPreference, visibility identity.ProfileShowcaseVisibility) []*protobuf.ProfileShowcaseSocialLink {
	entries := []*protobuf.ProfileShowcaseSocialLink{}
	for _, preference := range preferences {
		if preference.ShowcaseVisibility != visibility {
			continue
		}

		entries = append(entries, &protobuf.ProfileShowcaseSocialLink{
			Text:  preference.Text,
			Url:   preference.URL,
			Order: uint32(preference.Order),
		})
	}
	return entries
}

func (m *Messenger) fromProfileShowcaseCommunityProto(senderPubKey *ecdsa.PublicKey, messages []*protobuf.ProfileShowcaseCommunity) []*identity.ProfileShowcaseCommunity {
	// NOTE: no requests to the network are allowed to be made here, called in the receiver thread
	entries := []*identity.ProfileShowcaseCommunity{}
	for _, message := range messages {
		entry := &identity.ProfileShowcaseCommunity{
			CommunityID: message.CommunityId,
			Order:       int(message.Order),
			Grant:       message.Grant,
		}

		entries = append(entries, entry)
	}
	return entries
}

func (m *Messenger) fromProfileShowcaseAccountProto(messages []*protobuf.ProfileShowcaseAccount) []*identity.ProfileShowcaseAccount {
	// NOTE: no requests to the network are allowed to be made here, called in the receiver thread
	entries := []*identity.ProfileShowcaseAccount{}
	for _, entry := range messages {
		entries = append(entries, &identity.ProfileShowcaseAccount{
			Address: entry.Address,
			Name:    entry.Name,
			ColorID: entry.ColorId,
			Emoji:   entry.Emoji,
			Order:   int(entry.Order),
		})
	}
	return entries
}

func (m *Messenger) fromProfileShowcaseCollectibleProto(messages []*protobuf.ProfileShowcaseCollectible) []*identity.ProfileShowcaseCollectible {
	// NOTE: no requests to the network are allowed to be made here, called in the receiver thread
	entries := []*identity.ProfileShowcaseCollectible{}
	for _, message := range messages {
		entry := &identity.ProfileShowcaseCollectible{
			ContractAddress: message.ContractAddress,
			ChainID:         message.ChainId,
			TokenID:         message.TokenId,
			Order:           int(message.Order),
		}
		entries = append(entries, entry)
	}
	return entries
}

func (m *Messenger) fromProfileShowcaseVerifiedTokenProto(messages []*protobuf.ProfileShowcaseVerifiedToken) []*identity.ProfileShowcaseVerifiedToken {
	// NOTE: no requests to the network are allowed to be made here, called in the receiver thread
	entries := []*identity.ProfileShowcaseVerifiedToken{}
	for _, entry := range messages {
		entries = append(entries, &identity.ProfileShowcaseVerifiedToken{
			Symbol: entry.Symbol,
			Order:  int(entry.Order),
		})
	}
	return entries
}

func (m *Messenger) fromProfileShowcaseUnverifiedTokenProto(messages []*protobuf.ProfileShowcaseUnverifiedToken) []*identity.ProfileShowcaseUnverifiedToken {
	// NOTE: no requests to the network are allowed to be made here, called in the receiver thread
	entries := []*identity.ProfileShowcaseUnverifiedToken{}
	for _, entry := range messages {
		entries = append(entries, &identity.ProfileShowcaseUnverifiedToken{
			ContractAddress: entry.ContractAddress,
			ChainID:         entry.ChainId,
			Order:           int(entry.Order),
		})
	}
	return entries
}

func (m *Messenger) fromProfileShowcaseSocialLinkProto(messages []*protobuf.ProfileShowcaseSocialLink) []*identity.ProfileShowcaseSocialLink {
	// NOTE: no requests to the network are allowed to be made here, called in the receiver thread
	entries := []*identity.ProfileShowcaseSocialLink{}
	for _, entry := range messages {
		entries = append(entries, &identity.ProfileShowcaseSocialLink{
			Text:  entry.Text,
			URL:   entry.Url,
			Order: int(entry.Order),
		})
	}
	return entries
}

func (m *Messenger) SetProfileShowcasePreferences(preferences *identity.ProfileShowcasePreferences, sync bool) error {
	clock, _ := m.getLastClockWithRelatedChat()
	preferences.Clock = clock
	return m.setProfileShowcasePreferences(preferences, sync)
}

func (m *Messenger) setProfileShowcasePreferences(preferences *identity.ProfileShowcasePreferences, sync bool) error {
	err := identity.Validate(preferences)
	if err != nil {
		return err
	}

	err = m.validateCollectiblesOwnership(preferences.Accounts, preferences.Collectibles)
	if err != nil {
		return err
	}

	err = m.persistence.SaveProfileShowcasePreferences(preferences)
	if err != nil {
		return err
	}

	if sync {
		err = m.syncProfileShowcasePreferences(context.Background(), m.dispatchMessage)
		if err != nil {
			return err
		}
	}

	return m.DispatchProfileShowcase()
}

func (m *Messenger) DispatchProfileShowcase() error {
	err := m.publishContactCode()
	if err != nil {
		return err
	}
	return nil
}

func (m *Messenger) GetProfileShowcasePreferences() (*identity.ProfileShowcasePreferences, error) {
	return m.persistence.GetProfileShowcasePreferences()
}

func (m *Messenger) GetProfileShowcaseForContact(contactID string, validate bool) (*identity.ProfileShowcase, error) {
	profileShowcase, err := m.persistence.GetProfileShowcaseForContact(contactID)
	if err != nil {
		return nil, err
	}

	if !validate {
		return profileShowcase, nil
	}

	contactPubKey, err := common.HexToPubkey(contactID)
	if err != nil {
		return nil, err
	}

	profileShowcase.Communities, err = m.validateCommunitiesMembership(profileShowcase.Communities, contactPubKey)
	if err != nil {
		return nil, err
	}

	// TODO: validate collectibles & assets ownership, https://github.com/status-im/status-desktop/issues/14129

	return profileShowcase, nil
}

func (m *Messenger) GetProfileShowcaseAccountsByAddress(address string) ([]*identity.ProfileShowcaseAccount, error) {
	return m.persistence.GetProfileShowcaseAccountsByAddress(address)
}

func (m *Messenger) GetProfileShowcaseSocialLinksLimit() (int, error) {
	return identity.MaxProfileShowcaseSocialLinksLimit, nil
}

func (m *Messenger) GetProfileShowcaseEntriesLimit() (int, error) {
	return identity.MaxProfileShowcaseEntriesLimit, nil
}

func (m *Messenger) EncryptProfileShowcaseEntriesWithContactPubKeys(entries *protobuf.ProfileShowcaseEntries, contacts []*Contact) (*protobuf.ProfileShowcaseEntriesEncrypted, error) {
	// Make AES key
	AESKey := make([]byte, 32)
	_, err := crand.Read(AESKey)
	if err != nil {
		return nil, err
	}

	// Encrypt showcase entries with the AES key
	data, err := proto.Marshal(entries)
	if err != nil {
		return nil, err
	}

	encrypted, err := common.Encrypt(data, AESKey, crand.Reader)
	if err != nil {
		return nil, err
	}

	eAESKeys := [][]byte{}
	// Sign for each contact
	for _, contact := range contacts {
		var pubK *ecdsa.PublicKey
		var sharedKey []byte
		var eAESKey []byte

		pubK, err = contact.PublicKey()
		if err != nil {
			return nil, err
		}
		// Generate a Diffie-Helman (DH) between the sender private key and the recipient's public key
		sharedKey, err = common.MakeECDHSharedKey(m.identity, pubK)
		if err != nil {
			return nil, err
		}

		// Encrypt the main AES key with AES encryption using the DH key
		eAESKey, err = common.Encrypt(AESKey, sharedKey, crand.Reader)
		if err != nil {
			return nil, err
		}

		eAESKeys = append(eAESKeys, eAESKey)
	}

	return &protobuf.ProfileShowcaseEntriesEncrypted{
		EncryptedEntries: encrypted,
		EncryptionKeys:   eAESKeys,
	}, nil
}

func (m *Messenger) DecryptProfileShowcaseEntriesWithPubKey(senderPubKey *ecdsa.PublicKey, encrypted *protobuf.ProfileShowcaseEntriesEncrypted) (*protobuf.ProfileShowcaseEntries, error) {
	for _, eAESKey := range encrypted.EncryptionKeys {
		// Generate a Diffie-Helman (DH) between the recipient's private key and the sender's public key
		sharedKey, err := common.MakeECDHSharedKey(m.identity, senderPubKey)
		if err != nil {
			return nil, err
		}

		// Decrypt the main encryption AES key with AES encryption using the DH key
		dAESKey, err := common.Decrypt(eAESKey, sharedKey)
		if err != nil {
			if err.Error() == ErrCipherMessageAutentificationFailed {
				continue
			}
			return nil, err
		}
		if dAESKey == nil {
			return nil, errorDecryptingPayloadEncryptionKey
		}

		// Decrypt profile entries with the newly decrypted main encryption AES key
		entriesData, err := common.Decrypt(encrypted.EncryptedEntries, dAESKey)
		if err != nil {
			return nil, err
		}

		entries := &protobuf.ProfileShowcaseEntries{}
		err = proto.Unmarshal(entriesData, entries)
		if err != nil {
			return nil, err
		}

		return entries, nil
	}

	// Return empty if no matching key found
	return &protobuf.ProfileShowcaseEntries{}, nil
}

func (m *Messenger) GetProfileShowcaseForSelfIdentity() (*protobuf.ProfileShowcase, error) {
	preferences, err := m.GetProfileShowcasePreferences()
	if err != nil {
		return nil, err
	}

	forEveryone := &protobuf.ProfileShowcaseEntries{
		Communities:      m.toProfileShowcaseCommunityProto(preferences.Communities, identity.ProfileShowcaseVisibilityEveryone),
		Accounts:         m.toProfileShowcaseAccountProto(preferences.Accounts, identity.ProfileShowcaseVisibilityEveryone),
		Collectibles:     m.toProfileShowcaseCollectibleProto(preferences.Collectibles, identity.ProfileShowcaseVisibilityEveryone),
		VerifiedTokens:   m.toProfileShowcaseVerifiedTokensProto(preferences.VerifiedTokens, identity.ProfileShowcaseVisibilityEveryone),
		UnverifiedTokens: m.toProfileShowcaseUnverifiedTokensProto(preferences.UnverifiedTokens, identity.ProfileShowcaseVisibilityEveryone),
		SocialLinks:      m.toProfileShowcaseSocialLinksProto(preferences.SocialLinks, identity.ProfileShowcaseVisibilityEveryone),
	}

	forContacts := &protobuf.ProfileShowcaseEntries{
		Communities:      m.toProfileShowcaseCommunityProto(preferences.Communities, identity.ProfileShowcaseVisibilityContacts),
		Accounts:         m.toProfileShowcaseAccountProto(preferences.Accounts, identity.ProfileShowcaseVisibilityContacts),
		Collectibles:     m.toProfileShowcaseCollectibleProto(preferences.Collectibles, identity.ProfileShowcaseVisibilityContacts),
		VerifiedTokens:   m.toProfileShowcaseVerifiedTokensProto(preferences.VerifiedTokens, identity.ProfileShowcaseVisibilityContacts),
		UnverifiedTokens: m.toProfileShowcaseUnverifiedTokensProto(preferences.UnverifiedTokens, identity.ProfileShowcaseVisibilityContacts),
		SocialLinks:      m.toProfileShowcaseSocialLinksProto(preferences.SocialLinks, identity.ProfileShowcaseVisibilityContacts),
	}

	forIDVerifiedContacts := &protobuf.ProfileShowcaseEntries{
		Communities:      m.toProfileShowcaseCommunityProto(preferences.Communities, identity.ProfileShowcaseVisibilityIDVerifiedContacts),
		Accounts:         m.toProfileShowcaseAccountProto(preferences.Accounts, identity.ProfileShowcaseVisibilityIDVerifiedContacts),
		Collectibles:     m.toProfileShowcaseCollectibleProto(preferences.Collectibles, identity.ProfileShowcaseVisibilityIDVerifiedContacts),
		VerifiedTokens:   m.toProfileShowcaseVerifiedTokensProto(preferences.VerifiedTokens, identity.ProfileShowcaseVisibilityIDVerifiedContacts),
		UnverifiedTokens: m.toProfileShowcaseUnverifiedTokensProto(preferences.UnverifiedTokens, identity.ProfileShowcaseVisibilityIDVerifiedContacts),
		SocialLinks:      m.toProfileShowcaseSocialLinksProto(preferences.SocialLinks, identity.ProfileShowcaseVisibilityIDVerifiedContacts),
	}

	mutualContacts := []*Contact{}
	iDVerifiedContacts := []*Contact{}

	m.allContacts.Range(func(_ string, contact *Contact) (shouldContinue bool) {
		if contact.mutual() {
			mutualContacts = append(mutualContacts, contact)
			if contact.IsVerified() {
				iDVerifiedContacts = append(iDVerifiedContacts, contact)
			}
		}
		return true
	})

	forContactsEncrypted, err := m.EncryptProfileShowcaseEntriesWithContactPubKeys(forContacts, mutualContacts)
	if err != nil {
		return nil, err
	}

	forIDVerifiedContactsEncrypted, err := m.EncryptProfileShowcaseEntriesWithContactPubKeys(forIDVerifiedContacts, iDVerifiedContacts)
	if err != nil {
		return nil, err
	}

	return &protobuf.ProfileShowcase{
		ForEveryone:           forEveryone,
		ForContacts:           forContactsEncrypted,
		ForIdVerifiedContacts: forIDVerifiedContactsEncrypted,
	}, nil
}

func (m *Messenger) BuildProfileShowcaseFromIdentity(state *ReceivedMessageState, message *protobuf.ProfileShowcase) error {
	senderPubKey := state.CurrentMessageState.PublicKey
	contactID := state.CurrentMessageState.Contact.ID

	communities := []*identity.ProfileShowcaseCommunity{}
	accounts := []*identity.ProfileShowcaseAccount{}
	collectibles := []*identity.ProfileShowcaseCollectible{}
	verifiedTokens := []*identity.ProfileShowcaseVerifiedToken{}
	unverifiedTokens := []*identity.ProfileShowcaseUnverifiedToken{}
	socialLinks := []*identity.ProfileShowcaseSocialLink{}

	communities = append(communities, m.fromProfileShowcaseCommunityProto(senderPubKey, message.ForEveryone.Communities)...)
	accounts = append(accounts, m.fromProfileShowcaseAccountProto(message.ForEveryone.Accounts)...)
	collectibles = append(collectibles, m.fromProfileShowcaseCollectibleProto(message.ForEveryone.Collectibles)...)
	verifiedTokens = append(verifiedTokens, m.fromProfileShowcaseVerifiedTokenProto(message.ForEveryone.VerifiedTokens)...)
	unverifiedTokens = append(unverifiedTokens, m.fromProfileShowcaseUnverifiedTokenProto(message.ForEveryone.UnverifiedTokens)...)
	socialLinks = append(socialLinks, m.fromProfileShowcaseSocialLinkProto(message.ForEveryone.SocialLinks)...)

	forContacts, err := m.DecryptProfileShowcaseEntriesWithPubKey(senderPubKey, message.ForContacts)
	if err != nil {
		return err
	}

	if forContacts != nil {
		communities = append(communities, m.fromProfileShowcaseCommunityProto(senderPubKey, forContacts.Communities)...)
		accounts = append(accounts, m.fromProfileShowcaseAccountProto(forContacts.Accounts)...)
		collectibles = append(collectibles, m.fromProfileShowcaseCollectibleProto(forContacts.Collectibles)...)
		verifiedTokens = append(verifiedTokens, m.fromProfileShowcaseVerifiedTokenProto(forContacts.VerifiedTokens)...)
		unverifiedTokens = append(unverifiedTokens, m.fromProfileShowcaseUnverifiedTokenProto(forContacts.UnverifiedTokens)...)
		socialLinks = append(socialLinks, m.fromProfileShowcaseSocialLinkProto(forContacts.SocialLinks)...)
	}

	forIDVerifiedContacts, err := m.DecryptProfileShowcaseEntriesWithPubKey(senderPubKey, message.ForIdVerifiedContacts)
	if err != nil {
		return err
	}

	if forIDVerifiedContacts != nil {
		communities = append(communities, m.fromProfileShowcaseCommunityProto(senderPubKey, forIDVerifiedContacts.Communities)...)
		accounts = append(accounts, m.fromProfileShowcaseAccountProto(forIDVerifiedContacts.Accounts)...)
		collectibles = append(collectibles, m.fromProfileShowcaseCollectibleProto(forIDVerifiedContacts.Collectibles)...)
		verifiedTokens = append(verifiedTokens, m.fromProfileShowcaseVerifiedTokenProto(forIDVerifiedContacts.VerifiedTokens)...)
		unverifiedTokens = append(unverifiedTokens, m.fromProfileShowcaseUnverifiedTokenProto(forIDVerifiedContacts.UnverifiedTokens)...)
		socialLinks = append(socialLinks, m.fromProfileShowcaseSocialLinkProto(forIDVerifiedContacts.SocialLinks)...)
	}

	sortProfileEntyByOrder(communities, func(i int) int { return communities[i].Order })
	sortProfileEntyByOrder(accounts, func(i int) int { return accounts[i].Order })
	sortProfileEntyByOrder(collectibles, func(i int) int { return collectibles[i].Order })
	sortProfileEntyByOrder(verifiedTokens, func(i int) int { return verifiedTokens[i].Order })
	sortProfileEntyByOrder(unverifiedTokens, func(i int) int { return unverifiedTokens[i].Order })
	sortProfileEntyByOrder(socialLinks, func(i int) int { return socialLinks[i].Order })

	newShowcase := &identity.ProfileShowcase{
		ContactID:        contactID,
		Communities:      communities,
		Accounts:         accounts,
		Collectibles:     collectibles,
		VerifiedTokens:   verifiedTokens,
		UnverifiedTokens: unverifiedTokens,
		SocialLinks:      socialLinks,
	}

	oldShowcase, err := m.persistence.GetProfileShowcaseForContact(contactID)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(newShowcase, oldShowcase) {
		return nil
	}

	err = m.persistence.ClearProfileShowcaseForContact(contactID)
	if err != nil {
		return err
	}

	err = m.persistence.SaveProfileShowcaseForContact(newShowcase)
	if err != nil {
		return err
	}

	return nil
}

func (m *Messenger) UpdateProfileShowcaseWalletAccount(account *accounts.Account) error {
	profileAccount, err := m.persistence.GetProfileShowcaseAccountPreference(account.Address.Hex())
	if err != nil {
		return err
	}

	if profileAccount == nil {
		// No corresponding profile entry, exit
		return nil
	}

	return m.DispatchProfileShowcase()
}

func (m *Messenger) DeleteProfileShowcaseWalletAccount(account *accounts.Account) error {
	deleted, err := m.persistence.DeleteProfileShowcaseAccountPreference(account.Address.Hex())
	if err != nil {
		return err
	}

	if deleted {
		return m.DispatchProfileShowcase()
	}
	return nil
}

func (m *Messenger) DeleteProfileShowcaseCommunity(community *communities.Community) error {
	deleted, err := m.persistence.DeleteProfileShowcaseCommunityPreference(community.IDString())
	if err != nil {
		return err
	}

	if deleted {
		return m.DispatchProfileShowcase()
	}
	return nil
}

func (m *Messenger) saveProfileShowcasePreferencesProto(p *protobuf.SyncProfileShowcasePreferences, shouldSync bool) (*identity.ProfileShowcasePreferences, error) {
	if p == nil {
		return nil, nil
	}
	preferences := FromProfileShowcasePreferencesProto(p)
	return preferences, m.setProfileShowcasePreferences(preferences, shouldSync)
}

func (m *Messenger) syncProfileShowcasePreferences(ctx context.Context, rawMessageHandler RawMessageHandler) error {
	preferences, err := m.GetProfileShowcasePreferences()
	if err != nil {
		return err
	}

	syncMessage := ToProfileShowcasePreferencesProto(preferences)
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	_, chat := m.getLastClockWithRelatedChat()
	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_PROFILE_SHOWCASE_PREFERENCES,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	return err
}
