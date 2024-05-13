package protocol

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/status-im/status-go/protocol/wakusync"

	"github.com/status-im/status-go/protocol/identity"

	"github.com/status-im/status-go/eth-node/types"
	waku2 "github.com/status-im/status-go/wakuv2"

	"golang.org/x/exp/maps"

	"github.com/stretchr/testify/suite"

	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/protocol/tt"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var hexRunes = []rune("0123456789abcdef")

// WaitOnMessengerResponse Wait until the condition is true or the timeout is reached.
func WaitOnMessengerResponse(m *Messenger, condition func(*MessengerResponse) bool, errorMessage string) (*MessengerResponse, error) {
	response := &MessengerResponse{}
	err := tt.RetryWithBackOff(func() error {
		var err error
		r, err := m.RetrieveAll()
		if err != nil {
			panic(err)
		}
		if err := response.Merge(r); err != nil {
			panic(err)
		}

		if err == nil && !condition(response) {
			err = errors.New(errorMessage)
		}
		return err
	})
	return response, err
}

type MessengerSignalsHandlerMock struct {
	MessengerSignalsHandler

	responseChan                 chan *MessengerResponse
	communityFoundChan           chan *communities.Community
	wakuBackedUpDataResponseChan chan *wakusync.WakuBackedUpDataResponse
}

func (m *MessengerSignalsHandlerMock) SendWakuFetchingBackupProgress(response *wakusync.WakuBackedUpDataResponse) {
	m.wakuBackedUpDataResponseChan <- response
}
func (m *MessengerSignalsHandlerMock) SendWakuBackedUpProfile(*wakusync.WakuBackedUpDataResponse)  {}
func (m *MessengerSignalsHandlerMock) SendWakuBackedUpSettings(*wakusync.WakuBackedUpDataResponse) {}
func (m *MessengerSignalsHandlerMock) SendWakuBackedUpKeypair(*wakusync.WakuBackedUpDataResponse)  {}
func (m *MessengerSignalsHandlerMock) SendWakuBackedUpWatchOnlyAccount(*wakusync.WakuBackedUpDataResponse) {
}

func (m *MessengerSignalsHandlerMock) MessengerResponse(response *MessengerResponse) {
	// Non-blocking send
	select {
	case m.responseChan <- response:
	default:
	}
}

func (m *MessengerSignalsHandlerMock) MessageDelivered(chatID string, messageID string) {}

func (m *MessengerSignalsHandlerMock) CommunityInfoFound(community *communities.Community) {
	select {
	case m.communityFoundChan <- community:
	default:
	}
}

func WaitOnSignaledSendWakuFetchingBackupProgress(m *Messenger, condition func(*wakusync.WakuBackedUpDataResponse) bool, errorMessage string) (*wakusync.WakuBackedUpDataResponse, error) {
	interval := 500 * time.Millisecond
	timeoutChan := time.After(10 * time.Second)

	if m.config.messengerSignalsHandler != nil {
		return nil, errors.New("messengerSignalsHandler already provided/mocked")
	}

	responseChan := make(chan *wakusync.WakuBackedUpDataResponse, 1000)
	m.config.messengerSignalsHandler = &MessengerSignalsHandlerMock{
		wakuBackedUpDataResponseChan: responseChan,
	}

	defer func() {
		m.config.messengerSignalsHandler = nil
	}()

	for {
		_, err := m.RetrieveAll()
		if err != nil {
			return nil, err
		}

		select {
		case r := <-responseChan:
			if condition(r) {
				return r, nil
			}
		case <-timeoutChan:
			return nil, errors.New("timed out: " + errorMessage)
		default: // No immediate response, rest & loop back to retrieve again
			time.Sleep(interval)
		}
	}
}

func WaitOnSignaledMessengerResponse(m *Messenger, condition func(*MessengerResponse) bool, errorMessage string) (*MessengerResponse, error) {
	interval := 500 * time.Millisecond
	timeoutChan := time.After(10 * time.Second)

	if m.config.messengerSignalsHandler != nil {
		return nil, errors.New("messengerSignalsHandler already provided/mocked")
	}

	responseChan := make(chan *MessengerResponse, 64)
	m.config.messengerSignalsHandler = &MessengerSignalsHandlerMock{
		responseChan: responseChan,
	}

	defer func() {
		m.config.messengerSignalsHandler = nil
	}()

	for {
		_, err := m.RetrieveAll()
		if err != nil {
			return nil, err
		}

		select {
		case r := <-responseChan:
			if condition(r) {
				return r, nil
			}

		case <-timeoutChan:
			return nil, errors.New(errorMessage)

		default: // No immediate response, rest & loop back to retrieve again
			time.Sleep(interval)
		}
	}
}

