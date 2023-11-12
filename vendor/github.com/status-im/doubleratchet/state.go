package doubleratchet

// TODO: During each DH ratchet step a new ratchet key pair and sending chain are generated.
// As the sending chain is not needed right away, these steps could be deferred until the party
// is about to send a new message.

import (
	"fmt"
)

// The double ratchet state.
type State struct {
	Crypto Crypto

	// DH Ratchet public key (the remote key).
	DHr Key

	// DH Ratchet key pair (the self ratchet key).
	DHs DHPair

	// Symmetric ratchet root chain.
	RootCh kdfRootChain

	// Symmetric ratchet sending and receiving chains.
	SendCh, RecvCh kdfChain

	// Number of messages in previous sending chain.
	PN uint32

	// Dictionary of skipped-over message keys, indexed by ratchet public key or header key
	// and message number.
	MkSkipped KeysStorage

	// The maximum number of message keys that can be skipped in a single chain.
	// WithMaxSkip should be set high enough to tolerate routine lost or delayed messages,
	// but low enough that a malicious sender can't trigger excessive recipient computation.
	MaxSkip uint

	// Receiving header key and next header key. Only used for header encryption.
	HKr, NHKr Key

	// Sending header key and next header key. Only used for header encryption.
	HKs, NHKs Key

	// How long we keep messages keys, counted in number of messages received,
	// for example if MaxKeep is 5 we only keep the last 5 messages keys, deleting everything n - 5.
	MaxKeep uint

	// Max number of message keys per session, older keys will be deleted in FIFO fashion
	MaxMessageKeysPerSession int

	// The number of the current ratchet step.
	Step uint

	// KeysCount the number of keys generated for decrypting
	KeysCount uint
}

func DefaultState(sharedKey Key) State {
	c := DefaultCrypto{}

	return State{
		DHs:    dhPair{},
		Crypto: c,
		RootCh: kdfRootChain{CK: sharedKey, Crypto: c},
		// Populate CKs and CKr with sharedKey so that both parties could send and receive
		// messages from the very beginning.
		SendCh:                   kdfChain{CK: sharedKey, Crypto: c},
		RecvCh:                   kdfChain{CK: sharedKey, Crypto: c},
		MkSkipped:                &KeysStorageInMemory{},
		MaxSkip:                  1000,
		MaxMessageKeysPerSession: 2000,
		MaxKeep:                  2000,
		KeysCount:                0,
	}
}

func (s *State) applyOptions(opts []option) error {
	for i := range opts {
		if err := opts[i](s); err != nil {
			return fmt.Errorf("failed to apply option: %s", err)
		}
	}
	return nil
}

func newState(sharedKey Key, opts ...option) (State, error) {
	if sharedKey == nil {
		return State{}, fmt.Errorf("sharedKey mustn't be empty")
	}

	s := DefaultState(sharedKey)
	if err := s.applyOptions(opts); err != nil {
		return State{}, err
	}

	return s, nil
}

// dhRatchet performs a single ratchet step.
func (s *State) dhRatchet(m MessageHeader) error {
	s.PN = s.SendCh.N
	s.DHr = m.DH
	s.HKs = s.NHKs
	s.HKr = s.NHKr

	recvSecret, err := s.Crypto.DH(s.DHs, s.DHr)
	if err != nil {
		return fmt.Errorf("failed to generate dh recieve ratchet secret: %s", err)
	}
	s.RecvCh, s.NHKr = s.RootCh.step(recvSecret)

	s.DHs, err = s.Crypto.GenerateDH()
	if err != nil {
		return fmt.Errorf("failed to generate dh pair: %s", err)
	}

	sendSecret, err := s.Crypto.DH(s.DHs, s.DHr)
	if err != nil {
		return fmt.Errorf("failed to generate dh send ratchet secret: %s", err)
	}
	s.SendCh, s.NHKs = s.RootCh.step(sendSecret)

	return nil
}

type skippedKey struct {
	key Key
	nr  uint
	mk  Key
	seq uint
}

// skipMessageKeys skips message keys in the current receiving chain.
func (s *State) skipMessageKeys(key Key, until uint) ([]skippedKey, error) {
	if until < uint(s.RecvCh.N) {
		return nil, fmt.Errorf("bad until: probably an out-of-order message that was deleted")
	}

	if uint(s.RecvCh.N)+s.MaxSkip < until {
		return nil, fmt.Errorf("too many messages")
	}

	skipped := []skippedKey{}
	for uint(s.RecvCh.N) < until {
		mk := s.RecvCh.step()
		skipped = append(skipped, skippedKey{
			key: key,
			nr:  uint(s.RecvCh.N - 1),
			mk:  mk,
			seq: s.KeysCount,
		})
		// Increment key count
		s.KeysCount++

	}
	return skipped, nil
}

func (s *State) applyChanges(sc State, sessionID []byte, skipped []skippedKey) error {
	*s = sc
	for _, skipped := range skipped {
		if err := s.MkSkipped.Put(sessionID, skipped.key, skipped.nr, skipped.mk, skipped.seq); err != nil {
			return err
		}
	}

	if err := s.MkSkipped.TruncateMks(sessionID, s.MaxMessageKeysPerSession); err != nil {
		return err
	}
	if s.KeysCount >= s.MaxKeep {
		if err := s.MkSkipped.DeleteOldMks(sessionID, s.KeysCount-s.MaxKeep); err != nil {
			return err
		}
	}

	return nil
}
