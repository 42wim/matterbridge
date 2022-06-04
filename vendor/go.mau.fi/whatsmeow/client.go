// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package whatsmeow implements a client for interacting with the WhatsApp web multidevice API.
package whatsmeow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"go.mau.fi/whatsmeow/appstate"
	waBinary "go.mau.fi/whatsmeow/binary"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/socket"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/util/keys"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// EventHandler is a function that can handle events from WhatsApp.
type EventHandler func(evt interface{})
type nodeHandler func(node *waBinary.Node)

var nextHandlerID uint32

type wrappedEventHandler struct {
	fn EventHandler
	id uint32
}

// Client contains everything necessary to connect to and interact with the WhatsApp web API.
type Client struct {
	Store   *store.Device
	Log     waLog.Logger
	recvLog waLog.Logger
	sendLog waLog.Logger

	socket     *socket.NoiseSocket
	socketLock sync.RWMutex
	socketWait chan struct{}

	isLoggedIn            uint32
	expectedDisconnectVal uint32
	EnableAutoReconnect   bool
	LastSuccessfulConnect time.Time
	AutoReconnectErrors   int

	sendActiveReceipts uint32

	// EmitAppStateEventsOnFullSync can be set to true if you want to get app state events emitted
	// even when re-syncing the whole state.
	EmitAppStateEventsOnFullSync bool

	appStateProc     *appstate.Processor
	appStateSyncLock sync.Mutex

	historySyncNotifications  chan *waProto.HistorySyncNotification
	historySyncHandlerStarted uint32

	uploadPreKeysLock sync.Mutex
	lastPreKeyUpload  time.Time

	mediaConnCache *MediaConn
	mediaConnLock  sync.Mutex

	responseWaiters     map[string]chan<- *waBinary.Node
	responseWaitersLock sync.Mutex

	nodeHandlers      map[string]nodeHandler
	handlerQueue      chan *waBinary.Node
	eventHandlers     []wrappedEventHandler
	eventHandlersLock sync.RWMutex

	messageRetries     map[string]int
	messageRetriesLock sync.Mutex

	appStateKeyRequests     map[string]time.Time
	appStateKeyRequestsLock sync.RWMutex

	messageSendLock sync.Mutex

	privacySettingsCache atomic.Value

	groupParticipantsCache     map[types.JID][]types.JID
	groupParticipantsCacheLock sync.Mutex
	userDevicesCache           map[types.JID][]types.JID
	userDevicesCacheLock       sync.Mutex

	recentMessagesMap  map[recentMessageKey]*waProto.Message
	recentMessagesList [recentMessagesSize]recentMessageKey
	recentMessagesPtr  int
	recentMessagesLock sync.RWMutex

	sessionRecreateHistory     map[types.JID]time.Time
	sessionRecreateHistoryLock sync.Mutex
	// GetMessageForRetry is used to find the source message for handling retry receipts
	// when the message is not found in the recently sent message cache.
	GetMessageForRetry func(requester, to types.JID, id types.MessageID) *waProto.Message
	// PreRetryCallback is called before a retry receipt is accepted.
	// If it returns false, the accepting will be cancelled and the retry receipt will be ignored.
	PreRetryCallback func(receipt *events.Receipt, id types.MessageID, retryCount int, msg *waProto.Message) bool

	// Should untrusted identity errors be handled automatically? If true, the stored identity and existing signal
	// sessions will be removed on untrusted identity errors, and an events.IdentityChange will be dispatched.
	// If false, decrypting a message from untrusted devices will fail.
	AutoTrustIdentity bool

	uniqueID  string
	idCounter uint32

	proxy socket.Proxy
	http  *http.Client
}

// Size of buffer for the channel that all incoming XML nodes go through.
// In general it shouldn't go past a few buffered messages, but the channel is big to be safe.
const handlerQueueSize = 2048

