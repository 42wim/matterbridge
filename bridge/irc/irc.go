package birc

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/lrstanley/girc"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"
	"github.com/saintfish/chardet"
	"io"
	"io/ioutil"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Birc struct {
	i               *girc.Client
	Nick            string
	names           map[string][]string
	connected       chan struct{}
	Local           chan config.Message // local queue for flood control
	FirstConnection bool

	*config.BridgeConfig
}

func New(cfg *config.BridgeConfig) bridge.Bridger {
	b := &Birc{}
	b.BridgeConfig = cfg
	b.Nick = b.Config.Nick
	b.names = make(map[string][]string)
	b.connected = make(chan struct{})
	if b.Config.MessageDelay == 0 {
		b.Config.MessageDelay = 1300
	}
	if b.Config.MessageQueue == 0 {
		b.Config.MessageQueue = 30
	}
	if b.Config.MessageLength == 0 {
		b.Config.MessageLength = 400
	}
	b.FirstConnection = true
	return b
}

func (b *Birc) Command(msg *config.Message) string {
	switch msg.Text {
	case "!users":
		b.i.Handlers.Add(girc.RPL_NAMREPLY, b.storeNames)
		b.i.Handlers.Add(girc.RPL_ENDOFNAMES, b.endNames)
		b.i.Cmd.SendRaw("NAMES " + msg.Channel)
	}
	return ""
}

func (b *Birc) Connect() error {
	b.Local = make(chan config.Message, b.Config.MessageQueue+10)
	b.Log.Infof("Connecting %s", b.Config.Server)
	server, portstr, err := net.SplitHostPort(b.Config.Server)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		return err
	}
	// fix strict user handling of girc
	user := b.Config.Nick
	for !girc.IsValidUser(user) {
		if len(user) == 1 {
			user = "matterbridge"
			break
		}
		user = user[1:]
	}

	i := girc.New(girc.Config{
		Server:     server,
		ServerPass: b.Config.Password,
		Port:       port,
		Nick:       b.Config.Nick,
		User:       user,
		Name:       b.Config.Nick,
		SSL:        b.Config.UseTLS,
		TLSConfig:  &tls.Config{InsecureSkipVerify: b.Config.SkipTLSVerify, ServerName: server},
		PingDelay:  time.Minute,
	})

	if b.Config.UseSASL {
		i.Config.SASL = &girc.SASLPlain{b.Config.NickServNick, b.Config.NickServPassword}
	}

	i.Handlers.Add(girc.RPL_WELCOME, b.handleNewConnection)
	i.Handlers.Add(girc.RPL_ENDOFMOTD, b.handleOtherAuth)
	i.Handlers.Add(girc.ALL_EVENTS, b.handleOther)
	go func() {
		for {
			if err := i.Connect(); err != nil {
				b.Log.Errorf("error: %s", err)
				b.Log.Info("reconnecting in 30 seconds...")
				time.Sleep(30 * time.Second)
				i.Handlers.Clear(girc.RPL_WELCOME)
				i.Handlers.Add(girc.RPL_WELCOME, func(client *girc.Client, event girc.Event) {
					b.Remote <- config.Message{Username: "system", Text: "rejoin", Channel: "", Account: b.Account, Event: config.EVENT_REJOIN_CHANNELS}
					// set our correct nick on reconnect if necessary
					b.Nick = event.Source.Name
				})
			} else {
				return
			}
		}
	}()
	b.i = i
	select {
	case <-b.connected:
		b.Log.Info("Connection succeeded")
	case <-time.After(time.Second * 30):
		return fmt.Errorf("connection timed out")
	}
	//i.Debug = false
	if b.Config.DebugLevel == 0 {
		i.Handlers.Clear(girc.ALL_EVENTS)
	}
	go b.doSend()
	return nil
}

func (b *Birc) Disconnect() error {
	//b.i.Disconnect()
	close(b.Local)
	return nil
}

