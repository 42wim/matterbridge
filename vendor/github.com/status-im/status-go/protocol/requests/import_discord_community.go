package requests

import (
	"errors"
)

var (
	ErrImportDiscordCommunityMissingFilesToImport = errors.New("import-discord-community: missing files to import")
)

type ImportDiscordCommunity struct {
	CreateCommunity
	FilesToImport []string
	From          int64
}

func (u *ImportDiscordCommunity) Validate() error {
	if len(u.FilesToImport) == 0 {
		return ErrImportDiscordCommunityMissingFilesToImport
	}

	return u.CreateCommunity.Validate()
}

func (u *ImportDiscordCommunity) ToCreateCommunityRequest() *CreateCommunity {
	return &u.CreateCommunity
}