// NewClient initializes a new WhatsApp web client.
//
// The logger can be nil, it will default to a no-op logger.
//
// The device store must be set. A default SQL-backed implementation is available in the store/sqlstore package.
//     container, err := sqlstore.New("sqlite3", "file:yoursqlitefile.db?_foreign_keys=on", nil)
//     if err != nil {
//         panic(err)
//     }
//     // If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
//     deviceStore, err := container.GetFirstDevice()
//     if err != nil {
//         panic(err)
//     }
//     client := whatsmeow.NewClient(deviceStore, nil)
func NewClient(deviceStore *store.Device, log waLog.Logger) *Client {
	if log == nil {
		log = waLog.Noop
	}
	randomBytes := make([]byte, 2)
	_, _ = rand.Read(randomBytes)
	cli := &Client{
		http: &http.Client{
			Transport: (http.DefaultTransport.(*http.Transport)).Clone(),
		},
		proxy:           http.ProxyFromEnvironment,
		Store:           deviceStore,
		Log:             log,
		recvLog:         log.Sub("Recv"),
		sendLog:         log.Sub("Send"),
		uniqueID:        fmt.Sprintf("%d.%d-", randomBytes[0], randomBytes[1]),
		responseWaiters: make(map[string]chan<- *waBinary.Node),
		eventHandlers:   make([]wrappedEventHandler, 0, 1),
		messageRetries:  make(map[string]int),
		handlerQueue:    make(chan *waBinary.Node, handlerQueueSize),
		appStateProc:    appstate.NewProcessor(deviceStore, log.Sub("AppState")),
		socketWait:      make(chan struct{}),

		historySyncNotifications: make(chan *waProto.HistorySyncNotification, 32),

		groupParticipantsCache: make(map[types.JID][]types.JID),
		userDevicesCache:       make(map[types.JID][]types.JID),

		recentMessagesMap:      make(map[recentMessageKey]*waProto.Message, recentMessagesSize),
		sessionRecreateHistory: make(map[types.JID]time.Time),
		GetMessageForRetry:     func(requester, to types.JID, id types.MessageID) *waProto.Message { return nil },
		appStateKeyRequests:    make(map[string]time.Time),

		EnableAutoReconnect: true,
		AutoTrustIdentity:   true,
	}
	cli.nodeHandlers = map[string]nodeHandler{
		"message":      cli.handleEncryptedMessage,
		"receipt":      cli.handleReceipt,
		"call":         cli.handleCallEvent,
		"chatstate":    cli.handleChatState,
		"presence":     cli.handlePresence,
		"notification": cli.handleNotification,
		"success":      cli.handleConnectSuccess,
		"failure":      cli.handleConnectFailure,
		"stream:error": cli.handleStreamError,
		"iq":           cli.handleIQ,
		"ib":           cli.handleIB,
		// Apparently there's also an <error> node which can have a code=479 and means "Invalid stanza sent (smax-invalid)"
	}
	return cli
}

// SetProxyAddress is a helper method that parses a URL string and calls SetProxy.
//
// Returns an error if url.Parse fails to parse the given address.
func (cli *Client) SetProxyAddress(addr string) error {
	parsed, err := url.Parse(addr)
	if err != nil {
		return err
	}
	cli.SetProxy(http.ProxyURL(parsed))
	return nil
}

// SetProxy sets the proxy to use for WhatsApp web websocket connections and media uploads/downloads.
//
// Must be called before Connect() to take effect in the websocket connection.
// If you want to change the proxy after connecting, you must call Disconnect() and then Connect() again manually.
//
// By default, the client will find the proxy from the https_proxy environment variable like Go's net/http does.
//
// To disable reading proxy info from environment variables, explicitly set the proxy to nil:
//   cli.SetProxy(nil)
//
// To use a different proxy for the websocket and media, pass a function that checks the request path or headers:
//  cli.SetProxy(func(r *http.Request) (*url.URL, error) {
//      if r.URL.Host == "web.whatsapp.com" && r.URL.Path == "/ws/chat" {
//          return websocketProxyURL, nil
//      } else {
//          return mediaProxyURL, nil
//      }
//  })
func (cli *Client) SetProxy(proxy socket.Proxy) {
	cli.proxy = proxy
	cli.http.Transport.(*http.Transport).Proxy = proxy
}

func (cli *Client) getSocketWaitChan() <-chan struct{} {
	cli.socketLock.RLock()
	ch := cli.socketWait
	cli.socketLock.RUnlock()
	return ch
}

func (cli *Client) closeSocketWaitChan() {
	cli.socketLock.Lock()
	close(cli.socketWait)
	cli.socketWait = make(chan struct{})
	cli.socketLock.Unlock()
}

