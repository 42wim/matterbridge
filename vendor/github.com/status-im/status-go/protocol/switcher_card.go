package protocol

type SwitcherCard struct {
	CardID   string `json:"cardId,omitempty"`
	Type     int    `json:"type"`
	Clock    uint64 `json:"clock"`
	ScreenID string `json:"screenId"`
}
