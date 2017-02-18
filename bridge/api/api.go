package api

import (
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/zfjagann/golang-ring"
	"net/http"
	"sync"
)

type Api struct {
	Config   *config.Protocol
	Remote   chan config.Message
	Account  string
	Messages ring.Ring
	sync.RWMutex
}

type ApiMessage struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

var flog *log.Entry
var protocol = "api"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Api {
	b := &Api{}
	e := echo.New()
	b.Messages = ring.Ring{}
	b.Messages.SetCapacity(cfg.Buffer)
	b.Config = &cfg
	b.Account = account
	b.Remote = c
	e.GET("/api/messages", b.handleMessages)
	e.POST("/api/message", b.handlePostMessage)
	go func() {
		flog.Fatal(e.Start(cfg.BindAddress))
	}()
	return b
}

func (b *Api) Connect() error {
	return nil
}
func (b *Api) Disconnect() error {
	return nil

}
func (b *Api) JoinChannel(channel string) error {
	return nil

}

func (b *Api) Send(msg config.Message) error {
	b.Lock()
	defer b.Unlock()
	b.Messages.Enqueue(&msg)
	return nil
}

func (b *Api) handlePostMessage(c echo.Context) error {
	message := &ApiMessage{}
	if err := c.Bind(message); err != nil {
		return err
	}
	b.Remote <- config.Message{
		Text:     message.Text,
		Username: message.Username,
		Channel:  "api",
		Avatar:   message.Avatar,
		Account:  b.Account,
	}
	return c.JSON(http.StatusOK, message)
}

func (b *Api) handleMessages(c echo.Context) error {
	b.Lock()
	defer b.Unlock()
	for _, msg := range b.Messages.Values() {
		c.JSONPretty(http.StatusOK, msg, " ")
	}
	b.Messages = ring.Ring{}
	return nil
}
