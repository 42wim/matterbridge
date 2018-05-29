package bxmpp

import (
	"crypto/tls"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/jpillora/backoff"
	"github.com/rs/xid"
	"github.com/matterbridge/go-xmpp"
	"strings"
	"time"
)

type Bxmpp struct {
	xc      *xmpp.Client
	xmppMap map[string]string
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bxmpp{Config: cfg}
	b.xmppMap = make(map[string]string)
	return b
}

func (b *Bxmpp) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	b.xc, err = b.createXMPP()
	if err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}
	b.Log.Info("Connection succeeded")
	go func() {
		initial := true
		bf := &backoff.Backoff{
			Min:    time.Second,
			Max:    5 * time.Minute,
			Jitter: true,
		}
		for {
			if initial {
				b.handleXMPP()
				initial = false
			}
			d := bf.Duration()
			b.Log.Infof("Disconnected. Reconnecting in %s", d)
			time.Sleep(d)
			b.xc, err = b.createXMPP()
			if err == nil {
				b.Remote <- config.Message{Username: "system", Text: "rejoin", Channel: "", Account: b.Account, Event: config.EVENT_REJOIN_CHANNELS}
				b.handleXMPP()
				bf.Reset()
			}
		}
	}()
	return nil
}

func (b *Bxmpp) Disconnect() error {
	return nil
}

func (b *Bxmpp) JoinChannel(channel config.ChannelInfo) error {
	b.xc.JoinMUCNoHistory(channel.Name+"@"+b.GetString("Muc"), b.GetString("Nick"))
	return nil
}

func (b *Bxmpp) Send(msg config.Message) (string, error) {
	var msgid = ""
	var msgreplaceid = ""
	// ignore delete messages
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}
	b.Log.Debugf("=> Receiving %#v", msg)

	// Upload a file (in xmpp case send the upload URL because xmpp has no native upload support)
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.xc.Send(xmpp.Chat{Type: "groupchat", Remote: rmsg.Channel + "@" + b.GetString("Muc"), Text: rmsg.Username + rmsg.Text})
		}
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg)
		}
	}

	msgid = xid.New().String()
	if msg.ID != "" {
		msgid = msg.ID
		msgreplaceid = msg.ID
	}
	// Post normal message
	_, err := b.xc.Send(xmpp.Chat{Type: "groupchat", Remote: msg.Channel + "@" + b.GetString("Muc"), Text: msg.Username + msg.Text, ID: msgid, ReplaceID: msgreplaceid})
	if err != nil {
		return "", err
	}
	return msgid, nil
}

func (b *Bxmpp) createXMPP() (*xmpp.Client, error) {
	tc := new(tls.Config)
	tc.InsecureSkipVerify = b.GetBool("SkipTLSVerify")
	tc.ServerName = strings.Split(b.GetString("Server"), ":")[0]
	options := xmpp.Options{
		Host:                         b.GetString("Server"),
		User:                         b.GetString("Jid"),
		Password:                     b.GetString("Password"),
		NoTLS:                        true,
		StartTLS:                     true,
		TLSConfig:                    tc,
		Debug:                        b.GetBool("debug"),
		Logger:                       b.Log.Writer(),
		Session:                      true,
		Status:                       "",
		StatusMessage:                "",
		Resource:                     "",
		InsecureAllowUnencryptedAuth: false,
	}
	var err error
	b.xc, err = options.NewClient()
	return b.xc, err
}

func (b *Bxmpp) xmppKeepAlive() chan bool {
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

func (b *Bxmpp) handleXMPP() error {
	var ok bool
	var msgid string
	done := b.xmppKeepAlive()
	defer close(done)
	for {
		m, err := b.xc.Recv()
		if err != nil {
			return err
		}
		switch v := m.(type) {
		case xmpp.Chat:
			if v.Type == "groupchat" {
				b.Log.Debugf("== Receiving %#v", v)
				// skip invalid messages
				if b.skipMessage(v) {
					continue
				}
				msgid = v.ID
				if v.ReplaceID != "" {
					msgid = v.ReplaceID
				}
				rmsg := config.Message{Username: b.parseNick(v.Remote), Text: v.Text, Channel: b.parseChannel(v.Remote), Account: b.Account, UserID: v.Remote, ID: msgid}

				// check if we have an action event
				rmsg.Text, ok = b.replaceAction(rmsg.Text)
				if ok {
					rmsg.Event = config.EVENT_USER_ACTION
				}
				b.Log.Debugf("<= Sending message from %s on %s to gateway", rmsg.Username, b.Account)
				b.Log.Debugf("<= Message is %#v", rmsg)
				b.Remote <- rmsg
			}
		case xmpp.Presence:
			// do nothing
		}
	}
}

func (b *Bxmpp) replaceAction(text string) (string, bool) {
	if strings.HasPrefix(text, "/me ") {
		return strings.Replace(text, "/me ", "", -1), true
	}
	return text, false
}

// handleUploadFile handles native upload of files
func (b *Bxmpp) handleUploadFile(msg *config.Message) (string, error) {
	var urldesc = ""

	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if fi.Comment != "" {
			msg.Text += fi.Comment + ": "
		}
		if fi.URL != "" {
			msg.Text = fi.URL
			if fi.Comment != "" {
				msg.Text = fi.Comment + ": " + fi.URL
				urldesc = fi.Comment
			}
		}
		_, err := b.xc.Send(xmpp.Chat{Type: "groupchat", Remote: msg.Channel + "@" + b.GetString("Muc"), Text: msg.Username + msg.Text})
		if err != nil {
			return "", err
		}
		if fi.URL != "" {
			b.xc.SendOOB(xmpp.Chat{Type: "groupchat", Remote: msg.Channel + "@" + b.GetString("Muc"), Ooburl: fi.URL, Oobdesc: urldesc})
		}
	}
	return "", nil
}

func (b *Bxmpp) parseNick(remote string) string {
	s := strings.Split(remote, "@")
	if len(s) > 0 {
		s = strings.Split(s[1], "/")
		if len(s) == 2 {
			return s[1] // nick
		}
	}
	return ""
}

func (b *Bxmpp) parseChannel(remote string) string {
	s := strings.Split(remote, "@")
	if len(s) >= 2 {
		return s[0] // channel
	}
	return ""
}

// skipMessage skips messages that need to be skipped
func (b *Bxmpp) skipMessage(message xmpp.Chat) bool {
	// skip messages from ourselves
	if b.parseNick(message.Remote) == b.GetString("Nick") {
		return true
	}

	// skip empty messages
	if message.Text == "" {
		return true
	}

	// skip subject messages
	if strings.Contains(message.Text, "</subject>") {
		return true
	}

	// skip delayed messages
	t := time.Time{}
	return message.Stamp != t
}
