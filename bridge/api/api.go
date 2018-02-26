package api

import (
	"encoding/json"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zfjagann/golang-ring"
	"net/http"
	"sync"
	"time"
)

type Api struct {
	Messages ring.Ring
	sync.RWMutex
	*config.BridgeConfig
}

type ApiMessage struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	UserID   string `json:"userid"`
	Avatar   string `json:"avatar"`
	Gateway  string `json:"gateway"`
}

func New(cfg *config.BridgeConfig) bridge.Bridger {
	b := &Api{BridgeConfig: cfg}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	b.Messages = ring.Ring{}
	b.Messages.SetCapacity(b.Config.Buffer)
	if b.Config.Token != "" {
		e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
			return key == b.Config.Token, nil
		}))
	}
	e.GET("/api/messages", b.handleMessages)
	e.GET("/api/stream", b.handleStream)
	e.POST("/api/message", b.handlePostMessage)
	go func() {
		if b.Config.BindAddress == "" {
			b.Log.Fatalf("No BindAddress configured.")
		}
		b.Log.Infof("Listening on %s", b.Config.BindAddress)
		b.Log.Fatal(e.Start(b.Config.BindAddress))
	}()
	return b
}

func (b *Api) Connect() error {
	return nil
}
func (b *Api) Disconnect() error {
	return nil

}
func (b *Api) JoinChannel(channel config.ChannelInfo) error {
	return nil

}

func (b *Api) Send(msg config.Message) (string, error) {
	b.Lock()
	defer b.Unlock()
	// ignore delete messages
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}
	b.Messages.Enqueue(&msg)
	return "", nil
}

func (b *Api) handlePostMessage(c echo.Context) error {
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

func (b *Api) handleMessages(c echo.Context) error {
	b.Lock()
	defer b.Unlock()
	c.JSONPretty(http.StatusOK, b.Messages.Values(), " ")
	b.Messages = ring.Ring{}
	return nil
}

func (b *Api) handleStream(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
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
