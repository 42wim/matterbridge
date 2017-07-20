package bxmpp

import (
	"crypto/tls"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"github.com/jpillora/backoff"
	"github.com/mattn/go-xmpp"

	"strings"
	"time"
)

type Bxmpp struct {
	xc      *xmpp.Client
	xmppMap map[string]string
	Config  *config.Protocol
	Remote  chan config.Message
	Account string
}

var flog *log.Entry
var protocol = "xmpp"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Bxmpp {
	b := &Bxmpp{}
	b.xmppMap = make(map[string]string)
	b.Config = &cfg
	b.Account = account
	b.Remote = c
	return b
}

func (b *Bxmpp) Connect() error {
	var err error
	flog.Infof("Connecting %s", b.Config.Server)
	b.xc, err = b.createXMPP()
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	flog.Info("Connection succeeded")
	go func() {
		initial := true
		bf := &backoff.Backoff{
			Min:    time.Second,
			Max:    5 * time.Minute,
			Jitter: true,
		}
		for {
			if initial {
				b.handleXmpp()
				initial = false
			}
			d := bf.Duration()
			flog.Infof("Disconnected. Reconnecting in %s", d)
			time.Sleep(d)
			b.xc, err = b.createXMPP()
			if err == nil {
				b.Remote <- config.Message{Username: "system", Text: "rejoin", Channel: "", Account: b.Account, Event: config.EVENT_REJOIN_CHANNELS}
				b.handleXmpp()
				bf.Reset()
			}
		}
	}()
	return nil
}

func (b *Bxmpp) Disconnect() error {
	return nil
}

func (b *Bxmpp) JoinChannel(channel string) error {
	b.xc.JoinMUCNoHistory(channel+"@"+b.Config.Muc, b.Config.Nick)
	return nil
}

func (b *Bxmpp) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	b.xc.Send(xmpp.Chat{Type: "groupchat", Remote: msg.Channel + "@" + b.Config.Muc, Text: msg.Username + msg.Text})
	return nil
}

func (b *Bxmpp) createXMPP() (*xmpp.Client, error) {
	tc := new(tls.Config)
	tc.InsecureSkipVerify = b.Config.SkipTLSVerify
	tc.ServerName = strings.Split(b.Config.Server, ":")[0]
	options := xmpp.Options{
		Host:      b.Config.Server,
		User:      b.Config.Jid,
		Password:  b.Config.Password,
		NoTLS:     true,
		StartTLS:  true,
		TLSConfig: tc,

		//StartTLS:      false,
		Debug:                        true,
		Session:                      true,
		Status:                       "",
		StatusMessage:                "",
		Resource:                     "",
		InsecureAllowUnencryptedAuth: false,
		//InsecureAllowUnencryptedAuth: true,
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
				flog.Debugf("PING")
				err := b.xc.PingC2S("", "")
				if err != nil {
					flog.Debugf("PING failed %#v", err)
				}
			case <-done:
				return
			}
		}
	}()
	return done
}

func (b *Bxmpp) handleXmpp() error {
	done := b.xmppKeepAlive()
	defer close(done)
	nodelay := time.Time{}
	for {
		m, err := b.xc.Recv()
		if err != nil {
			return err
		}
		switch v := m.(type) {
		case xmpp.Chat:
			var channel, nick string
			if v.Type == "groupchat" {
				s := strings.Split(v.Remote, "@")
				if len(s) >= 2 {
					channel = s[0]
				}
				s = strings.Split(s[1], "/")
				if len(s) == 2 {
					nick = s[1]
				}
				if nick != b.Config.Nick && v.Stamp == nodelay && v.Text != "" {
					flog.Debugf("Sending message from %s on %s to gateway", nick, b.Account)
					b.Remote <- config.Message{Username: nick, Text: v.Text, Channel: channel, Account: b.Account, UserID: v.Remote}
				}
			}
		case xmpp.Presence:
			// do nothing
		}
	}
}
