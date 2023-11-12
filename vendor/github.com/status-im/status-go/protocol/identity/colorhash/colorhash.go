package colorhash

import (
	"math/big"

	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/protocol/identity"
)

const (
	colorHashSegmentMaxLen = 5
	colorHashColorsCount   = 32
)

var colorHashAlphabet [][]int

func GenerateFor(pubkey string) (hash multiaccounts.ColorHash, err error) {
	if len(colorHashAlphabet) == 0 {
		colorHashAlphabet = makeColorHashAlphabet(colorHashSegmentMaxLen, colorHashColorsCount)
	}

	compressedKey, err := identity.ToCompressedKey(pubkey)
	if err != nil {
		return nil, err
	}

	slices, err := identity.Slices(compressedKey)
	if err != nil {
		return nil, err
	}

	return toColorHash(new(big.Int).SetBytes(slices[2]), &colorHashAlphabet, colorHashColorsCount), nil
}

// [[1 0] [1 1] [1 2] ... [units, colors-1]]
// [3 12] => 3 units length, 12 color index
func makeColorHashAlphabet(units, colors int) (res [][]int) {
	res = make([][]int, units*colors)
	idx := 0
	for i := 0; i < units; i++ {
		for j := 0; j < colors; j++ {
			res[idx] = make([]int, 2)
			res[idx][0] = i + 1
			res[idx][1] = j
			idx++
		}
	}
	return
}

func toColorHash(value *big.Int, alphabet *[][]int, colorsCount int) (hash multiaccounts.ColorHash) {
	alphabetLen := len(*alphabet)
	indexes := identity.ToBigBase(value, uint64(alphabetLen))
	hash = make(multiaccounts.ColorHash, len(indexes))
	for i, v := range indexes {
		hash[i] = [2]int{}
		hash[i][0] = (*alphabet)[v][0]
		hash[i][1] = (*alphabet)[v][1]
	}

	// colors can't repeat themselves
	// this makes color hash not fully collision resistant
	prevColorIdx := hash[0][1]
	hashLen := len(hash)
	for i := 1; i < hashLen; i++ {
		colorIdx := hash[i][1]
		if colorIdx == prevColorIdx {
			hash[i][1] = (colorIdx + 1) % colorsCount
		}
		prevColorIdx = hash[i][1]
	}

	return
}
