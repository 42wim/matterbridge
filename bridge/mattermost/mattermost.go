package bmattermost

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
	"strings"
)

//type Bridge struct {
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
	*config.Config
	Plus   bool
	Remote chan config.Message
}

type FancyLog struct {
	irc  *log.Entry
	mm   *log.Entry
	xmpp *log.Entry
}

var flog FancyLog

const Legacy = "legacy"

func init() {
	flog.irc = log.WithFields(log.Fields{"module": "irc"})
	flog.mm = log.WithFields(log.Fields{"module": "mattermost"})
	flog.xmpp = log.WithFields(log.Fields{"module": "xmpp"})
}

func New(cfg *config.Config, c chan config.Message) *Bmattermost {
	b := &Bmattermost{}
	b.Config = cfg
	b.Remote = c
	b.Plus = cfg.General.Plus
	b.mmMap = make(map[string]string)
	if !b.Plus {
		b.mh = matterhook.New(b.Config.Mattermost.URL,
			matterhook.Config{InsecureSkipVerify: b.Config.Mattermost.SkipTLSVerify,
				BindAddress: b.Config.Mattermost.BindAddress})
	} else {
		b.mc = matterclient.New(b.Config.Mattermost.Login, b.Config.Mattermost.Password,
			b.Config.Mattermost.Team, b.Config.Mattermost.Server)
		b.mc.SkipTLSVerify = b.Config.Mattermost.SkipTLSVerify
		b.mc.NoTLS = b.Config.Mattermost.NoTLS
		flog.mm.Infof("Trying login %s (team: %s) on %s", b.Config.Mattermost.Login, b.Config.Mattermost.Team, b.Config.Mattermost.Server)
		err := b.mc.Login()
		if err != nil {
			flog.mm.Fatal("Can not connect", err)
		}
		flog.mm.Info("Login ok")
		b.mc.JoinChannel(b.Config.Mattermost.Channel)
		for _, val := range b.Config.Channel {
			b.mc.JoinChannel(val.Mattermost)
		}
		go b.mc.WsReceiver()
	}
	go b.handleMatter()
	return b
}

func (b *Bmattermost) Command(cmd string) string {
	return ""
}

func (b *Bmattermost) Name() string {
	return "mattermost"
}

func (b *Bmattermost) Send(msg config.Message) error {
	flog.mm.Infof("mattermost send %#v", msg)
	if msg.Origin != "mattermost" {
		username := msg.Username + ": "
		if b.Config.Mattermost.RemoteNickFormat != "" {
			username = strings.Replace(b.Config.Mattermost.RemoteNickFormat, "{NICK}", msg.Username, -1)
		}
		return b.SendType(username, msg.Text, msg.Channel, "")
	}
	return nil
}

func (b *Bmattermost) SendType(nick string, message string, channel string, mtype string) error {
	if b.Config.Mattermost.PrefixMessagesWithNick {
		/*if IsMarkup(message) {
			message = nick + "\n\n" + message
		} else {
		*/
		message = nick + " " + message
		//}
	}
	if !b.Plus {
		matterMessage := matterhook.OMessage{IconURL: b.Config.Mattermost.IconURL}
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
	flog.mm.Infof("Choosing Mattermost connection type %s", b.Plus)
	mchan := make(chan *MMMessage)
	if b.Plus {
		go b.handleMatterClient(mchan)
	} else {
		go b.handleMatterHook(mchan)
	}
	flog.mm.Info("Start listening for Mattermost messages")
	for message := range mchan {
		/*
			if b.ignoreMessage(message.Username, message.Text, "mattermost") {
				continue
			}
		*/
		texts := strings.Split(message.Text, "\n")
		for _, text := range texts {
			flog.mm.Debug("Sending message from " + message.Username + " to " + message.Channel)
			b.Remote <- config.Message{Text: text, Username: message.Username, Channel: message.Channel, Origin: "mattermost"}
		}
	}
}

func (b *Bmattermost) handleMatterClient(mchan chan *MMMessage) {
	for message := range b.mc.MessageChan {
		// do not post our own messages back to irc
		if message.Raw.Action == "posted" && b.mc.User.Username != message.Username {
			flog.mm.Debugf("receiving from matterclient %#v", message)
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

func (b *Bmattermost) formatnicks(nicks []string, continued bool) string {
	switch b.Config.Mattermost.NickFormatter {
	case "table":
		return tableformatter(nicks, b.nicksPerRow(), continued)
	default:
		return plainformatter(nicks, b.nicksPerRow())
	}
}

func (b *Bmattermost) nicksPerRow() int {
	if b.Config.Mattermost.NicksPerRow < 1 {
		return 4
	}
	return b.Config.Mattermost.NicksPerRow
}
