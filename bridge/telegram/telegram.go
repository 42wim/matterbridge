package btelegram

import (
	"bytes"
	"html"
	"strconv"

	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/russross/blackfriday"
)

type Btelegram struct {
	c       *tgbotapi.BotAPI
	Config  *config.Protocol
	Remote  chan config.Message
	Account string
}

var flog *log.Entry
var protocol = "telegram"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Btelegram {
	b := &Btelegram{}
	b.Config = &cfg
	b.Remote = c
	b.Account = account
	return b
}

func (b *Btelegram) Connect() error {
	var err error
	flog.Info("Connecting")
	b.c, err = tgbotapi.NewBotAPI(b.Config.Token)
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	updates, err := b.c.GetUpdatesChan(tgbotapi.NewUpdate(0))
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	flog.Info("Connection succeeded")
	go b.handleRecv(updates)
	return nil
}

func (b *Btelegram) JoinChannel(channel string) error {
	return nil
}

type customHtml struct {
	blackfriday.Renderer
}

func (options *customHtml) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()

	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("\n")
}

func (options *customHtml) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString("<pre>")

	out.WriteString(html.EscapeString(string(text)))
	out.WriteString("</pre>\n")
}

func (b *Btelegram) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	chatid, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return err
	}

	parsed := blackfriday.Markdown([]byte(msg.Text),
		&customHtml{blackfriday.HtmlRenderer(blackfriday.HTML_USE_XHTML, "", "")},
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS|
			blackfriday.EXTENSION_FENCED_CODE|
			blackfriday.EXTENSION_AUTOLINK|
			blackfriday.EXTENSION_SPACE_HEADERS|
			blackfriday.EXTENSION_HEADER_IDS|
			blackfriday.EXTENSION_DEFINITION_LISTS)

	m := tgbotapi.NewMessage(chatid, msg.Username+string(parsed))
	m.ParseMode = "HTML"
	_, err = b.c.Send(m)
	return err
}

func (b *Btelegram) handleRecv(updates <-chan tgbotapi.Update) {
	for update := range updates {
		if update.Message == nil {
			continue
		}
		flog.Debugf("Sending message from %s on %s to gateway", update.Message.From.UserName, b.Account)
		b.Remote <- config.Message{Username: update.Message.From.UserName, Text: update.Message.Text, Channel: strconv.FormatInt(update.Message.Chat.ID, 10), Account: b.Account}
	}
}
