// +build darwin linux

package tcp

import (
	"errors"

	"golang.org/x/sys/unix"
)

// ErrTimeout indicates I/O timeout
var ErrTimeout = &timeoutError{}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "I/O timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

// ErrConnect is an error occurs while connecting to the host
// To get the detail of underlying error, lookup ErrorCode() in 'man 2 connect'
type ErrConnect struct {
	error
}

// newErrConnect returns a ErrConnect with given error code
func newErrConnect(errCode int) *ErrConnect {
	return &ErrConnect{unix.Errno(errCode)}
}

// ErrCheckerAlreadyStarted indicates there is another instance of CheckingLoop running.
var ErrCheckerAlreadyStarted = errors.New("Checker was already started")
