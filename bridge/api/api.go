package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"gopkg.in/olahol/melody.v1"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	ring "github.com/zfjagann/golang-ring"
)

type API struct {
	Messages ring.Ring
	sync.RWMutex
	*bridge.Config
	mrouter *melody.Melody
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

	b.mrouter = melody.New()
	b.mrouter.HandleMessage(func(s *melody.Session, msg []byte) {
		message := config.Message{}
		err := json.Unmarshal(msg, &message)
		if err != nil {
			b.Log.Errorf("failed to decode message from byte[] '%s'", string(msg))
			return
		}
		b.handleWebsocketMessage(message, s)
	})
	b.mrouter.HandleConnect(func(session *melody.Session) {
		greet := b.getGreeting()
		data, err := json.Marshal(greet)
		if err != nil {
			b.Log.Errorf("failed to encode message '%v'", greet)
			return
		}
		err = session.Write(data)
		if err != nil {
			b.Log.Errorf("failed to write message '%s'", string(data))
			return
		}
		// TODO: send message history buffer from `b.Messages` here
	})

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
	b.Log.Debugf("enqueueing message from %s on ring buffer", msg.Username)
	b.Messages.Enqueue(msg)

	data, err := json.Marshal(msg)
	if err != nil {
		b.Log.Errorf("failed to encode message  '%s'", msg)
	}
	_ = b.mrouter.Broadcast(data)
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
		// TODO: this causes issues, messages should be broadcasted to all connected clients
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

func (b *API) handleWebsocketMessage(message config.Message, s *melody.Session) {
	message.Channel = "api"
	message.Protocol = "api"
	message.Account = b.Account
	message.ID = ""
	message.Timestamp = time.Now()

	data, err := json.Marshal(message)
	if err != nil {
		b.Log.Errorf("failed to encode message for loopback '%v'", message)
		return
	}
	_ = b.mrouter.BroadcastOthers(data, s)

	b.Log.Debugf("Sending websocket message from %s on %s to gateway", message.Username, "api")
	b.Remote <- message
}

func (b *API) handleWebsocket(c echo.Context) error {
	err := b.mrouter.HandleRequest(c.Response(), c.Request())
	if err != nil {
		b.Log.Errorf("error in websocket handling  '%v'", err)
		return err
	}

	return nil
}
