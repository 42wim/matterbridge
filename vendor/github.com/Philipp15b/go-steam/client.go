package steam

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Philipp15b/go-steam/cryptoutil"
	"github.com/Philipp15b/go-steam/netutil"
	. "github.com/Philipp15b/go-steam/protocol"
	. "github.com/Philipp15b/go-steam/protocol/protobuf"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	. "github.com/Philipp15b/go-steam/steamid"
)

// Represents a client to the Steam network.
// Always poll events from the channel returned by Events() or receiving messages will stop.
// All access, unless otherwise noted, should be threadsafe.
//
// When a FatalErrorEvent is emitted, the connection is automatically closed. The same client can be used to reconnect.
// Other errors don't have any effect.
type Client struct {
	// these need to be 64 bit aligned for sync/atomic on 32bit
	sessionId    int32
	_            uint32
	steamId      uint64
	currentJobId uint64

	Auth          *Auth
	Social        *Social
	Web           *Web
	Notifications *Notifications
	Trading       *Trading
	GC            *GameCoordinator

	events        chan interface{}
	handlers      []PacketHandler
	handlersMutex sync.RWMutex

	tempSessionKey []byte

	ConnectionTimeout time.Duration

	mutex     sync.RWMutex // guarding conn and writeChan
	conn      connection
	writeChan chan IMsg
	writeBuf  *bytes.Buffer
	heartbeat *time.Ticker
}

type PacketHandler interface {
	HandlePacket(*Packet)
}

func NewClient() *Client {
	client := &Client{
		events:   make(chan interface{}, 3),
		writeBuf: new(bytes.Buffer),
	}
	client.Auth = &Auth{client: client}
	client.RegisterPacketHandler(client.Auth)
	client.Social = newSocial(client)
	client.RegisterPacketHandler(client.Social)
	client.Web = &Web{client: client}
	client.RegisterPacketHandler(client.Web)
	client.Notifications = newNotifications(client)
	client.RegisterPacketHandler(client.Notifications)
	client.Trading = &Trading{client: client}
	client.RegisterPacketHandler(client.Trading)
	client.GC = newGC(client)
	client.RegisterPacketHandler(client.GC)
	return client
}

// Get the event channel. By convention all events are pointers, except for errors.
// It is never closed.
func (c *Client) Events() <-chan interface{} {
	return c.events
}

func (c *Client) Emit(event interface{}) {
	c.events <- event
}

// Emits a FatalErrorEvent formatted with fmt.Errorf and disconnects.
func (c *Client) Fatalf(format string, a ...interface{}) {
	c.Emit(FatalErrorEvent(fmt.Errorf(format, a...)))
	c.Disconnect()
}

// Emits an error formatted with fmt.Errorf.
func (c *Client) Errorf(format string, a ...interface{}) {
	c.Emit(fmt.Errorf(format, a...))
}

// Registers a PacketHandler that receives all incoming packets.
func (c *Client) RegisterPacketHandler(handler PacketHandler) {
	c.handlersMutex.Lock()
	defer c.handlersMutex.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *Client) GetNextJobId() JobId {
	return JobId(atomic.AddUint64(&c.currentJobId, 1))
}

func (c *Client) SteamId() SteamId {
	return SteamId(atomic.LoadUint64(&c.steamId))
}

func (c *Client) SessionId() int32 {
	return atomic.LoadInt32(&c.sessionId)
}

func (c *Client) Connected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.conn != nil
}

// Connects to a random Steam server and returns its address.
// If this client is already connected, it is disconnected first.
// This method tries to use an address from the Steam Directory and falls
// back to the built-in server list if the Steam Directory can't be reached.
// If you want to connect to a specific server, use `ConnectTo`.
func (c *Client) Connect() *netutil.PortAddr {
	var server *netutil.PortAddr

	// try to initialize the directory cache
	if !steamDirectoryCache.IsInitialized() {
		_ = steamDirectoryCache.Initialize()
	}
	if steamDirectoryCache.IsInitialized() {
		server = steamDirectoryCache.GetRandomCM()
	} else {
		server = GetRandomCM()
	}

	c.ConnectTo(server)
	return server
}

// Connects to a specific server.
// You may want to use one of the `GetRandom*CM()` functions in this package.
// If this client is already connected, it is disconnected first.
func (c *Client) ConnectTo(addr *netutil.PortAddr) {
	c.ConnectToBind(addr, nil)
}

// Connects to a specific server, and binds to a specified local IP
// If this client is already connected, it is disconnected first.
func (c *Client) ConnectToBind(addr *netutil.PortAddr, local *net.TCPAddr) {
	c.Disconnect()

	conn, err := dialTCP(local, addr.ToTCPAddr())
	if err != nil {
		c.Fatalf("Connect failed: %v", err)
		return
	}
	c.conn = conn
	c.writeChan = make(chan IMsg, 5)

	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) Disconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return
	}

	c.conn.Close()
	c.conn = nil
	if c.heartbeat != nil {
		c.heartbeat.Stop()
	}
	close(c.writeChan)
	c.Emit(&DisconnectedEvent{})

}

