// +build windows

package tcp

import "errors"

// ErrTimeout indicates I/O timeout
var ErrTimeout = &timeoutError{}

type timeoutError struct{}

type ErrConnect struct {
	error
}

var ErrCheckerAlreadyStarted = errors.New("Checker was already started")
