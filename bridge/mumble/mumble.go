package bmumble

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	stripmd "github.com/writeas/go-strip-markdown"

	// We need to import the 'data' package as an implicit dependency.
	// See: https://godoc.org/github.com/paulrosania/go-charset/charset
	_ "github.com/paulrosania/go-charset/data"
)

type Bmumble struct {
	client             *gumble.Client
	Nick               string
	Host               string
	Channel            *uint32
	local              chan config.Message
	running            chan error
	connected          chan gumble.DisconnectEvent
	serverConfigUpdate chan gumble.ServerConfigEvent
	serverConfig       gumble.ServerConfigEvent
	tlsConfig          tls.Config

	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmumble{}
	b.Config = cfg
	b.Nick = b.GetString("Nick")
	b.local = make(chan config.Message)
	b.running = make(chan error)
	b.connected = make(chan gumble.DisconnectEvent)
	b.serverConfigUpdate = make(chan gumble.ServerConfigEvent)
	return b
}

func (b *Bmumble) Connect() error {
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	host, portstr, err := net.SplitHostPort(b.GetString("Server"))
	if err != nil {
		return err
	}
	b.Host = host
	_, err = strconv.Atoi(portstr)
	if err != nil {
		return err
	}

	if err = b.buildTLSConfig(); err != nil {
		return err
	}

	go b.doSend()
	go b.connectLoop()
	err = <-b.running
	return err
}

func (b *Bmumble) Disconnect() error {
	return b.client.Disconnect()
}

func (b *Bmumble) JoinChannel(channel config.ChannelInfo) error {
	cid, err := strconv.ParseUint(channel.Name, 10, 32)
	if err != nil {
		return err
	}
	channelID := uint32(cid)
	if b.Channel != nil && *b.Channel != channelID {
		b.Log.Fatalf("Cannot join channel ID '%d', already joined to channel ID %d", channelID, *b.Channel)
		return errors.New("the Mumble bridge can only join a single channel")
	}
	b.Channel = &channelID
	return b.doJoin(b.client, channelID)
}

func (b *Bmumble) Send(msg config.Message) (string, error) {
	// Only process text messages
	b.Log.Debugf("=> Received local message %#v", msg)
	if msg.Event != "" && msg.Event != config.EventUserAction && msg.Event != config.EventJoinLeave {
		return "", nil
	}

	attachments := b.extractFiles(&msg)
	b.local <- msg
	for _, a := range attachments {
		b.local <- a
	}
	return "", nil
}

