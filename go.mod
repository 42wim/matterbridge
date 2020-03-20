module github.com/42wim/matterbridge

require (
	github.com/42wim/go-gitter v0.0.0-20170828205020-017310c2d557
	github.com/Baozisoftware/qrcode-terminal-go v0.0.0-20170407111555-c0650d8dff0f
	github.com/Jeffail/gabs v1.1.1 // indirect
	github.com/Philipp15b/go-steam v1.0.1-0.20190816133340-b04c5a83c1c0
	github.com/Rhymen/go-whatsapp v0.1.0
	github.com/d5/tengo/v2 v2.0.2
	github.com/dfordsoft/golib v0.0.0-20180902042739-76ee6ab99bec
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.5-0.20181225215658-ec221ba9ea45+incompatible
	github.com/gomarkdown/markdown v0.0.0-20200127000047-1813ea067497
	github.com/google/gops v0.3.6
	github.com/gopackage/ddp v0.0.0-20170117053602-652027933df4 // indirect
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/golang-lru v0.5.3
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0
	github.com/keybase/go-keybase-chat-bot v0.0.0-20200226211841-4e48f3eaef3e
	github.com/labstack/echo/v4 v4.1.13
	github.com/lrstanley/girc v0.0.0-20190801035559-4fc93959e1a7
	github.com/matterbridge/Rocket.Chat.Go.SDK v0.0.0-20190210153444-cc9d05784d5d
	github.com/matterbridge/discordgo v0.18.1-0.20200308151012-aa40f01cbcc3
	github.com/matterbridge/emoji v2.1.1-0.20191117213217-af507f6b02db+incompatible
	github.com/matterbridge/go-xmpp v0.0.0-20180529212104-cd19799fba91
	github.com/matterbridge/gomatrix v0.0.0-20200209224845-c2104d7936a6
	github.com/matterbridge/gozulipbot v0.0.0-20190212232658-7aa251978a18
	github.com/matterbridge/logrus-prefixed-formatter v0.0.0-20180806162718-01618749af61
	github.com/matterbridge/msgraph.go v0.0.0-20200308150230-9e043fe9dbaa
	github.com/mattermost/mattermost-server v5.5.0+incompatible
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mattn/godown v0.0.0-20180312012330-2e9e17e0ea51
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mreiferson/go-httpclient v0.0.0-20160630210159-31f0106b4474 // indirect
	github.com/mrexodia/wray v0.0.0-20160318003008-78a2c1f284ff // indirect
	github.com/nelsonken/gomf v0.0.0-20180504123937-a9dd2f9deae9
	github.com/nicksnyder/go-i18n v1.4.0 // indirect
	github.com/onsi/ginkgo v1.6.0 // indirect
	github.com/onsi/gomega v1.4.1 // indirect
	github.com/paulrosania/go-charset v0.0.0-20190326053356-55c9d7a5834c
	github.com/pborman/uuid v0.0.0-20160216163710-c55201b03606 // indirect
	github.com/rs/xid v1.2.1
	github.com/russross/blackfriday v1.5.2
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca
	github.com/shazow/ssh-chat v1.8.3-0.20200308224626-80ddf1f43a98
	github.com/sirupsen/logrus v1.4.2
	github.com/slack-go/slack v0.6.3-0.20200228121756-f56d616d5901
	github.com/spf13/viper v1.6.1
	github.com/stretchr/testify v1.4.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	github.com/zfjagann/golang-ring v0.0.0-20190106091943-a88bb6aef447
	golang.org/x/image v0.0.0-20191214001246-9130b4cfad52
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

//replace github.com/bwmarrin/discordgo v0.20.2 => github.com/matterbridge/discordgo v0.18.1-0.20200109173909-ed873362fa43

//replace github.com/yaegashi/msgraph.go => github.com/matterbridge/msgraph.go v0.0.0-20191226214848-9e5d9c08a4e1

go 1.13
