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
	*config.BridgeConfig
}

func New(cfg *config.BridgeConfig) bridge.Bridger {
	return &Brocketchat{BridgeConfig: cfg}
}

func (b *Brocketchat) Command(cmd string) string {
	return ""
}

func (b *Brocketchat) Connect() error {
	b.Log.Info("Connecting webhooks")
	b.mh = matterhook.New(b.Config.WebhookURL,
		matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
			DisableServer: true})
	b.rh = rockethook.New(b.Config.WebhookURL, rockethook.Config{BindAddress: b.Config.WebhookBindAddress})
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
	b.Log.Debugf("Receiving %#v", msg)
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL, Channel: rmsg.Channel, UserName: rmsg.Username,
				Text: rmsg.Text}
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

	matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL}
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
		if message.UserName == b.Config.Nick {
			continue
		}
		b.Log.Debugf("Sending message from %s on %s to gateway", message.UserName, b.Account)
		b.Remote <- config.Message{Text: message.Text, Username: message.UserName, Channel: message.ChannelName, Account: b.Account, UserID: message.UserID}
	}
}
