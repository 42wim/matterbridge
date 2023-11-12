package statecontrol

// TODO refactor into the pairing package once the backend dependencies have been removed.

import (
	"fmt"
	"sync"
)

var (
	ErrProcessStateManagerAlreadyPairing = fmt.Errorf("cannot start new LocalPairing session, already pairing")
	ErrProcessStateManagerAlreadyPaired  = func(sessionName string) error {
		return fmt.Errorf("given connection string already successfully used '%s'", sessionName)
	}
)

// ProcessStateManager represents a g
type ProcessStateManager struct {
	pairing     bool
	pairingLock sync.Mutex

	// sessions represents a map[string]bool:
	// where string is a ConnectionParams string and bool is the transfer success state of that connection string
	sessions sync.Map
}

// IsPairing returns if the ProcessStateManager is currently in pairing mode
func (psm *ProcessStateManager) IsPairing() bool {
	psm.pairingLock.Lock()
	defer psm.pairingLock.Unlock()
	return psm.pairing
}

// SetPairing sets the ProcessStateManager pairing state
func (psm *ProcessStateManager) SetPairing(state bool) {
	psm.pairingLock.Lock()
	defer psm.pairingLock.Unlock()
	psm.pairing = state
}

// RegisterSession stores a sessionName with the default false value.
// In practice a sessionName will be a ConnectionParams string provided by the server mode device.
// The boolean value represents whether the ConnectionParams string session resulted in a successful transfer.
func (psm *ProcessStateManager) RegisterSession(sessionName string) {
	psm.sessions.Store(sessionName, false)
}

// CompleteSession updates a transfer session with a given transfer success state only if the session is registered.
func (psm *ProcessStateManager) CompleteSession(sessionName string) {
	r, c := psm.GetSession(sessionName)
	if r && !c {
		psm.sessions.Store(sessionName, true)
	}
}

// GetSession returns two booleans for a given sessionName.
// These represent if the sessionName has been registered and if it has resulted in a successful transfer
func (psm *ProcessStateManager) GetSession(sessionName string) (bool, bool) {
	completed, registered := psm.sessions.Load(sessionName)
	if !registered {
		return registered, false
	}
	return registered, completed.(bool)
}

// StartPairing along with StopPairing are the core functions of the ProcessStateManager
// This function takes a sessionName, which in practice is a ConnectionParams string, and attempts to init pairing state management.
// The function will return an error if the ProcessStateManager is already currently pairing or if the sessionName was previously successful.
func (psm *ProcessStateManager) StartPairing(sessionName string) error {
	if psm.IsPairing() {
		return ErrProcessStateManagerAlreadyPairing
	}

	registered, completed := psm.GetSession(sessionName)
	if completed {
		return ErrProcessStateManagerAlreadyPaired(sessionName)
	}
	if !registered {
		psm.RegisterSession(sessionName)
	}

	psm.SetPairing(true)
	return nil
}

// StopPairing takes a sessionName and an error, if the error is nil the sessionName will be recorded as successful
// the pairing state of the ProcessStateManager is set to false.
func (psm *ProcessStateManager) StopPairing(sessionName string, err error) {
	if err == nil {
		psm.CompleteSession(sessionName)
	}
	psm.SetPairing(false)
}
