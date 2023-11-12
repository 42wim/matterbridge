package pubsub

import (
	"fmt"

	pb "github.com/libp2p/go-libp2p-pubsub/pb"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// MessageSignaturePolicy describes if signatures are produced, expected, and/or verified.
type MessageSignaturePolicy uint8

// LaxSign and LaxNoSign are deprecated. In the future msgSigning and msgVerification can be unified.
const (
	// msgSigning is set when the locally produced messages must be signed
	msgSigning MessageSignaturePolicy = 1 << iota
	// msgVerification is set when external messages must be verfied
	msgVerification
)

const (
	// StrictSign produces signatures and expects and verifies incoming signatures
	StrictSign = msgSigning | msgVerification
	// StrictNoSign does not produce signatures and drops and penalises incoming messages that carry one
	StrictNoSign = msgVerification
	// LaxSign produces signatures and validates incoming signatures iff one is present
	// Deprecated: it is recommend to either strictly enable, or strictly disable, signatures.
	LaxSign = msgSigning
	// LaxNoSign does not produce signatures and validates incoming signatures iff one is present
	// Deprecated: it is recommend to either strictly enable, or strictly disable, signatures.
	LaxNoSign = 0
)

// mustVerify is true when a message signature must be verified.
// If signatures are not expected, verification checks if the signature is absent.
func (policy MessageSignaturePolicy) mustVerify() bool {
	return policy&msgVerification != 0
}

// mustSign is true when messages should be signed, and incoming messages are expected to have a signature.
func (policy MessageSignaturePolicy) mustSign() bool {
	return policy&msgSigning != 0
}

const SignPrefix = "libp2p-pubsub:"

func verifyMessageSignature(m *pb.Message) error {
	pubk, err := messagePubKey(m)
	if err != nil {
		return err
	}

	xm := *m
	xm.Signature = nil
	xm.Key = nil
	bytes, err := xm.Marshal()
	if err != nil {
		return err
	}

	bytes = withSignPrefix(bytes)

	valid, err := pubk.Verify(bytes, m.Signature)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func messagePubKey(m *pb.Message) (crypto.PubKey, error) {
	var pubk crypto.PubKey

	pid, err := peer.IDFromBytes(m.From)
	if err != nil {
		return nil, err
	}

	if m.Key == nil {
		// no attached key, it must be extractable from the source ID
		pubk, err = pid.ExtractPublicKey()
		if err != nil {
			return nil, fmt.Errorf("cannot extract signing key: %s", err.Error())
		}
		if pubk == nil {
			return nil, fmt.Errorf("cannot extract signing key")
		}
	} else {
		pubk, err = crypto.UnmarshalPublicKey(m.Key)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal signing key: %s", err.Error())
		}

		// verify that the source ID matches the attached key
		if !pid.MatchesPublicKey(pubk) {
			return nil, fmt.Errorf("bad signing key; source ID %s doesn't match key", pid)
		}
	}

	return pubk, nil
}

func signMessage(pid peer.ID, key crypto.PrivKey, m *pb.Message) error {
	bytes, err := m.Marshal()
	if err != nil {
		return err
	}

	bytes = withSignPrefix(bytes)

	sig, err := key.Sign(bytes)
	if err != nil {
		return err
	}

	m.Signature = sig

	pk, _ := pid.ExtractPublicKey()
	if pk == nil {
		pubk, err := crypto.MarshalPublicKey(key.GetPublic())
		if err != nil {
			return err
		}
		m.Key = pubk
	}

	return nil
}

func withSignPrefix(bytes []byte) []byte {
	return append([]byte(SignPrefix), bytes...)
}
