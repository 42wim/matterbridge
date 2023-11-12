package protocol

import "github.com/status-im/status-go/protocol/requests"

func (m *Messenger) UpsertSwitcherCard(request *requests.UpsertSwitcherCard) error {
	if err := request.Validate(); err != nil {
		return err
	}

	switcherCard := SwitcherCard{
		CardID:   request.CardID,
		Type:     request.Type,
		Clock:    request.Clock,
		ScreenID: request.ScreenID,
	}

	return m.persistence.UpsertSwitcherCard(switcherCard)
}

func (m *Messenger) DeleteSwitcherCard(cardID string) error {
	return m.persistence.DeleteSwitcherCard(cardID)
}

func (m *Messenger) SwitcherCards() ([]SwitcherCard, error) {
	return m.persistence.SwitcherCards()
}
