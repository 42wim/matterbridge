package emojihash

import (
	"bufio"
	"bytes"
	"errors"
	"math/big"
	"strings"

	"github.com/status-im/status-go/protocol/identity"
	"github.com/status-im/status-go/static"
)

const (
	emojiAlphabetLen = 2757 // 20bytes of data described by 14 emojis requires at least 2757 length alphabet
	emojiHashLen     = 14
)

var emojisAlphabet []string

func GenerateFor(pubkey string) ([]string, error) {
	if len(emojisAlphabet) == 0 {
		alphabet, err := loadAlphabet()
		if err != nil {
			return nil, err
		}
		emojisAlphabet = *alphabet
	}

	compressedKey, err := identity.ToCompressedKey(pubkey)
	if err != nil {
		return nil, err
	}

	slices, err := identity.Slices(compressedKey)
	if err != nil {
		return nil, err
	}

	return toEmojiHash(new(big.Int).SetBytes(slices[1]), emojiHashLen, &emojisAlphabet)
}

func loadAlphabet() (*[]string, error) {
	data, err := static.Asset("emojis.txt")
	if err != nil {
		return nil, err
	}

	alphabet := make([]string, 0, emojiAlphabetLen)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		alphabet = append(alphabet, strings.Replace(scanner.Text(), "\n", "", -1))
	}

	// current alphabet contains more emojis than needed, just in case some emojis needs to be removed
	// make sure only necessary part is loaded
	if len(alphabet) > emojiAlphabetLen {
		alphabet = alphabet[:emojiAlphabetLen]
	}

	return &alphabet, nil
}

func toEmojiHash(value *big.Int, hashLen int, alphabet *[]string) (hash []string, err error) {
	valueBitLen := value.BitLen()
	alphabetLen := new(big.Int).SetInt64(int64(len(*alphabet)))

	indexes := identity.ToBigBase(value, alphabetLen.Uint64())
	if hashLen == 0 {
		hashLen = len(indexes)
	} else if hashLen > len(indexes) {
		prependLen := hashLen - len(indexes)
		for i := 0; i < prependLen; i++ {
			indexes = append([](uint64){0}, indexes...)
		}
	}

	// alphabetLen^hashLen
	possibleCombinations := new(big.Int).Exp(alphabetLen, new(big.Int).SetInt64(int64(hashLen)), nil)

	// 2^valueBitLen
	requiredCombinations := new(big.Int).Exp(new(big.Int).SetInt64(2), new(big.Int).SetInt64(int64(valueBitLen)), nil)

	if possibleCombinations.Cmp(requiredCombinations) == -1 {
		return nil, errors.New("alphabet or hash length is too short to encode given value")
	}

	for _, v := range indexes {
		hash = append(hash, (*alphabet)[v])
	}

	return hash, nil
}
