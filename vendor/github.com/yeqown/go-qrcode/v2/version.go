package qrcode

import (
	"errors"
	"log"
	"strconv"

	// "github.com/skip2/go-qrcode/bitset"
	"github.com/yeqown/reedsolomon/binary"
)

// ecLevel error correction level
type ecLevel int

const (
	// ErrorCorrectionLow :Level L: 7% error recovery.
	ErrorCorrectionLow ecLevel = iota + 1

	// ErrorCorrectionMedium :Level M: 15% error recovery. Good default choice.
	ErrorCorrectionMedium

	// ErrorCorrectionQuart :Level Q: 25% error recovery.
	ErrorCorrectionQuart

	// ErrorCorrectionHighest :Level H: 30% error recovery.
	ErrorCorrectionHighest

	formatInfoBitsNum = 15 // format info bits num
	verInfoBitsNum    = 18 // version info length bits num
)

var (
	errInvalidErrorCorrectionLevel = errors.New("invalid error correction level")
	errAnalyzeVersionFailed        = errors.New("could not match version! " +
		"check your content length is in limitation of encode mode and error correction level")
	errMissMatchedVersion    = errors.New("could not match version")
	errMissMatchedEncodeType = errors.New("could not match the encode type")
	// versions                 []version
	// Each QR Code contains a 15-bit Format Information qrbool.  The 15 bits
	// consist of 5 data bits concatenated with 10 error correction bits.
	//
	// The 5 data bits consist of:
	// - 2 bits for the error correction level (L=01, M=00, G=11, H=10).
	// - 3 bits for the data mask pattern identifier.
	//
	// formatBitSequence is a mapping from the 5 data bits to the completed 15-bit
	// Format Information qrbool.
	//
	// For example, a QR Code using error correction level L, and data mask
	// pattern identifier 001:
	//
	// 01 | 001 = 01001 = 0x9
	// formatBitSequence[0x9].qrCode = 0x72f3 = 111001011110011
	formatBitSequence = []struct {
		regular uint32
		micro   uint32
	}{
		{0x5412, 0x4445}, {0x5125, 0x4172}, {0x5e7c, 0x4e2b}, {0x5b4b, 0x4b1c},
		{0x45f9, 0x55ae}, {0x40ce, 0x5099}, {0x4f97, 0x5fc0}, {0x4aa0, 0x5af7},
		{0x77c4, 0x6793}, {0x72f3, 0x62a4}, {0x7daa, 0x6dfd}, {0x789d, 0x68ca},
		{0x662f, 0x7678}, {0x6318, 0x734f}, {0x6c41, 0x7c16}, {0x6976, 0x7921},
		{0x1689, 0x06de}, {0x13be, 0x03e9}, {0x1ce7, 0x0cb0}, {0x19d0, 0x0987},
		{0x0762, 0x1735}, {0x0255, 0x1202}, {0x0d0c, 0x1d5b}, {0x083b, 0x186c},
		{0x355f, 0x2508}, {0x3068, 0x203f}, {0x3f31, 0x2f66}, {0x3a06, 0x2a51},
		{0x24b4, 0x34e3}, {0x2183, 0x31d4}, {0x2eda, 0x3e8d}, {0x2bed, 0x3bba},
	}

	// QR Codes version 7 and higher contain an 18-bit version Information qrbool,
	// consisting of a 6 data bits and 12 error correction bits.
	//
	// versionBitSequence is a mapping from QR Code version to the completed
	// 18-bit version Information qrbool.
	//
	// For example, a QR code of version 7:
	// versionBitSequence[0x7] = 0x07c94 = 000111110010010100
	versionBitSequence = []uint32{
		0x00000, 0x00000, 0x00000, 0x00000, 0x00000, 0x00000, 0x00000, 0x07c94,
		0x085bc, 0x09a99, 0x0a4d3, 0x0bbf6, 0x0c762, 0x0d847, 0x0e60d, 0x0f928,
		0x10b78, 0x1145d, 0x12a17, 0x13532, 0x149a6, 0x15683, 0x168c9, 0x177ec,
		0x18ec4, 0x191e1, 0x1afab, 0x1b08e, 0x1cc1a, 0x1d33f, 0x1ed75, 0x1f250,
		0x209d5, 0x216f0, 0x228ba, 0x2379f, 0x24b0b, 0x2542e, 0x26a64, 0x27541, 0x28c69,
	}
)

// capacity struct includes data type max capacity
type capacity struct {
	Numeric      int `json:"n"` // num capacity
	AlphaNumeric int `json:"a"` // char capacity
	Byte         int `json:"b"` // byte capacity (utf-8 also)
	JP           int `json:"j"` // Japanese capacity
}