func (b *Birc) JoinChannel(channel config.ChannelInfo) error {
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
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}

	b.Log.Debugf("Receiving %#v", msg)

	// Execute a command
	if strings.HasPrefix(msg.Text, "!") {
		b.Command(&msg)
	}

	// convert to specified charset
	if b.Config.Charset != "" {
		buf := new(bytes.Buffer)
		w, err := charset.NewWriter(b.Config.Charset, buf)
		if err != nil {
			b.Log.Errorf("charset from utf-8 conversion failed: %s", err)
			return "", err
		}
		fmt.Fprintf(w, msg.Text)
		w.Close()
		msg.Text = buf.String()
	}

	// Handle files
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.Local <- rmsg
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
				b.Local <- config.Message{Text: msg.Text, Username: msg.Username, Channel: msg.Channel, Event: msg.Event}
			}
			return "", nil
		}
	}

	// split long messages on messageLength, to avoid clipped messages #281
	if b.Config.MessageSplit {
		msg.Text = helper.SplitStringLength(msg.Text, b.Config.MessageLength)
	}
	for _, text := range strings.Split(msg.Text, "\n") {
		if len(text) > b.Config.MessageLength {
			text = text[:b.Config.MessageLength-len(" <message clipped>")]
			if r, size := utf8.DecodeLastRuneInString(text); r == utf8.RuneError {
				text = text[:len(text)-size]
			}
			text += " <message clipped>"
		}
		if len(b.Local) < b.Config.MessageQueue {
			if len(b.Local) == b.Config.MessageQueue-1 {
				text = text + " <message clipped>"
			}
			b.Local <- config.Message{Text: text, Username: msg.Username, Channel: msg.Channel, Event: msg.Event}
		} else {
			b.Log.Debugf("flooding, dropping message (queue at %d)", len(b.Local))
		}
	}
	return "", nil
}

func (b *Birc) doSend() {
	rate := time.Millisecond * time.Duration(b.Config.MessageDelay)
	throttle := time.NewTicker(rate)
	for msg := range b.Local {
		<-throttle.C
		if msg.Event == config.EVENT_USER_ACTION {
			b.i.Cmd.Action(msg.Channel, msg.Username+msg.Text)
		} else {
			b.i.Cmd.Message(msg.Channel, msg.Username+msg.Text)
		}
	}
}

func (b *Birc) endNames(client *girc.Client, event girc.Event) {
	channel := event.Params[1]
	sort.Strings(b.names[channel])
	maxNamesPerPost := (300 / b.nicksPerRow()) * b.nicksPerRow()
	continued := false
	for len(b.names[channel]) > maxNamesPerPost {
		b.Remote <- config.Message{Username: b.Nick, Text: b.formatnicks(b.names[channel][0:maxNamesPerPost], continued),
			Channel: channel, Account: b.Account}
		b.names[channel] = b.names[channel][maxNamesPerPost:]
		continued = true
	}
	b.Remote <- config.Message{Username: b.Nick, Text: b.formatnicks(b.names[channel], continued),
		Channel: channel, Account: b.Account}
	b.names[channel] = nil
	b.i.Handlers.Clear(girc.RPL_NAMREPLY)
	b.i.Handlers.Clear(girc.RPL_ENDOFNAMES)
}

func (b *Birc) handleNewConnection(client *girc.Client, event girc.Event) {
	b.Log.Debug("Registering callbacks")
	i := b.i
	b.Nick = event.Params[0]

	i.Handlers.Add(girc.RPL_ENDOFMOTD, b.handleOtherAuth)
	i.Handlers.Add("PRIVMSG", b.handlePrivMsg)
	i.Handlers.Add("CTCP_ACTION", b.handlePrivMsg)
	i.Handlers.Add(girc.RPL_TOPICWHOTIME, b.handleTopicWhoTime)
	i.Handlers.Add(girc.NOTICE, b.handleNotice)
	i.Handlers.Add("JOIN", b.handleJoinPart)
	i.Handlers.Add("PART", b.handleJoinPart)
	i.Handlers.Add("QUIT", b.handleJoinPart)
	i.Handlers.Add("KICK", b.handleJoinPart)
	// we are now fully connected
	b.connected <- struct{}{}
}

func (b *Birc) handleJoinPart(client *girc.Client, event girc.Event) {
	if len(event.Params) == 0 {
		b.Log.Debugf("handleJoinPart: empty Params? %#v", event)
		return
	}
	channel := strings.ToLower(event.Params[0])
	if event.Command == "KICK" {
		b.Log.Infof("Got kicked from %s by %s", channel, event.Source.Name)
		time.Sleep(time.Duration(b.Config.RejoinDelay) * time.Second)
		b.Remote <- config.Message{Username: "system", Text: "rejoin", Channel: channel, Account: b.Account, Event: config.EVENT_REJOIN_CHANNELS}
		return
	}
	if event.Command == "QUIT" {
		if event.Source.Name == b.Nick && strings.Contains(event.Trailing, "Ping timeout") {
			b.Log.Infof("%s reconnecting ..", b.Account)
			b.Remote <- config.Message{Username: "system", Text: "reconnect", Channel: channel, Account: b.Account, Event: config.EVENT_FAILURE}
			return
		}
	}
	if event.Source.Name != b.Nick {
		b.Log.Debugf("Sending JOIN_LEAVE event from %s to gateway", b.Account)
		msg := config.Message{Username: "system", Text: event.Source.Name + " " + strings.ToLower(event.Command) + "s", Channel: channel, Account: b.Account, Event: config.EVENT_JOIN_LEAVE}
		b.Log.Debugf("Message is %#v", msg)
		b.Remote <- msg
		return
	}
	b.Log.Debugf("handle %#v", event)
}

