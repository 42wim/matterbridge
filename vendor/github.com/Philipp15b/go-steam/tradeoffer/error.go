package tradeoffer

import (
	"fmt"
)

// SteamError can be returned by Create, Accept, Decline and Cancel methods.
// It means we got response from steam, but it was in unknown format
// or request was declined.
type SteamError struct {
	msg string
}

func (e *SteamError) Error() string {
	return e.msg
}
func newSteamErrorf(format string, a ...interface{}) *SteamError {
	return &SteamError{fmt.Sprintf(format, a...)}
}