func (b *Bmumble) buildTLSConfig() error {
	b.tlsConfig = tls.Config{}
	// Load TLS client certificate keypair required for registered user authentication
	if cpath := b.GetString("TLSClientCertificate"); cpath != "" {
		if ckey := b.GetString("TLSClientKey"); ckey != "" {
			cert, err := tls.LoadX509KeyPair(cpath, ckey)
			if err != nil {
				return err
			}
			b.tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}
	// Load TLS CA used for server verification.  If not provided, the Go system trust anchor is used
	if capath := b.GetString("TLSCACertificate"); capath != "" {
		ca, err := ioutil.ReadFile(capath)
		if err != nil {
			return err
		}
		b.tlsConfig.RootCAs = x509.NewCertPool()
		b.tlsConfig.RootCAs.AppendCertsFromPEM(ca)
	}
	b.tlsConfig.InsecureSkipVerify = b.GetBool("SkipTLSVerify")
	return nil
}

func (b *Bmumble) connectLoop() {
	firstConnect := true
	for {
		err := b.doConnect()
		if firstConnect {
			b.running <- err
		}
		if err != nil {
			b.Log.Errorf("Connection to server failed: %#v", err)
			if firstConnect {
				break
			} else {
				b.Log.Info("Retrying in 10s")
				time.Sleep(10 * time.Second)
				continue
			}
		}
		firstConnect = false
		d := <-b.connected
		switch d.Type {
		case gumble.DisconnectError:
			b.Log.Errorf("Lost connection to the server (%s), attempting reconnect", d.String)
			continue
		case gumble.DisconnectKicked:
			b.Log.Errorf("Kicked from the server (%s), attempting reconnect", d.String)
			continue
		case gumble.DisconnectBanned:
			b.Log.Errorf("Banned from the server (%s), not attempting reconnect", d.String)
			close(b.connected)
			close(b.running)
			return
		case gumble.DisconnectUser:
			b.Log.Infof("Disconnect successful")
			close(b.connected)
			close(b.running)
			return
		}
	}
}

func (b *Bmumble) doConnect() error {
	// Create new gumble config and attach event handlers
	gumbleConfig := gumble.NewConfig()
	gumbleConfig.Attach(gumbleutil.Listener{
		ServerConfig: b.handleServerConfig,
		TextMessage:  b.handleTextMessage,
		Connect:      b.handleConnect,
		Disconnect:   b.handleDisconnect,
		UserChange:   b.handleUserChange,
	})
	gumbleConfig.Username = b.GetString("Nick")
	if password := b.GetString("Password"); password != "" {
		gumbleConfig.Password = password
	}

	registerNullCodecAsOpus()
	client, err := gumble.DialWithDialer(new(net.Dialer), b.GetString("Server"), gumbleConfig, &b.tlsConfig)
	if err != nil {
		return err
	}
	b.client = client
	return nil
}

func (b *Bmumble) doJoin(client *gumble.Client, channelID uint32) error {
	channel, ok := client.Channels[channelID]
	if !ok {
		return fmt.Errorf("no channel with ID %d", channelID)
	}
	client.Self.Move(channel)
	return nil
}

func (b *Bmumble) doSend() {
	// Message sending loop that makes sure server-side
	// restrictions and client-side message traits don't conflict
	// with each other.
	for {
		select {
		case serverConfig := <-b.serverConfigUpdate:
			b.Log.Debugf("Received server config update: AllowHTML=%#v, MaximumMessageLength=%#v", serverConfig.AllowHTML, serverConfig.MaximumMessageLength)
			b.serverConfig = serverConfig
		case msg := <-b.local:
			b.processMessage(&msg)
		}
	}
}

func (b *Bmumble) processMessage(msg *config.Message) {
	b.Log.Debugf("Processing message %s", msg.Text)

	allowHTML := true
	if b.serverConfig.AllowHTML != nil {
		allowHTML = *b.serverConfig.AllowHTML
	}

	// If this is a specially generated image message, send it unmodified
	if msg.Event == "mumble_image" {
		if allowHTML {
			b.client.Self.Channel.Send(msg.Username+msg.Text, false)
		} else {
			b.Log.Info("Can't send image, server does not allow HTML messages")
		}
		return
	}

	// Don't process empty messages
	if len(msg.Text) == 0 {
		return
	}
	// If HTML is allowed, convert markdown into HTML, otherwise strip markdown
	if allowHTML {
		msg.Text = helper.ParseMarkdown(msg.Text)
	} else {
		msg.Text = stripmd.Strip(msg.Text)
	}

	// If there is a maximum message length, split and truncate the lines
	var msgLines []string
	if maxLength := b.serverConfig.MaximumMessageLength; maxLength != nil {
		if *maxLength != 0 { // Some servers will have unlimited message lengths.
			// Not doing this makes underflows happen.
			msgLines = helper.GetSubLines(msg.Text, *maxLength-len(msg.Username), b.GetString("MessageClipped"))
		} else {
			msgLines = helper.GetSubLines(msg.Text, 0, b.GetString("MessageClipped"))
		}
	} else {
		msgLines = helper.GetSubLines(msg.Text, 0, b.GetString("MessageClipped"))
	}
	// Send the individual lines
	for i := range msgLines {
		// Remove unnecessary newline character, since either way we're sending it as individual lines
		msgLines[i] = strings.TrimSuffix(msgLines[i], "\n")
		b.client.Self.Channel.Send(msg.Username+msgLines[i], false)
	}
}
