package bslack

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

const minimumRefreshInterval = 10 * time.Second

type users struct {
	log *logrus.Entry
	sc  *slack.Client

	users           map[string]*slack.User
	usersMutex      sync.RWMutex
	usersSyncPoints map[string]chan struct{}

	refreshInProgress bool
	earliestRefresh   time.Time
	refreshMutex      sync.Mutex
}

func newUserManager(log *logrus.Entry, sc *slack.Client) *users {
	return &users{
		log:             log,
		sc:              sc,
		users:           make(map[string]*slack.User),
		usersSyncPoints: make(map[string]chan struct{}),
		earliestRefresh: time.Now(),
	}
}

func (b *users) getUser(id string) *slack.User {
	b.usersMutex.RLock()
	user, ok := b.users[id]
	b.usersMutex.RUnlock()
	if ok {
		return user
	}
	b.populateUser(id)
	b.usersMutex.RLock()
	defer b.usersMutex.RUnlock()

	return b.users[id]
}

func (b *users) getUsername(id string) string {
	if user := b.getUser(id); user != nil {
		if user.Profile.DisplayName != "" {
			return user.Profile.DisplayName
		}
		return user.Name
	}
	b.log.Warnf("Could not find user with ID '%s'", id)
	return ""
}

func (b *users) getAvatar(id string) string {
	if user := b.getUser(id); user != nil {
		return user.Profile.Image48
	}
	return ""
}

func (b *users) populateUser(userID string) {
	for {
		b.usersMutex.Lock()
		_, exists := b.users[userID]
		if exists {
			// already in cache
			b.usersMutex.Unlock()
			return
		}

		if syncPoint, ok := b.usersSyncPoints[userID]; ok {
			// Another goroutine is already populating this user for us so wait on it to finish.
			b.usersMutex.Unlock()
			<-syncPoint
			// We do not return and iterate again to check that the entry does indeed exist
			// in case the previous query failed for some reason.
		} else {
			b.usersSyncPoints[userID] = make(chan struct{})
			defer func() {
				// Wake up any waiting goroutines and remove the synchronization point.
				close(b.usersSyncPoints[userID])
				delete(b.usersSyncPoints, userID)
			}()
			break
		}
	}

	// Do not hold the lock while fetching information from Slack
	// as this might take an unbounded amount of time.
	b.usersMutex.Unlock()

	user, err := b.sc.GetUserInfo(userID)
	if err != nil {
		b.log.Debugf("GetUserInfo failed for %v: %v", userID, err)
		return
	}

	b.usersMutex.Lock()
	defer b.usersMutex.Unlock()

	// Register user information.
	b.users[userID] = user
}

func (b *users) invalidateUser(userID string) {
	b.usersMutex.Lock()
	defer b.usersMutex.Unlock()
	delete(b.users, userID)
}

func (b *users) populateUsers(wait bool) {
	b.refreshMutex.Lock()
	if !wait && (time.Now().Before(b.earliestRefresh) || b.refreshInProgress) {
		b.log.Debugf("Not refreshing user list as it was done less than %v ago.", minimumRefreshInterval)
		b.refreshMutex.Unlock()

		return
	}
	for b.refreshInProgress {
		b.refreshMutex.Unlock()
		time.Sleep(time.Second)
		b.refreshMutex.Lock()
	}
	b.refreshInProgress = true
	b.refreshMutex.Unlock()

	newUsers := map[string]*slack.User{}
	pagination := b.sc.GetUsersPaginated(slack.GetUsersOptionLimit(200))
	count := 0
	for {
		var err error
		pagination, err = pagination.Next(context.Background())
		time.Sleep(time.Second)
		if err != nil {
			if pagination.Done(err) {
				break
			}

			if err = handleRateLimit(b.log, err); err != nil {
				b.log.Errorf("Could not retrieve users: %#v", err)
				return
			}
			continue
		}

		for i := range pagination.Users {
			newUsers[pagination.Users[i].ID] = &pagination.Users[i]
		}
		b.log.Debugf("getting %d users", len(pagination.Users))
		count++
		// more > 2000 users, slack will complain and ratelimit. break
		if count > 10 {
			b.log.Info("Large slack detected > 2000 users, skipping loading complete userlist.")
			break
		}
	}

	b.usersMutex.Lock()
	defer b.usersMutex.Unlock()
	b.users = newUsers

	b.refreshMutex.Lock()
	defer b.refreshMutex.Unlock()
	b.earliestRefresh = time.Now().Add(minimumRefreshInterval)
	b.refreshInProgress = false
}

