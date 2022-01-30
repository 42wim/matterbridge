package signalerror

import "errors"

var (
	ErrNoSenderKeyStatesInRecord = errors.New("no sender key states in record")
	ErrNoSenderKeyStateForID     = errors.New("no sender key state for key ID")
)

var (
	ErrUntrustedIdentity = errors.New("untrusted identity")
	ErrNoSignedPreKey    = errors.New("no signed prekey found in bundle")
	ErrInvalidSignature  = errors.New("invalid signature on device key")
	ErrNoOneTimeKeyFound = errors.New("prekey store didn't return one-time key")
)

var (
	ErrNoValidSessions      = errors.New("no valid sessions")
	ErrUninitializedSession = errors.New("uninitialized session")
	ErrWrongMessageVersion  = errors.New("wrong message version")
	ErrTooFarIntoFuture     = errors.New("message index is over 2000 messages into the future")
	ErrOldCounter           = errors.New("received message with old counter")
	ErrNoSessionForUser     = errors.New("no session found for user")
)

var (
	ErrSenderKeyStateVerificationFailed = errors.New("sender key state failed verification with given public key")
	ErrNoSenderKeyForUser               = errors.New("no sender key")
)

var (
	ErrOldMessageVersion     = errors.New("too old message version")
	ErrUnknownMessageVersion = errors.New("unknown message version")
	ErrIncompleteMessage     = errors.New("incomplete message")
)

var ErrBadMAC = errors.New("mismatching MAC in signal message")
