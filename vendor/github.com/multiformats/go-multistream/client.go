package multistream

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

// ErrNotSupported is the error returned when the muxer doesn't support
// the protocols tried for the handshake.
type ErrNotSupported[T StringLike] struct {

	// Slice of protocols that were not supported by the muxer
	Protos []T
}

func (e ErrNotSupported[T]) Error() string {
	return fmt.Sprintf("protocols not supported: %v", e.Protos)
}

func (e ErrNotSupported[T]) Is(target error) bool {
	_, ok := target.(ErrNotSupported[T])
	return ok
}

// ErrNoProtocols is the error returned when the no protocols have been
// specified.
var ErrNoProtocols = errors.New("no protocols specified")

const (
	tieBreakerPrefix = "select:"
	initiator        = "initiator"
	responder        = "responder"
)

// SelectProtoOrFail performs the initial multistream handshake
// to inform the muxer of the protocol that will be used to communicate
// on this ReadWriteCloser. It returns an error if, for example,
// the muxer does not know how to handle this protocol.
func SelectProtoOrFail[T StringLike](proto T, rwc io.ReadWriteCloser) (err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			fmt.Fprintf(os.Stderr, "caught panic: %s\n%s\n", rerr, debug.Stack())
			err = fmt.Errorf("panic selecting protocol: %s", rerr)
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		var buf bytes.Buffer
		if err := delitmWriteAll(&buf, []byte(ProtocolID), []byte(proto)); err != nil {
			errCh <- err
			return
		}
		_, err := io.Copy(rwc, &buf)
		errCh <- err
	}()
	// We have to read *both* errors.
	err1 := readMultistreamHeader(rwc)
	err2 := readProto(proto, rwc)
	if werr := <-errCh; werr != nil {
		return werr
	}
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// SelectOneOf will perform handshakes with the protocols on the given slice
// until it finds one which is supported by the muxer.
func SelectOneOf[T StringLike](protos []T, rwc io.ReadWriteCloser) (proto T, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			fmt.Fprintf(os.Stderr, "caught panic: %s\n%s\n", rerr, debug.Stack())
			err = fmt.Errorf("panic selecting one of protocols: %s", rerr)
		}
	}()

	if len(protos) == 0 {
		return "", ErrNoProtocols
	}

	// Use SelectProtoOrFail to pipeline the /multistream/1.0.0 handshake
	// with an attempt to negotiate the first protocol. If that fails, we
	// can continue negotiating the rest of the protocols normally.
	//
	// This saves us a round trip.
	switch err := SelectProtoOrFail(protos[0], rwc); err.(type) {
	case nil:
		return protos[0], nil
	case ErrNotSupported[T]: // try others
	default:
		return "", err
	}
	proto, err = selectProtosOrFail(protos[1:], rwc)
	if _, ok := err.(ErrNotSupported[T]); ok {
		return "", ErrNotSupported[T]{protos}
	}
	return proto, err
}

const simOpenProtocol = "/libp2p/simultaneous-connect"

// SelectWithSimopenOrFail performs protocol negotiation with the simultaneous open extension.
// The returned boolean indicator will be true if we should act as a server.
func SelectWithSimopenOrFail[T StringLike](protos []T, rwc io.ReadWriteCloser) (proto T, isServer bool, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			fmt.Fprintf(os.Stderr, "caught panic: %s\n%s\n", rerr, debug.Stack())
			err = fmt.Errorf("panic selecting protocol with simopen: %s", rerr)
		}
	}()

	if len(protos) == 0 {
		return "", false, ErrNoProtocols
	}

	werrCh := make(chan error, 1)
	go func() {
		var buf bytes.Buffer
		if err := delitmWriteAll(&buf, []byte(ProtocolID), []byte(simOpenProtocol), []byte(protos[0])); err != nil {
			werrCh <- err
			return
		}

		_, err := io.Copy(rwc, &buf)
		werrCh <- err
	}()

	if err := readMultistreamHeader(rwc); err != nil {
		return "", false, err
	}

	tok, err := ReadNextToken[T](rwc)
	if err != nil {
		return "", false, err
	}

	if err = <-werrCh; err != nil {
		return "", false, err
	}

	switch tok {
	case simOpenProtocol:
		// simultaneous open
		return simOpen(protos, rwc)
	case "na":
		// client open
		proto, err := clientOpen(protos, rwc)
		if err != nil {
			return "", false, err
		}
		return proto, false, nil
	default:
		return "", false, fmt.Errorf("unexpected response: %s", tok)
	}
}

