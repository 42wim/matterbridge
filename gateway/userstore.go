package gateway

import (
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
	OptOut OptOutStatus
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
