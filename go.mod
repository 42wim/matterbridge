module github.com/42wim/matterbridge

require (
	github.com/42wim/go-gitter v0.0.0-20170828205020-017310c2d557
	github.com/Baozisoftware/qrcode-terminal-go v0.0.0-20170407111555-c0650d8dff0f
	github.com/Jeffail/gabs v1.1.1 // indirect
	github.com/Philipp15b/go-steam v1.0.1-0.20190816133340-b04c5a83c1c0
	github.com/Rhymen/go-whatsapp v0.0.3-0.20191003184814-fc3f792c814c
	github.com/bwmarrin/discordgo v0.19.0
	// github.com/bwmarrin/discordgo v0.19.0
	github.com/d5/tengo v1.24.8
	github.com/dfordsoft/golib v0.0.0-20180902042739-76ee6ab99bec
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.5-0.20181225215658-ec221ba9ea45+incompatible
	github.com/google/gops v0.3.6
	github.com/gopackage/ddp v0.0.0-20170117053602-652027933df4 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20180628210949-0892b62f0d9f // indirect
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/golang-lru v0.5.3
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/jpillora/backoff v0.0.0-20180909062703-3050d21c67d7
	github.com/jtolds/gls v4.2.1+incompatible // indirect
	github.com/keybase/go-keybase-chat-bot v0.0.0-20190816161829-561f10822eb2
	github.com/labstack/echo/v4 v4.1.10
	github.com/lrstanley/girc v0.0.0-20190801035559-4fc93959e1a7
	github.com/matterbridge/Rocket.Chat.Go.SDK v0.0.0-20190210153444-cc9d05784d5d
	github.com/matterbridge/go-xmpp v0.0.0-20180529212104-cd19799fba91
	github.com/matterbridge/gomatrix v0.0.0-20191026211822-6fc7accd00ca
	github.com/matterbridge/gozulipbot v0.0.0-20190212232658-7aa251978a18
	github.com/matterbridge/logrus-prefixed-formatter v0.0.0-20180806162718-01618749af61
	github.com/mattermost/mattermost-server v5.5.0+incompatible
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mreiferson/go-httpclient v0.0.0-20160630210159-31f0106b4474 // indirect
	github.com/mrexodia/wray v0.0.0-20160318003008-78a2c1f284ff // indirect
	github.com/nelsonken/gomf v0.0.0-20180504123937-a9dd2f9deae9
	github.com/nicksnyder/go-i18n v1.4.0 // indirect
	github.com/nlopes/slack v0.6.0
	github.com/onsi/ginkgo v1.6.0 // indirect
	github.com/onsi/gomega v1.4.1 // indirect
	github.com/paulrosania/go-charset v0.0.0-20190326053356-55c9d7a5834c
	github.com/pborman/uuid v0.0.0-20160216163710-c55201b03606 // indirect
	github.com/peterhellberg/emojilib v0.0.0-20190124112554-c18758d55320
	github.com/rs/xid v1.2.1
	github.com/russross/blackfriday v1.5.2
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca
	github.com/shazow/ssh-chat v0.0.0-20190125184227-81d7e1686296
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/assertions v0.0.0-20180803164922-886ec427f6b9 // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a // indirect
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	github.com/zfjagann/golang-ring v0.0.0-20190304061218-d34796e0a6c2
	gitlab.com/golang-commonmark/html v0.0.0-20180917080848-cfaf75183c4a // indirect
	gitlab.com/golang-commonmark/linkify v0.0.0-20191026162114-a0c2df6c8f82 // indirect
	gitlab.com/golang-commonmark/markdown v0.0.0-20181102083822-772775880e1f
	gitlab.com/golang-commonmark/mdurl v0.0.0-20180912090424-e5bce34c34f2 // indirect
	gitlab.com/golang-commonmark/puny v0.0.0-20180912090636-2cd490539afe // indirect
	gitlab.com/opennota/wd v0.0.0-20180912061657-c5d65f63c638 // indirect
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586 // indirect
	golang.org/x/image v0.0.0-20190902063713-cb417be4ba39
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

replace github.com/bwmarrin/discordgo v0.19.0 => github.com/matterbridge/discordgo v0.0.0-20191026232317-01823f4ebba4

go 1.13
