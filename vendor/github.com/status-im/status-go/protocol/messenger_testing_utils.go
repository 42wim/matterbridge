package protocol

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"time"

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

	responseChan       chan *MessengerResponse
	communityFoundChan chan *communities.Community
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

func WaitOnSignaledMessengerResponse(m *Messenger, condition func(*MessengerResponse) bool, errorMessage string) (*MessengerResponse, error) {
	interval := 500 * time.Millisecond
	timeoutChan := time.After(10 * time.Second)

	if m.config.messengerSignalsHandler != nil {
		return nil, errors.New("messengerSignalsHandler already provided/mocked")
	}

	responseChan := make(chan *MessengerResponse, 1)
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
			return nil, errors.New(errorMessage)

		case <-timeoutChan:
			return nil, errors.New("timed out: " + errorMessage)

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
