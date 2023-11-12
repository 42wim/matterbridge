package requests

import "errors"

var ErrUpsertSwitcherCardInvalidCardID = errors.New("upsert-switcher-card: invalid card id")

type UpsertSwitcherCard struct {
	CardID   string `json:"cardId,omitempty"`
	Type     int    `json:"type"`
	Clock    uint64 `json:"clock"`
	ScreenID string `json:"screenId"`
}

func (a *UpsertSwitcherCard) Validate() error {
	if len(a.CardID) == 0 {
		return ErrUpsertSwitcherCardInvalidCardID
	}

	return nil
}