// group contains fields to generate ECBlocks
// and append _defaultPadding bit
type group struct {
	// NumBlocks num of blocks
	NumBlocks int `json:"nbs"`

	// NumDataCodewords Number of data codewords.
	NumDataCodewords int `json:"ndcs"`

	// ECBlockwordsPerBlock ...
	ECBlockwordsPerBlock int `json:"ecbs_pb"`
}

// version ...
type version struct {
	// version code 1-40
	Ver int `json:"ver"`

	// ECLevel error correction 0, 1, 2, 3
	ECLevel ecLevel `json:"eclv"`

	// Cap includes each type's max capacity (specified by `Ver` and `ecLevel`)
	// ref to: https://www.thonky.com/qr-code-tutorial/character-capacities
	Cap capacity `json:"cap"`

	// RemainderBits remainder bits need to append finally
	RemainderBits int `json:"rembits"`

	// groups info to generate
	// ref to: https://www.thonky.com/qr-code-tutorial/error-correction-table
	// numGroup = len(Groups)
	Groups []group `json:"groups"`
}

// Dimension ...
func (v version) Dimension() int {
	return v.Ver*4 + 17
}

// NumTotalCodewords total data codewords
func (v version) NumTotalCodewords() int {
	var total int
	for _, g := range v.Groups {
		total = total + (g.NumBlocks * g.NumDataCodewords)
	}
	return total
}

// NumGroups ... need group num. ref to version config file
func (v version) NumGroups() int {
	return len(v.Groups)
}

// TotalNumBlocks ... total data blocks num, ref to version config file
func (v version) TotalNumBlocks() int {
	var total int
	for _, g := range v.Groups {
		total = total + g.NumBlocks
	}
	return total
}

// VerInfo version info bitset
func (v version) verInfo() *binary.Binary {
	if v.Ver < 7 {
		return nil
	}

	result := binary.New()
	result.AppendUint32(versionBitSequence[v.Ver], verInfoBitsNum)

	return result
}

// formatInfo returns the 15-bit Format Information qrbool for a QR
// code.
func (v version) formatInfo(maskPattern int) *binary.Binary {
	formatID := 0

	switch v.ECLevel {
	case ErrorCorrectionLow:
		formatID = 0x08 // 0b01000
	case ErrorCorrectionMedium:
		formatID = 0x00 // 0b00000
	case ErrorCorrectionQuart:
		formatID = 0x18 // 0b11000
	case ErrorCorrectionHighest:
		formatID = 0x10 // 0b10000
	default:
		log.Panicf("Invalid level %d", v.ECLevel)
	}

	if maskPattern < 0 || maskPattern > 7 {
		log.Panicf("Invalid maskPattern %d", maskPattern)
	}

	formatID |= maskPattern & 0x7
	result := binary.New()
	result.AppendUint32(formatBitSequence[formatID].regular, formatInfoBitsNum)
	return result
}

var emptyVersion = version{Ver: -1}

// binarySearchVersion speed up searching target version in versions.
// low, high to set the left and right bound of the search range (min:0 to max:159).
// compare represents the function to compare the target version with the cursor version.
// negative means lower direction, positive means higher direction, zero mean hit.
func binarySearchVersion(low, high int, compare func(*version) int) (hit version, found bool) {
	// left low and high in a valid range
	if low > high || low > _VERSIONS_ITEM_COUNT || high < 0 {
		return emptyVersion, false
	}

	if low < 0 {
		low = 0
	}
	if high >= _VERSIONS_ITEM_COUNT {
		high = len(versions) - 1
	}

	for low <= high {
		mid := (low + high) / 2
		r := compare(&versions[mid])
		if r == 0 {
			hit = versions[mid]
			found = true
			break
		}

		if r > 0 {
			// move toward higher direction
			low = mid + 1
		} else {
			// move toward lower direction
			high = mid
		}
	}

	return hit, found
}

// defaultBinaryCompare built-in compare function for binary search.
func defaultBinaryCompare(ver int, ec ecLevel) func(cursor *version) int {
	return func(cursor *version) int {
		switch r := ver - cursor.Ver; r {
		case 0:
		default:
			// v is bigger return positive; otherwise return negative.
			return r
		}

		return int(ec - cursor.ECLevel)
	}
}

// loadVersion get version config by specified version indicator and error correction level.
// we can speed up this process, by shrink the range to search.
func loadVersion(lv int, ec ecLevel) version {
	// each version only has 4 items in versions array,
	// and them are ordered[ASC] already.
	high := lv*4 - 1
	low := (lv - 1) * 4

	for i := low; i <= high; i++ {
		if versions[i].ECLevel == ec {
			return versions[i]
		}
	}
	panic(errMissMatchedVersion)
}

