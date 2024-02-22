package nack

import "errors"

// ErrInvalidSize is returned by newReceiveLog/newSendBuffer, when an incorrect buffer size is supplied.
var ErrInvalidSize = errors.New("invalid buffer size")

var errPacketReleased = errors.New("could not retain packet, already released")
