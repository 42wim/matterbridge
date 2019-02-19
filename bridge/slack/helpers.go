package bslack

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/nlopes/slack"
)

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

func (b *Bslack) getChannel(channel string) (*slack.Channel, error) {
	if strings.HasPrefix(channel, "ID:") {
		return b.getChannelByID(strings.TrimPrefix(channel, "ID:"))
	}
	return b.getChannelByName(channel)
}

func (b *Bslack) getChannelByName(name string) (*slack.Channel, error) {
	return b.getChannelBy(name, b.channelsByName)
}

func (b *Bslack) getChannelByID(ID string) (*slack.Channel, error) {
	return b.getChannelBy(ID, b.channelsByID)
}

func (b *Bslack) getChannelBy(lookupKey string, lookupMap map[string]*slack.Channel) (*slack.Channel, error) {
	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	if channel, ok := lookupMap[lookupKey]; ok {
		return channel, nil
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, lookupKey)
}

const minimumRefreshInterval = 10 * time.Second

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

			if err = b.handleRateLimit(err); err != nil {
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
			if err = b.handleRateLimit(err); err != nil {
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

// populateReceivedMessage shapes the initial Matterbridge message that we will forward to the
// router before we apply message-dependent modifications.
func (b *Bslack) populateReceivedMessage(ev *slack.MessageEvent) (*config.Message, error) {
	// Use our own func because rtm.GetChannelInfo doesn't work for private channels.
	channel, err := b.getChannelByID(ev.Channel)
	if err != nil {
		return nil, err
	}

	rmsg := &config.Message{
		Text:     ev.Text,
		Channel:  channel.Name,
		Account:  b.Account,
		ID:       ev.Timestamp,
		Extra:    make(map[string][]interface{}),
		ParentID: ev.ThreadTimestamp,
		Protocol: b.Protocol,
	}
	if b.useChannelID {
		rmsg.Channel = "ID:" + channel.ID
	}

	// Handle 'edit' messages.
	if ev.SubMessage != nil && !b.GetBool(editDisableConfig) {
		rmsg.ID = ev.SubMessage.Timestamp
		if ev.SubMessage.ThreadTimestamp != ev.SubMessage.Timestamp {
			b.Log.Debugf("SubMessage %#v", ev.SubMessage)
			rmsg.Text = ev.SubMessage.Text + b.GetString(editSuffixConfig)
		}
	}

	// For edits, only submessage has thread ts.
	// Ensures edits to threaded messages maintain their prefix hint on the
	// unthreaded end.
	if ev.SubMessage != nil {
		rmsg.ParentID = ev.SubMessage.ThreadTimestamp
	}

	if err = b.populateMessageWithUserInfo(ev, rmsg); err != nil {
		return nil, err
	}
	return rmsg, err
}

func (b *Bslack) populateMessageWithUserInfo(ev *slack.MessageEvent, rmsg *config.Message) error {
	if ev.SubType == sMessageDeleted || ev.SubType == sFileComment {
		return nil
	}

	// First, deal with bot-originating messages but only do so when not using webhooks: we
	// would not be able to distinguish which bot would be sending them.
	if err := b.populateMessageWithBotInfo(ev, rmsg); err != nil {
		return err
	}

	// Second, deal with "real" users if we have the necessary information.
	var userID string
	switch {
	case ev.User != "":
		userID = ev.User
	case ev.SubMessage != nil && ev.SubMessage.User != "":
		userID = ev.SubMessage.User
	default:
		return nil
	}

	user := b.getUser(userID)
	if user == nil {
		return fmt.Errorf("could not find information for user with id %s", ev.User)
	}

	rmsg.UserID = user.ID
	rmsg.Username = user.Name
	if user.Profile.DisplayName != "" {
		rmsg.Username = user.Profile.DisplayName
	}
	return nil
}

func (b *Bslack) populateMessageWithBotInfo(ev *slack.MessageEvent, rmsg *config.Message) error {
	if ev.BotID == "" || b.GetString(outgoingWebhookConfig) != "" {
		return nil
	}

	var err error
	var bot *slack.Bot
	for {
		bot, err = b.rtm.GetBotInfo(ev.BotID)
		if err == nil {
			break
		}

		if err = b.handleRateLimit(err); err != nil {
			b.Log.Errorf("Could not retrieve bot information: %#v", err)
			return err
		}
	}
	b.Log.Debugf("Found bot %#v", bot)

	if bot.Name != "" {
		rmsg.Username = bot.Name
		if ev.Username != "" {
			rmsg.Username = ev.Username
		}
		rmsg.UserID = bot.ID
	}
	return nil
}

var (
	mentionRE        = regexp.MustCompile(`<@([a-zA-Z0-9]+)>`)
	channelRE        = regexp.MustCompile(`<#[a-zA-Z0-9]+\|(.+?)>`)
	variableRE       = regexp.MustCompile(`<!((?:subteam\^)?[a-zA-Z0-9]+)(?:\|@?(.+?))?>`)
	urlRE            = regexp.MustCompile(`<(.*?)(\|.*?)?>`)
	codeFenceRE      = regexp.MustCompile(`(?m)^` + "```" + `\w+$`)
	topicOrPurposeRE = regexp.MustCompile(`(?s)(@.+) (cleared|set)(?: the)? channel (topic|purpose)(?:: (.*))?`)
)

func (b *Bslack) extractTopicOrPurpose(text string) (string, string) {
	r := topicOrPurposeRE.FindStringSubmatch(text)
	if len(r) == 5 {
		action, updateType, extracted := r[2], r[3], r[4]
		switch action {
		case "set":
			return updateType, extracted
		case "cleared":
			return updateType, ""
		}
	}
	b.Log.Warnf("Encountered channel topic or purpose change message with unexpected format: %s", text)
	return "unknown", ""
}

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

func (b *Bslack) replaceCodeFence(text string) string {
	return codeFenceRE.ReplaceAllString(text, "```")
}

func (b *Bslack) handleRateLimit(err error) error {
	rateLimit, ok := err.(*slack.RateLimitedError)
	if !ok {
		return err
	}
	b.Log.Infof("Rate-limited by Slack. Sleeping for %v", rateLimit.RetryAfter)
	time.Sleep(rateLimit.RetryAfter)
	return nil
}

// getUsersInConversation returns an array of userIDs that are members of channelID
func (b *Bslack) getUsersInConversation(channelID string) ([]string, error) {
	channelMembers := []string{}
	for {
		queryParams := &slack.GetUsersInConversationParameters{
			ChannelID: channelID,
		}

		members, nextCursor, err := b.sc.GetUsersInConversation(queryParams)
		if err != nil {
			if err = b.handleRateLimit(err); err != nil {
				return channelMembers, fmt.Errorf("Could not retrieve users in channels: %#v", err)
			}
			continue
		}

		channelMembers = append(channelMembers, members...)

		if nextCursor == "" {
			break
		}
		queryParams.Cursor = nextCursor
	}
	return channelMembers, nil
}
