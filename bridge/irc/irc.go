package birc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/lrstanley/girc"
	stripmd "github.com/writeas/go-strip-markdown"

	// We need to import the 'data' package as an implicit dependency.
	// See: https://godoc.org/github.com/paulrosania/go-charset/charset
	_ "github.com/paulrosania/go-charset/data"
)

type Birc struct {
	i                                         *girc.Client
	Nick                                      string
	names                                     map[string][]string
	connected                                 chan error
	Local                                     chan config.Message // local queue for flood control
	FirstConnection, authDone                 bool
	MessageDelay, MessageQueue, MessageLength int
	channels                                  map[string]bool

	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Birc{}
	b.Config = cfg
	b.Nick = b.GetString("Nick")
	b.names = make(map[string][]string)
	b.connected = make(chan error)
	b.channels = make(map[string]bool)

	if b.GetInt("MessageDelay") == 0 {
		b.MessageDelay = 1300
	} else {
		b.MessageDelay = b.GetInt("MessageDelay")
	}
	if b.GetInt("MessageQueue") == 0 {
		b.MessageQueue = 30
	} else {
		b.MessageQueue = b.GetInt("MessageQueue")
	}
	if b.GetInt("MessageLength") == 0 {
		b.MessageLength = 400
	} else {
		b.MessageLength = b.GetInt("MessageLength")
	}
	b.FirstConnection = true
	return b
}

func (b *Birc) Command(msg *config.Message) string {
	if msg.Text == "!users" {
		b.i.Handlers.Add(girc.RPL_NAMREPLY, b.storeNames)
		b.i.Handlers.Add(girc.RPL_ENDOFNAMES, b.endNames)
		b.i.Cmd.SendRaw("NAMES " + msg.Channel) //nolint:errcheck
	}
	return ""
}

func (b *Birc) Connect() error {
	if b.GetBool("UseSASL") && b.GetString("TLSClientCertificate") != "" {
		return errors.New("you can't enable SASL and TLSClientCertificate at the same time")
	}

	b.Local = make(chan config.Message, b.MessageQueue+10)
	b.Log.Infof("Connecting %s", b.GetString("Server"))

	i, err := b.getClient()
	if err != nil {
		return err
	}

	if b.GetBool("UseSASL") {
		i.Config.SASL = &girc.SASLPlain{
			User: b.GetString("NickServNick"),
			Pass: b.GetString("NickServPassword"),
		}
	}

	i.Handlers.Add(girc.RPL_WELCOME, b.handleNewConnection)
	i.Handlers.Add(girc.RPL_ENDOFMOTD, b.handleOtherAuth)
	i.Handlers.Add(girc.ERR_NOMOTD, b.handleOtherAuth)
	i.Handlers.Add(girc.ALL_EVENTS, b.handleOther)
	b.i = i

	go b.doConnect()

	err = <-b.connected
	if err != nil {
		return fmt.Errorf("connection failed %s", err)
	}
	b.Log.Info("Connection succeeded")
	b.FirstConnection = false
	if b.GetInt("DebugLevel") == 0 {
		i.Handlers.Clear(girc.ALL_EVENTS)
	}
	go b.doSend()
	return nil
}

func (b *Birc) Disconnect() error {
	b.i.Close()
	close(b.Local)
	return nil
}

func (b *Birc) JoinChannel(channel config.ChannelInfo) error {
	b.channels[channel.Name] = true
	// need to check if we have nickserv auth done before joining channels
	for {
		if b.authDone {
			break
		}
		time.Sleep(time.Second)
	}
	if channel.Options.Key != "" {
		b.Log.Debugf("using key %s for channel %s", channel.Options.Key, channel.Name)
		b.i.Cmd.JoinKey(channel.Name, channel.Options.Key)
	} else {
		b.i.Cmd.Join(channel.Name)
	}
	return nil
}

