package bxmpp

import (
	"regexp"

	"github.com/42wim/matterbridge/bridge/config"
)

var pathRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

// GetAvatar constructs a URL for a given user-avatar if it is available in the cache.
func getAvatar(av map[string]string, userid string, general *config.Protocol) string {
	if hash, ok := av[userid]; ok {
		// NOTE: This does not happen in bridge/helper/helper.go but messes up XMPP
		id := pathRegex.ReplaceAllString(userid, "_")
		return general.MediaServerDownload + "/" + hash + "/" + id + ".png"
	}
	return ""
}

func (b *Bxmpp) cacheAvatar(msg *config.Message) string {
	fi := msg.Extra["file"][0].(config.FileInfo)
	/* if we have a sha we have successfully uploaded the file to the media server,
	so we can now cache the sha */
	if fi.SHA != "" {
		b.Log.Debugf("Added %s to %s in avatarMap", fi.SHA, msg.UserID)
		b.avatarMap[msg.UserID] = fi.SHA
	}
	return ""
}