func (cli *Client) WaitForConnection(timeout time.Duration) bool {
	timeoutChan := time.After(timeout)
	cli.socketLock.RLock()
	for cli.socket == nil || !cli.socket.IsConnected() || !cli.IsLoggedIn() {
		ch := cli.socketWait
		cli.socketLock.RUnlock()
		select {
		case <-ch:
		case <-timeoutChan:
			return false
		}
		cli.socketLock.RLock()
	}
	cli.socketLock.RUnlock()
	return true
}

// Connect connects the client to the WhatsApp web websocket. After connection, it will either
// authenticate if there's data in the device store, or emit a QREvent to set up a new link.
func (cli *Client) Connect() error {
	cli.socketLock.Lock()
	defer cli.socketLock.Unlock()
	if cli.socket != nil {
		if !cli.socket.IsConnected() {
			cli.unlockedDisconnect()
		} else {
			return ErrAlreadyConnected
		}
	}

	cli.resetExpectedDisconnect()
	fs := socket.NewFrameSocket(cli.Log.Sub("Socket"), socket.WAConnHeader, cli.proxy)
	if err := fs.Connect(); err != nil {
		fs.Close(0)
		return err
	} else if err = cli.doHandshake(fs, *keys.NewKeyPair()); err != nil {
		fs.Close(0)
		return fmt.Errorf("noise handshake failed: %w", err)
	}
	go cli.keepAliveLoop(cli.socket.Context())
	go cli.handlerQueueLoop(cli.socket.Context())
	return nil
}

// IsLoggedIn returns true after the client is successfully connected and authenticated on WhatsApp.
func (cli *Client) IsLoggedIn() bool {
	return atomic.LoadUint32(&cli.isLoggedIn) == 1
}

func (cli *Client) onDisconnect(ns *socket.NoiseSocket, remote bool) {
	ns.Stop(false)
	cli.socketLock.Lock()
	defer cli.socketLock.Unlock()
	if cli.socket == ns {
		cli.socket = nil
		cli.clearResponseWaiters(xmlStreamEndNode)
		if !cli.isExpectedDisconnect() && remote {
			cli.Log.Debugf("Emitting Disconnected event")
			go cli.dispatchEvent(&events.Disconnected{})
			go cli.autoReconnect()
		} else if remote {
			cli.Log.Debugf("OnDisconnect() called, but it was expected, so not emitting event")
		} else {
			cli.Log.Debugf("OnDisconnect() called after manual disconnection")
		}
	} else {
		cli.Log.Debugf("Ignoring OnDisconnect on different socket")
	}
}

func (cli *Client) expectDisconnect() {
	atomic.StoreUint32(&cli.expectedDisconnectVal, 1)
}

func (cli *Client) resetExpectedDisconnect() {
	atomic.StoreUint32(&cli.expectedDisconnectVal, 0)
}

func (cli *Client) isExpectedDisconnect() bool {
	return atomic.LoadUint32(&cli.expectedDisconnectVal) == 1
}

func (cli *Client) autoReconnect() {
	if !cli.EnableAutoReconnect || cli.Store.ID == nil {
		return
	}
	for {
		autoReconnectDelay := time.Duration(cli.AutoReconnectErrors) * 2 * time.Second
		cli.Log.Debugf("Automatically reconnecting after %v", autoReconnectDelay)
		cli.AutoReconnectErrors++
		time.Sleep(autoReconnectDelay)
		err := cli.Connect()
		if errors.Is(err, ErrAlreadyConnected) {
			cli.Log.Debugf("Connect() said we're already connected after autoreconnect sleep")
			return
		} else if err != nil {
			cli.Log.Errorf("Error reconnecting after autoreconnect sleep: %v", err)
		} else {
			return
		}
	}
}

// IsConnected checks if the client is connected to the WhatsApp web websocket.
// Note that this doesn't check if the client is authenticated. See the IsLoggedIn field for that.
func (cli *Client) IsConnected() bool {
	cli.socketLock.RLock()
	connected := cli.socket != nil && cli.socket.IsConnected()
	cli.socketLock.RUnlock()
	return connected
}

// Disconnect disconnects from the WhatsApp web websocket.
//
// This will not emit any events, the Disconnected event is only used when the
// connection is closed by the server or a network error.
func (cli *Client) Disconnect() {
	if cli.socket == nil {
		return
	}
	cli.socketLock.Lock()
	cli.unlockedDisconnect()
	cli.socketLock.Unlock()
}