func (b *Birc) Send(msg config.Message) (string, error) {
	// ignore delete messages
	if msg.Event == config.EventMsgDelete {
		return "", nil
	}

	b.Log.Debugf("=> Receiving %#v", msg)

	// we can be in between reconnects #385
	if !b.i.IsConnected() {
		b.Log.Error("Not connected to server, dropping message")
		return "", nil
	}

	// Execute a command
	if strings.HasPrefix(msg.Text, "!") {
		b.Command(&msg)
	}

	// convert to specified charset
	if err := b.handleCharset(&msg); err != nil {
		return "", err
	}

	// handle files, return if we're done here
	if ok := b.handleFiles(&msg); ok {
		return "", nil
	}

	var msgLines []string
	if b.GetBool("StripMarkdown") {
		msg.Text = stripmd.Strip(msg.Text)
	}

	if b.GetBool("MessageSplit") {
		msgLines = helper.GetSubLines(msg.Text, b.MessageLength, b.GetString("MessageClipped"))
	} else {
		msgLines = helper.GetSubLines(msg.Text, 0, b.GetString("MessageClipped"))
	}
	for i := range msgLines {
		if len(b.Local) >= b.MessageQueue {
			b.Log.Debugf("flooding, dropping message (queue at %d)", len(b.Local))
			return "", nil
		}

		msg.Text = msgLines[i]
		b.Local <- msg
	}
	return "", nil
}

func (b *Birc) doConnect() {
	for {
		if err := b.i.Connect(); err != nil {
			b.Log.Errorf("disconnect: error: %s", err)
			if b.FirstConnection {
				b.connected <- err
				return
			}
		} else {
			b.Log.Info("disconnect: client requested quit")
		}
		b.Log.Info("reconnecting in 30 seconds...")
		time.Sleep(30 * time.Second)
		b.i.Handlers.Clear(girc.RPL_WELCOME)
		b.i.Handlers.Add(girc.RPL_WELCOME, func(client *girc.Client, event girc.Event) {
			b.Remote <- config.Message{Username: "system", Text: "rejoin", Channel: "", Account: b.Account, Event: config.EventRejoinChannels}
			// set our correct nick on reconnect if necessary
			b.Nick = event.Source.Name
		})
	}
}

// Sanitize nicks for RELAYMSG: replace IRC characters with special meanings with "-"
func sanitizeNick(nick string) string {
	sanitize := func(r rune) rune {
		if strings.ContainsRune("!+%@&#$:'\"?*,. ", r) {
			return '-'
		}
		return r
	}
	return strings.Map(sanitize, nick)
}

func (b *Birc) doSend() {
	rate := time.Millisecond * time.Duration(b.MessageDelay)
	throttle := time.NewTicker(rate)
	for msg := range b.Local {
		<-throttle.C
		username := msg.Username
		// Optional support for the proposed RELAYMSG extension, described at
		// https://github.com/jlu5/ircv3-specifications/blob/master/extensions/relaymsg.md
		// nolint:nestif
		if (b.i.HasCapability("overdrivenetworks.com/relaymsg") || b.i.HasCapability("draft/relaymsg")) &&
			b.GetBool("UseRelayMsg") {
			username = sanitizeNick(username)
			text := msg.Text

			// Work around girc chomping leading commas on single word messages?
			if strings.HasPrefix(text, ":") && !strings.ContainsRune(text, ' ') {
				text = ":" + text
			}

			if msg.Event == config.EventUserAction {
				b.i.Cmd.SendRawf("RELAYMSG %s %s :\x01ACTION %s\x01", msg.Channel, username, text) //nolint:errcheck
			} else {
				b.Log.Debugf("Sending RELAYMSG to channel %s: nick=%s", msg.Channel, username)
				b.i.Cmd.SendRawf("RELAYMSG %s %s :%s", msg.Channel, username, text) //nolint:errcheck
			}
		} else {
			if b.GetBool("Colornicks") {
				checksum := crc32.ChecksumIEEE([]byte(msg.Username))
				colorCode := checksum%14 + 2 // quick fix - prevent white or black color codes
				username = fmt.Sprintf("\x03%02d%s\x0F", colorCode, msg.Username)
			}
			switch msg.Event {
			case config.EventUserAction:
				b.i.Cmd.Action(msg.Channel, username+msg.Text)
			case config.EventNoticeIRC:
				b.Log.Debugf("Sending notice to channel %s", msg.Channel)
				b.i.Cmd.Notice(msg.Channel, username+msg.Text)
			default:
				b.Log.Debugf("Sending to channel %s", msg.Channel)
				b.i.Cmd.Message(msg.Channel, username+msg.Text)
			}
		}
	}
}

