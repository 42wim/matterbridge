package bslack

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// populateReceivedMessage shapes the initial Matterbridge message that we will forward to the
// router before we apply message-dependent modifications.
func (b *Bslack) populateReceivedMessage(ev *slack.MessageEvent) (*config.Message, error) {
	// Use our own func because rtm.GetChannelInfo doesn't work for private channels.
	channel, err := b.channels.getChannelByID(ev.Channel)
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

	user := b.users.getUser(userID)
	if user == nil {
		return fmt.Errorf("could not find information for user with id %s", ev.User)
	}

	rmsg.UserID = user.ID
	rmsg.Username = user.Name
	if user.Profile.DisplayName != "" {
		rmsg.Username = user.Profile.DisplayName
	}
	if b.GetBool("UseFullName") && user.Profile.RealName != "" {
		rmsg.Username = user.Profile.RealName
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

		if err = handleRateLimit(b.Log, err); err != nil {
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
	urlRE            = regexp.MustCompile(`<([^<\|]+)\|([^>]+)>`)
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
		if username := b.users.getUsername(userID); userID != "" {
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
	return urlRE.ReplaceAllString(text, "[${2}](${1})")
}

func (b *Bslack) replaceb0rkedMarkDown(text string) string {
	// taken from https://github.com/mattermost/mattermost-server/blob/master/app/slackimport.go
	//
	regexReplaceAllString := []struct {
		regex *regexp.Regexp
		rpl   string
	}{
		// bold
		{
			regexp.MustCompile(`(^|[\s.;,])\*(\S[^*\n]+)\*`),
			"$1**$2**",
		},
		// strikethrough
		{
			regexp.MustCompile(`(^|[\s.;,])\~(\S[^~\n]+)\~`),
			"$1~~$2~~",
		},
		// single paragraph blockquote
		// Slack converts > character to &gt;
		{
			regexp.MustCompile(`(?sm)^&gt;`),
			">",
		},
	}
	for _, rule := range regexReplaceAllString {
		text = rule.regex.ReplaceAllString(text, rule.rpl)
	}
	return text
}

func (b *Bslack) replaceCodeFence(text string) string {
	return codeFenceRE.ReplaceAllString(text, "```")
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
			if err = handleRateLimit(b.Log, err); err != nil {
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

func handleRateLimit(log *logrus.Entry, err error) error {
	rateLimit, ok := err.(*slack.RateLimitedError)
	if !ok {
		return err
	}
	log.Infof("Rate-limited by Slack. Sleeping for %v", rateLimit.RetryAfter)
	time.Sleep(rateLimit.RetryAfter)
	return nil
}
