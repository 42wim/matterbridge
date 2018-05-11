// Package ic convert text between CJK and UTF-8 in pure Go way
package ic

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

var (
	transformers = map[string]encoding.Encoding{
		"gbk":         simplifiedchinese.GBK,
		"cp936":       simplifiedchinese.GBK,
		"windows-936": simplifiedchinese.GBK,
		"gb18030":     simplifiedchinese.GB18030,
		"gb2312":      simplifiedchinese.HZGB2312,
		"big5":        traditionalchinese.Big5,
		"big-5":       traditionalchinese.Big5,
		"cp950":       traditionalchinese.Big5,
		"euc-kr":      korean.EUCKR,
		"euckr":       korean.EUCKR,
		"cp949":       korean.EUCKR,
		"euc-jp":      japanese.EUCJP,
		"eucjp":       japanese.EUCJP,
		"shift-jis":   japanese.ShiftJIS,
		"iso-2022-jp": japanese.ISO2022JP,
		"cp932":       japanese.ISO2022JP,
		"windows-31j": japanese.ISO2022JP,
	}
)

// ToUTF8 convert from CJK encoding to UTF-8
func ToUTF8(from string, s []byte) ([]byte, error) {
	var reader *transform.Reader

	transformer, ok := transformers[strings.ToLower(from)]
	if !ok {
		return s, errors.New("Unsupported encoding " + from)
	}
	reader = transform.NewReader(bytes.NewReader(s), transformer.NewDecoder())

	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// FromUTF8 convert from UTF-8 encoding to CJK encoding
func FromUTF8(to string, s []byte) ([]byte, error) {
	var reader *transform.Reader

	transformer, ok := transformers[strings.ToLower(to)]
	if !ok {
		return s, errors.New("Unsupported encoding " + to)
	}
	reader = transform.NewReader(bytes.NewReader(s), transformer.NewEncoder())

	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