// validateInput validates the server/port/nick configuration. Returns a *girc.Client if successful
func (b *Birc) getClient() (*girc.Client, error) {
	server, portstr, err := net.SplitHostPort(b.GetString("Server"))
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		return nil, err
	}
	user := b.GetString("UserName")
	if user == "" {
		user = b.GetString("Nick")
	}
	// fix strict user handling of girc
	for !girc.IsValidUser(user) {
		if len(user) == 1 || len(user) == 0 {
			user = "matterbridge"
			break
		}
		user = user[1:]
	}
	realName := b.GetString("RealName")
	if realName == "" {
		realName = b.GetString("Nick")
	}

	debug := ioutil.Discard
	if b.GetInt("DebugLevel") == 2 {
		debug = b.Log.Writer()
	}

	pingDelay, err := time.ParseDuration(b.GetString("pingdelay"))
	if err != nil || pingDelay == 0 {
		pingDelay = time.Minute
	}

	b.Log.Debugf("setting pingdelay to %s", pingDelay)

	tlsConfig, err := b.getTLSConfig()
	if err != nil {
		return nil, err
	}

	i := girc.New(girc.Config{
		Server:     server,
		ServerPass: b.GetString("Password"),
		Port:       port,
		Nick:       b.GetString("Nick"),
		User:       user,
		Name:       realName,
		SSL:        b.GetBool("UseTLS"),
		Bind:       b.GetString("Bind"),
		TLSConfig:  tlsConfig,
		PingDelay:  pingDelay,
		// skip gIRC internal rate limiting, since we have our own throttling
		AllowFlood:    true,
		Debug:         debug,
		SupportedCaps: map[string][]string{"overdrivenetworks.com/relaymsg": nil, "draft/relaymsg": nil},
	})
	return i, nil
}

func (b *Birc) endNames(client *girc.Client, event girc.Event) {
	channel := event.Params[1]
	sort.Strings(b.names[channel])
	maxNamesPerPost := (300 / b.nicksPerRow()) * b.nicksPerRow()
	for len(b.names[channel]) > maxNamesPerPost {
		b.Remote <- config.Message{
			Username: b.Nick, Text: b.formatnicks(b.names[channel][0:maxNamesPerPost]),
			Channel: channel, Account: b.Account,
		}
		b.names[channel] = b.names[channel][maxNamesPerPost:]
	}
	b.Remote <- config.Message{
		Username: b.Nick, Text: b.formatnicks(b.names[channel]),
		Channel: channel, Account: b.Account,
	}
	b.names[channel] = nil
	b.i.Handlers.Clear(girc.RPL_NAMREPLY)
	b.i.Handlers.Clear(girc.RPL_ENDOFNAMES)
}

func (b *Birc) skipPrivMsg(event girc.Event) bool {
	// Our nick can be changed
	b.Nick = b.i.GetNick()

	// freenode doesn't send 001 as first reply
	if event.Command == "NOTICE" && len(event.Params) != 2 {
		return true
	}
	// don't forward queries to the bot
	if event.Params[0] == b.Nick {
		return true
	}
	// don't forward message from ourself
	if event.Source != nil {
		if event.Source.Name == b.Nick {
			return true
		}
	}
	// don't forward messages we sent via RELAYMSG
	if relayedNick, ok := event.Tags.Get("draft/relaymsg"); ok && relayedNick == b.Nick {
		return true
	}
	// This is the old name of the cap sent in spoofed messages; I've kept this in
	// for compatibility reasons
	if relayedNick, ok := event.Tags.Get("relaymsg"); ok && relayedNick == b.Nick {
		return true
	}
	return false
}

func (b *Birc) nicksPerRow() int {
	return 4
}

func (b *Birc) storeNames(client *girc.Client, event girc.Event) {
	channel := event.Params[2]
	b.names[channel] = append(
		b.names[channel],
		strings.Split(strings.TrimSpace(event.Last()), " ")...)
}

func (b *Birc) formatnicks(nicks []string) string {
	return strings.Join(nicks, ", ") + " currently on IRC"
}

func (b *Birc) getTLSConfig() (*tls.Config, error) {
	server, _, _ := net.SplitHostPort(b.GetString("server"))

	tlsConfig := &tls.Config{
		InsecureSkipVerify: b.GetBool("skiptlsverify"), //nolint:gosec
		ServerName:         server,
	}

	if filename := b.GetString("TLSClientCertificate"); filename != "" {
		cert, err := tls.LoadX509KeyPair(filename, filename)
		if err != nil {
			return nil, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
