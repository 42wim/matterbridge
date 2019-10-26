package bdiscord

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/bwmarrin/discordgo"
)

func (b *Bdiscord) messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) { //nolint:unparam
	rmsg := config.Message{Account: b.Account, ID: m.ID, Event: config.EventMsgDelete, Text: config.EventMsgDelete}
	rmsg.Channel = b.getChannelName(m.ChannelID)
	if b.useChannelID {
		rmsg.Channel = "ID:" + m.ChannelID
	}
	b.Log.Debugf("<= Sending message from %s to gateway", b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

// TODO(qaisjp): if other bridges support bulk deletions, it could be fanned out centrally
func (b *Bdiscord) messageDeleteBulk(s *discordgo.Session, m *discordgo.MessageDeleteBulk) { //nolint:unparam
	for _, msgID := range m.Messages {
		rmsg := config.Message{
			Account: b.Account,
			ID:      msgID,
			Event:   config.EventMsgDelete,
			Text:    config.EventMsgDelete,
			Channel: "ID:" + m.ChannelID,
		}

		if !b.useChannelID {
			rmsg.Channel = b.getChannelName(m.ChannelID)
		}

		b.Log.Debugf("<= Sending message from %s to gateway", b.Account)
		b.Log.Debugf("<= Message is %#v", rmsg)
		b.Remote <- rmsg
	}
}

func (b *Bdiscord) messageTyping(s *discordgo.Session, m *discordgo.TypingStart) {
	if !b.GetBool("ShowUserTyping") {
		return
	}

	rmsg := config.Message{Account: b.Account, Event: config.EventUserTyping}
	rmsg.Channel = b.getChannelName(m.ChannelID)
	if b.useChannelID {
		rmsg.Channel = "ID:" + m.ChannelID
	}
	b.Remote <- rmsg
}

func (b *Bdiscord) messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) { //nolint:unparam
	if b.GetBool("EditDisable") {
		return
	}
	// only when message is actually edited
	if m.Message.EditedTimestamp != "" {
		b.Log.Debugf("Sending edit message")
		m.Content += b.GetString("EditSuffix")
		msg := &discordgo.MessageCreate{
			Message: m.Message,
		}
		b.messageCreate(s, msg)
	}
}

func (b *Bdiscord) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) { //nolint:unparam
	var err error

	// not relay our own messages
	if m.Author.Username == b.nick {
		return
	}
	// if using webhooks, do not relay if it's ours
	if b.useWebhook() && m.Author.Bot && b.isWebhookID(m.Author.ID) {
		return
	}

	// add the url of the attachments to content
	if len(m.Attachments) > 0 {
		for _, attach := range m.Attachments {
			m.Content = m.Content + "\n" + attach.URL
		}
	}

	rmsg := config.Message{Account: b.Account, Avatar: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".jpg", UserID: m.Author.ID, ID: m.ID}

	if m.Content != "" {
		b.Log.Debugf("== Receiving event %#v", m.Message)
		m.Message.Content = b.stripCustomoji(m.Message.Content)
		m.Message.Content = b.replaceChannelMentions(m.Message.Content)
		rmsg.Text, err = m.ContentWithMoreMentionsReplaced(b.c)
		if err != nil {
			b.Log.Errorf("ContentWithMoreMentionsReplaced failed: %s", err)
			rmsg.Text = m.ContentWithMentionsReplaced()
		}
	}

	// set channel name
	rmsg.Channel = b.getChannelName(m.ChannelID)
	if b.useChannelID {
		rmsg.Channel = "ID:" + m.ChannelID
	}

	// set username
	if !b.GetBool("UseUserName") {
		rmsg.Username = b.getNick(m.Author, m.GuildID)
	} else {
		rmsg.Username = m.Author.Username
		if b.GetBool("UseDiscriminator") {
			rmsg.Username += "#" + m.Author.Discriminator
		}
	}

	// if we have embedded content add it to text
	if b.GetBool("ShowEmbeds") && m.Message.Embeds != nil {
		for _, embed := range m.Message.Embeds {
			rmsg.Text = rmsg.Text + "embed: " + embed.Title + " - " + embed.Description + " - " + embed.URL + "\n"
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

	b.Log.Debugf("<= Sending message from %s on %s to gateway", m.Author.Username, b.Account)
	b.Log.Debugf("<= Message is %#v", rmsg)
	b.Remote <- rmsg
}

func (b *Bdiscord) memberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
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
