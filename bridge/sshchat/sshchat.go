package bsshchat

import (
	"bufio"
	"io"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/shazow/ssh-chat/sshd"
	log "github.com/sirupsen/logrus"
)

type Bsshchat struct {
	r *bufio.Scanner
	w io.WriteCloser
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bsshchat{Config: cfg}
}

func (b *Bsshchat) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	go func() {
		err = sshd.ConnectShell(b.GetString("Server"), b.GetString("Nick"), func(r io.Reader, w io.WriteCloser) error {
			b.r = bufio.NewScanner(r)
			b.w = w
			b.r.Scan()
			w.Write([]byte("/theme mono\r\n"))
			b.handleSSHChat()
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
	if msg.Event == config.EventMsgDelete {
		return "", nil
	}
	b.Log.Debugf("=> Receiving %#v", msg)
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
					if fi.Comment != "" {
						msg.Text = fi.Comment + ": " + fi.URL
					}
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

func (b *Bsshchat) handleSSHChat() error {
	/*
		done := b.sshchatKeepAlive()
		defer close(done)
	*/
	wait := true
	for {
		if b.r.Scan() {
			// ignore messages from ourselves
			if !strings.Contains(b.r.Text(), "\033[K") {
				continue
			}
			res := strings.Split(stripPrompt(b.r.Text()), ":")
			if res[0] == "-> Set theme" {
				wait = false
				log.Debugf("mono found, allowing")
				continue
			}
			if !wait {
				b.Log.Debugf("<= Message %#v", res)
				rmsg := config.Message{Username: res[0], Text: strings.Join(res[1:], ":"), Channel: "sshchat", Account: b.Account, UserID: "nick"}
				b.Remote <- rmsg
			}
		}
	}
}
