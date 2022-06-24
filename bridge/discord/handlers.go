package bdiscord

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

func (b *Bdiscord) messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) { //nolint:unparam
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring messageDelete because it originates from a different guild")
		return
	}
	rmsg := config.Message{Account: b.Account, ID: m.ID, Event: config.EventMsgDelete, Text: config.EventMsgDelete}
	rmsg.Channel = b.getChannelName(m.ChannelID)

	b.Log.Debugf("<= Sending message from %s to gateway", b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

// TODO(qaisjp): if other bridges support bulk deletions, it could be fanned out centrally
func (b *Bdiscord) messageDeleteBulk(s *discordgo.Session, m *discordgo.MessageDeleteBulk) { //nolint:unparam
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring messageDeleteBulk because it originates from a different guild")
		return
	}
	for _, msgID := range m.Messages {
		rmsg := config.Message{
			Account: b.Account,
			ID:      msgID,
			Event:   config.EventMsgDelete,
			Text:    config.EventMsgDelete,
			Channel: b.getChannelName(m.ChannelID),
		}

		b.Log.Debugf("<= Sending message from %s to gateway", b.Account)
		b.Log.Debugf("<= Message is %#v", rmsg)
		b.Remote <- rmsg
	}
}

func (b *Bdiscord) messageEvent(s *discordgo.Session, m *discordgo.Event) {
	b.Log.Debug(spew.Sdump(m.Struct))
}

func (b *Bdiscord) messageTyping(s *discordgo.Session, m *discordgo.TypingStart) {
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring messageTyping because it originates from a different guild")
		return
	}
	if !b.GetBool("ShowUserTyping") {
		return
	}

	// Ignore our own typing messages
	if m.UserID == b.userID {
		return
	}

	rmsg := config.Message{Account: b.Account, Event: config.EventUserTyping}
	rmsg.Channel = b.getChannelName(m.ChannelID)
	b.Remote <- rmsg
}

func (b *Bdiscord) messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) { //nolint:unparam
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring messageUpdate because it originates from a different guild")
		return
	}
	if b.GetBool("EditDisable") {
		return
	}
	// only when message is actually edited
	if m.Message.EditedTimestamp != nil {
		b.Log.Debugf("Sending edit message")
		m.Content += b.GetString("EditSuffix")
		msg := &discordgo.MessageCreate{
			Message: m.Message,
		}
		b.messageCreate(s, msg)
	}
}

func (b *Bdiscord) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) { //nolint:unparam
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring messageCreate because it originates from a different guild")
		return
	}
	var err error

	// not relay our own messages
	if m.Author.Username == b.nick {
		return
	}
	// if using webhooks, do not relay if it's ours
	if m.Author.Bot && b.transmitter.HasWebhook(m.Author.ID) {
		return
	}

	// add the url of the attachments to content
	if len(m.Attachments) > 0 {
		for _, attach := range m.Attachments {
			m.Content = m.Content + "\n" + attach.URL
		}
	}

	rmsg := config.Message{Account: b.Account, Avatar: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".jpg", UserID: m.Author.ID, ID: m.ID}

	b.Log.Debugf("== Receiving event %#v", m.Message)

	if m.Content != "" {
		m.Message.Content = b.replaceChannelMentions(m.Message.Content)
		rmsg.Text, err = m.ContentWithMoreMentionsReplaced(b.c)
		if err != nil {
			b.Log.Errorf("ContentWithMoreMentionsReplaced failed: %s", err)
			rmsg.Text = m.ContentWithMentionsReplaced()
		}
	}

	// set channel name
	rmsg.Channel = b.getChannelName(m.ChannelID)

	fromWebhook := m.WebhookID != ""
	if !fromWebhook && !b.GetBool("UseUserName") {
		rmsg.Username = b.getNick(m.Author, m.GuildID)
	} else {
		rmsg.Username = m.Author.Username
		if !fromWebhook && b.GetBool("UseDiscriminator") {
			rmsg.Username += "#" + m.Author.Discriminator
		}
	}

	// if we have embedded content add it to text
	if b.GetBool("ShowEmbeds") && m.Message.Embeds != nil {
		for _, embed := range m.Message.Embeds {
			rmsg.Text += handleEmbed(embed)
		}
	}

	// no empty messages
	if rmsg.Text == "" {
		return
	}

	// do we have a /me action
	var ok bool
	rmsg.Text, ok = b.replaceAction(rmsg.Text)
	if ok {
		rmsg.Event = config.EventUserAction
	}

	// Replace emotes
	rmsg.Text = replaceEmotes(rmsg.Text)

	// Add our parent id if it exists, and if it's not referring to a message in another channel
	if ref := m.MessageReference; ref != nil && ref.ChannelID == m.ChannelID {
		rmsg.ParentID = ref.MessageID
	}

	b.Log.Debugf("<= Sending message from %s on %s to gateway", m.Author.Username, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

func (b *Bdiscord) memberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring memberUpdate because it originates from a different guild")
		return
	}
	if m.Member == nil {
		b.Log.Warnf("Received member update with no member information: %#v", m)
	}

	b.membersMutex.Lock()
	defer b.membersMutex.Unlock()

	if currMember, ok := b.userMemberMap[m.Member.User.ID]; ok {
		b.Log.Debugf(
			"%s: memberupdate: user %s (nick %s) changes nick to %s",
			b.Account,
			m.Member.User.Username,
			b.userMemberMap[m.Member.User.ID].Nick,
			m.Member.Nick,
		)
		delete(b.nickMemberMap, currMember.User.Username)
		delete(b.nickMemberMap, currMember.Nick)
		delete(b.userMemberMap, m.Member.User.ID)
	}
	b.userMemberMap[m.Member.User.ID] = m.Member
	b.nickMemberMap[m.Member.User.Username] = m.Member
	if m.Member.Nick != "" {
		b.nickMemberMap[m.Member.Nick] = m.Member
	}
}

func (b *Bdiscord) memberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring memberAdd because it originates from a different guild")
		return
	}
	if b.GetBool("nosendjoinpart") {
		return
	}
	if m.Member == nil {
		b.Log.Warnf("Received member update with no member information: %#v", m)
		return
	}
	username := m.Member.User.Username
	if m.Member.Nick != "" {
		username = m.Member.Nick
	}

	rmsg := config.Message{
		Account:  b.Account,
		Event:    config.EventJoinLeave,
		Username: "system",
		Text:     username + " joins",
	}
	b.Log.Debugf("<= Sending message from %s to gateway", b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

func (b *Bdiscord) memberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	if m.GuildID != b.guildID {
		b.Log.Debugf("Ignoring memberRemove because it originates from a different guild")
		return
	}
	if b.GetBool("nosendjoinpart") {
		return
	}
	if m.Member == nil {
		b.Log.Warnf("Received member update with no member information: %#v", m)
		return
	}
	username := m.Member.User.Username
	if m.Member.Nick != "" {
		username = m.Member.Nick
	}

	rmsg := config.Message{
		Account:  b.Account,
		Event:    config.EventJoinLeave,
		Username: "system",
		Text:     username + " leaves",
	}
	b.Log.Debugf("<= Sending message from %s to gateway", b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

func handleEmbed(embed *discordgo.MessageEmbed) string {
	var t []string
	var result string

	t = append(t, embed.Title)
	t = append(t, embed.Description)
	t = append(t, embed.URL)

	i := 0
	for _, e := range t {
		if e == "" {
			continue
		}

		i++
		if i == 1 {
			result += " embed: " + e
			continue
		}

		result += " - " + e
	}

	if result != "" {
		result += "\n"
	}

	return result
}