// Disconnect closes the websocket connection.
func (cli *Client) unlockedDisconnect() {
	if cli.socket != nil {
		cli.socket.Stop(true)
		cli.socket = nil
		cli.clearResponseWaiters(xmlStreamEndNode)
	}
}

// Logout sends a request to unlink the device, then disconnects from the websocket and deletes the local device store.
//
// If the logout request fails, the disconnection and local data deletion will not happen either.
// If an error is returned, but you want to force disconnect/clear data, call Client.Disconnect() and Client.Store.Delete() manually.
//
// Note that this will not emit any events. The LoggedOut event is only used for external logouts
// (triggered by the user from the main device or by WhatsApp servers).
func (cli *Client) Logout() error {
	if cli.Store.ID == nil {
		return ErrNotLoggedIn
	}
	_, err := cli.sendIQ(infoQuery{
		Namespace: "md",
		Type:      "set",
		To:        types.ServerJID,
		Content: []waBinary.Node{{
			Tag: "remove-companion-device",
			Attrs: waBinary.Attrs{
				"jid":    *cli.Store.ID,
				"reason": "user_initiated",
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("error sending logout request: %w", err)
	}
	cli.Disconnect()
	err = cli.Store.Delete()
	if err != nil {
		return fmt.Errorf("error deleting data from store: %w", err)
	}
	return nil
}

// AddEventHandler registers a new function to receive all events emitted by this client.
//
// The returned integer is the event handler ID, which can be passed to RemoveEventHandler to remove it.
//
// All registered event handlers will receive all events. You should use a type switch statement to
// filter the events you want:
//   func myEventHandler(evt interface{}) {
//       switch v := evt.(type) {
//       case *events.Message:
//           fmt.Println("Received a message!")
//       case *events.Receipt:
//           fmt.Println("Received a receipt!")
//       }
//   }
//
// If you want to access the Client instance inside the event handler, the recommended way is to
// wrap the whole handler in another struct:
//   type MyClient struct {
//       WAClient *whatsmeow.Client
//       eventHandlerID uint32
//   }
//
//   func (mycli *MyClient) register() {
//       mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.myEventHandler)
//   }
//
//   func (mycli *MyClient) myEventHandler(evt interface{}) {
//        // Handle event and access mycli.WAClient
//   }
func (cli *Client) AddEventHandler(handler EventHandler) uint32 {
	nextID := atomic.AddUint32(&nextHandlerID, 1)
	cli.eventHandlersLock.Lock()
	cli.eventHandlers = append(cli.eventHandlers, wrappedEventHandler{handler, nextID})
	cli.eventHandlersLock.Unlock()
	return nextID
}

// RemoveEventHandler removes a previously registered event handler function.
// If the function with the given ID is found, this returns true.
//
// N.B. Do not run this directly from an event handler. That would cause a deadlock because the
// event dispatcher holds a read lock on the event handler list, and this method wants a write lock
// on the same list. Instead run it in a goroutine:
//   func (mycli *MyClient) myEventHandler(evt interface{}) {
//       if noLongerWantEvents {
//           go mycli.WAClient.RemoveEventHandler(mycli.eventHandlerID)
//       }
//   }
func (cli *Client) RemoveEventHandler(id uint32) bool {
	cli.eventHandlersLock.Lock()
	defer cli.eventHandlersLock.Unlock()
	for index := range cli.eventHandlers {
		if cli.eventHandlers[index].id == id {
			if index == 0 {
				cli.eventHandlers[0].fn = nil
				cli.eventHandlers = cli.eventHandlers[1:]
				return true
			} else if index < len(cli.eventHandlers)-1 {
				copy(cli.eventHandlers[index:], cli.eventHandlers[index+1:])
			}
			cli.eventHandlers[len(cli.eventHandlers)-1].fn = nil
			cli.eventHandlers = cli.eventHandlers[:len(cli.eventHandlers)-1]
			return true
		}
	}
	return false
}

// RemoveEventHandlers removes all event handlers that have been registered with AddEventHandler
func (cli *Client) RemoveEventHandlers() {
	cli.eventHandlersLock.Lock()
	cli.eventHandlers = make([]wrappedEventHandler, 0, 1)
	cli.eventHandlersLock.Unlock()
}

func (cli *Client) handleFrame(data []byte) {
	decompressed, err := waBinary.Unpack(data)
	if err != nil {
		cli.Log.Warnf("Failed to decompress frame: %v", err)
		cli.Log.Debugf("Errored frame hex: %s", hex.EncodeToString(data))
		return
	}
	node, err := waBinary.Unmarshal(decompressed)
	if err != nil {
		cli.Log.Warnf("Failed to decode node in frame: %v", err)
		cli.Log.Debugf("Errored frame hex: %s", hex.EncodeToString(decompressed))
		return
	}
	cli.recvLog.Debugf("%s", node.XMLString())
	if node.Tag == "xmlstreamend" {
		if !cli.isExpectedDisconnect() {
			cli.Log.Warnf("Received stream end frame")
		}
		// TODO should we do something else?
	} else if cli.receiveResponse(node) {
		// handled
	} else if _, ok := cli.nodeHandlers[node.Tag]; ok {
		select {
		case cli.handlerQueue <- node:
		default:
			cli.Log.Warnf("Handler queue is full, message ordering is no longer guaranteed")
			go func() {
				cli.handlerQueue <- node
			}()
		}
	} else {
		cli.Log.Debugf("Didn't handle WhatsApp node %s", node.Tag)
	}
}

func (cli *Client) handlerQueueLoop(ctx context.Context) {
	for {
		select {
		case node := <-cli.handlerQueue:
			cli.nodeHandlers[node.Tag](node)
		case <-ctx.Done():
			return
		}
	}
}

func (cli *Client) sendNodeAndGetData(node waBinary.Node) ([]byte, error) {
	cli.socketLock.RLock()
	sock := cli.socket
	cli.socketLock.RUnlock()
	if sock == nil {
		return nil, ErrNotConnected
	}

	payload, err := waBinary.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node: %w", err)
	}

	cli.sendLog.Debugf("%s", node.XMLString())
	return payload, sock.SendFrame(payload)
}

func (cli *Client) sendNode(node waBinary.Node) error {
	_, err := cli.sendNodeAndGetData(node)
	return err
}

func (cli *Client) dispatchEvent(evt interface{}) {
	cli.eventHandlersLock.RLock()
	defer func() {
		cli.eventHandlersLock.RUnlock()
		err := recover()
		if err != nil {
			cli.Log.Errorf("Event handler panicked while handling a %T: %v\n%s", evt, err, debug.Stack())
		}
	}()
	for _, handler := range cli.eventHandlers {
		handler.fn(evt)
	}
}

// ParseWebMessage parses a WebMessageInfo object into *events.Message to match what real-time messages have.
//
// The chat JID can be found in the Conversation data:
//   chatJID, err := types.ParseJID(conv.GetId())
//   for _, historyMsg := range conv.GetMessages() {
//     evt, err := cli.ParseWebMessage(chatJID, historyMsg.GetMessage())
//     yourNormalEventHandler(evt)
//   }
func (cli *Client) ParseWebMessage(chatJID types.JID, webMsg *waProto.WebMessageInfo) (*events.Message, error) {
	info := types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat:     chatJID,
			IsFromMe: webMsg.GetKey().GetFromMe(),
			IsGroup:  chatJID.Server == types.GroupServer,
		},
		ID:        webMsg.GetKey().GetId(),
		PushName:  webMsg.GetPushName(),
		Timestamp: time.Unix(int64(webMsg.GetMessageTimestamp()), 0),
	}
	var err error
	if info.IsFromMe {
		info.Sender = cli.Store.ID.ToNonAD()
	} else if chatJID.Server == types.DefaultUserServer {
		info.Sender = chatJID
	} else if webMsg.GetParticipant() != "" {
		info.Sender, err = types.ParseJID(webMsg.GetParticipant())
	} else if webMsg.GetKey().GetParticipant() != "" {
		info.Sender, err = types.ParseJID(webMsg.GetKey().GetParticipant())
	} else {
		return nil, fmt.Errorf("couldn't find sender of message %s", info.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse sender of message %s: %v", info.ID, err)
	}
	evt := &events.Message{
		RawMessage: webMsg.GetMessage(),
		Info:       info,
	}
	evt.UnwrapRaw()
	return evt, nil
}
