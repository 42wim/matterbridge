package bsshchat

import (
	"bufio"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/shazow/ssh-chat/sshd"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
)

type Bsshchat struct {
	r *bufio.Scanner
	w io.WriteCloser
	*config.BridgeConfig
}

func New(cfg *config.BridgeConfig) bridge.Bridger {
	return &Bsshchat{BridgeConfig: cfg}
}

func (b *Bsshchat) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.Config.Server)
	go func() {
		err = sshd.ConnectShell(b.Config.Server, b.Config.Nick, func(r io.Reader, w io.WriteCloser) error {
			b.r = bufio.NewScanner(r)
			b.w = w
			b.r.Scan()
			w.Write([]byte("/theme mono\r\n"))
			b.handleSshChat()
			return nil
		})
	}()
	if err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}
	b.Log.Info("Connection succeeded")
	return nil
}

func (b *Bsshchat) Disconnect() error {
	return nil
}

func (b *Bsshchat) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Bsshchat) Send(msg config.Message) (string, error) {
	// ignore delete messages
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}
	b.Log.Debugf("Receiving %#v", msg)
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.w.Write([]byte(rmsg.Username + rmsg.Text + "\r\n"))
		}
		if len(msg.Extra["file"]) > 0 {
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				if fi.Comment != "" {
					msg.Text += fi.Comment + ": "
				}
				if fi.URL != "" {
					msg.Text = fi.URL
				}
				b.w.Write([]byte(msg.Username + msg.Text))
			}
			return "", nil
		}
	}
	b.w.Write([]byte(msg.Username + msg.Text + "\r\n"))
	return "", nil
}

/*
func (b *Bsshchat) sshchatKeepAlive() chan bool {
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(90 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				b.Log.Debugf("PING")
				err := b.xc.PingC2S("", "")
				if err != nil {
					b.Log.Debugf("PING failed %#v", err)
				}
			case <-done:
				return
			}
		}
	}()
	return done
}
*/

func stripPrompt(s string) string {
	pos := strings.LastIndex(s, "\033[K")
	if pos < 0 {
		return s
	}
	return s[pos+3:]
}

func (b *Bsshchat) handleSshChat() error {
	/*
		done := b.sshchatKeepAlive()
		defer close(done)
	*/
	wait := true
	for {
		if b.r.Scan() {
			res := strings.Split(stripPrompt(b.r.Text()), ":")
			if res[0] == "-> Set theme" {
				wait = false
				log.Debugf("mono found, allowing")
				continue
			}
			if !wait {
				b.Log.Debugf("message %#v", res)
				rmsg := config.Message{Username: res[0], Text: strings.Join(res[1:], ":"), Channel: "sshchat", Account: b.Account, UserID: "nick"}
				b.Remote <- rmsg
			}
		}
	}
}
