package dtls

import (
	"context"

	"github.com/pion/dtls/v2/pkg/crypto/prf"
	"github.com/pion/dtls/v2/pkg/protocol"
	"github.com/pion/dtls/v2/pkg/protocol/alert"
	"github.com/pion/dtls/v2/pkg/protocol/handshake"
	"github.com/pion/dtls/v2/pkg/protocol/recordlayer"
)

func flight5bParse(ctx context.Context, c flightConn, state *State, cache *handshakeCache, cfg *handshakeConfig) (flightVal, *alert.Alert, error) {
	_, msgs, ok := cache.fullPullMap(state.handshakeRecvSequence-1,
		handshakeCachePullRule{handshake.TypeFinished, cfg.initialEpoch + 1, false, false},
	)
	if !ok {
		// No valid message received. Keep reading
		return 0, nil, nil
	}

	if _, ok = msgs[handshake.TypeFinished].(*handshake.MessageFinished); !ok {
		return 0, &alert.Alert{Level: alert.Fatal, Description: alert.InternalError}, nil
	}

	// Other party may re-transmit the last flight. Keep state to be flight5b.
	return flight5b, nil, nil
}

func flight5bGenerate(c flightConn, state *State, cache *handshakeCache, cfg *handshakeConfig) ([]*packet, *alert.Alert, error) { //nolint:gocognit
	var pkts []*packet

	pkts = append(pkts,
		&packet{
			record: &recordlayer.RecordLayer{
				Header: recordlayer.Header{
					Version: protocol.Version1_2,
				},
				Content: &protocol.ChangeCipherSpec{},
			},
		})

	if len(state.localVerifyData) == 0 {
		plainText := cache.pullAndMerge(
			handshakeCachePullRule{handshake.TypeClientHello, cfg.initialEpoch, true, false},
			handshakeCachePullRule{handshake.TypeServerHello, cfg.initialEpoch, false, false},
			handshakeCachePullRule{handshake.TypeFinished, cfg.initialEpoch + 1, false, false},
		)

		var err error
		state.localVerifyData, err = prf.VerifyDataClient(state.masterSecret, plainText, state.cipherSuite.HashFunc())
		if err != nil {
			return nil, &alert.Alert{Level: alert.Fatal, Description: alert.InternalError}, err
		}
	}

	pkts = append(pkts,
		&packet{
			record: &recordlayer.RecordLayer{
				Header: recordlayer.Header{
					Version: protocol.Version1_2,
					Epoch:   1,
				},
				Content: &handshake.Handshake{
					Message: &handshake.MessageFinished{
						VerifyData: state.localVerifyData,
					},
				},
			},
			shouldEncrypt:            true,
			resetLocalSequenceNumber: true,
		})

	return pkts, nil, nil
}
