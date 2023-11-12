package connection

const InvalidTimestamp = int64(-1)

type StateValue int

const (
	StateValueUnknown StateValue = iota
	StateValueConnected
	StateValueDisconnected
)

type State struct {
	Value         StateValue `json:"value"`
	LastCheckedAt int64      `json:"last_checked_at"`
	LastSuccessAt int64      `json:"last_success_at"`
}

func NewState() State {
	return State{
		Value:         StateValueUnknown,
		LastCheckedAt: InvalidTimestamp,
		LastSuccessAt: InvalidTimestamp,
	}
}