func (b *Birc) handleNotice(client *girc.Client, event girc.Event) {
	if strings.Contains(event.String(), "This nickname is registered") && event.Source.Name == b.Config.NickServNick {
		b.i.Cmd.Message(b.Config.NickServNick, "IDENTIFY "+b.Config.NickServPassword)
	} else {
		b.handlePrivMsg(client, event)
	}
}

func (b *Birc) handleOther(client *girc.Client, event girc.Event) {
	if b.Config.DebugLevel == 1 {
		if event.Command != "CLIENT_STATE_UPDATED" &&
			event.Command != "CLIENT_GENERAL_UPDATED" {
			b.Log.Debugf("%#v", event.String())
		}
		return
	}
	switch event.Command {
	case "372", "375", "376", "250", "251", "252", "253", "254", "255", "265", "266", "002", "003", "004", "005":
		return
	}
	b.Log.Debugf("%#v", event.String())
}

func (b *Birc) handleOtherAuth(client *girc.Client, event girc.Event) {
	if strings.EqualFold(b.Config.NickServNick, "Q@CServe.quakenet.org") {
		b.Log.Debugf("Authenticating %s against %s", b.Config.NickServUsername, b.Config.NickServNick)
		b.i.Cmd.Message(b.Config.NickServNick, "AUTH "+b.Config.NickServUsername+" "+b.Config.NickServPassword)
	}
}

func (b *Birc) skipPrivMsg(event girc.Event) bool {
	// Our nick can be changed
	b.Nick = b.i.GetNick()

	// freenode doesn't send 001 as first reply
	if event.Command == "NOTICE" {
		return true
	}
	// don't forward queries to the bot
	if event.Params[0] == b.Nick {
		return true
	}
	// don't forward message from ourself
	if event.Source.Name == b.Nick {
		return true
	}
	return false
}

func (b *Birc) handlePrivMsg(client *girc.Client, event girc.Event) {
	if b.skipPrivMsg(event) {
		return
	}
	rmsg := config.Message{Username: event.Source.Name, Channel: strings.ToLower(event.Params[0]), Account: b.Account, UserID: event.Source.Ident + "@" + event.Source.Host}
	b.Log.Debugf("Receiving PRIVMSG: %s %s %#v", event.Source.Name, event.Trailing, event)

	// set action event
	if event.IsAction() {
		rmsg.Event = config.EVENT_USER_ACTION
	}

	// strip action, we made an event if it was an action
	rmsg.Text += event.StripAction()

	// strip IRC colors
	re := regexp.MustCompile(`[[:cntrl:]](?:\d{1,2}(?:,\d{1,2})?)?`)
	rmsg.Text = re.ReplaceAllString(rmsg.Text, "")

	// start detecting the charset
	var r io.Reader
	var err error
	mycharset := b.Config.Charset
	if mycharset == "" {
		// detect what were sending so that we convert it to utf-8
		detector := chardet.NewTextDetector()
		result, err := detector.DetectBest([]byte(rmsg.Text))
		if err != nil {
			b.Log.Infof("detection failed for rmsg.Text: %#v", rmsg.Text)
			return
		}
		b.Log.Debugf("detected %s confidence %#v", result.Charset, result.Confidence)
		mycharset = result.Charset
		// if we're not sure, just pick ISO-8859-1
		if result.Confidence < 80 {
			mycharset = "ISO-8859-1"
		}
	}
	r, err = charset.NewReader(mycharset, strings.NewReader(rmsg.Text))
	if err != nil {
		b.Log.Errorf("charset to utf-8 conversion failed: %s", err)
		return
	}
	output, _ := ioutil.ReadAll(r)
	rmsg.Text = string(output)

	b.Log.Debugf("Sending message from %s on %s to gateway", event.Params[0], b.Account)
	b.Remote <- rmsg
}

func (b *Birc) handleTopicWhoTime(client *girc.Client, event girc.Event) {
	parts := strings.Split(event.Params[2], "!")
	t, err := strconv.ParseInt(event.Params[3], 10, 64)
	if err != nil {
		b.Log.Errorf("Invalid time stamp: %s", event.Params[3])
	}
	user := parts[0]
	if len(parts) > 1 {
		user += " [" + parts[1] + "]"
	}
	b.Log.Debugf("%s: Topic set by %s [%s]", event.Command, user, time.Unix(t, 0))
}

func (b *Birc) nicksPerRow() int {
	return 4
}

func (b *Birc) storeNames(client *girc.Client, event girc.Event) {
	channel := event.Params[2]
	b.names[channel] = append(
		b.names[channel],
		strings.Split(strings.TrimSpace(event.Trailing), " ")...)
}

func (b *Birc) formatnicks(nicks []string, continued bool) string {
	return plainformatter(nicks, b.nicksPerRow())
}
