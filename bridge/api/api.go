package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zfjagann/golang-ring"
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
	e.GET("/api/health", b.handleHealthcheck)
	e.GET("/api/messages", b.handleMessages)
	e.GET("/api/stream", b.handleStream)
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

func (b *API) handleStream(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	greet := config.Message{
		Event:     config.EventAPIConnected,
		Timestamp: time.Now(),
	}
	if err := json.NewEncoder(c.Response()).Encode(greet); err != nil {
		return err
	}
	c.Response().Flush()
	closeNotifier := c.Response().CloseNotify()
	for {
		select {
		case <-closeNotifier:
			return nil
		default:
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
}
