package torrent

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/anacrolix/torrent/mse"
	pp "github.com/anacrolix/torrent/peer_protocol"
)

// Wraps a raw connection and provides the interface we want for using the
// connection in the message loop.
type deadlineReader struct {
	nc net.Conn
	r  io.Reader
}

func (r deadlineReader) Read(b []byte) (int, error) {
	// Keep-alives should be received every 2 mins. Give a bit of gracetime.
	err := r.nc.SetReadDeadline(time.Now().Add(150 * time.Second))
	if err != nil {
		return 0, fmt.Errorf("error setting read deadline: %s", err)
	}
	return r.r.Read(b)
}

// Handles stream encryption for inbound connections.
func handleEncryption(
	rw io.ReadWriter,
	skeys mse.SecretKeyIter,
	policy HeaderObfuscationPolicy,
	selector mse.CryptoSelector,
) (
	ret io.ReadWriter,
	headerEncrypted bool,
	cryptoMethod mse.CryptoMethod,
	err error,
) {
	// Tries to start an unencrypted stream.
	if !policy.RequirePreferred || !policy.Preferred {
		var protocol [len(pp.Protocol)]byte
		_, err = io.ReadFull(rw, protocol[:])
		if err != nil {
			return
		}
		// Put the protocol back into the stream.
		rw = struct {
			io.Reader
			io.Writer
		}{
			io.MultiReader(bytes.NewReader(protocol[:]), rw),
			rw,
		}
		if string(protocol[:]) == pp.Protocol {
			ret = rw
			return
		}
		if policy.RequirePreferred {
			// We are here because we require unencrypted connections.
			err = fmt.Errorf("unexpected protocol string %q and header obfuscation disabled", protocol)
			return
		}
	}
	headerEncrypted = true
	ret, cryptoMethod, err = mse.ReceiveHandshake(rw, skeys, selector)
	return
}

type PeerExtensionBits = pp.PeerExtensionBits