func WaitOnSignaledCommunityFound(m *Messenger, action func(), condition func(community *communities.Community) bool, timeout time.Duration, errorMessage string) error {
	timeoutChan := time.After(timeout)

	if m.config.messengerSignalsHandler != nil {
		return errors.New("messengerSignalsHandler already provided/mocked")
	}

	communityFoundChan := make(chan *communities.Community, 1)
	m.config.messengerSignalsHandler = &MessengerSignalsHandlerMock{
		communityFoundChan: communityFoundChan,
	}

	defer func() {
		m.config.messengerSignalsHandler = nil
	}()

	// Call the action after setting up the mock
	action()

	// Wait for condition after
	for {
		select {
		case c := <-communityFoundChan:
			if condition(c) {
				return nil
			}
		case <-timeoutChan:
			return errors.New("timed out: " + errorMessage)
		}
	}
}

func WaitForConnectionStatus(s *suite.Suite, waku *waku2.Waku, action func() bool) {
	subscription := waku.SubscribeToConnStatusChanges()
	defer subscription.Unsubscribe()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Action should return the desired online status
	wantedOnline := action()

	for {
		select {
		case status := <-subscription.C:
			if status.IsOnline == wantedOnline {
				return
			}
		case <-ctx.Done():
			s.Require().Fail(fmt.Sprintf("timeout waiting for waku connection status '%t'", wantedOnline))
			return
		}
	}
}

func hasAllPeers(m map[string]types.WakuV2Peer, checkSlice []string) bool {
	for _, check := range checkSlice {
		if _, ok := m[check]; !ok {
			return false
		}
	}
	return true
}

func WaitForPeersConnected(s *suite.Suite, waku *waku2.Waku, action func() []string) {
	subscription := waku.SubscribeToConnStatusChanges()
	defer subscription.Unsubscribe()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Action should return the desired peer ID
	peerIDs := action()
	if hasAllPeers(waku.Peers(), peerIDs) {
		return
	}

	for {
		select {
		case status := <-subscription.C:
			if hasAllPeers(status.Peers, peerIDs) {
				// Give some time for p2p events, otherwise might look like peer is available, but fail to send a message.
				time.Sleep(100 * time.Millisecond)
				return
			}
		case <-ctx.Done():
			s.Require().Fail(fmt.Sprintf("timeout waiting for peers connected '%+v'", peerIDs))
			return
		}
	}
}

func FindFirstByContentType(messages []*common.Message, contentType protobuf.ChatMessage_ContentType) *common.Message {
	for _, message := range messages {
		if message.ContentType == contentType {
			return message
		}
	}
	return nil
}

func PairDevices(s *suite.Suite, device1, device2 *Messenger) {
	// Send pairing data
	response, err := device1.SendPairInstallation(context.Background(), nil)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Len(response.Chats(), 1)
	s.False(response.Chats()[0].Active)

	i, ok := device1.allInstallations.Load(device1.installationID)
	s.Require().True(ok)

	// Wait for the message to reach its destination
	response, err = WaitOnMessengerResponse(
		device2,
		func(r *MessengerResponse) bool {
			for _, installation := range r.Installations {
				if installation.ID == device1.installationID {
					return installation.InstallationMetadata != nil &&
						i.InstallationMetadata.Name == installation.InstallationMetadata.Name &&
						i.InstallationMetadata.DeviceType == installation.InstallationMetadata.DeviceType
				}
			}
			return false

		},
		"installation not received",
	)
	s.Require().NoError(err)
	s.Require().NotNil(response)

	// Ensure installation is enabled
	err = device2.EnableInstallation(device1.installationID)
	s.Require().NoError(err)
}

func SetSettingsAndWaitForChange(s *suite.Suite, messenger *Messenger, timeout time.Duration,
	actionCallback func(), eventCallback func(*SelfContactChangeEvent) bool) {

	allEventsReceived := false
	channel := messenger.SubscribeToSelfContactChanges()
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for !allEventsReceived {
			select {
			case event := <-channel:
				allEventsReceived = eventCallback(event)
			case <-time.After(timeout):
				return
			}
		}
	}()

	actionCallback()

	wg.Wait()

	s.Require().True(allEventsReceived)
}

