package bslack

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

func (b *Bslack) getUsername(id string) string {
	for _, u := range b.users {
		if u.ID == id {
			if u.Profile.DisplayName != "" {
				return u.Profile.DisplayName
			}
			return u.Name
		}
	}
	b.Log.Warnf("Could not find user with ID '%s'", id)
	return ""
}

func (b *Bslack) getAvatar(userid string) string {
	for _, u := range b.users {
		if userid == u.ID {
			return u.Profile.Image48
		}
	}
	return ""
}

func (b *Bslack) getChannel(channel string) (*slack.Channel, error) {
	if strings.HasPrefix(channel, "ID:") {
		return b.getChannelByID(strings.TrimPrefix(channel, "ID:"))
	}
	return b.getChannelByName(channel)
}

func (b *Bslack) getChannelByName(name string) (*slack.Channel, error) {
	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	if channel, ok := b.channelsByName[name]; ok {
		return channel, nil
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, name)
}

func (b *Bslack) getChannelByID(ID string) (*slack.Channel, error) {
	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	if channel, ok := b.channelsByID[ID]; ok {
		return channel, nil
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, ID)
}

func (b *Bslack) populateUsers() {
	users, err := b.sc.GetUsers()
	if err != nil {
		b.Log.Errorf("Could not reload users: %#v", err)
		return
	}

	newUsers := map[string]*slack.User{}
	for i := range users {
		// Use array index for pointer, not the copy
		// See: https://stackoverflow.com/a/29498133/504018
		newUsers[users[i].ID] = &users[i]
	}

	b.usersMutex.Lock()
	defer b.usersMutex.Unlock()
	b.users = newUsers
}

func (b *Bslack) populateChannels() {
	newChannelsByID := map[string]*slack.Channel{}
	newChannelsByName := map[string]*slack.Channel{}

	// We only retrieve public and private channels, not IMs
	// and MPIMs as those do not have a channel name.
	queryParams := &slack.GetConversationsParameters{
		ExcludeArchived: "true",
		Types:           []string{"public_channel,private_channel"},
	}
	for {
		channels, nextCursor, err := b.sc.GetConversations(queryParams)
		if err != nil {
			b.Log.Errorf("Could not reload channels: %#v", err)
			return
		}
		for i := 0; i < len(channels); i++ {
			newChannelsByID[channels[i].ID] = &channels[i]
			newChannelsByName[channels[i].Name] = &channels[i]
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
}

var (
	mentionRE  = regexp.MustCompile(`<@([a-zA-Z0-9]+)>`)
	channelRE  = regexp.MustCompile(`<#[a-zA-Z0-9]+\|(.+?)>`)
	variableRE = regexp.MustCompile(`<!((?:subteam\^)?[a-zA-Z0-9]+)(?:\|@?(.+?))?>`)
	urlRE      = regexp.MustCompile(`<(.*?)(\|.*?)?>`)
)

// @see https://api.slack.com/docs/message-formatting#linking_to_channels_and_users
func (b *Bslack) replaceMention(text string) string {
	replaceFunc := func(match string) string {
		userID := strings.Trim(match, "@<>")
		if username := b.getUsername(userID); userID != "" {
			return "@" + username
		}
		return match
	}
	return mentionRE.ReplaceAllStringFunc(text, replaceFunc)
}

// @see https://api.slack.com/docs/message-formatting#linking_to_channels_and_users
func (b *Bslack) replaceChannel(text string) string {
	for _, r := range channelRE.FindAllStringSubmatch(text, -1) {
		text = strings.Replace(text, r[0], "#"+r[1], 1)
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#variables
func (b *Bslack) replaceVariable(text string) string {
	for _, r := range variableRE.FindAllStringSubmatch(text, -1) {
		if r[2] != "" {
			text = strings.Replace(text, r[0], "@"+r[2], 1)
		} else {
			text = strings.Replace(text, r[0], "@"+r[1], 1)
		}
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#linking_to_urls
func (b *Bslack) replaceURL(text string) string {
	for _, r := range urlRE.FindAllStringSubmatch(text, -1) {
		if len(strings.TrimSpace(r[2])) == 1 { // A display text separator was found, but the text was blank
			text = strings.Replace(text, r[0], "", 1)
		} else {
			text = strings.Replace(text, r[0], r[1], 1)
		}
	}
	return text
}
