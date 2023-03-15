package gateway

import (
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
)

// returns true if a command was registered (therefore a should not be relayed
func (r *Router) handleCommand(msg *config.Message) bool {
	help := `!optout - opt out from all message relaying
	!optoutmedia - only opt out from relaying attachments
	!optin - opt back into chat relaying
	!setwelcome - set channel welcome message (admin)
	!unsetwelcome - clear channel welcome message (admin)
	!help - display this message`

	isAdmin := r.isAdmin(msg)
	addTextFromCaptions(msg)
	cmd := msg.Text

	switch {
	case cmd == "!help":
		r.logger.Debug("!help")
		r.replyCmd(msg, help)
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
	case cmd == "!optin":
		r.logger.Debugf("!optin: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptIn)
	case cmd == "!optout":
		r.logger.Debugf("!optout: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptOut)
	case cmd == "!optoutmedia":
		r.logger.Debugf("!optoutmedia: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptOutMediaOnly)
	case isAdmin && strings.HasPrefix(cmd, "!setwelcome"):
		r.logger.Debugf("!setwelcome: %s - %+v", msg.Channel, msg)
		r.handleWelcomeCmd(msg, msg)
	case isAdmin && strings.HasPrefix(cmd, "!unsetwelcome"):
		r.logger.Debugf("!unsetwelcome: %s", msg.Channel)
		r.handleWelcomeCmd(msg, nil)
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

	if r.getWelcomeMessage(msg.Channel) == nil {
		r.replyCmd(msg, "No welcome message configured, set with !setwelcome")
		return
	}

	r.handleEventWelcome(msg)
}
