package bmumble

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"strconv"
//	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
	stripmd "github.com/writeas/go-strip-markdown"

	// We need to import the 'data' package as an implicit dependency.
	// See: https://godoc.org/github.com/paulrosania/go-charset/charset
	_ "github.com/paulrosania/go-charset/data"
)


type Bmumble struct {
        client               *gumble.Client
	Nick                 string
	Host                 string
	Channel              string
	local                chan config.Message
	running              chan error
	connected            chan disconnect
	serverConfigUpdate   chan *gumble.ServerConfigEvent
	serverConfig         gumble.ServerConfigEvent
	tlsConfig            tls.Config

	*bridge.Config
}


type disconnect struct {
	Reason  gumble.DisconnectType
	Message string
}


func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmumble{}
	b.Config = cfg
	b.Nick = b.GetString("Nick")
	b.local = make(chan config.Message)
	b.running = make(chan error)
	b.connected = make(chan disconnect)
	b.serverConfigUpdate = make(chan *gumble.ServerConfigEvent)
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

	go b.doSend()
	go b.connectLoop()
	err = <-b.running
	return err

}

func (b *Bmumble) Disconnect() error {
	b.client.Disconnect()
	return nil
}

func (b *Bmumble) JoinChannel(channel config.ChannelInfo) error {
	if b.Channel != "" {
		b.Log.Fatalf("Cannot join channel '%s', already joined to channel '%'s", channel.Name, b.Channel)
		return errors.New("The Mumble bridge can only join a single channel")
	}
	b.Channel = channel.Name
	return b.doJoin(b.client, channel.Name)
}


func (b *Bmumble) Send(msg config.Message) (string, error) {
	// Only process text messages
	b.Log.Debugf("=> Received local message %#v", msg)
	if msg.Event != "" && msg.Event != config.EventUserAction {
		return "", nil
	}
	
	b.local <- msg
	return "", nil
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
		switch d.Reason {
		case gumble.DisconnectError:
			b.Log.Errorf("Lost connection to the server (%s), attempting reconnect", d.Message)
			continue
		case gumble.DisconnectKicked:
			b.Log.Errorf("Kicked from the server (%s), attempting reconnect", d.Message)
			continue
		case gumble.DisconnectBanned:
			b.Log.Errorf("Banned from the server (%s), not attempting reconnect", d.Message)
			break
		case gumble.DisconnectUser:
			b.Log.Infof("Disconnect successful")
			break
		}
	}
	close(b.connected)
	close(b.running)
}

func (b *Bmumble) doConnect() error {

	// Create new gumble config and attach event handlers
	gumbleConfig := gumble.NewConfig()
	gumbleConfig.Attach(gumbleutil.Listener{
		ServerConfig: b.handleServerConfig,
		TextMessage: b.handleTextMessage,
		Connect: b.handleConnect,
		Disconnect: b.handleDisconnect,
		UserChange: b.handleUserChange,
	})
	if b.GetInt("DebugLevel") == 0 {
		gumbleConfig.Attach(b.makeDebugHandler())
	}
	gumbleConfig.Username = b.GetString("Nick")
	if password := b.GetString("Password"); password != "" {
		gumbleConfig.Password = password
	}
	
	client, err := gumble.DialWithDialer(new(net.Dialer), b.GetString("Server"), gumbleConfig, &b.tlsConfig)
	if err != nil {
		return err
	}
	b.client = client
	return nil
}

func (b *Bmumble) doJoin(client *gumble.Client, name string) error {
	c := client.Channels.Find(name)
	if c == nil {
		return errors.New("No such channel: " + name)
	}
	client.Self.Move(c)
	b.Channel = c.Name
	return nil
}


func (b *Bmumble) doSend() {
	// Message sending loop that makes sure server-side
	// restrictions and client-side message traits don't conflict
	// with each other.
	for {
		select {
		case config := <-b.serverConfigUpdate:
			b.Log.Debugf("Received server config update: AllowHTML=%d, MaxMessageLength=%d", config.AllowHTML, config.MaximumMessageLength)
			b.serverConfig = *config
		case msg := <-b.local:
			b.processMessage(&msg)
		}
	}
}


func (b *Bmumble) processMessage(msg *config.Message) {
	b.Log.Debugf("Processing message %s", msg.Text)
	
	// If HTML is allowed, convert markdown into HTML, otherwise strip markdown
	if allowHtml := b.serverConfig.AllowHTML; allowHtml == nil || !*allowHtml  {
		msg.Text = helper.ParseMarkdown(msg.Text)
	} else {
		msg.Text = stripmd.Strip(msg.Text)
	}
	
	// If there is a maximum message length, split and truncate the lines
	var msgLines []string
	if maxLength := b.serverConfig.MaximumMessageLength; maxLength != nil {
		msgLines = helper.GetSubLines(msg.Text, *maxLength)
	} else {
		msgLines = helper.GetSubLines(msg.Text, 0)
	}
	// Send the individual lindes
	for i := range msgLines {
		b.Log.Debugf("Sending line: %s", msgLines[i])
		b.client.Self.Channel.Send(msg.Username + msgLines[i], false)
	}

}
