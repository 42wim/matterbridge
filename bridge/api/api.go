package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second // TODO: 60 seconds

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type API struct {
	send chan config.Message
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
	b.send = make(chan config.Message, b.GetInt("Buffer"))
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
	return nil

}

func (b *API) Send(msg config.Message) (string, error) {
	b.Lock()
	defer b.Unlock()
	// ignore delete messages
	if msg.Event == config.EventMsgDelete {
		return "", nil
	}
	b.send <- msg
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
	// collect all messages until the channel has no more messages in the buffer
	var messages []config.Message
	loop: for {
		select {
		case msg := <- b.send:
			messages = append(messages, msg)
		default:
			break loop
		}
	}
	// TODO: get all messages from send channel
	c.JSONPretty(http.StatusOK, messages, " ")
	// TODO: clear send channel ?
	//b.send = make(chan config.Message, b.GetInt("Buffer"))
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
	for {
		select {
		// block until channel has message
		case msg := <- b.send:
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

func (b *API) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		b.Log.Debug("closing websocket")
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg := <-b.send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := conn.WriteJSON(msg)
			if err != nil {
				b.Log.Errorf("error: %v", err)
				return
			}
		case <-ticker.C:
			b.Log.Debug("sending ping")
			conn.SetWriteDeadline(time.Now().Add(writeWait))
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
		conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(
		func(string) error {
			b.Log.Debug("received pong")
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

    for {
		message := config.Message{}
		//err := conn.ReadJSON(&message)
		//if err != nil {
		//	b.Log.Errorf("error: %v", err)
		//	return
		//}
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			b.Log.Errorf("error: %v", err)
			return
		}
		err = json.NewDecoder(bytes.NewReader(messageBytes)).Decode(&message)
		if err != nil {
			if err == io.EOF {
				// One value is expected in the message.
				err = io.ErrUnexpectedEOF
			}
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

	go b.writePump(conn)
	go b.readPump(conn)

	return nil
}
