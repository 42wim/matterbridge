package harmony

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/harmony-development/shibshib"
	chatv1 "github.com/harmony-development/shibshib/gen/chat/v1"
	typesv1 "github.com/harmony-development/shibshib/gen/harmonytypes/v1"
	profilev1 "github.com/harmony-development/shibshib/gen/profile/v1"
)

type cachedProfile struct {
	data        *profilev1.GetProfileResponse
	lastUpdated time.Time
}

type Bharmony struct {
	*bridge.Config

	c            *shibshib.Client
	profileCache map[uint64]cachedProfile
}

func uToStr(in uint64) string {
	return strconv.FormatUint(in, 10)
}

func strToU(in string) (uint64, error) {
	return strconv.ParseUint(in, 10, 64)
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bharmony{
		Config:       cfg,
		profileCache: map[uint64]cachedProfile{},
	}

	return b
}

func (b *Bharmony) getProfile(u uint64) (*profilev1.GetProfileResponse, error) {
	if v, ok := b.profileCache[u]; ok && time.Since(v.lastUpdated) < time.Minute*10 {
		return v.data, nil
	}

	resp, err := b.c.ProfileKit.GetProfile(&profilev1.GetProfileRequest{
		UserId: u,
	})
	if err != nil {
		if v, ok := b.profileCache[u]; ok {
			return v.data, nil
		}
		return nil, err
	}
	b.profileCache[u] = cachedProfile{
		data:        resp,
		lastUpdated: time.Now(),
	}
	return resp, nil
}

func (b *Bharmony) avatarFor(m *chatv1.Message) string {
	if m.Overrides != nil {
		return m.Overrides.GetAvatar()
	}

	profi, err := b.getProfile(m.AuthorId)
	if err != nil {
		return ""
	}

	return b.c.TransformHMCURL(profi.Profile.GetUserAvatar())
}

func (b *Bharmony) usernameFor(m *chatv1.Message) string {
	if m.Overrides != nil {
		return m.Overrides.GetUsername()
	}

	profi, err := b.getProfile(m.AuthorId)
	if err != nil {
		return ""
	}

	return profi.Profile.UserName
}

func (b *Bharmony) toMessage(msg *shibshib.LocatedMessage) config.Message {
	message := config.Message{}
	message.Account = b.Account
	message.UserID = uToStr(msg.Message.AuthorId)
	message.Avatar = b.avatarFor(msg.Message)
	message.Username = b.usernameFor(msg.Message)
	message.Channel = uToStr(msg.ChannelID)
	message.ID = uToStr(msg.MessageId)

	switch content := msg.Message.Content.Content.(type) {
	case *chatv1.Content_EmbedMessage:
		message.Text = "Embed"
	case *chatv1.Content_AttachmentMessage:
		var s strings.Builder
		for idx, attach := range content.AttachmentMessage.Files {
			s.WriteString(b.c.TransformHMCURL(attach.Id))
			if idx < len(content.AttachmentMessage.Files)-1 {
				s.WriteString(", ")
			}
		}
		message.Text = s.String()
	case *chatv1.Content_PhotoMessage:
		var s strings.Builder
		for idx, attach := range content.PhotoMessage.GetPhotos() {
			s.WriteString(attach.GetCaption().GetText())
			s.WriteString("\n")
			s.WriteString(b.c.TransformHMCURL(attach.GetHmc()))
			if idx < len(content.PhotoMessage.GetPhotos())-1 {
				s.WriteString("\n\n")
			}
		}
		message.Text = s.String()
	case *chatv1.Content_TextMessage:
		message.Text = content.TextMessage.Content.Text
	}

	return message
}

func (b *Bharmony) outputMessages() {
	for {
		msg := <-b.c.EventsStream()

		if msg.Message.AuthorId == b.c.UserID {
			continue
		}

		b.Remote <- b.toMessage(msg)
	}
}

func (b *Bharmony) GetUint64(conf string) uint64 {
	num, err := strToU(b.GetString(conf))
	if err != nil {
		log.Fatal(err)
	}

	return num
}

func (b *Bharmony) Connect() (err error) {
	b.c, err = shibshib.NewClient(b.GetString("Homeserver"), b.GetString("Token"), b.GetUint64("UserID"))
	if err != nil {
		return
	}
	b.c.SubscribeToGuild(b.GetUint64("Community"))

	go b.outputMessages()

	return nil
}

func (b *Bharmony) send(msg config.Message) (id string, err error) {
	msgChan, err := strToU(msg.Channel)
	if err != nil {
		return
	}

	retID, err := b.c.ChatKit.SendMessage(&chatv1.SendMessageRequest{
		GuildId:   b.GetUint64("Community"),
		ChannelId: msgChan,
		Content: &chatv1.Content{
			Content: &chatv1.Content_TextMessage{
				TextMessage: &chatv1.Content_TextContent{
					Content: &chatv1.FormattedText{
						Text: msg.Text,
					},
				},
			},
		},
		Overrides: &chatv1.Overrides{
			Username: &msg.Username,
			Avatar:   &msg.Avatar,
			Reason:   &chatv1.Overrides_Bridge{Bridge: &typesv1.Empty{}},
		},
		InReplyTo: nil,
		EchoId:    nil,
		Metadata:  nil,
	})
	if err != nil {
		err = fmt.Errorf("send: error sending message: %w", err)
		log.Println(err.Error())
	}

	return uToStr(retID.MessageId), err
}

func (b *Bharmony) delete(msg config.Message) (id string, err error) {
	msgChan, err := strToU(msg.Channel)
	if err != nil {
		return "", err
	}

	msgID, err := strToU(msg.ID)
	if err != nil {
		return "", err
	}

	_, err = b.c.ChatKit.DeleteMessage(&chatv1.DeleteMessageRequest{
		GuildId:   b.GetUint64("Community"),
		ChannelId: msgChan,
		MessageId: msgID,
	})
	return "", err
}

func (b *Bharmony) typing(msg config.Message) (id string, err error) {
	msgChan, err := strToU(msg.Channel)
	if err != nil {
		return "", err
	}

	_, err = b.c.ChatKit.Typing(&chatv1.TypingRequest{
		GuildId:   b.GetUint64("Community"),
		ChannelId: msgChan,
	})
	return "", err
}

func (b *Bharmony) Send(msg config.Message) (id string, err error) {
	switch msg.Event {
	case "":
		return b.send(msg)
	case config.EventMsgDelete:
		return b.delete(msg)
	case config.EventUserTyping:
		return b.typing(msg)
	default:
		return "", nil
	}
}

func (b *Bharmony) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Bharmony) Disconnect() error {
	return nil
}