// analyzeVersion the raw text, and then decide which version should be chosen
// according to the text length , error correction level and encode mode to choose the
// closest capacity of version.
//
// check out http://muyuchengfeng.xyz/%E4%BA%8C%E7%BB%B4%E7%A0%81-%E5%AD%97%E7%AC%A6%E5%AE%B9%E9%87%8F%E8%A1%A8/
// for more details.
func analyzeVersion(raw []byte, ec ecLevel, mode encMode) (*version, error) {
	step := 0
	switch ec {
	case ErrorCorrectionLow:
		step = 0
	case ErrorCorrectionMedium:
		step = 1
	case ErrorCorrectionQuart:
		step = 2
	case ErrorCorrectionHighest:
		step = 3
	default:
		return nil, errInvalidErrorCorrectionLevel
	}

	want, mark := len(raw), 0
	for ; step < 160; step += 4 {

		switch mode {
		case EncModeNumeric:
			mark = versions[step].Cap.Numeric
		case EncModeAlphanumeric:
			mark = versions[step].Cap.AlphaNumeric
		case EncModeByte:
			mark = versions[step].Cap.Byte
		case EncModeJP:
			mark = versions[step].Cap.JP
		default:
			return nil, errMissMatchedEncodeType
		}

		if mark >= want {
			return &versions[step], nil
		}
	}
	debugLogf("mismatched version, version's length: %d, ec: %v", len(versions), ec)

	return nil, errAnalyzeVersionFailed
}

var (
	// https://www.thonky.com/qr-code-tutorial/alignment-pattern-locations
	// DONE(@yeqown): add more version
	alignPatternLocation = map[int][]int{
		2:  {6, 18},
		3:  {6, 22},
		4:  {6, 26},
		5:  {6, 30},
		6:  {6, 34},
		7:  {6, 22, 38},
		8:  {6, 24, 42},
		9:  {6, 26, 46},
		10: {6, 28, 50},
		11: {6, 30, 54},
		12: {6, 32, 58},
		13: {6, 34, 62},
		14: {6, 26, 46, 66},
		15: {6, 26, 48, 70},
		16: {6, 26, 50, 74},
		17: {6, 30, 54, 78},
		18: {6, 30, 56, 82},
		19: {6, 30, 58, 86},
		20: {6, 34, 62, 90},
		21: {6, 28, 50, 72, 94},
		22: {6, 26, 50, 74, 98},
		23: {6, 30, 54, 78, 102},
		24: {6, 28, 54, 80, 106},
		25: {6, 32, 58, 84, 110},
		26: {6, 30, 58, 86, 114},
		27: {6, 34, 62, 90, 118},
		28: {6, 26, 50, 74, 98, 122},
		29: {6, 30, 54, 78, 102, 126},
		30: {6, 26, 52, 78, 104, 130},
		31: {6, 30, 56, 82, 108, 134},
		32: {6, 34, 60, 86, 112, 138},
		33: {6, 30, 58, 86, 114, 142},
		34: {6, 34, 62, 90, 118, 146},
		35: {6, 30, 54, 78, 102, 126, 150},
		36: {6, 24, 50, 76, 102, 128, 154},
		37: {6, 28, 54, 80, 106, 132, 158},
		38: {6, 32, 58, 84, 110, 136, 162},
		39: {6, 26, 54, 82, 110, 138, 166},
		40: {6, 30, 58, 86, 114, 142, 170},
	}

	alignPatternCache = map[int][]loc{}
)

// loc point position(x,y)
type loc struct {
	X int // for width
	Y int // for height
}

// loadAlignmentPatternLoc ...
func loadAlignmentPatternLoc(ver int) (locs []loc) {
	if ver < 2 {
		return
	}
	var ok bool
	if locs, ok = alignPatternCache[ver]; ok {
		return
	}

	dimension := ver*4 + 17
	positions, ok := alignPatternLocation[ver]
	if !ok {
		panic("could not found align at version: " + strconv.Itoa(ver))
	}

	for _, pos1 := range positions {
		for _, pos2 := range positions {
			if !valid(pos1, pos2, dimension) {
				continue
			}
			locs = append(locs, loc{X: pos1, Y: pos2})
		}
	}
	alignPatternCache[ver] = locs
	return
}

// x, y center position x,y so
func valid(x, y, dimension int) bool {
	// valid left-top
	if (x-2) < 7 && (y-2) < 7 {
		return false
	}
	// valid right-top
	if (x+2) > dimension-7 && (y-2) < 7 {
		return false
	}
	// valid left-bottom
	if (x-2) < 7 && (y+2) > dimension-7 {
		return false
	}

	return true
}