func clientOpen[T StringLike](protos []T, rwc io.ReadWriteCloser) (T, error) {
	// check to see if we selected the pipelined protocol
	tok, err := ReadNextToken[T](rwc)
	if err != nil {
		return "", err
	}

	switch tok {
	case protos[0]:
		return tok, nil
	case "na":
		proto, err := selectProtosOrFail(protos[1:], rwc)
		if _, ok := err.(ErrNotSupported[T]); ok {
			return "", ErrNotSupported[T]{protos}
		}
		return proto, err
	default:
		return "", fmt.Errorf("unexpected response: %s", tok)
	}
}

func selectProtosOrFail[T StringLike](protos []T, rwc io.ReadWriteCloser) (T, error) {
	for _, p := range protos {
		err := trySelect(p, rwc)
		switch err := err.(type) {
		case nil:
			return p, nil
		case ErrNotSupported[T]:
		default:
			return "", err
		}
	}
	return "", ErrNotSupported[T]{protos}
}

func simOpen[T StringLike](protos []T, rwc io.ReadWriteCloser) (T, bool, error) {
	randBytes := make([]byte, 8)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", false, err
	}
	myNonce := binary.LittleEndian.Uint64(randBytes)

	werrCh := make(chan error, 1)
	go func() {
		myselect := []byte(tieBreakerPrefix + strconv.FormatUint(myNonce, 10))
		err := delimWriteBuffered(rwc, myselect)
		werrCh <- err
	}()

	// skip exactly one protocol
	// see https://github.com/multiformats/go-multistream/pull/42#discussion_r558757135
	_, err = ReadNextToken[T](rwc)
	if err != nil {
		return "", false, err
	}

	// read the tie breaker nonce
	tok, err := ReadNextToken[T](rwc)
	if err != nil {
		return "", false, err
	}
	if !strings.HasPrefix(string(tok), tieBreakerPrefix) {
		return "", false, errors.New("tie breaker nonce not sent with the correct prefix")
	}

	if err = <-werrCh; err != nil {
		return "", false, err
	}

	peerNonce, err := strconv.ParseUint(string(tok[len(tieBreakerPrefix):]), 10, 64)
	if err != nil {
		return "", false, err
	}

	var iamserver bool

	if peerNonce == myNonce {
		return "", false, errors.New("failed client selection; identical nonces")
	}
	iamserver = peerNonce > myNonce

	var proto T
	if iamserver {
		proto, err = simOpenSelectServer(protos, rwc)
	} else {
		proto, err = simOpenSelectClient(protos, rwc)
	}

	return proto, iamserver, err
}

func simOpenSelectServer[T StringLike](protos []T, rwc io.ReadWriteCloser) (T, error) {
	werrCh := make(chan error, 1)
	go func() {
		err := delimWriteBuffered(rwc, []byte(responder))
		werrCh <- err
	}()

	tok, err := ReadNextToken[T](rwc)
	if err != nil {
		return "", err
	}
	if tok != initiator {
		return "", fmt.Errorf("unexpected response: %s", tok)
	}
	if err = <-werrCh; err != nil {
		return "", err
	}
	for {
		tok, err = ReadNextToken[T](rwc)

		if err == io.EOF {
			return "", ErrNotSupported[T]{protos}
		}

		if err != nil {
			return "", err
		}

		for _, p := range protos {
			if tok == p {
				err = delimWriteBuffered(rwc, []byte(p))
				if err != nil {
					return "", err
				}

				return p, nil
			}
		}

		err = delimWriteBuffered(rwc, []byte("na"))
		if err != nil {
			return "", err
		}
	}

}

func simOpenSelectClient[T StringLike](protos []T, rwc io.ReadWriteCloser) (T, error) {
	werrCh := make(chan error, 1)
	go func() {
		err := delimWriteBuffered(rwc, []byte(initiator))
		werrCh <- err
	}()

	tok, err := ReadNextToken[T](rwc)
	if err != nil {
		return "", err
	}
	if tok != responder {
		return "", fmt.Errorf("unexpected response: %s", tok)
	}
	if err = <-werrCh; err != nil {
		return "", err
	}

	return selectProtosOrFail(protos, rwc)
}

func readMultistreamHeader(r io.Reader) error {
	tok, err := ReadNextToken[string](r)
	if err != nil {
		return err
	}

	if tok != ProtocolID {
		return errors.New("received mismatch in protocol id")
	}
	return nil
}

func trySelect[T StringLike](proto T, rwc io.ReadWriteCloser) error {
	err := delimWriteBuffered(rwc, []byte(proto))
	if err != nil {
		return err
	}
	return readProto(proto, rwc)
}

func readProto[T StringLike](proto T, r io.Reader) error {
	tok, err := ReadNextToken[T](r)
	if err != nil {
		return err
	}

	switch tok {
	case proto:
		return nil
	case "na":
		return ErrNotSupported[T]{[]T{proto}}
	default:
		return fmt.Errorf("unrecognized response: %s", tok)
	}
}
