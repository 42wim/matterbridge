package bsshchat

import (
	"bufio"
	"io"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/shazow/ssh-chat/sshd"
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
	b.Log.Infof("Connecting %s", b.GetString("Server"))

	// connHandler will be called by 'sshd.ConnectShell()' below
	// once the connection is established in order to handle it.
	connErr := make(chan error, 1) // Needs to be buffered.
	connSignal := make(chan struct{})
	connHandler := func(r io.Reader, w io.WriteCloser) error {
		b.r = bufio.NewScanner(r)
		b.r.Scan()
		b.w = w
		if _, err := b.w.Write([]byte("/theme mono\r\n/quiet\r\n")); err != nil {
			return err
		}
		close(connSignal) // Connection is established so we can signal the success.
		return b.handleSSHChat()
	}

	go func() {
		// As a successful connection will result in this returning after the Connection
		// method has already returned point we NEED to have a buffered channel to still
		// be able to write.
		connErr <- sshd.ConnectShell(b.GetString("Server"), b.GetString("Nick"), connHandler)
	}()

	select {
	case err := <-connErr:
		b.Log.Error("Connection failed")
		return err
	case <-connSignal:
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
			if _, err := b.w.Write([]byte(rmsg.Username + rmsg.Text + "\r\n")); err != nil {
				b.Log.Errorf("Could not send extra message: %#v", err)
			}
		}
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg)
		}
	}
	_, err := b.w.Write([]byte(msg.Username + msg.Text + "\r\n"))
	return "", err
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
			if strings.Contains(b.r.Text(), "Rate limiting is in effect") {
				continue
			}
			// skip our own messages
			if !strings.HasPrefix(b.r.Text(), "["+b.GetString("Nick")+"] \x1b") {
				continue
			}
			res := strings.Split(stripPrompt(b.r.Text()), ":")
			if res[0] == "-> Set theme" {
				wait = false
				b.Log.Debugf("mono found, allowing")
				continue
			}
			if !wait {
				b.Log.Debugf("<= Message %#v", res)
				rmsg := config.Message{Username: res[0], Text: strings.TrimSpace(strings.Join(res[1:], ":")), Channel: "sshchat", Account: b.Account, UserID: "nick"}
				b.Remote <- rmsg
			}
		}
	}
}

func (b *Bsshchat) handleUploadFile(msg *config.Message) (string, error) {
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
		if _, err := b.w.Write([]byte(msg.Username + msg.Text + "\r\n")); err != nil {
			b.Log.Errorf("Could not send file message: %#v", err)
		}
	}
	return "", nil
}
