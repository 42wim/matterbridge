package protocol

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	accountJson "github.com/status-im/status-go/account/json"
	"github.com/status-im/status-go/api/multiformat"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/identity"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/verification"
)

type ContactRequestState int

const (
	ContactRequestStateNone ContactRequestState = iota
	ContactRequestStateMutual
	ContactRequestStateSent
	// Received is a confusing state, we should use
	// sent for both, since they are now stored in different
	// states
	ContactRequestStateReceived
	ContactRequestStateDismissed
)

type MutualStateUpdateType int

const (
	MutualStateUpdateTypeSent MutualStateUpdateType = iota + 1
	MutualStateUpdateTypeAdded
	MutualStateUpdateTypeRemoved
)

// ContactDeviceInfo is a struct containing information about a particular device owned by a contact
type ContactDeviceInfo struct {
	// The installation id of the device
	InstallationID string `json:"id"`
	// Timestamp represents the last time we received this info
	Timestamp int64 `json:"timestamp"`
	// FCMToken is to be used for push notifications
	FCMToken string `json:"fcmToken"`
}

func (c *Contact) CanonicalImage(profilePicturesVisibility settings.ProfilePicturesVisibilityType) string {
	if profilePicturesVisibility == settings.ProfilePicturesVisibilityNone || (profilePicturesVisibility == settings.ProfilePicturesVisibilityContactsOnly && !c.added()) {
		return c.Identicon
	}

	if largeImage, ok := c.Images[images.LargeDimName]; ok {
		imageBase64, err := largeImage.GetDataURI()
		if err == nil {
			return imageBase64
		}
	}

	if thumbImage, ok := c.Images[images.SmallDimName]; ok {
		imageBase64, err := thumbImage.GetDataURI()
		if err == nil {
			return imageBase64
		}
	}

	return c.Identicon
}

type VerificationStatus int

const (
	VerificationStatusUNVERIFIED VerificationStatus = iota
	VerificationStatusVERIFYING
	VerificationStatusVERIFIED
)

// Contact has information about a "Contact"
type Contact struct {
	// ID of the contact. It's a hex-encoded public key (prefixed with 0x).
	ID string `json:"id"`
	// Ethereum address of the contact
	Address string `json:"address,omitempty"`
	// ENS name of contact
	EnsName string `json:"name,omitempty"`
	// EnsVerified whether we verified the name of the contact
	ENSVerified bool `json:"ensVerified"`
	// Generated username name of the contact
	Alias string `json:"alias,omitempty"`
	// Identicon generated from public key
	Identicon string `json:"identicon"`
	// LastUpdated is the last time we received an update from the contact
	// updates should be discarded if last updated is less than the one stored
	LastUpdated uint64 `json:"lastUpdated"`

	// LastUpdatedLocally is the last time we updated the contact locally
	LastUpdatedLocally uint64 `json:"lastUpdatedLocally"`

	LocalNickname string `json:"localNickname,omitempty"`

	// Display name of the contact
	DisplayName string `json:"displayName"`

	// Bio - description of the contact (tell us about yourself)
	Bio string `json:"bio"`

	SocialLinks identity.SocialLinks `json:"socialLinks"`

	Images map[string]images.IdentityImage `json:"images"`

	Blocked bool `json:"blocked"`

	// ContactRequestRemoteState is the state of the contact request
	// on the contact's end
	ContactRequestRemoteState ContactRequestState `json:"contactRequestRemoteState"`
	// ContactRequestRemoteClock is the clock for incoming contact requests
	ContactRequestRemoteClock uint64 `json:"contactRequestRemoteClock"`

	// ContactRequestLocalState is the state of the contact request
	// on our end
	ContactRequestLocalState ContactRequestState `json:"contactRequestLocalState"`
	// ContactRequestLocalClock is the clock for outgoing contact requests
	ContactRequestLocalClock uint64 `json:"contactRequestLocalClock"`

	IsSyncing bool
	Removed   bool

	VerificationStatus VerificationStatus       `json:"verificationStatus"`
	TrustStatus        verification.TrustStatus `json:"trustStatus"`
}

func (c Contact) IsVerified() bool {
	return c.VerificationStatus == VerificationStatusVERIFIED
}

func (c Contact) IsVerifying() bool {
	return c.VerificationStatus == VerificationStatusVERIFYING
}

func (c Contact) IsUnverified() bool {
	return c.VerificationStatus == VerificationStatusUNVERIFIED
}

