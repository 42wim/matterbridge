package bls12381

import (
	"math/big"
)

func wnaf(e0 *big.Int, window uint) []int64 {
	e := new(big.Int).Set(e0)
	zero := big.NewInt(0)
	if e.Cmp(zero) == 0 {
		return []int64{}
	}
	max := int64(1 << window)
	midpoint := int64(1 << (window - 1))
	modulusMask := uint64(1<<window) - 1
	var out []int64
	for e.Cmp(zero) != 0 {
		var z int64
		if e.Bit(0)&1 == 1 {
			maskedBits := int64(e.Uint64() & modulusMask)
			if maskedBits > midpoint {
				z = maskedBits - max
				e.Add(e, new(big.Int).SetInt64(0-z))
			} else {
				z = maskedBits
				e.Sub(e, new(big.Int).SetInt64(z))
			}
		} else {
			z = 0
		}
		out = append(out, z)
		e.Rsh(e, 1)
	}
	return out
}
