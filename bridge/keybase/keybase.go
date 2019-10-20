package bkeybase

import (
	"strconv"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

// Bkeybase bridge structure
type Bkeybase struct {
	kbc     *kbchat.API
	user    string
	channel string
	team    string
	*bridge.Config
}

// New initializes Bkeybase object and sets team
func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bkeybase{Config: cfg}
	b.team = b.Config.GetString("Team")
	return b
}

// Connect starts keybase API and listener loop
func (b *Bkeybase) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Team"))

	// use default keybase location (`keybase`)
	b.kbc, err = kbchat.Start(kbchat.RunOptions{})
	if err != nil {
		return err
	}
	b.user = b.kbc.GetUsername()
	b.Log.Info("Connection succeeded")
	go b.handleKeybase()
	return nil
}

// Disconnect doesn't do anything for now
func (b *Bkeybase) Disconnect() error {
	return nil
}

// JoinChannel sets channel name in struct
func (b *Bkeybase) JoinChannel(channel config.ChannelInfo) error {
	if _, err := b.kbc.JoinChannel(b.team, channel.Name); err != nil {
		return err
	}
	b.channel = channel.Name
	return nil
}

// Send receives bridge messages and sends them to Keybase chat room
func (b *Bkeybase) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// Handle /me events
	if msg.Event == config.EventUserAction {
		msg.Text = "_" + msg.Text + "_"
	}

	// Delete message if we have an ID
	// Delete message not supported by keybase go library yet

	// Edit message if we have an ID
	// kbchat lib does not support message editing yet

	if len(msg.Extra["file"]) > 0 {
		// Upload a file
		dir, err := ioutil.TempDir("", "matterbridge")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(dir)
		fname := msg.Extra["file"][0].(config.FileInfo).Name
		fpath := filepath.Join(dir, msg.Extra["file"][0].(config.FileInfo).Name)
		if err := ioutil.WriteFile(fpath, *msg.Extra["file"][0].(config.FileInfo).Data, 0600); err != nil {
			return "", err
		}

		resp, err := b.kbc.SendAttachmentByTeam(b.team, fpath, fname, &b.channel)
		return strconv.Itoa(resp.Result.MsgID), err
	} else {
		// Send regular message
		resp, err := b.kbc.SendMessageByTeamName(b.team, msg.Username+msg.Text, &b.channel)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(resp.Result.MsgID), err
	}
}
