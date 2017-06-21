// The GsBot package contains some useful utilites for working with the
// steam package. It implements authentication with sentries, server lists and
// logging messages and events.
//
// Every module is optional and requires an instance of the GsBot struct.
// Should a module have a `HandlePacket` method, you must register it with the
// steam.Client with `RegisterPacketHandler`. Any module with a `HandleEvent`
// method must be integrated into your event loop and should be called for each
// event you receive.
package gsbot

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/netutil"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/davecgh/go-spew/spew"
)

// Base structure holding common data among GsBot modules.
type GsBot struct {
	Client *steam.Client
	Log    *log.Logger
}

// Creates a new GsBot with a new steam.Client where logs are written to stdout.
func Default() *GsBot {
	return &GsBot{
		steam.NewClient(),
		log.New(os.Stdout, "", 0),
	}
}

// This module handles authentication. It logs on automatically after a ConnectedEvent
// and saves the sentry data to a file which is also used for logon if available.
// If you're logging on for the first time Steam may require an authcode. You can then
// connect again with the new logon details.
type Auth struct {
	bot             *GsBot
	details         *LogOnDetails
	sentryPath      string
	machineAuthHash []byte
}

func NewAuth(bot *GsBot, details *LogOnDetails, sentryPath string) *Auth {
	return &Auth{
		bot:        bot,
		details:    details,
		sentryPath: sentryPath,
	}
}

type LogOnDetails struct {
	Username      string
	Password      string
	AuthCode      string
	TwoFactorCode string
}

// This is called automatically after every ConnectedEvent, but must be called once again manually
// with an authcode if Steam requires it when logging on for the first time.
func (a *Auth) LogOn(details *LogOnDetails) {
	a.details = details
	sentry, err := ioutil.ReadFile(a.sentryPath)
	if err != nil {
		a.bot.Log.Printf("Error loading sentry file from path %v - This is normal if you're logging in for the first time.\n", a.sentryPath)
	}
	a.bot.Client.Auth.LogOn(&steam.LogOnDetails{
		Username:       details.Username,
		Password:       details.Password,
		SentryFileHash: sentry,
		AuthCode:       details.AuthCode,
		TwoFactorCode:  details.TwoFactorCode,
	})
}

func (a *Auth) HandleEvent(event interface{}) {
	switch e := event.(type) {
	case *steam.ConnectedEvent:
		a.LogOn(a.details)
	case *steam.LoggedOnEvent:
		a.bot.Log.Printf("Logged on (%v) with SteamId %v and account flags %v", e.Result, e.ClientSteamId, e.AccountFlags)
	case *steam.MachineAuthUpdateEvent:
		a.machineAuthHash = e.Hash
		err := ioutil.WriteFile(a.sentryPath, e.Hash, 0666)
		if err != nil {
			panic(err)
		}
	}
}

// This module saves the server list from ClientCMListEvent and uses
// it when you call `Connect()`.
type ServerList struct {
	bot      *GsBot
	listPath string
}

func NewServerList(bot *GsBot, listPath string) *ServerList {
	return &ServerList{
		bot,
		listPath,
	}
}

func (s *ServerList) HandleEvent(event interface{}) {
	switch e := event.(type) {
	case *steam.ClientCMListEvent:
		d, err := json.Marshal(e.Addresses)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(s.listPath, d, 0666)
		if err != nil {
			panic(err)
		}
	}
}

func (s *ServerList) Connect() (bool, error) {
	return s.ConnectBind(nil)
}

func (s *ServerList) ConnectBind(laddr *net.TCPAddr) (bool, error) {
	d, err := ioutil.ReadFile(s.listPath)
	if err != nil {
		s.bot.Log.Println("Connecting to random server.")
		s.bot.Client.Connect()
		return false, nil
	}
	var addrs []*netutil.PortAddr
	err = json.Unmarshal(d, &addrs)
	if err != nil {
		return false, err
	}
	raddr := addrs[rand.Intn(len(addrs))]
	s.bot.Log.Printf("Connecting to %v from server list\n", raddr)
	s.bot.Client.ConnectToBind(raddr, laddr)
	return true, nil
}

// This module logs incoming packets and events to a directory.
type Debug struct {
	packetId, eventId uint64
	bot               *GsBot
	base              string
}

func NewDebug(bot *GsBot, base string) (*Debug, error) {
	base = path.Join(base, fmt.Sprint(time.Now().Unix()))
	err := os.MkdirAll(path.Join(base, "events"), 0700)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(path.Join(base, "packets"), 0700)
	if err != nil {
		return nil, err
	}
	return &Debug{
		0, 0,
		bot,
		base,
	}, nil
}

func (d *Debug) HandlePacket(packet *protocol.Packet) {
	d.packetId++
	name := path.Join(d.base, "packets", fmt.Sprintf("%d_%d_%s", time.Now().Unix(), d.packetId, packet.EMsg))

	text := packet.String() + "\n\n" + hex.Dump(packet.Data)
	err := ioutil.WriteFile(name+".txt", []byte(text), 0666)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(name+".bin", packet.Data, 0666)
	if err != nil {
		panic(err)
	}
}

func (d *Debug) HandleEvent(event interface{}) {
	d.eventId++
	name := fmt.Sprintf("%d_%d_%s.txt", time.Now().Unix(), d.eventId, name(event))
	err := ioutil.WriteFile(path.Join(d.base, "events", name), []byte(spew.Sdump(event)), 0666)
	if err != nil {
		panic(err)
	}
}

func name(obj interface{}) string {
	val := reflect.ValueOf(obj)
	ind := reflect.Indirect(val)
	if ind.IsValid() {
		return ind.Type().Name()
	} else {
		return val.Type().Name()
	}
}
