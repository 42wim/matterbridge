package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	_ "github.com/42wim/matterbridge/docs" // required by swagger
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/swaggo/echo-swagger"
	"github.com/zfjagann/golang-ring"
)

// @title Matterbridge API

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

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization
func New(cfg *bridge.Config) bridge.Bridger {
	b := &API{Config: cfg}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Pre(middleware.RemoveTrailingSlash())
	b.Messages = ring.Ring{}
	if b.GetInt("Buffer") != 0 {
		b.Messages.SetCapacity(b.GetInt("Buffer"))
	}
	if b.GetString("Token") != "" {
		e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, c echo.Context) (bool, error) {
				return key == b.GetString("Token"), nil
			},
			Skipper: func(c echo.Context) bool {
				for _, path := range strings.Fields("/ /api /api/health /swagger /swagger/*") {
					if c.Path() == path {
						return true
					}
				}
				return false
			},
		}))
	}
	for _, path := range strings.Fields("/ /api /swagger") {
		e.GET(path, b.handleDocsRedirect)
	}
	e.GET("/swagger/*", echoSwagger.WrapHandler)
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

// handleHealthcheck godoc
// @Summary Checks if the server is alive.
// @Success 200 {string} string
// @Router /health [get]
func (b *API) handleHealthcheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (b *API) handleDocsRedirect(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}

// handlePostMessage godoc
// @Description Required fields: text, gateway. Optional fields: username, avatar.
// @Summary Create/Update a message
// @Accept json
// @Produce json
// @Param message body config.Message true "Message object to create"
// @Success 200 {object} config.Message
// @Security ApiKeyAuth
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
// @Security ApiKeyAuth
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
// @Security ApiKeyAuth
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

// Must go at the bottom due to bug where endpoint descriptions override.
// @description A read/write API for the Matterbridge chat bridge.
