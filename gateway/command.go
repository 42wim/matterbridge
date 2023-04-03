package gateway

import (
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
)

func (r *Router) handleHelp(msg *config.Message) {
	help := `*!help* - display this message
*!optout* - opt out from all message relaying
*!optoutmedia* - only opt out from relaying attachments
*!optin* - opt back into chat relaying
*!setname {NewName}* - use "NewName" as a nickname
*!unsetname* - clear custom name on the bridge
*!setavatar {imageURL}* - use "imageURL" as a custom avatar (will only show up on discord - must be a public image url)
*!unsetavatar* - clear custom avatar on the bridge
*!showstatus* - show your current user preferences`

	isAdmin := r.isAdmin(msg)

	if msg.Channel != msg.UserID {
		help = "Note that commands can be run from any chat - *including DM'ing the bot.*\n\n" + help
	}
	if isAdmin {
		help += `

*Admin Commands:*
*!setwelcome {WelcomeMsg}* - set channel welcome message (supports attachments)
*!unsetwelcome* - clear channel welcome message`
	}

	if msg.Protocol != "whatsapp" && msg.Protocol != "discord" {
		help = strings.ReplaceAll(help, "*", "")
	}

	r.replyCmd(msg, help)
}

// returns true if a command was registered (therefore a should not be relayed
func (r *Router) handleCommand(msg *config.Message) bool {
	isAdmin := r.isAdmin(msg)
	addTextFromCaptions(msg) // todo: figure out if this would cause a bug
	cmd := msg.Text

	switch {
	case cmd == "!help":
		r.logger.Debug("!help")
		r.handleHelp(msg)
	case cmd == "!optin":
		r.logger.Debugf("!optin: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptIn)
	case cmd == "!optout":
		r.logger.Debugf("!optout: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptOut)
	case cmd == "!optoutmedia":
		r.logger.Debugf("!optoutmedia: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptOutMediaOnly)
	case strings.HasPrefix(cmd, "!setname"):
		r.logger.Debugf("%s - %s", cmd, msg.UserID)
		r.handleNameCmd(msg, cmd)
	case cmd == "!unsetname":
		r.logger.Debugf("!unsetname: %s", msg.UserID)
		r.handleNameCmd(msg, "!setname ") // bit of a hack lol
	case strings.HasPrefix(cmd, "!setavatar"):
		r.logger.Debugf("%s - %s", cmd, msg.UserID)
		r.handleAvatarCmd(msg, cmd)
	case cmd == "!unsetavatar":
		r.logger.Debugf("!unsetname: %s", msg.UserID)
		r.handleAvatarCmd(msg, "!setavatar ") // bit of a hack lol
	case cmd == "!showstatus":
		r.logger.Debugf("!showstatus: %s", msg.UserID)
		r.handleStatusCmd(msg)
	// ! ------- admin commands -------
	case isAdmin && strings.HasPrefix(cmd, "!setwelcome"):
		r.logger.Debugf("!setwelcome: %s - %+v", msg.Channel, msg)
		r.handleWelcomeCmd(msg, msg)
	case isAdmin && strings.HasPrefix(cmd, "!unsetwelcome"):
		r.logger.Debugf("!unsetwelcome: %s", msg.Channel)
		r.handleWelcomeCmd(msg, nil)
	// ! ------- debug commands -------
	case cmd == "!chatId":
		r.logger.Infof("!chatId: %s", msg.Channel)
	case cmd == "!userId":
		r.logger.Infof("!userId: %s", msg.UserID)
	case cmd == "!ping":
		r.logger.Debug("!pong: %s,%s", msg.Channel, msg.UserID)
		r.replyCmd(msg, "pong!")
	case cmd == "!pingdm":
		r.logger.Debug("!pongdm: %s,%s", msg.Channel, msg.UserID)
		r.replyDM(msg, "pong!")
	case cmd == "!echowelcome":
		r.logger.Debugf("!echowelcome: %s,%s", msg.Channel, msg.UserID)
		r.handleEchoWelcomeCmd(msg)
	default:
		return false
	}

	return true
}

func (r *Router) isAdmin(msg *config.Message) bool {
	admins, _ := r.GetStringSlice("Admins")

	for _, ID := range admins {
		if msg.UserID == ID {
			return true
		}
	}
	return false
}

func addTextFromCaptions(msg *config.Message) {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)

		msg.Text += fi.Comment
	}
}

func (r *Router) replyCmd(msg *config.Message, str string) {
	srcBridge := r.getBridge(msg.Account)

	reply := config.Message{
		Text:     str,
		Channel:  msg.Channel,
		Account:  msg.Account,
		Username: "",
		UserID:   "",
		Protocol: msg.Protocol,
		Gateway:  msg.Gateway,
		ParentID: msg.ID,
	}

	srcBridge.Send(reply)
}

func (r *Router) replyDM(msg *config.Message, str string) {
	srcBridge := r.getBridge(msg.Account)

	reply := config.Message{
		Text:     str,
		Channel:  msg.UserID,
		Account:  msg.Account,
		Username: "",
		UserID:   "",
		Protocol: msg.Protocol,
		Gateway:  msg.Gateway,
	}

	srcBridge.Send(reply)
}

func (r *Router) sendDM(msg *config.Message, dmChannel string) {
	srcBridge := r.getBridge(msg.Account)

	msg.Channel = dmChannel
	msg.Username = ""
	msg.UserID = ""
	msg.ID = ""
	msg.Event = ""

	srcBridge.Send(*msg)
}

func (r *Router) handleOptOutCmd(msg *config.Message, newStatus OptOutStatus) {
	err := r.setOptOutStatus(msg.UserID, newStatus)

	reply := "Successfully set message relay preferences."
	if err != nil {
		reply = "Error setting message relay preferences, try again later or contact the moderators."
	}

	r.replyCmd(msg, reply)
}

func (r *Router) handleNameCmd(msg *config.Message, cmd string) {

	newName := strings.Replace(cmd, "!setname ", "", 1)

	err := r.setUserName(msg.UserID, newName)

	reply := "Successfully set new name: " + r.getUserName(msg)
	if err != nil {
		reply = "Error setting nickname, try again later or contact the moderators."
	}

	r.replyCmd(msg, reply)
}

func (r *Router) handleAvatarCmd(msg *config.Message, cmd string) {

	newAvatar := strings.Replace(cmd, "!setavatar ", "", 1)

	err := r.setAvatar(msg.UserID, newAvatar)

	reply := "Successfully set new avatar: " + r.getAvatar(msg)
	if err != nil {
		reply = "Error setting avatar, try again later or contact the moderators."
	}

	r.replyCmd(msg, reply)
}

func (r *Router) handleStatusCmd(msg *config.Message) {

	reply := r.getUserPreferencesStr(msg)
	r.replyCmd(msg, reply)
}

func (r *Router) handleWelcomeCmd(msg *config.Message, welcomeMsg *config.Message) {

	if welcomeMsg != nil {
		welcomeMsg.Text = strings.Replace(welcomeMsg.Text, "!setwelcome ", "", 1)

		for i, f := range welcomeMsg.Extra["file"] {
			fi := f.(config.FileInfo)
			fi.Comment = strings.Replace(fi.Comment, "!setwelcome ", "", 1)

			welcomeMsg.Extra["file"][i] = fi
		}
	}

	err := r.setWelcomeMessage(msg.Channel, welcomeMsg)

	reply := "Successfully set welcome message for channel."
	if welcomeMsg == nil {
		reply = "Successfully removed welcome message for channel."
	}
	if err != nil {
		reply = "Error setting channel welcome message, try again later or contact the moderators."
	}

	r.replyCmd(msg, reply)
}

func (r *Router) handleEchoWelcomeCmd(msg *config.Message) {
	msg.Event = config.EventWelcomeMsg

	srcBridge := r.getBridge(msg.Account)
	str := srcBridge.Channels[getChannelID(msg)].Options.WelcomeMessage

	if r.getWelcomeMessage(msg.Channel) == nil && str == "" {
		r.replyCmd(msg, "No welcome message configured, set with !setwelcome")
		return
	}

	r.handleEventWelcome(msg)
}
