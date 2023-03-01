package gateway

import (
	"github.com/42wim/matterbridge/bridge/config"
)

// returns true if a command was registered (therefore a should not be relayed
func (r *Router) handleCommand(msg *config.Message) bool {
	switch text := msg.Text; text {
	case "!chatId":
		r.logger.Infof("!chatId: %s", msg.Channel)
	case "!optin":
		r.logger.Debugf("!optin: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptIn)
	case "!optout":
		r.logger.Debugf("!optout: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptOut)
	case "!optoutmedia":
		r.logger.Debugf("!optoutmedia: %s", msg.UserID)
		r.handleOptOutCmd(msg, OptOutMediaOnly)
	case "!help":
		r.logger.Debug("!help")
		help := `!optout - opt out from all message relaying
!optoutmedia - only opt out from relaying attachments
!optin - opt back into chat relaying
!help - display this message`

		r.replyCmd(msg, help)
	case "!ping":
		r.logger.Debug("!pong:")
		r.replyCmd(msg, "pong!")
	default:
		return false
	}
	return true
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

func (r *Router) handleOptOutCmd(msg *config.Message, newStaus OptOutStatus) {
	err := r.setOptOutStatus(msg.UserID, newStaus)

	reply := "Successfully set message relay preferences."
	if err != nil {
		reply = "Error setting message relay preferences, try again later or contact the moderators."
	}

	r.replyCmd(msg, reply)
}
