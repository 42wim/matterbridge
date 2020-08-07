package api

import (
	"encoding/json"
	"github.com/grafov/bcast"
	"net/http"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/gorilla/websocket"
	//"github.com/grafov/bcast"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	ring "github.com/zfjagann/golang-ring"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second // TODO: 60 seconds

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type API struct {
	Messages       ring.Ring
	group          bcast.Group
	//messageMember  bcast.Member
	//messageChannel chan config.Message
	//streamMember   bcast.Member
	//streamChannel  chan config.Message
	sync.RWMutex
	*bridge.Config
}

type Message struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	UserID   string `json:"userid"`
	Avatar   string `json:"avatar"`
	Gateway  string `json:"gateway"`
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &API{Config: cfg}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	b.group = *bcast.NewGroup()
	go b.group.Broadcast(0) // TODO: cancel this group broadcast at some point ?

	b.Messages = ring.Ring{}
	if b.GetInt("Buffer") != 0 {
		b.Messages.SetCapacity(b.GetInt("Buffer"))
	} else {
		// TODO: set default capacity ?
	}

	if b.GetString("Token") != "" {
		e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
			return key == b.GetString("Token"), nil
		}))
	}

	// Set RemoteNickFormat to a sane default
	if !b.IsKeySet("RemoteNickFormat") {
		b.Log.Debugln("RemoteNickFormat is unset, defaulting to \"{NICK}\"")
		b.Config.Config.Viper().Set(b.GetConfigKey("RemoteNickFormat"), "{NICK}")
	}

	e.GET("/api/health", b.handleHealthcheck)
	e.GET("/api/messages", b.handleMessages)
	e.GET("/api/stream", b.handleStream)
	e.GET("/api/websocket", b.handleWebsocket)
	e.POST("/api/message", b.handlePostMessage)
	go func() {
		if b.GetString("BindAddress") == "" {
			b.Log.Fatalf("No BindAddress configured.")
		}
		b.Log.Infof("Listening on %s", b.GetString("BindAddress"))
		b.Log.Fatal(e.Start(b.GetString("BindAddress")))
	}()
	return b
}

func (b *API) Connect() error {
	return nil
}
func (b *API) Disconnect() error {
	return nil

}
func (b *API) JoinChannel(channel config.ChannelInfo) error {
	// we could have a `chan config.Message` for each text channel here, instead of hardcoded "api"
	return nil

}

func (b *API) Send(msg config.Message) (string, error) {
	b.Lock()
	defer b.Unlock()
	// ignore delete messages
	if msg.Event == config.EventMsgDelete {
		return "", nil
	}
	b.Log.Debugf("enqueueing message from %s to group broadcast", msg.Username)
	b.Messages.Enqueue(msg)
	b.group.Send(msg)
	return "", nil
}

func (b *API) handleHealthcheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (b *API) handlePostMessage(c echo.Context) error {
	message := config.Message{}
	if err := c.Bind(&message); err != nil {
		return err
	}
	// these values are fixed
	message.Channel = "api"
	message.Protocol = "api"
	message.Account = b.Account
	message.ID = ""
	message.Timestamp = time.Now()
	b.Log.Debugf("Sending message from %s on %s to gateway", message.Username, "api")
	b.Remote <- message
	return c.JSON(http.StatusOK, message)
}

func (b *API) handleMessages(c echo.Context) error {
	b.Lock()
	defer b.Unlock()
	_ = c.JSONPretty(http.StatusOK, b.Messages.Values(), " ")
	// not clearing history.. intentionally
	//b.Messages = ring.Ring{}
	return nil
}

func (b *API) getGreeting() config.Message {
	return config.Message{
		Event:     config.EventAPIConnected,
		Timestamp: time.Now(),
	}
}

func (b *API) handleStream(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	greet := b.getGreeting()
	if err := json.NewEncoder(c.Response()).Encode(greet); err != nil {
		return err
	}
	c.Response().Flush()

	// TODO: currently this skips sending history
	// TODO: send history from ringbuffer ?

	member := *b.group.Join()
	defer func() {
		// i hope this will properly close it..
		member.Close()
	}()

	for {
		select {
		// block until channel has message
		case msg := <-member.Read:
			if err := json.NewEncoder(c.Response()).Encode(msg); err != nil {
				return err
			}
			c.Response().Flush()
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (b *API) handleWebsocketMessage(message config.Message) {
	message.Channel = "api"
	message.Protocol = "api"
	message.Account = b.Account
	message.ID = ""
	message.Timestamp = time.Now()

	b.Log.Debugf("Sending websocket message from %s on %s to gateway", message.Username, "api")
	b.Remote <- message
}

func (b *API) writePump(conn *websocket.Conn, member bcast.Member) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		b.Log.Debug("closing websocket")
		ticker.Stop()
		_ = conn.Close()
		member.Close()
	}()

	for {
		select {
		case msg := <-member.Read:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			b.Log.Debugf("sending message %v", msg)
			err := conn.WriteJSON(msg)
			if err != nil {
				b.Log.Errorf("error: %v", err)
				return
			}
		case <-ticker.C:
			b.Log.Debug("sending ping")
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				b.Log.Errorf("error: %v", err)
				return
			}
		}
	}
}

func (b *API) readPump(conn *websocket.Conn) {
	defer func() {
		b.Log.Debug("closing websocket")
		_ = conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(
		func(string) error {
			b.Log.Debug("received pong")
			_ = conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		message := config.Message{}
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				b.Log.Errorf("Websocket closed unexpectedly: %v", err)
			}
			return
		}
		b.handleWebsocketMessage(message)
	}
}

func (b *API) handleWebsocket(c echo.Context) error {
	u := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := u.Upgrade(c.Response().Writer, c.Request(), nil)
	//websocket.Upgrade(c.Response().Writer, c.Request(), nil, 1024, 1024)
	if err != nil {
		return err
	}

	greet := b.getGreeting()
	_ = conn.WriteJSON(greet)

	// TODO: maybe send all history as single message as json array ?

	// send all messages from history
	for _, msg := range b.Messages.Values() {
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		b.Log.Debugf("sending message %v", msg)
		err := conn.WriteJSON(msg)
		if err != nil {
			b.Log.Errorf("error: %v", err)
			break
		}
	}

	member := *b.group.Join()

	go b.writePump(conn, member)
	go b.readPump(conn)

	return nil
}
