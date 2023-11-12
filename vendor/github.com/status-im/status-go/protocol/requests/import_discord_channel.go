package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

// errrors
var (
	ErrImportDiscordChannelMissingFilesToImport = errors.New("import-discord-channel: missing files to import")
	ErrImportDiscordChannelChannelIDIsEmpty     = errors.New("import-discord-channel: discord channel id is empty")
	ErrImportDiscordChannelCommunityIDIsEmpty   = errors.New("import-discord-channel: community id is empty")
)

type ImportDiscordChannel struct {
	Name             string         `json:"name"`
	DiscordChannelID string         `json:"discordChannelID"`
	CommunityID      types.HexBytes `json:"communityId"`
	Description      string         `json:"description"`
	Color            string         `json:"color"`
	Emoji            string         `json:"emoji"`
	FilesToImport    []string       `json:"filesToImport"`
	From             int64          `json:"from"`
}

func (r *ImportDiscordChannel) Validate() error {
	if len(r.FilesToImport) == 0 {
		return ErrImportDiscordChannelMissingFilesToImport
	}

	if len(r.DiscordChannelID) == 0 {
		return ErrImportDiscordChannelChannelIDIsEmpty
	}

	if len(r.CommunityID) == 0 {
		return ErrImportDiscordChannelCommunityIDIsEmpty
	}

	return nil
}
