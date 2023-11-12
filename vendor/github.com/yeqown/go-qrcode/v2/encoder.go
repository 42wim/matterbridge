// Package qrcode ...
// encoder.go working for data encoding
package qrcode

import (
	"fmt"
	"log"

	"github.com/yeqown/reedsolomon/binary"
)

// encMode ...
type encMode uint

const (
	// a qrbool of EncModeAuto will trigger a detection of the letter set from the input data,
	EncModeAuto = 0
	// EncModeNone mode ...
	EncModeNone encMode = 1 << iota
	// EncModeNumeric mode ...
	EncModeNumeric
	// EncModeAlphanumeric mode ...
	EncModeAlphanumeric
	// EncModeByte mode ...
	EncModeByte
	// EncModeJP mode ...
	EncModeJP
)

var (
	paddingByte1, _ = binary.NewFromBinaryString("11101100")
	paddingByte2, _ = binary.NewFromBinaryString("00010001")
)

// getEncModeName ...
func getEncModeName(mode encMode) string {
	switch mode {
	case EncModeNone:
		return "none"
	case EncModeNumeric:
		return "numeric"
	case EncModeAlphanumeric:
		return "alphanumeric"
	case EncModeByte:
		return "byte"
	case EncModeJP:
		return "japan"
	default:
		return "unknown"
	}
}

// getEncodeModeIndicator ...
func getEncodeModeIndicator(mode encMode) *binary.Binary {
	switch mode {
	case EncModeNumeric:
		return binary.New(false, false, false, true)
	case EncModeAlphanumeric:
		return binary.New(false, false, true, false)
	case EncModeByte:
		return binary.New(false, true, false, false)
	case EncModeJP:
		return binary.New(true, false, false, false)
	default:
		panic("no indicator")
	}
}

// encoder ... data to bit stream ...
type encoder struct {
	// self init
	dst  *binary.Binary
	data []byte // raw input data

	// initial params
	mode encMode // encode mode
	ecLv ecLevel // error correction level

	// self load
	version version // QR version ref
}

func newEncoder(m encMode, ec ecLevel, v version) *encoder {
	return &encoder{
		dst:     nil,
		data:    nil,
		mode:    m,
		ecLv:    ec,
		version: v,
	}
}

// Encode ...
// 1. encode raw data into bitset
// 2. append _defaultPadding data
//
func (e *encoder) Encode(byts []byte) (*binary.Binary, error) {
	e.dst = binary.New()
	e.data = byts

	// append mode indicator symbol
	indicator := getEncodeModeIndicator(e.mode)
	e.dst.Append(indicator)
	// append chars length counter bits symbol
	e.dst.AppendUint32(uint32(len(byts)), e.charCountBits())

	// encode data with specified mode
	switch e.mode {
	case EncModeNumeric:
		e.encodeNumeric()
	case EncModeAlphanumeric:
		e.encodeAlphanumeric()
	case EncModeByte:
		e.encodeByte()
	case EncModeJP:
		panic("this has not been finished")
	}

	// fill and _defaultPadding bits
	e.breakUpInto8bit()

	return e.dst, nil
}

// 0001b mode indicator
func (e *encoder) encodeNumeric() {
	if e.dst == nil {
		log.Println("e.dst is nil")
		return
	}
	for i := 0; i < len(e.data); i += 3 {
		charsRemaining := len(e.data) - i

		var value uint32
		bitsUsed := 1

		for j := 0; j < charsRemaining && j < 3; j++ {
			value *= 10
			value += uint32(e.data[i+j] - 0x30)
			bitsUsed += 3
		}
		e.dst.AppendUint32(value, bitsUsed)
	}
}

// 0010b mode indicator
func (e *encoder) encodeAlphanumeric() {
	if e.dst == nil {
		log.Println("e.dst is nil")
		return
	}
	for i := 0; i < len(e.data); i += 2 {
		charsRemaining := len(e.data) - i

		var value uint32
		for j := 0; j < charsRemaining && j < 2; j++ {
			value *= 45
			value += encodeAlphanumericCharacter(e.data[i+j])
		}

		bitsUsed := 6
		if charsRemaining > 1 {
			bitsUsed = 11
		}

		e.dst.AppendUint32(value, bitsUsed)
	}
}

// 0100b mode indicator
func (e *encoder) encodeByte() {
	if e.dst == nil {
		log.Println("e.dst is nil")
		return
	}
	for _, b := range e.data {
		_ = e.dst.AppendByte(b, 8)
	}
}

