package doubleratchet

// KDFer performs key derivation functions for chains.
type KDFer interface {
	// KdfRK returns a pair (32-byte root key, 32-byte chain key) as the output of applying
	// a KDF keyed by a 32-byte root key rk to a Diffie-Hellman output dhOut.
	KdfRK(rk, dhOut Key) (rootKey, chainKey, newHeaderKey Key)

	// KdfCK returns a pair (32-byte chain key, 32-byte message key) as the output of applying
	// a KDF keyed by a 32-byte chain key ck to some constant.
	KdfCK(ck Key) (chainKey, msgKey Key)
}

type kdfChain struct {
	Crypto KDFer

	// 32-byte chain key.
	CK Key

	// Messages count in the chain.
	N uint32
}

// step performs symmetric ratchet step and returns a new message key.
func (c *kdfChain) step() Key {
	var mk Key
	c.CK, mk = c.Crypto.KdfCK(c.CK)
	c.N++
	return mk
}

type kdfRootChain struct {
	Crypto KDFer

	// 32-byte kdfChain key.
	CK Key
}

// step performs symmetric ratchet step and returns a new chain and new header key.
func (c *kdfRootChain) step(kdfInput Key) (ch kdfChain, nhk Key) {
	ch = kdfChain{
		Crypto: c.Crypto,
	}
	c.CK, ch.CK, nhk = c.Crypto.KdfRK(c.CK, kdfInput)
	return ch, nhk
}
