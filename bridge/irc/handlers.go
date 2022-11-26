package birc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/lrstanley/girc"
	"github.com/paulrosania/go-charset/charset"
	"github.com/saintfish/chardet"

	// We need to import the 'data' package as an implicit dependency.
	// See: https://godoc.org/github.com/paulrosania/go-charset/charset
	_ "github.com/paulrosania/go-charset/data"
)

func (b *Birc) handleCharset(msg *config.Message) error {
	if b.GetString("Charset") != "" {
		switch b.GetString("Charset") {
		case "gbk", "gb18030", "gb2312", "big5", "euc-kr", "euc-jp", "shift-jis", "iso-2022-jp":
			msg.Text = toUTF8(b.GetString("Charset"), msg.Text)
		default:
			buf := new(bytes.Buffer)
			w, err := charset.NewWriter(b.GetString("Charset"), buf)
			if err != nil {
				b.Log.Errorf("charset to utf-8 conversion failed: %s", err)
				return err
			}
			fmt.Fprint(w, msg.Text)
			w.Close()
			msg.Text = buf.String()
		}
	}
	return nil
}

// handleFiles returns true if we have handled the files, otherwise return false
func (b *Birc) handleFiles(msg *config.Message) bool {
	if msg.Extra == nil {
		return false
	}
	for _, rmsg := range helper.HandleExtra(msg, b.General) {
		b.Local <- rmsg
	}
	if len(msg.Extra["file"]) == 0 {
		return false
	}
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if fi.Comment != "" {
			msg.Text += fi.Comment + " : "
		}
		if fi.URL != "" {
			msg.Text = fi.URL
			if fi.Comment != "" {
				msg.Text = fi.Comment + " : " + fi.URL
			}
		}
		b.Local <- config.Message{Text: msg.Text, Username: msg.Username, Channel: msg.Channel, Event: msg.Event}
	}
	return true
}

func (b *Birc) handleInvite(client *girc.Client, event girc.Event) {
	if len(event.Params) != 2 {
		return
	}

	channel := event.Params[1]

	b.Log.Debugf("got invite for %s", channel)

	if _, ok := b.channels[channel]; ok {
		b.i.Cmd.Join(channel)
	}
}

func (b *Birc) handleJoinPart(client *girc.Client, event girc.Event) {
	if len(event.Params) == 0 {
		b.Log.Debugf("handleJoinPart: empty Params? %#v", event)
		return
	}
	channel := strings.ToLower(event.Params[0])
	if event.Command == "KICK" && event.Params[1] == b.Nick {
		b.Log.Infof("Got kicked from %s by %s", channel, event.Source.Name)
		time.Sleep(time.Duration(b.GetInt("RejoinDelay")) * time.Second)
		b.Remote <- config.Message{Username: "system", Text: "rejoin", Channel: channel, Account: b.Account, Event: config.EventRejoinChannels}
		return
	}
	if event.Command == "QUIT" {
		if event.Source.Name == b.Nick && strings.Contains(event.Last(), "Ping timeout") {
			b.Log.Infof("%s reconnecting ..", b.Account)
			b.Remote <- config.Message{Username: "system", Text: "reconnect", Channel: channel, Account: b.Account, Event: config.EventFailure}
			return
		}
	}
	if event.Source.Name != b.Nick {
		if b.GetBool("nosendjoinpart") {
			return
		}
		msg := config.Message{Username: "system", Text: event.Source.Name + " " + strings.ToLower(event.Command) + "s", Channel: channel, Account: b.Account, Event: config.EventJoinLeave}
		if b.GetBool("verbosejoinpart") {
			b.Log.Debugf("<= Sending verbose JOIN_LEAVE event from %s to gateway", b.Account)
			msg = config.Message{Username: "system", Text: event.Source.Name + " (" + event.Source.Ident + "@" + event.Source.Host + ") " + strings.ToLower(event.Command) + "s", Channel: channel, Account: b.Account, Event: config.EventJoinLeave}
		} else {
			b.Log.Debugf("<= Sending JOIN_LEAVE event from %s to gateway", b.Account)
		}
		b.Log.Debugf("<= Message is %#v", msg)
		b.Remote <- msg
		return
	}
	b.Log.Debugf("handle %#v", event)
}