type channels struct {
	log *logrus.Entry
	sc  *slack.Client

	channelsByID   map[string]*slack.Channel
	channelsByName map[string]*slack.Channel
	channelsMutex  sync.RWMutex

	channelMembers      map[string][]string
	channelMembersMutex sync.RWMutex

	refreshInProgress bool
	earliestRefresh   time.Time
	refreshMutex      sync.Mutex
}

func newChannelManager(log *logrus.Entry, sc *slack.Client) *channels {
	return &channels{
		log:             log,
		sc:              sc,
		channelsByID:    make(map[string]*slack.Channel),
		channelsByName:  make(map[string]*slack.Channel),
		earliestRefresh: time.Now(),
	}
}

func (b *channels) getChannel(channel string) (*slack.Channel, error) {
	if strings.HasPrefix(channel, "ID:") {
		return b.getChannelByID(strings.TrimPrefix(channel, "ID:"))
	}
	return b.getChannelByName(channel)
}

func (b *channels) getChannelByName(name string) (*slack.Channel, error) {
	return b.getChannelBy(name, b.channelsByName)
}

func (b *channels) getChannelByID(id string) (*slack.Channel, error) {
	return b.getChannelBy(id, b.channelsByID)
}

func (b *channels) getChannelBy(lookupKey string, lookupMap map[string]*slack.Channel) (*slack.Channel, error) {
	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	if channel, ok := lookupMap[lookupKey]; ok {
		return channel, nil
	}
	return nil, fmt.Errorf("channel %s not found", lookupKey)
}

func (b *channels) getChannelMembers(users *users) config.ChannelMembers {
	b.channelMembersMutex.RLock()
	defer b.channelMembersMutex.RUnlock()

	membersInfo := config.ChannelMembers{}
	for channelID, members := range b.channelMembers {
		for _, member := range members {
			channelName := ""
			userName := ""
			userNick := ""
			user := users.getUser(member)
			if user != nil {
				userName = user.Name
				userNick = user.Profile.DisplayName
			}
			channel, _ := b.getChannelByID(channelID)
			if channel != nil {
				channelName = channel.Name
			}
			memberInfo := config.ChannelMember{
				Username:    userName,
				Nick:        userNick,
				UserID:      member,
				ChannelID:   channelID,
				ChannelName: channelName,
			}
			membersInfo = append(membersInfo, memberInfo)
		}
	}
	return membersInfo
}

func (b *channels) registerChannel(channel slack.Channel) {
	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()

	b.channelsByID[channel.ID] = &channel
	b.channelsByName[channel.Name] = &channel
}

func (b *channels) populateChannels(wait bool) {
	b.refreshMutex.Lock()
	if !wait && (time.Now().Before(b.earliestRefresh) || b.refreshInProgress) {
		b.log.Debugf("Not refreshing channel list as it was done less than %v seconds ago.", minimumRefreshInterval)
		b.refreshMutex.Unlock()
		return
	}
	for b.refreshInProgress {
		b.refreshMutex.Unlock()
		time.Sleep(time.Second)
		b.refreshMutex.Lock()
	}
	b.refreshInProgress = true
	b.refreshMutex.Unlock()

	newChannelsByID := map[string]*slack.Channel{}
	newChannelsByName := map[string]*slack.Channel{}
	newChannelMembers := make(map[string][]string)

	// We only retrieve public and private channels, not IMs
	// and MPIMs as those do not have a channel name.
	queryParams := &slack.GetConversationsParameters{
		ExcludeArchived: true,
		Types:           []string{"public_channel,private_channel"},
		Limit:           1000,
	}
	for {
		channels, nextCursor, err := b.sc.GetConversations(queryParams)
		if err != nil {
			if err = handleRateLimit(b.log, err); err != nil {
				b.log.Errorf("Could not retrieve channels: %#v", err)
				return
			}
			continue
		}

		for i := range channels {
			newChannelsByID[channels[i].ID] = &channels[i]
			newChannelsByName[channels[i].Name] = &channels[i]
			// also find all the members in every channel
			// comment for now, issues on big slacks
			/*
				members, err := b.getUsersInConversation(channels[i].ID)
				if err != nil {
					if err = b.handleRateLimit(err); err != nil {
						b.Log.Errorf("Could not retrieve channel members: %#v", err)
						return
					}
					continue
				}
				newChannelMembers[channels[i].ID] = members
			*/
		}

		if nextCursor == "" {
			break
		}
		queryParams.Cursor = nextCursor
	}

	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()
	b.channelsByID = newChannelsByID
	b.channelsByName = newChannelsByName

	b.channelMembersMutex.Lock()
	defer b.channelMembersMutex.Unlock()
	b.channelMembers = newChannelMembers

	b.refreshMutex.Lock()
	defer b.refreshMutex.Unlock()
	b.earliestRefresh = time.Now().Add(minimumRefreshInterval)
	b.refreshInProgress = false
}
