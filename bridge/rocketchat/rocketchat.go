package brocketchat

import (
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/hook/rockethook"
	"github.com/42wim/matterbridge/matterhook"
)

type MMhook struct {
	mh *matterhook.Client
	rh *rockethook.Client
}

type Brocketchat struct {
	MMhook
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Brocketchat{Config: cfg}
}

func (b *Brocketchat) Command(cmd string) string {
	return ""
}

func (b *Brocketchat) Connect() error {
	b.Log.Info("Connecting webhooks")
	b.mh = matterhook.New(b.GetString("WebhookURL"),
		matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
			DisableServer: true})
	b.rh = rockethook.New(b.GetString("WebhookURL"), rockethook.Config{BindAddress: b.GetString("WebhookBindAddress")})
	go b.handleRocketHook()
	return nil
}

func (b *Brocketchat) Disconnect() error {
	return nil

}

func (b *Brocketchat) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Brocketchat) Send(msg config.Message) (string, error) {
	// ignore delete messages
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}
	b.Log.Debugf("=> Receiving %#v", msg)
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			iconURL := config.GetIconURL(&rmsg, b.GetString("iconurl"))
			matterMessage := matterhook.OMessage{IconURL: iconURL, Channel: rmsg.Channel, UserName: rmsg.Username, Text: rmsg.Text}
			b.mh.Send(matterMessage)
		}
		if len(msg.Extra["file"]) > 0 {
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				if fi.URL != "" {
					msg.Text += fi.URL
				}
			}
		}
	}

	iconURL := config.GetIconURL(&msg, b.GetString("iconurl"))
	matterMessage := matterhook.OMessage{IconURL: iconURL}
	matterMessage.Channel = msg.Channel
	matterMessage.UserName = msg.Username
	matterMessage.Type = ""
	matterMessage.Text = msg.Text
	err := b.mh.Send(matterMessage)
	if err != nil {
		b.Log.Info(err)
		return "", err
	}
	return "", nil
}

func (b *Brocketchat) handleRocketHook() {
	for {
		message := b.rh.Receive()
		b.Log.Debugf("Receiving from rockethook %#v", message)
		// do not loop
		if message.UserName == b.GetString("Nick") {
			continue
		}
		b.Log.Debugf("<= Sending message from %s on %s to gateway", message.UserName, b.Account)
		b.Remote <- config.Message{Text: message.Text, Username: message.UserName, Channel: message.ChannelName, Account: b.Account, UserID: message.UserID}
	}
}