func (b *Birc) handleNewConnection(client *girc.Client, event girc.Event) {
	b.Log.Debug("Registering callbacks")
	i := b.i
	b.Nick = event.Params[0]

	i.Handlers.AddBg("PRIVMSG", b.handlePrivMsg)
	i.Handlers.AddBg("CTCP_ACTION", b.handlePrivMsg)
	i.Handlers.Add(girc.RPL_TOPICWHOTIME, b.handleTopicWhoTime)
	i.Handlers.AddBg(girc.NOTICE, b.handleNotice)
	i.Handlers.AddBg("JOIN", b.handleJoinPart)
	i.Handlers.AddBg("PART", b.handleJoinPart)
	i.Handlers.AddBg("QUIT", b.handleJoinPart)
	i.Handlers.AddBg("KICK", b.handleJoinPart)
	i.Handlers.Add("INVITE", b.handleInvite)
}

func (b *Birc) handleNickServ() {
	if !b.GetBool("UseSASL") && b.GetString("NickServNick") != "" && b.GetString("NickServPassword") != "" {
		b.Log.Debugf("Sending identify to nickserv %s", b.GetString("NickServNick"))
		b.i.Cmd.Message(b.GetString("NickServNick"), "IDENTIFY "+b.GetString("NickServPassword"))
	}
	if strings.EqualFold(b.GetString("NickServNick"), "Q@CServe.quakenet.org") {
		b.Log.Debugf("Authenticating %s against %s", b.GetString("NickServUsername"), b.GetString("NickServNick"))
		b.i.Cmd.Message(b.GetString("NickServNick"), "AUTH "+b.GetString("NickServUsername")+" "+b.GetString("NickServPassword"))
	}
	// give nickserv some slack
	time.Sleep(time.Second * 5)
	b.authDone = true
}

func (b *Birc) handleNotice(client *girc.Client, event girc.Event) {
	if strings.Contains(event.String(), "This nickname is registered") && event.Source.Name == b.GetString("NickServNick") {
		b.handleNickServ()
	} else {
		b.handlePrivMsg(client, event)
	}
}

func (b *Birc) handleOther(client *girc.Client, event girc.Event) {
	if b.GetInt("DebugLevel") == 1 {
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
	b.handleNickServ()
	b.handleRunCommands()
	// we are now fully connected
	// only send on first connection
	if b.FirstConnection {
		b.connected <- nil
	}
}

func (b *Birc) handlePrivMsg(client *girc.Client, event girc.Event) {
	if b.skipPrivMsg(event) {
		return
	}

	rmsg := config.Message{
		Username: event.Source.Name,
		Channel:  strings.ToLower(event.Params[0]),
		Account:  b.Account,
		UserID:   event.Source.Ident + "@" + event.Source.Host,
	}

	b.Log.Debugf("== Receiving PRIVMSG: %s %s %#v", event.Source.Name, event.Last(), event)

	// set action event
	if event.IsAction() {
		rmsg.Event = config.EventUserAction
	}

	// set NOTICE event
	if event.Command == "NOTICE" {
		rmsg.Event = config.EventNoticeIRC
	}

	// strip action, we made an event if it was an action
	rmsg.Text += event.StripAction()

	// start detecting the charset
	mycharset := b.GetString("Charset")
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
	switch mycharset {
	case "gbk", "gb18030", "gb2312", "big5", "euc-kr", "euc-jp", "shift-jis", "iso-2022-jp":
		rmsg.Text = toUTF8(b.GetString("Charset"), rmsg.Text)
	default:
		r, err := charset.NewReader(mycharset, strings.NewReader(rmsg.Text))
		if err != nil {
			b.Log.Errorf("charset to utf-8 conversion failed: %s", err)
			return
		}
		output, _ := ioutil.ReadAll(r)
		rmsg.Text = string(output)
	}

	b.Log.Debugf("<= Sending message from %s on %s to gateway", event.Params[0], b.Account)
	b.Remote <- rmsg
}

func (b *Birc) handleRunCommands() {
	for _, cmd := range b.GetStringSlice("RunCommands") {
		cmd = strings.ReplaceAll(cmd, "{BOTNICK}", b.Nick)
		if err := b.i.Cmd.SendRaw(cmd); err != nil {
			b.Log.Errorf("RunCommands %s failed: %s", cmd, err)
		}
		time.Sleep(time.Second)
	}
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
