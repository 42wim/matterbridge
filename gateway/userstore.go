package gateway

import (
	"fmt"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/bbolt"
	"github.com/philippgille/gokv/encoding"
)

type OptOutStatus int64

const (
	OptIn           OptOutStatus = 0
	OptOut          OptOutStatus = 1
	OptOutMediaOnly OptOutStatus = 2
)

type UserData struct {
	OptOut   OptOutStatus
	UserName string
	Avatar   string
}

func (r *Router) getUserStore(path string) gokv.Store {
	options := bbolt.Options{
		BucketName: "UserData",
		Path:       path,
		Codec:      encoding.Gob,
	}

	store, err := bbolt.NewStore(options)
	if err != nil {
		r.logger.Errorf("Could not connect to db: %s", path)
	}

	return store
}

func (r *Router) getUserPreferencesStr(msg *config.Message) string {
	optStr := getOptStr(r.getOptOutStatus(msg.UserID))
	userName := r.getUserName(msg)
	avatar := r.getAvatar(msg)
	if avatar == "" {
		avatar = "None"
	}

	status := fmt.Sprintf(`User Preferences:
OptIn Status: %s
UserName: %s
Avatar: %s
`, optStr, userName, avatar)

	return status
}

func (r *Router) getOptOutStatus(UserID string) OptOutStatus {
	userdata := new(UserData)
	found, err := r.UserStore.Get(UserID, userdata)
	if err != nil {
		r.logger.Error(err)
	}

	if found {
		return userdata.OptOut
	}

	return OptIn
}

func (r *Router) setOptOutStatus(UserID string, newStatus OptOutStatus) error {
	userdata := new(UserData)
	r.UserStore.Get(UserID, userdata)

	userdata.OptOut = newStatus

	err := r.UserStore.Set(UserID, userdata)
	if err != nil {
		r.logger.Errorf(err.Error())
	}
	return err
}

func (r *Router) getUserName(msg *config.Message) string {
	userdata := new(UserData)
	found, err := r.UserStore.Get(msg.UserID, userdata)
	if err != nil {
		r.logger.Error(err)
	}

	if found && userdata.UserName != "" {
		return userdata.UserName
	}

	return msg.Username
}

func (r *Router) setUserName(UserID string, newName string) error {
	userdata := new(UserData)
	r.UserStore.Get(UserID, userdata)

	userdata.UserName = newName

	err := r.UserStore.Set(UserID, userdata)
	if err != nil {
		r.logger.Errorf(err.Error())
	}
	return err
}

func (r *Router) getAvatar(msg *config.Message) string {
	userdata := new(UserData)
	found, err := r.UserStore.Get(msg.UserID, userdata)
	if err != nil {
		r.logger.Error(err)
	}

	if found && userdata.Avatar != "" {
		return userdata.Avatar
	}

	return msg.Avatar
}

func (r *Router) setAvatar(UserID string, newAvatar string) error {
	userdata := new(UserData)
	r.UserStore.Get(UserID, userdata)

	userdata.Avatar = newAvatar

	err := r.UserStore.Set(UserID, userdata)
	if err != nil {
		r.logger.Errorf(err.Error())
	}
	return err
}

func getOptStr(status OptOutStatus) string {
	switch status {
	case OptIn:
		return "Opt In"
	case OptOut:
		return "Opt Out"
	case OptOutMediaOnly:
		return "Opt Out - Attachments Only"
	}

	return "Unknown"
}
