package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	ring "github.com/zfjagann/golang-ring"
)

type API struct {
	Messages ring.Ring
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
	b.Messages = ring.Ring{}
	if b.GetInt("Buffer") != 0 {
		b.Messages.SetCapacity(b.GetInt("Buffer"))
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
	return nil

}

func (b *API) Send(msg config.Message) (string, error) {
	b.Lock()
	defer b.Unlock()
	// ignore delete messages
	if msg.Event == config.EventMsgDelete {
		return "", nil
	}
	b.Messages.Enqueue(&msg)
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
	c.JSONPretty(http.StatusOK, b.Messages.Values(), " ")
	b.Messages = ring.Ring{}
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
		msg := b.Messages.Dequeue()
		if msg != nil {
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
	for {
		msg := b.Messages.Dequeue()
		if msg != nil {
			err := conn.WriteJSON(msg)
			if err != nil {
				break
			}
		}
	}
}

func (b *API) readPump(conn *websocket.Conn) {
	for {
		message := config.Message{}
		err := conn.ReadJSON(&message)
		if err != nil {
			break
		}
		b.handleWebsocketMessage(message)
	}
}

func (b *API) handleWebsocket(c echo.Context) error {
	conn, err := websocket.Upgrade(c.Response().Writer, c.Request(), nil, 1024, 1024)
	if err != nil {
		return err
	}

	greet := b.getGreeting()
	_ = conn.WriteJSON(greet)

	go b.writePump(conn)
	go b.readPump(conn)

	return nil
}