func (c Contact) IsUntrustworthy() bool {
	return c.TrustStatus == verification.TrustStatusUNTRUSTWORTHY
}

func (c Contact) IsTrusted() bool {
	return c.TrustStatus == verification.TrustStatusTRUSTED
}

func (c Contact) PublicKey() (*ecdsa.PublicKey, error) {
	b, err := types.DecodeHex(c.ID)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPubkey(b)
}

func (c *Contact) Block(clock uint64) {
	c.Blocked = true
	c.DismissContactRequest(clock)
	c.Removed = true
}

func (c *Contact) BlockDesktop() {
	c.Blocked = true
}

func (c *Contact) Unblock(clock uint64) {
	c.Blocked = false
	// Reset the contact request flow
	c.RetractContactRequest(clock)
}

func (c *Contact) added() bool {
	return c.ContactRequestLocalState == ContactRequestStateSent
}

func (c *Contact) hasAddedUs() bool {
	return c.ContactRequestRemoteState == ContactRequestStateReceived
}

func (c *Contact) mutual() bool {
	return c.added() && c.hasAddedUs()
}

func (c *Contact) active() bool {
	return c.mutual() && !c.Blocked
}

func (c *Contact) dismissed() bool {
	return c.ContactRequestLocalState == ContactRequestStateDismissed
}

func (c *Contact) names() []string {
	var names []string

	if c.LocalNickname != "" {
		names = append(names, c.LocalNickname)
	}

	if c.ENSVerified && len(c.EnsName) != 0 {
		names = append(names, c.EnsName)
	}

	if c.DisplayName != "" {
		names = append(names, c.DisplayName)
	}

	return append(names, c.Alias)

}

func (c *Contact) PrimaryName() string {
	return c.names()[0]
}

func (c *Contact) SecondaryName() string {
	// Only shown if the user has a nickname
	if c.LocalNickname == "" {
		return ""
	}
	names := c.names()
	if len(names) > 1 {
		return names[1]
	}
	return ""
}

type ContactRequestProcessingResponse struct {
	processed                 bool
	newContactRequestReceived bool
	sendBackState             bool
}

func (c *Contact) ContactRequestSent(clock uint64) ContactRequestProcessingResponse {
	if clock <= c.ContactRequestLocalClock {
		return ContactRequestProcessingResponse{}
	}

	c.ContactRequestLocalClock = clock
	c.ContactRequestLocalState = ContactRequestStateSent

	c.Removed = false

	return ContactRequestProcessingResponse{processed: true}
}

func (c *Contact) AcceptContactRequest(clock uint64) ContactRequestProcessingResponse {
	// We treat accept the same as sent, that's because accepting a contact
	// request that does not exist is possible if the instruction is coming from
	// a different device, we'd rather assume that a contact requested existed
	// and didn't reach our device than being in an inconsistent state
	return c.ContactRequestSent(clock)
}

func (c *Contact) RetractContactRequest(clock uint64) ContactRequestProcessingResponse {
	if clock <= c.ContactRequestLocalClock {
		return ContactRequestProcessingResponse{}
	}

	// This is a symmetric action, we set both local & remote clock
	// since we want everything before this point discarded, regardless
	// the side it was sent from
	c.ContactRequestLocalClock = clock
	c.ContactRequestLocalState = ContactRequestStateNone
	c.ContactRequestRemoteState = ContactRequestStateNone
	c.ContactRequestRemoteClock = clock
	c.Removed = true

	return ContactRequestProcessingResponse{processed: true}
}

func (c *Contact) DismissContactRequest(clock uint64) ContactRequestProcessingResponse {
	if clock <= c.ContactRequestLocalClock {
		return ContactRequestProcessingResponse{}
	}

	c.ContactRequestLocalClock = clock
	c.ContactRequestLocalState = ContactRequestStateDismissed

	return ContactRequestProcessingResponse{processed: true}
}

// Remote actions

func (c *Contact) contactRequestRetracted(clock uint64, fromSyncing bool, r ContactRequestProcessingResponse) ContactRequestProcessingResponse {
	if clock <= c.ContactRequestRemoteClock {
		return r
	}

	// This is a symmetric action, we set both local & remote clock
	// since we want everything before this point discarded, regardless
	// the side it was sent from. The only exception is when the contact
	// request has been explicitly dismissed, in which case we don't
	// change state
	if c.ContactRequestLocalState != ContactRequestStateDismissed && !fromSyncing {
		c.ContactRequestLocalClock = clock
		c.ContactRequestLocalState = ContactRequestStateNone
	}
	c.ContactRequestRemoteClock = clock
	c.ContactRequestRemoteState = ContactRequestStateNone
	r.processed = true
	return r
}