// Break Up into 8-bit Codewords and Add Pad Bytes if Necessary
func (e *encoder) breakUpInto8bit() {
	// fill ending code (max 4bit)
	// depends on max capacity of current version and EC level
	maxCap := e.version.NumTotalCodewords() * 8
	if less := maxCap - e.dst.Len(); less < 0 {
		err := fmt.Errorf(
			"wrong version(%d) cap(%d bits) and could not contain all bits: %d bits",
			e.version.Ver, maxCap, e.dst.Len(),
		)
		panic(err)
	} else if less < 4 {
		e.dst.AppendNumBools(less, false)
	} else {
		e.dst.AppendNumBools(4, false)
	}

	// append `0` to be 8 times bits length
	if mod := e.dst.Len() % 8; mod != 0 {
		e.dst.AppendNumBools(8-mod, false)
	}

	// _defaultPadding bytes
	// _defaultPadding byte 11101100 00010001
	if n := maxCap - e.dst.Len(); n > 0 {
		debugLogf("maxCap: %d, len: %d, less: %d", maxCap, e.dst.Len(), n)
		for i := 1; i <= (n / 8); i++ {
			if i%2 == 1 {
				e.dst.Append(paddingByte1)
			} else {
				e.dst.Append(paddingByte2)
			}
		}
	}
}

// 字符计数指示符位长字典
var charCountMap = map[string]int{
	"9_numeric":       10,
	"9_alphanumeric":  9,
	"9_byte":          8,
	"9_japan":         8,
	"26_numeric":      12,
	"26_alphanumeric": 11,
	"26_byte":         16,
	"26_japan":        10,
	"40_numeric":      14,
	"40_alphanumeric": 13,
	"40_byte":         16,
	"40_japan":        12,
}

// charCountBits
func (e *encoder) charCountBits() int {
	var lv int
	if v := e.version.Ver; v <= 9 {
		lv = 9
	} else if v <= 26 {
		lv = 26
	} else {
		lv = 40
	}
	pos := fmt.Sprintf("%d_%s", lv, getEncModeName(e.mode))
	return charCountMap[pos]
}

// v must be a QR Code defined alphanumeric character: 0-9, A-Z, SP, $%*+-./ or
// :. The characters are mapped to values in the range 0-44 respectively.
func encodeAlphanumericCharacter(v byte) uint32 {
	c := uint32(v)

	switch {
	case c >= '0' && c <= '9':
		// 0-9 encoded as 0-9.
		return c - '0'
	case c >= 'A' && c <= 'Z':
		// A-Z encoded as 10-35.
		return c - 'A' + 10
	case c == ' ':
		return 36
	case c == '$':
		return 37
	case c == '%':
		return 38
	case c == '*':
		return 39
	case c == '+':
		return 40
	case c == '-':
		return 41
	case c == '.':
		return 42
	case c == '/':
		return 43
	case c == ':':
		return 44
	default:
		log.Panicf("encodeAlphanumericCharacter() with non alphanumeric char %c", v)
	}

	return 0
}

// analyzeEncFunc returns true is current byte matched in current mode,
// otherwise means you should use a bigger character set to check.
type analyzeEncFunc func(byte) bool

// analyzeEncodeModeFromRaw try to detect letter set of input data,
// so that encoder can determine which mode should be use.
// reference: https://en.wikipedia.org/wiki/QR_code
//
// case1: only numbers, use EncModeNumeric.
// case2: could not use EncModeNumeric, but you can find all of them in character mapping, use EncModeAlphanumeric.
// case3: could not use EncModeAlphanumeric, but you can find all of them in ISO-8859-1 character set, use EncModeByte.
// case4: could not use EncModeByte, use EncModeJP, no more choice.
//
func analyzeEncodeModeFromRaw(raw []byte) encMode {
	analyzeFnMapping := map[encMode]analyzeEncFunc{
		EncModeNumeric:      analyzeNum,
		EncModeAlphanumeric: analyzeAlphaNum,
		EncModeByte:         nil,
		EncModeJP:           nil,
	}

	var (
		f    analyzeEncFunc
		mode = EncModeNumeric
	)

	// loop to check each character in raw data,
	// from low mode to higher while current mode could bearing the input data.
	for _, byt := range raw {
	reAnalyze:
		if f = analyzeFnMapping[mode]; f == nil {
			break
		}

		// issue#28 @borislavone reports this bug.
		// FIXED(@yeqown): next encMode analyzeVersionAuto func did not check the previous byte,
		// add goto statement to reanalyze previous byte which can't be analyzed in last encMode.
		if !f(byt) {
			mode <<= 1
			goto reAnalyze
		}
	}

	return mode
}

// analyzeNum is byt in num encMode
func analyzeNum(byt byte) bool {
	return byt >= '0' && byt <= '9'
}

// analyzeAlphaNum is byt in alpha number
func analyzeAlphaNum(byt byte) bool {
	if (byt >= '0' && byt <= '9') || (byt >= 'A' && byt <= 'Z') {
		return true
	}
	switch byt {
	case ' ', '$', '%', '*', '+', '-', '.', '/', ':':
		return true
	}
	return false
}

//// analyzeByte is byt in bytes.
//func analyzeByte(byt byte) qrbool {
//	return false
//}