func SetIdentityImagesAndWaitForChange(s *suite.Suite, messenger *Messenger, timeout time.Duration, actionCallback func()) {
	channel := messenger.SubscribeToSelfContactChanges()
	ok := false
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case event := <-channel:
			if event.ImagesChanged {
				ok = true
			}
		case <-time.After(timeout):
			return
		}
	}()

	actionCallback()

	wg.Wait()

	s.Require().True(ok)
}

func WaitForAvailableStoreNode(s *suite.Suite, m *Messenger, timeout time.Duration) {
	available := m.waitForAvailableStoreNode(timeout)
	s.Require().True(available)
}

func TearDownMessenger(s *suite.Suite, m *Messenger) {
	if m == nil {
		return
	}
	s.Require().NoError(m.Shutdown())
	if m.database != nil {
		s.Require().NoError(m.database.Close())
	}
	if m.multiAccounts != nil {
		s.Require().NoError(m.multiAccounts.Close())
	}
}

func randomInt(length int) int {
	max := big.NewInt(int64(length))
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return int(value.Int64())
}

func randomString(length int, runes []rune) string {
	out := make([]rune, length)
	for i := range out {
		out[i] = runes[randomInt(len(runes))] // nolint: gosec
	}
	return string(out)
}

func RandomLettersString(length int) string {
	return randomString(length, letterRunes)
}

func RandomColor() string {
	return "#" + randomString(6, hexRunes)
}

func RandomCommunityTags(count int) []string {
	all := maps.Keys(requests.TagsEmojies)
	tags := make([]string, 0, count)
	indexes := map[int]struct{}{}

	for len(indexes) != count {
		index := randomInt(len(all))
		indexes[index] = struct{}{}
	}

	for index := range indexes {
		tags = append(tags, all[index])
	}

	return tags
}

func RandomBytes(length int) []byte {
	out := make([]byte, length)
	_, err := rand.Read(out)
	if err != nil {
		panic(err)
	}
	return out
}

func DummyProfileShowcasePreferences(withCollectibles bool) *identity.ProfileShowcasePreferences {
	preferences := &identity.ProfileShowcasePreferences{
		Communities: []*identity.ProfileShowcaseCommunityPreference{
			{
				CommunityID:        "0x254254546768764565565",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityEveryone,
			},
			{
				CommunityID:        "0x865241434343432412343",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityContacts,
			},
		},
		Accounts: []*identity.ProfileShowcaseAccountPreference{
			{
				Address:            "0x0000000000000000000000000033433445133423",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityEveryone,
				Order:              0,
			},
			{
				Address:            "0x0000000000000000000000000032433445133424",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityContacts,
				Order:              1,
			},
		},
		VerifiedTokens: []*identity.ProfileShowcaseVerifiedTokenPreference{
			{
				Symbol:             "ETH",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityEveryone,
				Order:              1,
			},
			{
				Symbol:             "DAI",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityIDVerifiedContacts,
				Order:              2,
			},
			{
				Symbol:             "SNT",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityNoOne,
				Order:              3,
			},
		},
		UnverifiedTokens: []*identity.ProfileShowcaseUnverifiedTokenPreference{
			{
				ContractAddress:    "0x454525452023452",
				ChainID:            11155111,
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityEveryone,
				Order:              0,
			},
			{
				ContractAddress:    "0x12312323323233",
				ChainID:            1,
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityContacts,
				Order:              1,
			},
		},
		SocialLinks: []*identity.ProfileShowcaseSocialLinkPreference{
			&identity.ProfileShowcaseSocialLinkPreference{
				Text:               identity.TwitterID,
				URL:                "https://twitter.com/ethstatus",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityEveryone,
				Order:              1,
			},
			&identity.ProfileShowcaseSocialLinkPreference{
				Text:               identity.TwitterID,
				URL:                "https://twitter.com/StatusIMBlog",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityIDVerifiedContacts,
				Order:              2,
			},
			&identity.ProfileShowcaseSocialLinkPreference{
				Text:               identity.GithubID,
				URL:                "https://github.com/status-im",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityContacts,
				Order:              3,
			},
		},
	}

	if withCollectibles {
		preferences.Collectibles = []*identity.ProfileShowcaseCollectiblePreference{
			{
				ContractAddress:    "0x12378534257568678487683576",
				ChainID:            1,
				TokenID:            "12321389592999903",
				ShowcaseVisibility: identity.ProfileShowcaseVisibilityEveryone,
				Order:              0,
			},
		}
	} else {
		preferences.Collectibles = []*identity.ProfileShowcaseCollectiblePreference{}
	}

	return preferences
}
