package connection

import (
	"sync"
	"time"
)

type StateChangeCb func(State)

type Status struct {
	stateChangeCb StateChangeCb
	state         State
	stateLock     sync.RWMutex
}

func NewStatus() *Status {
	return &Status{
		state: NewState(),
	}
}

func (c *Status) SetStateChangeCb(stateChangeCb StateChangeCb) {
	c.stateChangeCb = stateChangeCb
}

func (c *Status) SetIsConnected(value bool) {
	now := time.Now().Unix()

	state := c.GetState()
	state.LastCheckedAt = now
	if value {
		state.LastSuccessAt = now
	}
	if value {
		state.Value = StateValueConnected
	} else {
		state.Value = StateValueDisconnected
	}

	c.SetState(state)
}

func (c *Status) ResetStateValue() {
	state := c.GetState()
	state.Value = StateValueUnknown
	c.SetState(state)
}

func (c *Status) GetState() State {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()

	return c.state
}

func (c *Status) SetState(state State) {
	c.stateLock.Lock()
	isStateChange := c.state.Value != state.Value
	c.state = state
	c.stateLock.Unlock()

	if isStateChange && c.stateChangeCb != nil {
		c.stateChangeCb(state)
	}
}

func (c *Status) GetStateValue() StateValue {
	return c.GetState().Value
}

func (c *Status) IsConnected() bool {
	return c.GetStateValue() == StateValueConnected
}
