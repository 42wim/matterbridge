package charset

import (
	"fmt"
	"unicode/utf8"
)

func init() {
	registerClass("big5", fromBig5, nil)
}

// Big5 consists of 89 fonts of 157 chars each
const (
	big5Max  = 13973
	big5Font = 157
	big5Data = "big5.dat"
)

type translateFromBig5 struct {
	font    int
	scratch []byte
	big5map []rune
}

func (p *translateFromBig5) Translate(data []byte, eof bool) (int, []byte, error) {
	p.scratch = p.scratch[:0]
	n := 0
	for len(data) > 0 {
		c := int(data[0])
		data = data[1:]
		n++
		if p.font == -1 {
			// idle state
			if c >= 0xa1 {
				p.font = c
				continue
			}
			if c == 26 {
				c = '\n'
			}
			continue
		}
		f := p.font
		p.font = -1
		r := utf8.RuneError
		switch {
		case c >= 64 && c <= 126:
			c -= 64
		case c >= 161 && c <= 254:
			c = c - 161 + 63
		default:
			// bad big5 char
			f = 255
		}
		if f <= 254 {
			f -= 161
			ix := f*big5Font + c
			if ix < len(p.big5map) {
				r = p.big5map[ix]
			}
			if r == -1 {
				r = utf8.RuneError
			}
		}
		p.scratch = appendRune(p.scratch, r)
	}
	return n, p.scratch, nil
}

type big5Key bool

func fromBig5(arg string) (Translator, error) {
	big5map, err := cache(big5Key(false), func() (interface{}, error) {
		data, err := readFile(big5Data)
		if err != nil {
			return nil, fmt.Errorf("charset: cannot open big5 data file: %v", err)
		}
		big5map := []rune(string(data))
		if len(big5map) != big5Max {
			return nil, fmt.Errorf("charset: corrupt big5 data")
		}
		return big5map, nil
	})
	if err != nil {
		return nil, err
	}
	return &translateFromBig5{big5map: big5map.([]rune), font: -1}, nil
}
