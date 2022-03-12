package transmitter

import (
	"github.com/bwmarrin/discordgo"
)

// isDiscordPermissionError returns false for nil, and true if a Discord RESTError with code discordgo.ErrorCodeMissionPermissions
func isDiscordPermissionError(err error) bool {
	if err == nil {
		return false
	}

	restErr, ok := err.(*discordgo.RESTError)
	if !ok {
		return false
	}

	return restErr.Message != nil && restErr.Message.Code == discordgo.ErrCodeMissingPermissions
}

// getDiscordUserID gets own user ID from state, and fallback on API request
func getDiscordUserID(session *discordgo.Session) (string, error) {
	if user := session.State.User; user != nil {
		return user.ID, nil
	}

	user, err := session.User("@me")
	if err != nil {
		return "", err
	}
	return user.ID, nil
}
