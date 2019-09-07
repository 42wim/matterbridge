package bxmpp

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/jpillora/backoff"
	"github.com/matterbridge/go-xmpp"
	"github.com/rs/xid"
)

type Bxmpp struct {
	*bridge.Config

	startTime time.Time
	xc        *xmpp.Client
	xmppMap   map[string]string
	connected bool
	sync.RWMutex
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bxmpp{
		Config:  cfg,
		xmppMap: make(map[string]string),
	}
}

func (b *Bxmpp) Connect() error {
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	if err := b.createXMPP(); err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}

	b.Log.Info("Connection succeeded")
	go b.manageConnection()
	return nil
}

func (b *Bxmpp) Disconnect() error {
	return nil
}

func (b *Bxmpp) JoinChannel(channel config.ChannelInfo) error {
	if channel.Options.Key != "" {
		b.Log.Debugf("using key %s for channel %s", channel.Options.Key, channel.Name)
		b.xc.JoinProtectedMUC(channel.Name+"@"+b.GetString("Muc"), b.GetString("Nick"), channel.Options.Key, xmpp.NoHistory, 0, nil)
	} else {
		b.xc.JoinMUCNoHistory(channel.Name+"@"+b.GetString("Muc"), b.GetString("Nick"))
	}
	return nil
}

func (b *Bxmpp) Send(msg config.Message) (string, error) {
	// should be fixed by using a cache instead of dropping
	if !b.Connected() {
		return "", fmt.Errorf("bridge %s not connected, dropping message %#v to bridge", b.Account, msg)
	}
	// ignore delete messages
	if msg.Event == config.EventMsgDelete {
		return "", nil
	}
	b.Log.Debugf("=> Receiving %#v", msg)

	// Upload a file (in XMPP case send the upload URL because XMPP has no native upload support).
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.Log.Debugf("=> Sending attachement message %#v", rmsg)
			if _, err := b.xc.Send(xmpp.Chat{
				Type:   "groupchat",
				Remote: rmsg.Channel + "@" + b.GetString("Muc"),
				Text:   rmsg.Username + rmsg.Text,
			}); err != nil {
				b.Log.WithError(err).Error("Unable to send message with share URL.")
			}
		}
		if len(msg.Extra["file"]) > 0 {
			return "", b.handleUploadFile(&msg)
		}
	}

	var msgReplaceID string
	msgID := xid.New().String()
	if msg.ID != "" {
		msgID = msg.ID
		msgReplaceID = msg.ID
	}
	// Post normal message.
	b.Log.Debugf("=> Sending message %#v", msg)
	if _, err := b.xc.Send(xmpp.Chat{
		Type:      "groupchat",
		Remote:    msg.Channel + "@" + b.GetString("Muc"),
		Text:      msg.Username + msg.Text,
		ID:        msgID,
		ReplaceID: msgReplaceID,
	}); err != nil {
		return "", err
	}
	return msgID, nil
}

func (b *Bxmpp) createXMPP() error {
	if !strings.Contains(b.GetString("Jid"), "@") {
		return fmt.Errorf("the Jid %s doesn't contain an @", b.GetString("Jid"))
	}
	tc := &tls.Config{
		ServerName:         strings.Split(b.GetString("Jid"), "@")[1],
		InsecureSkipVerify: b.GetBool("SkipTLSVerify"), // nolint: gosec
	}
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
	return err
}

func (b *Bxmpp) manageConnection() {
	b.setConnected(true)
	initial := true
	bf := &backoff.Backoff{
		Min:    time.Second,
		Max:    5 * time.Minute,
		Jitter: true,
	}

	// Main connection loop. Each iteration corresponds to a successful
	// connection attempt and the subsequent handling of the connection.
	for {
		if initial {
			initial = false
		} else {
			b.Remote <- config.Message{
				Username: "system",
				Text:     "rejoin",
				Channel:  "",
				Account:  b.Account,
				Event:    config.EventRejoinChannels,
			}
		}

		if err := b.handleXMPP(); err != nil {
			b.Log.WithError(err).Error("Disconnected.")
			b.setConnected(false)
		}

		// Reconnection loop using an exponential back-off strategy. We
		// only break out of the loop if we have successfully reconnected.
		for {
			d := bf.Duration()
			b.Log.Infof("Reconnecting in %s.", d)
			time.Sleep(d)

			b.Log.Infof("Reconnecting now.")
			if err := b.createXMPP(); err == nil {
				b.setConnected(true)
				bf.Reset()
				break
			}
			b.Log.Warn("Failed to reconnect.")
		}
	}
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
				if err := b.xc.PingC2S("", ""); err != nil {
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
	b.startTime = time.Now()

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

				// Skip invalid messages.
				if b.skipMessage(v) {
					continue
				}

				var event string
				if strings.Contains(v.Text, "has set the subject to:") {
					event = config.EventTopicChange
				}

				msgID := v.ID
				if v.ReplaceID != "" {
					msgID = v.ReplaceID
				}
				rmsg := config.Message{
					Username: b.parseNick(v.Remote),
					Text:     v.Text,
					Channel:  b.parseChannel(v.Remote),
					Account:  b.Account,
					UserID:   v.Remote,
					ID:       msgID,
					Event:    event,
				}

				// Check if we have an action event.
				var ok bool
				rmsg.Text, ok = b.replaceAction(rmsg.Text)
				if ok {
					rmsg.Event = config.EventUserAction
				}

				b.Log.Debugf("<= Sending message from %s on %s to gateway", rmsg.Username, b.Account)
				b.Log.Debugf("<= Message is %#v", rmsg)
				b.Remote <- rmsg
			}
		case xmpp.Presence:
			// Do nothing.
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
func (b *Bxmpp) handleUploadFile(msg *config.Message) error {
	var urlDesc string

	for _, file := range msg.Extra["file"] {
		fileInfo := file.(config.FileInfo)
		if fileInfo.Comment != "" {
			msg.Text += fileInfo.Comment + ": "
		}
		if fileInfo.URL != "" {
			msg.Text = fileInfo.URL
			if fileInfo.Comment != "" {
				msg.Text = fileInfo.Comment + ": " + fileInfo.URL
				urlDesc = fileInfo.Comment
			}
		}
		if _, err := b.xc.Send(xmpp.Chat{
			Type:   "groupchat",
			Remote: msg.Channel + "@" + b.GetString("Muc"),
			Text:   msg.Username + msg.Text,
		}); err != nil {
			return err
		}

		if fileInfo.URL != "" {
			if _, err := b.xc.SendOOB(xmpp.Chat{
				Type:    "groupchat",
				Remote:  msg.Channel + "@" + b.GetString("Muc"),
				Ooburl:  fileInfo.URL,
				Oobdesc: urlDesc,
			}); err != nil {
				b.Log.WithError(err).Warn("Failed to send share URL.")
			}
		}
	}
	return nil
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

	// do not show subjects on connect #732
	if strings.Contains(message.Text, "has set the subject to:") && time.Since(b.startTime) < time.Second*5 {
		return true
	}

	// skip delayed messages
	return !message.Stamp.IsZero() && time.Since(message.Stamp).Minutes() > 5
}

func (b *Bxmpp) setConnected(state bool) {
	b.Lock()
	b.connected = state
	defer b.Unlock()
}

func (b *Bxmpp) Connected() bool {
	b.RLock()
	defer b.RUnlock()
	return b.connected
}
