package harmony

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/harmony-development/shibshib"
	chatv1 "github.com/harmony-development/shibshib/gen/chat/v1"
	types "github.com/harmony-development/shibshib/gen/harmonytypes/v1"
)

type Bharmony struct {
	*bridge.Config

	c *shibshib.Client
}

func uToStr(in uint64) string {
	return strconv.FormatUint(in, 10)
}

func strToU(in string) (uint64, error) {
	return strconv.ParseUint(in, 10, 64)
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bharmony{Config: cfg}

	return b
}

func (b *Bharmony) outputMessages() {
	for {
		msg := <-b.c.EventsStream()

		if msg.AuthorId == b.c.UserID {
			continue
		}

		m := config.Message{}
		m.Account = b.Account
		m.UserID = uToStr(msg.AuthorId)
		m.Avatar = b.c.AvatarFor(msg)
		m.Username = b.c.UsernameFor(msg)
		m.Channel = uToStr(msg.ChannelId)
		m.ID = uToStr(msg.MessageId)

		switch e := msg.Content.Content.(type) {
		case *types.Content_EmbedMessage:
			m.Text = "Embed"
		case *types.Content_FilesMessage:
			var s strings.Builder
			s.WriteString("Files: ")
			for idx, attach := range e.FilesMessage.Attachments {
				s.WriteString(b.c.TransformHMCURL(attach.Id))
				if idx < len(e.FilesMessage.Attachments)-1 {
					s.WriteString(", ")
				}
			}
			m.Text = s.String()
		case *types.Content_TextMessage:
			m.Text = e.TextMessage.Content
		}

		b.Remote <- m
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

	if val, ok := msg.Extra["workaround"]; ok {
		msg.Username = val[0].(string)
	}

	retID, err := b.c.ChatKit.SendMessage(&chatv1.SendMessageRequest{
		GuildId:   b.GetUint64("Community"),
		ChannelId: msgChan,
		Content: &types.Content{
			Content: &types.Content_TextMessage{
				TextMessage: &types.ContentText{
					Content: msg.Text,
				},
			},
		},
		Overrides: &types.Override{
			Name:   msg.Username,
			Avatar: msg.Avatar,
			Reason: &types.Override_Bridge{Bridge: &emptypb.Empty{}},
		},
		InReplyTo: 0,
		EchoId:    0,
		Metadata:  nil,
	})

	return uToStr(retID.MessageId), fmt.Errorf("send: error sending message: %w", err)
}

func (b *Bharmony) Send(msg config.Message) (id string, err error) {
	switch msg.Event {
	case "":
		return b.send(msg)
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
