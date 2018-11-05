package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zfjagann/golang-ring"
	"github.com/swaggo/echo-swagger"
	_ "github.com/42wim/matterbridge/docs"
)

// @title Matterbridge API
// @description A read/write API for the Matterbridge chat bridge.

// @license.name Apache 2.0
// @license.url https://github.com/42wim/matterbridge/blob/master/LICENSE

// TODO @host
// @basePath /api

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

// @securityDefinitions.apikey
// @in header
// @name Authorization
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
	for _, path := range strings.Fields("/api /swagger /") {
		e.GET(path, b.handleDocsRedirect)
	}
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/swagger", b.handleDocsRedirect)
	e.GET("/", b.handleDocsRedirect)
	e.GET("/api", b.handleDocsRedirect)
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

func (b *Api) handleDocsRedirect(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}

// handlePostMessage godoc
// @Summary Create/Update a message
// @Accept json
// @Produce json
// @Param message body config.Message true "Message object to create"
// @Success 200 {object} config.Message
// @Router /message [post]
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

// handleMessages godoc
// @Summary List new messages
// @Produce json
// @Success 200 {array} config.Message
// @Router /messages [get]
func (b *API) handleMessages(c echo.Context) error {
	b.Lock()
	defer b.Unlock()
	c.JSONPretty(http.StatusOK, b.Messages.Values(), " ")
	b.Messages = ring.Ring{}
	return nil
}

// handleStream godoc
// @Summary Stream realtime messages
// @Produce json-stream
// @Success 200 {object} config.Message
// @Router /stream [get]
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