func (c *Contact) ContactRequestRetracted(clock uint64, fromSyncing bool) ContactRequestProcessingResponse {
	return c.contactRequestRetracted(clock, fromSyncing, ContactRequestProcessingResponse{})
}

func (c *Contact) contactRequestReceived(clock uint64, r ContactRequestProcessingResponse) ContactRequestProcessingResponse {
	if clock <= c.ContactRequestRemoteClock {
		return r
	}
	r.processed = true
	c.ContactRequestRemoteClock = clock
	switch c.ContactRequestRemoteState {
	case ContactRequestStateNone:
		r.newContactRequestReceived = true
	}
	c.ContactRequestRemoteState = ContactRequestStateReceived

	return r
}

func (c *Contact) ContactRequestReceived(clock uint64) ContactRequestProcessingResponse {
	return c.contactRequestReceived(clock, ContactRequestProcessingResponse{})
}

func (c *Contact) ContactRequestAccepted(clock uint64) ContactRequestProcessingResponse {
	if clock <= c.ContactRequestRemoteClock {
		return ContactRequestProcessingResponse{}
	}
	// We treat received and accepted in the same way
	// since the intention is clear on the other side
	// and there's no difference
	return c.ContactRequestReceived(clock)
}

func buildContactFromPkString(pkString string) (*Contact, error) {
	publicKeyBytes, err := types.DecodeHex(pkString)
	if err != nil {
		return nil, err
	}

	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return nil, err
	}

	return buildContact(pkString, publicKey)
}

func BuildContactFromPublicKey(publicKey *ecdsa.PublicKey) (*Contact, error) {
	id := common.PubkeyToHex(publicKey)
	return buildContact(id, publicKey)
}

func getShortenedCompressedKey(publicKey string) string {
	if len(publicKey) > 9 {
		firstPart := publicKey[0:3]
		ellipsis := "..."
		publicKeySize := len(publicKey)
		lastPart := publicKey[publicKeySize-6 : publicKeySize]
		abbreviatedKey := fmt.Sprintf("%s%s%s", firstPart, ellipsis, lastPart)
		return abbreviatedKey
	}
	return ""
}

func buildContact(publicKeyString string, publicKey *ecdsa.PublicKey) (*Contact, error) {
	compressedKey, err := multiformat.SerializeLegacyKey(common.PubkeyToHex(publicKey))
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(*publicKey)

	contact := &Contact{
		ID:      publicKeyString,
		Alias:   getShortenedCompressedKey(compressedKey),
		Address: types.EncodeHex(address[:]),
	}

	return contact, nil
}

func buildSelfContact(identity *ecdsa.PrivateKey, settings *accounts.Database, multiAccounts *multiaccounts.Database, account *multiaccounts.Account) (*Contact, error) {
	myPublicKeyString := types.EncodeHex(crypto.FromECDSAPub(&identity.PublicKey))

	c, err := buildContact(myPublicKeyString, &identity.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build contact: %w", err)
	}

	if settings != nil {
		if s, err := settings.GetSettings(); err == nil {
			c.DisplayName = s.DisplayName
			c.Bio = s.Bio
			if s.PreferredName != nil {
				c.EnsName = *s.PreferredName
			}
		}
		if socialLinks, err := settings.GetSocialLinks(); err != nil {
			c.SocialLinks = socialLinks
		}
	}

	if multiAccounts != nil && account != nil {
		if identityImages, err := multiAccounts.GetIdentityImages(account.KeyUID); err != nil {
			imagesMap := make(map[string]images.IdentityImage)
			for _, img := range identityImages {
				imagesMap[img.Name] = *img
			}

			c.Images = imagesMap
		}
	}

	return c, nil
}

func contactIDFromPublicKey(key *ecdsa.PublicKey) string {
	return types.EncodeHex(crypto.FromECDSAPub(key))
}

func contactIDFromPublicKeyString(key string) (string, error) {
	pubKey, err := common.HexToPubkey(key)
	if err != nil {
		return "", err
	}

	return contactIDFromPublicKey(pubKey), nil
}

