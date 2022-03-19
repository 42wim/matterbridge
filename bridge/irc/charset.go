package birc

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

var encoders = map[string]encoding.Encoding{
	"utf-8":       unicode.UTF8,
	"iso-2022-jp": japanese.ISO2022JP,
	"big5":        traditionalchinese.Big5,
	"gbk":         simplifiedchinese.GBK,
	"euc-kr":      korean.EUCKR,
	"gb2312":      simplifiedchinese.HZGB2312,
	"shift-jis":   japanese.ShiftJIS,
	"euc-jp":      japanese.EUCJP,
	"gb18030":     simplifiedchinese.GB18030,
}

func toUTF8(from string, input string) string {
	enc, ok := encoders[from]
	if !ok {
		return input
	}

	res, _ := enc.NewDecoder().String(input)
	return res
}
