package doubleratchet

import "fmt"

// option is a constructor option.
type option func(*State) error

// WithMaxSkip specifies the maximum number of skipped message in a single chain.
// nolint: golint
func WithMaxSkip(n int) option {
	return func(s *State) error {
		if n < 0 {
			return fmt.Errorf("n must be non-negative")
		}
		s.MaxSkip = uint(n)
		return nil
	}
}

// WithMaxKeep specifies how long we keep message keys, counted in number of messages received
// nolint: golint
func WithMaxKeep(n int) option {
	return func(s *State) error {
		if n < 0 {
			return fmt.Errorf("n must be non-negative")
		}
		s.MaxKeep = uint(n)
		return nil
	}
}

// WithMaxMessageKeysPerSession specifies the maximum number of message keys per session
// nolint: golint
func WithMaxMessageKeysPerSession(n int) option {
	return func(s *State) error {
		if n < 0 {
			return fmt.Errorf("n must be non-negative")
		}
		s.MaxMessageKeysPerSession = n
		return nil
	}
}

// WithKeysStorage replaces the default keys storage with the specified.
// nolint: golint
func WithKeysStorage(ks KeysStorage) option {
	return func(s *State) error {
		if ks == nil {
			return fmt.Errorf("KeysStorage mustn't be nil")
		}
		s.MkSkipped = ks
		return nil
	}
}

// WithCrypto replaces the default cryptographic supplement with the specified.
// nolint: golint
func WithCrypto(c Crypto) option {
	return func(s *State) error {
		if c == nil {
			return fmt.Errorf("Crypto mustn't be nil")
		}
		s.Crypto = c
		s.RootCh.Crypto = c
		s.SendCh.Crypto = c
		s.RecvCh.Crypto = c
		return nil
	}
}
