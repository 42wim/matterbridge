package bmattermost

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	log "github.com/Sirupsen/logrus"
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
	Remote   chan config.Message
	name     string
	origin   string
	protocol string
	TeamId   string
}

var flog *log.Entry
var protocol = "mattermost"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, origin string, c chan config.Message) *Bmattermost {
	b := &Bmattermost{}
	b.Config = &cfg
	b.origin = origin
	b.Remote = c
	b.protocol = "mattermost"
	b.name = cfg.Name
	b.mmMap = make(map[string]string)
	return b
}

func (b *Bmattermost) Command(cmd string) string {
	return ""
}

func (b *Bmattermost) Connect() error {
	if !b.Config.UseAPI {
		flog.Info("Connecting webhooks")
		b.mh = matterhook.New(b.Config.URL,
			matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
				BindAddress: b.Config.BindAddress})
	} else {
		b.mc = matterclient.New(b.Config.Login, b.Config.Password,
			b.Config.Team, b.Config.Server)
		b.mc.SkipTLSVerify = b.Config.SkipTLSVerify
		b.mc.NoTLS = b.Config.NoTLS
		flog.Infof("Connecting %s (team: %s) on %s", b.Config.Login, b.Config.Team, b.Config.Server)
		err := b.mc.Login()
		if err != nil {
			return err
		}
		flog.Info("Connection succeeded")
		b.TeamId = b.mc.GetTeamId()
		go b.mc.WsReceiver()
	}
	go b.handleMatter()
	return nil
}

func (b *Bmattermost) FullOrigin() string {
	return b.protocol + "." + b.origin
}

func (b *Bmattermost) JoinChannel(channel string) error {
	// we can only join channels using the API
	if b.Config.UseAPI {
		return b.mc.JoinChannel(b.mc.GetChannelId(channel, ""))
	}
	return nil
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
	flog.Debugf("Receiving %#v", msg)
	return b.SendType(msg.Username, msg.Text, msg.Channel, "")
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
	if !b.Config.UseAPI {
		matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL}
		matterMessage.Channel = channel
		matterMessage.UserName = nick
		matterMessage.Type = mtype
		matterMessage.Text = message
		err := b.mh.Send(matterMessage)
		if err != nil {
			flog.Info(err)
			return err
		}
		return nil
	}
	b.mc.PostMessage(b.mc.GetChannelId(channel, ""), message)
	return nil
}

func (b *Bmattermost) handleMatter() {
	flog.Debugf("Choosing API based Mattermost connection: %t", b.Config.UseAPI)
	mchan := make(chan *MMMessage)
	if b.Config.UseAPI {
		go b.handleMatterClient(mchan)
	} else {
		go b.handleMatterHook(mchan)
	}
	for message := range mchan {
		flog.Debugf("Sending message from %s on %s to gateway", message.Username, b.FullOrigin())
		b.Remote <- config.Message{Text: message.Text, Username: message.Username, Channel: message.Channel, Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
	}
}

func (b *Bmattermost) handleMatterClient(mchan chan *MMMessage) {
	for message := range b.mc.MessageChan {
		// do not post our own messages back to irc
		// only listen to message from our team
		if message.Raw.Event == "posted" && b.mc.User.Username != message.Username && message.Raw.TeamId == b.TeamId {
			flog.Debugf("Receiving from matterclient %#v", message)
			m := &MMMessage{}
			m.Username = message.Username
			m.Channel = message.Channel
			m.Text = message.Text
			if len(message.Post.Filenames) > 0 {
				for _, link := range b.mc.GetPublicLinks(message.Post.Filenames) {
					m.Text = m.Text + "\n" + link
				}
			}
			mchan <- m
		}
	}
}

func (b *Bmattermost) handleMatterHook(mchan chan *MMMessage) {
	for {
		message := b.mh.Receive()
		flog.Debugf("Receiving from matterhook %#v", message)
		m := &MMMessage{}
		m.Username = message.UserName
		m.Text = message.Text
		m.Channel = message.ChannelName
		mchan <- m
	}
}
