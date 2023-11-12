package api

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/status-im/status-go/eth-node/crypto"
)

// RunAsync runs the specified function asynchronously.
func RunAsync(f func() error) <-chan error {
	resp := make(chan error, 1)
	go func() {
		err := f()
		resp <- err
		close(resp)
	}()
	return resp
}

// HashMessage calculates the hash of a message to be safely signed by the keycard
// The hash is calulcated as
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func HashMessage(message string) ([]byte, error) {
	buf := bytes.NewBufferString("\x19Ethereum Signed Message:\n")
	if value, ok := decodeHexStrict(message); ok {
		if _, err := buf.WriteString(strconv.Itoa(len(value))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(value); err != nil {
			return nil, err
		}
	} else {
		if _, err := buf.WriteString(strconv.Itoa(len(message))); err != nil {
			return nil, err
		}
		if _, err := buf.WriteString(message); err != nil {
			return nil, err
		}
	}

	return crypto.Keccak256(buf.Bytes()), nil
}

func decodeHexStrict(s string) ([]byte, bool) {
	if !strings.HasPrefix(s, "0x") {
		return nil, false
	}

	value, err := hex.DecodeString(s[2:])
	if err != nil {
		return nil, false
	}

	return value, true
}
