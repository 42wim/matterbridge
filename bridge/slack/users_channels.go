package bslack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

const minimumRefreshInterval = 10 * time.Second

func (b *Bslack) getUser(id string) *slack.User {
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

func (b *Bslack) getUsername(id string) string {
	if user := b.getUser(id); user != nil {
		if user.Profile.DisplayName != "" {
			return user.Profile.DisplayName
		}
		return user.Name
	}
	b.Log.Warnf("Could not find user with ID '%s'", id)
	return ""
}

func (b *Bslack) getAvatar(id string) string {
	if user := b.getUser(id); user != nil {
		return user.Profile.Image48
	}
	return ""
}

func (b *Bslack) populateUser(userID string) {
	b.usersMutex.RLock()
	_, exists := b.users[userID]
	b.usersMutex.RUnlock()
	if exists {
		// already in cache
		return
	}

	user, err := b.sc.GetUserInfo(userID)
	if err != nil {
		b.Log.Debugf("GetUserInfo failed for %v: %v", userID, err)
		return
	}

	b.usersMutex.Lock()
	b.users[userID] = user
	b.usersMutex.Unlock()
}

func (b *Bslack) populateUsers(wait bool) {
	b.refreshMutex.Lock()
	if !wait && (time.Now().Before(b.earliestUserRefresh) || b.refreshInProgress) {
		b.Log.Debugf("Not refreshing user list as it was done less than %v ago.",
			minimumRefreshInterval)
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

			if err = handleRateLimit(b.Log, err); err != nil {
				b.Log.Errorf("Could not retrieve users: %#v", err)
				return
			}
			continue
		}

		for i := range pagination.Users {
			newUsers[pagination.Users[i].ID] = &pagination.Users[i]
		}
		b.Log.Debugf("getting %d users", len(pagination.Users))
		count++
		// more > 2000 users, slack will complain and ratelimit. break
		if count > 10 {
			b.Log.Info("Large slack detected > 2000 users, skipping loading complete userlist.")
			break
		}
	}

	b.usersMutex.Lock()
	defer b.usersMutex.Unlock()
	b.users = newUsers

	b.refreshMutex.Lock()
	defer b.refreshMutex.Unlock()
	b.earliestUserRefresh = time.Now().Add(minimumRefreshInterval)
	b.refreshInProgress = false
}

func (b *Bslack) getChannel(channel string) (*slack.Channel, error) {
	if strings.HasPrefix(channel, "ID:") {
		return b.getChannelByID(strings.TrimPrefix(channel, "ID:"))
	}
	return b.getChannelByName(channel)
}

func (b *Bslack) getChannelByName(name string) (*slack.Channel, error) {
	return b.getChannelBy(name, b.channelsByName)
}

func (b *Bslack) getChannelByID(id string) (*slack.Channel, error) {
	return b.getChannelBy(id, b.channelsByID)
}

func (b *Bslack) getChannelBy(lookupKey string, lookupMap map[string]*slack.Channel) (*slack.Channel, error) {
	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	if channel, ok := lookupMap[lookupKey]; ok {
		return channel, nil
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, lookupKey)
}

func (b *Bslack) populateChannels(wait bool) {
	b.refreshMutex.Lock()
	if !wait && (time.Now().Before(b.earliestChannelRefresh) || b.refreshInProgress) {
		b.Log.Debugf("Not refreshing channel list as it was done less than %v seconds ago.",
			minimumRefreshInterval)
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
		ExcludeArchived: "true",
		Types:           []string{"public_channel,private_channel"},
	}
	for {
		channels, nextCursor, err := b.sc.GetConversations(queryParams)
		if err != nil {
			if err = handleRateLimit(b.Log, err); err != nil {
				b.Log.Errorf("Could not retrieve channels: %#v", err)
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
	b.earliestChannelRefresh = time.Now().Add(minimumRefreshInterval)
	b.refreshInProgress = false
}