// Adds a message to the send queue. Modifications to the given message after
// writing are not allowed (possible race conditions).
//
// Writes to this client when not connected are ignored.
func (c *Client) Write(msg IMsg) {
	if cm, ok := msg.(IClientMsg); ok {
		cm.SetSessionId(c.SessionId())
		cm.SetSteamId(c.SteamId())
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.conn == nil {
		return
	}
	c.writeChan <- msg
}

func (c *Client) readLoop() {
	for {
		// This *should* be atomic on most platforms, but the Go spec doesn't guarantee it
		c.mutex.RLock()
		conn := c.conn
		c.mutex.RUnlock()
		if conn == nil {
			return
		}
		packet, err := conn.Read()

		if err != nil {
			c.Fatalf("Error reading from the connection: %v", err)
			return
		}
		c.handlePacket(packet)
	}
}

func (c *Client) writeLoop() {
	for {
		c.mutex.RLock()
		conn := c.conn
		c.mutex.RUnlock()
		if conn == nil {
			return
		}

		msg, ok := <-c.writeChan
		if !ok {
			return
		}

		err := msg.Serialize(c.writeBuf)
		if err != nil {
			c.writeBuf.Reset()
			c.Fatalf("Error serializing message %v: %v", msg, err)
			return
		}

		err = conn.Write(c.writeBuf.Bytes())

		c.writeBuf.Reset()

		if err != nil {
			c.Fatalf("Error writing message %v: %v", msg, err)
			return
		}
	}
}

func (c *Client) heartbeatLoop(seconds time.Duration) {
	if c.heartbeat != nil {
		c.heartbeat.Stop()
	}
	c.heartbeat = time.NewTicker(seconds * time.Second)
	for {
		_, ok := <-c.heartbeat.C
		if !ok {
			break
		}
		c.Write(NewClientMsgProtobuf(EMsg_ClientHeartBeat, new(CMsgClientHeartBeat)))
	}
	c.heartbeat = nil
}

func (c *Client) handlePacket(packet *Packet) {
	switch packet.EMsg {
	case EMsg_ChannelEncryptRequest:
		c.handleChannelEncryptRequest(packet)
	case EMsg_ChannelEncryptResult:
		c.handleChannelEncryptResult(packet)
	case EMsg_Multi:
		c.handleMulti(packet)
	case EMsg_ClientCMList:
		c.handleClientCMList(packet)
	}

	c.handlersMutex.RLock()
	defer c.handlersMutex.RUnlock()
	for _, handler := range c.handlers {
		handler.HandlePacket(packet)
	}
}

func (c *Client) handleChannelEncryptRequest(packet *Packet) {
	body := NewMsgChannelEncryptRequest()
	packet.ReadMsg(body)

	if body.Universe != EUniverse_Public {
		c.Fatalf("Invalid univserse %v!", body.Universe)
	}

	c.tempSessionKey = make([]byte, 32)
	rand.Read(c.tempSessionKey)
	encryptedKey := cryptoutil.RSAEncrypt(GetPublicKey(EUniverse_Public), c.tempSessionKey)

	payload := new(bytes.Buffer)
	payload.Write(encryptedKey)
	binary.Write(payload, binary.LittleEndian, crc32.ChecksumIEEE(encryptedKey))
	payload.WriteByte(0)
	payload.WriteByte(0)
	payload.WriteByte(0)
	payload.WriteByte(0)

	c.Write(NewMsg(NewMsgChannelEncryptResponse(), payload.Bytes()))
}

func (c *Client) handleChannelEncryptResult(packet *Packet) {
	body := NewMsgChannelEncryptResult()
	packet.ReadMsg(body)

	if body.Result != EResult_OK {
		c.Fatalf("Encryption failed: %v", body.Result)
		return
	}
	c.conn.SetEncryptionKey(c.tempSessionKey)
	c.tempSessionKey = nil

	c.Emit(&ConnectedEvent{})
}

func (c *Client) handleMulti(packet *Packet) {
	body := new(CMsgMulti)
	packet.ReadProtoMsg(body)

	payload := body.GetMessageBody()

	if body.GetSizeUnzipped() > 0 {
		r, err := gzip.NewReader(bytes.NewReader(payload))
		if err != nil {
			c.Errorf("handleMulti: Error while decompressing: %v", err)
			return
		}

		payload, err = ioutil.ReadAll(r)
		if err != nil {
			c.Errorf("handleMulti: Error while decompressing: %v", err)
			return
		}
	}

	pr := bytes.NewReader(payload)
	for pr.Len() > 0 {
		var length uint32
		binary.Read(pr, binary.LittleEndian, &length)
		packetData := make([]byte, length)
		pr.Read(packetData)
		p, err := NewPacket(packetData)
		if err != nil {
			c.Errorf("Error reading packet in Multi msg %v: %v", packet, err)
			continue
		}
		c.handlePacket(p)
	}
}

func (c *Client) handleClientCMList(packet *Packet) {
	body := new(CMsgClientCMList)
	packet.ReadProtoMsg(body)

	l := make([]*netutil.PortAddr, 0)
	for i, ip := range body.GetCmAddresses() {
		l = append(l, &netutil.PortAddr{
			readIp(ip),
			uint16(body.GetCmPorts()[i]),
		})
	}

	c.Emit(&ClientCMListEvent{l})
}

func readIp(ip uint32) net.IP {
	r := make(net.IP, 4)
	r[3] = byte(ip)
	r[2] = byte(ip >> 8)
	r[1] = byte(ip >> 16)
	r[0] = byte(ip >> 24)
	return r
}
