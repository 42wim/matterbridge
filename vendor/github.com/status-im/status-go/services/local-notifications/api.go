package localnotifications

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
)

func NewAPI(s *Service) *API {
	return &API{s}
}

type API struct {
	s *Service
}

func (api *API) NotificationPreferences(ctx context.Context) ([]NotificationPreference, error) {
	return api.s.db.GetPreferences()
}

func (api *API) SwitchWalletNotifications(ctx context.Context, preference bool) error {
	log.Debug("Switch Transaction Notification")
	err := api.s.db.ChangeWalletPreference(preference)
	if err != nil {
		return err
	}

	api.s.WatchingEnabled = preference

	return nil
}
