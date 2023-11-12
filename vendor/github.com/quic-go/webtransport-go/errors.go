package webtransport

import (
	"errors"
	"fmt"

	"github.com/quic-go/quic-go"
)

// StreamErrorCode is an error code used for stream termination.
type StreamErrorCode uint8

// SessionErrorCode is an error code for session termination.
type SessionErrorCode uint32

const (
	firstErrorCode = 0x52e4a40fa8db
	lastErrorCode  = 0x52e4a40fa9e2
)

func webtransportCodeToHTTPCode(n StreamErrorCode) quic.StreamErrorCode {
	return quic.StreamErrorCode(firstErrorCode) + quic.StreamErrorCode(n) + quic.StreamErrorCode(n/0x1e)
}

func httpCodeToWebtransportCode(h quic.StreamErrorCode) (StreamErrorCode, error) {
	if h < firstErrorCode || h > lastErrorCode {
		return 0, errors.New("error code outside of expected range")
	}
	if (h-0x21)%0x1f == 0 {
		return 0, errors.New("invalid error code")
	}
	shifted := h - firstErrorCode
	return StreamErrorCode(shifted - shifted/0x1f), nil
}

func isWebTransportError(e error) bool {
	if e == nil {
		return false
	}
	var strErr *quic.StreamError
	if !errors.As(e, &strErr) {
		return false
	}
	if strErr.ErrorCode == sessionCloseErrorCode {
		return true
	}
	_, err := httpCodeToWebtransportCode(strErr.ErrorCode)
	return err == nil
}

// WebTransportBufferedStreamRejectedErrorCode is the error code of the
// H3_WEBTRANSPORT_BUFFERED_STREAM_REJECTED error.
const WebTransportBufferedStreamRejectedErrorCode quic.StreamErrorCode = 0x3994bd84

// StreamError is the error that is returned from stream operations (Read, Write) when the stream is canceled.
type StreamError struct {
	ErrorCode StreamErrorCode
}

func (e *StreamError) Is(target error) bool {
	_, ok := target.(*StreamError)
	return ok
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("stream canceled with error code %d", e.ErrorCode)
}

// ConnectionError is a WebTransport connection error.
type ConnectionError struct {
	Remote    bool
	ErrorCode SessionErrorCode
	Message   string
}

var _ error = &ConnectionError{}

func (e *ConnectionError) Error() string { return e.Message }