func (c *Contact) ProcessSyncContactRequestState(remoteState ContactRequestState, remoteClock uint64, localState ContactRequestState, localClock uint64) {
	// We process the two separately, first local state
	switch localState {
	case ContactRequestStateDismissed:
		c.DismissContactRequest(localClock)
	case ContactRequestStateNone:
		c.RetractContactRequest(localClock)
	case ContactRequestStateSent:
		c.ContactRequestSent(localClock)
	}

	// and later remote state
	switch remoteState {
	case ContactRequestStateReceived:
		c.ContactRequestReceived(remoteClock)
	case ContactRequestStateNone:
		c.ContactRequestRetracted(remoteClock, true)
	}
}

func (c *Contact) MarshalJSON() ([]byte, error) {
	type Alias Contact
	type ContactType struct {
		*Alias
		Added               bool                `json:"added"`
		ContactRequestState ContactRequestState `json:"contactRequestState"`
		HasAddedUs          bool                `json:"hasAddedUs"`
		Mutual              bool                `json:"mutual"`
		Active              bool                `json:"active"`
		PrimaryName         string              `json:"primaryName"`
		SecondaryName       string              `json:"secondaryName,omitempty"`
	}

	item := ContactType{
		Alias: (*Alias)(c),
	}

	item.Added = c.added()
	item.HasAddedUs = c.hasAddedUs()
	item.Mutual = c.mutual()
	item.Active = c.active()
	item.PrimaryName = c.PrimaryName()
	item.SecondaryName = c.SecondaryName()

	if c.mutual() {
		item.ContactRequestState = ContactRequestStateMutual
	} else if c.dismissed() {
		item.ContactRequestState = ContactRequestStateDismissed
	} else if c.added() {
		item.ContactRequestState = ContactRequestStateSent
	} else if c.hasAddedUs() {
		item.ContactRequestState = ContactRequestStateReceived
	}
	ext, err := accountJson.ExtendStructWithPubKeyData(item.ID, item)
	if err != nil {
		return nil, err
	}

	return json.Marshal(ext)
}

// ContactRequestPropagatedStateReceived handles the propagation of state from
// the other end.
func (c *Contact) ContactRequestPropagatedStateReceived(state *protobuf.ContactRequestPropagatedState) ContactRequestProcessingResponse {

	// It's inverted, as their local states is our remote state
	expectedLocalState := ContactRequestState(state.RemoteState)
	expectedLocalClock := state.RemoteClock

	remoteState := ContactRequestState(state.LocalState)
	remoteClock := state.LocalClock

	response := ContactRequestProcessingResponse{}

	// If we notice that the state is not consistent, and their clock is
	// outdated, we send back the state so they can catch up.
	if expectedLocalClock < c.ContactRequestLocalClock && expectedLocalState != c.ContactRequestLocalState {
		response.processed = true
		response.sendBackState = true
	}

	// If they expect our state to be more up-to-date, we only
	// trust it if the state is set to None, in this case we can trust
	// it, since a retraction can be initiated by both parties
	if expectedLocalClock > c.ContactRequestLocalClock && c.ContactRequestLocalState != ContactRequestStateDismissed && expectedLocalState == ContactRequestStateNone {
		response.processed = true
		c.ContactRequestLocalClock = expectedLocalClock
		c.ContactRequestLocalState = ContactRequestStateNone
		// We set the remote state, as this was an implicit retraction
		// potentially, for example this could happen if they
		// sent a retraction earier, but we never received it,
		// or one of our paired devices has retracted the contact request
		// but we never synced with them.
		c.ContactRequestRemoteState = ContactRequestStateNone
	}

	// We always trust this
	if remoteClock > c.ContactRequestRemoteClock {
		if remoteState == ContactRequestStateSent {
			response = c.contactRequestReceived(remoteClock, response)
		} else if remoteState == ContactRequestStateNone {
			response = c.contactRequestRetracted(remoteClock, false, response)
		}
	}

	return response
}

func (c *Contact) ContactRequestPropagatedState() *protobuf.ContactRequestPropagatedState {
	return &protobuf.ContactRequestPropagatedState{
		LocalClock:  c.ContactRequestLocalClock,
		LocalState:  uint64(c.ContactRequestLocalState),
		RemoteClock: c.ContactRequestRemoteClock,
		RemoteState: uint64(c.ContactRequestRemoteState),
	}
}
