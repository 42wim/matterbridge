package charset

import (
	"fmt"
	"unicode/utf8"
)

func init() {
	registerClass("cp", fromCodePage, toCodePage)
}

type translateFromCodePage struct {
	byte2rune *[256]rune
	scratch   []byte
}

type cpKeyFrom string
type cpKeyTo string

func (p *translateFromCodePage) Translate(data []byte, eof bool) (int, []byte, error) {
	p.scratch = ensureCap(p.scratch, len(data)*utf8.UTFMax)[:0]
	buf := p.scratch
	for _, x := range data {
		r := p.byte2rune[x]
		if r < utf8.RuneSelf {
			buf = append(buf, byte(r))
			continue
		}
		size := utf8.EncodeRune(buf[len(buf):cap(buf)], r)
		buf = buf[0 : len(buf)+size]
	}
	return len(data), buf, nil
}

type toCodePageInfo struct {
	rune2byte map[rune]byte
	// same gives the number of runes at start of code page that map exactly to
	// unicode.
	same rune
}

type translateToCodePage struct {
	toCodePageInfo
	scratch []byte
}

func (p *translateToCodePage) Translate(data []byte, eof bool) (int, []byte, error) {
	p.scratch = ensureCap(p.scratch, len(data))
	buf := p.scratch[:0]

	for i := 0; i < len(data); {
		r := rune(data[i])
		size := 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeRune(data[i:])
			if size == 1 && !eof && !utf8.FullRune(data[i:]) {
				return i, buf, nil
			}
		}

		var b byte
		if r < p.same {
			b = byte(r)
		} else {
			var ok bool
			b, ok = p.rune2byte[r]
			if !ok {
				b = '?'
			}
		}
		buf = append(buf, b)
		i += size
	}
	return len(data), buf, nil
}

func fromCodePage(arg string) (Translator, error) {
	runes, err := cache(cpKeyFrom(arg), func() (interface{}, error) {
		data, err := readFile(arg)
		if err != nil {
			return nil, err
		}
		runes := []rune(string(data))
		if len(runes) != 256 {
			return nil, fmt.Errorf("charset: %q has wrong rune count (%d)", arg, len(runes))
		}
		r := new([256]rune)
		copy(r[:], runes)
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return &translateFromCodePage{byte2rune: runes.(*[256]rune)}, nil
}

func toCodePage(arg string) (Translator, error) {
	m, err := cache(cpKeyTo(arg), func() (interface{}, error) {
		data, err := readFile(arg)
		if err != nil {
			return nil, err
		}

		info := toCodePageInfo{
			rune2byte: make(map[rune]byte),
			same:      256,
		}
		atStart := true
		i := rune(0)
		for _, r := range string(data) {
			if atStart {
				if r == i {
					i++
					continue
				}
				info.same = i
				atStart = false
			}
			info.rune2byte[r] = byte(i)
			i++
		}
		// TODO fix tables
		// fmt.Printf("%s, same = %d\n", arg, info.same)
		if i != 256 {
			return nil, fmt.Errorf("charset: %q has wrong rune count (%d)", arg, i)
		}
		return info, nil
	})
	if err != nil {
		return nil, err
	}
	return &translateToCodePage{toCodePageInfo: m.(toCodePageInfo)}, nil
}
