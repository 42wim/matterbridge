package bmattermost

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"strings"
)

type MMhook struct {
	mh *matterhook.Client
}

type MMapi struct {
	mc            *matterclient.MMClient
	mmMap         map[string]string
	mmIgnoreNicks []string
}

type MMMessage struct {
	Text     string
	Channel  string
	Username string
}

type Bmattermost struct {
	MMhook
	MMapi
	Config   *config.Protocol
	Plus     bool
	Remote   chan config.Message
	name     string
	origin   string
	protocol string
}

type FancyLog struct {
	mm *log.Entry
}

var flog FancyLog

const Legacy = "legacy"

func init() {
	flog.mm = log.WithFields(log.Fields{"module": "mattermost"})
}

func New(cfg config.Protocol, origin string, c chan config.Message) *Bmattermost {
	b := &Bmattermost{}
	b.Config = &cfg
	b.origin = origin
	b.Remote = c
	b.protocol = "mattermost"
	b.name = cfg.Name
	b.Plus = cfg.UseAPI
	b.mmMap = make(map[string]string)
	return b
}

func (b *Bmattermost) Command(cmd string) string {
	return ""
}

func (b *Bmattermost) Connect() error {
	if !b.Plus {
		b.mh = matterhook.New(b.Config.URL,
			matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
				BindAddress: b.Config.BindAddress})
	} else {
		b.mc = matterclient.New(b.Config.Login, b.Config.Password,
			b.Config.Team, b.Config.Server)
		b.mc.SkipTLSVerify = b.Config.SkipTLSVerify
		b.mc.NoTLS = b.Config.NoTLS
		flog.mm.Infof("Trying login %s (team: %s) on %s", b.Config.Login, b.Config.Team, b.Config.Server)
		err := b.mc.Login()
		if err != nil {
			return err
		}
		flog.mm.Info("Login ok")
		/*
			b.mc.JoinChannel(b.Config.Channel)
			for _, val := range b.Config.Channel {
				b.mc.JoinChannel(val.Mattermost)
			}
		*/
		go b.mc.WsReceiver()
	}
	go b.handleMatter()
	return nil
}

func (b *Bmattermost) FullOrigin() string {
	return b.protocol + "." + b.origin
}

func (b *Bmattermost) JoinChannel(channel string) error {
	return b.mc.JoinChannel(channel)
}

func (b *Bmattermost) Name() string {
	return b.protocol + "." + b.origin
}

func (b *Bmattermost) Origin() string {
	return b.origin
}

func (b *Bmattermost) Protocol() string {
	return b.protocol
}

func (b *Bmattermost) Send(msg config.Message) error {
	flog.mm.Infof("mattermost send %#v", msg)
	if msg.Origin != b.origin {
		return b.SendType(msg.Username, msg.Text, msg.Channel, "")
	}
	return nil
}

func (b *Bmattermost) SendType(nick string, message string, channel string, mtype string) error {
	if b.Config.PrefixMessagesWithNick {
		/*if IsMarkup(message) {
			message = nick + "\n\n" + message
		} else {
		*/
		message = nick + " " + message
		//}
	}
	if !b.Plus {
		matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL}
		matterMessage.Channel = channel
		matterMessage.UserName = nick
		matterMessage.Type = mtype
		matterMessage.Text = message
		err := b.mh.Send(matterMessage)
		if err != nil {
			flog.mm.Info(err)
			return err
		}
		flog.mm.Debug("->mattermost channel: ", channel, " ", message)
		return nil
	}
	flog.mm.Debug("->mattermost channel plus: ", channel, " ", message)
	b.mc.PostMessage(b.mc.GetChannelId(channel, ""), message)
	return nil
}

func (b *Bmattermost) handleMatter() {
	flog.mm.Infof("Choosing API based Mattermost connection: %t", b.Plus)
	mchan := make(chan *MMMessage)
	if b.Plus {
		go b.handleMatterClient(mchan)
	} else {
		go b.handleMatterHook(mchan)
	}
	flog.mm.Info("Start listening for Mattermost messages")
	for message := range mchan {
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			flog.mm.Debug("Sending message from " + message.Username + " to " + message.Channel)
			b.Remote <- config.Message{Text: text, Username: message.Username, Channel: message.Channel, Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
		}
	}
}

func (b *Bmattermost) handleMatterClient(mchan chan *MMMessage) {
	for message := range b.mc.MessageChan {
		// do not post our own messages back to irc
		if message.Raw.Event == "posted" && b.mc.User.Username != message.Username {
			flog.mm.Debugf("receiving from matterclient %#v", message)
			flog.mm.Debugf("receiving from matterclient %#v", message.Raw)
			m := &MMMessage{}
			m.Username = message.Username
			m.Channel = message.Channel
			m.Text = message.Text
			mchan <- m
		}
	}
}

func (b *Bmattermost) handleMatterHook(mchan chan *MMMessage) {
	for {
		message := b.mh.Receive()
		flog.mm.Debugf("receiving from matterhook %#v", message)
		m := &MMMessage{}
		m.Username = message.UserName
		m.Text = message.Text
		m.Channel = message.ChannelName
		mchan <- m
	}
}
