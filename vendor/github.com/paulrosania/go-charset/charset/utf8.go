package charset

import (
	"unicode/utf8"
)

func init() {
	registerClass("utf8", toUTF8, toUTF8)
}

type translateToUTF8 struct {
	scratch []byte
}

var errorBytes = []byte(string(utf8.RuneError))

const errorRuneLen = len(string(utf8.RuneError))

func (p *translateToUTF8) Translate(data []byte, eof bool) (int, []byte, error) {
	p.scratch = ensureCap(p.scratch, (len(data))*errorRuneLen)
	buf := p.scratch[:0]
	for i := 0; i < len(data); {
		// fast path for ASCII
		if b := data[i]; b < utf8.RuneSelf {
			buf = append(buf, b)
			i++
			continue
		}
		_, size := utf8.DecodeRune(data[i:])
		if size == 1 {
			if !eof && !utf8.FullRune(data) {
				// When DecodeRune has converted only a single
				// byte, we know there must be some kind of error
				// because we know the byte's not ASCII.
				// If we aren't at EOF, and it's an incomplete
				// rune encoding, then we return to process
				// the final bytes in a subsequent call.
				return i, buf, nil
			}
			buf = append(buf, errorBytes...)
		} else {
			buf = append(buf, data[i:i+size]...)
		}
		i += size
	}
	return len(data), buf, nil
}

func toUTF8(arg string) (Translator, error) {
	return new(translateToUTF8), nil
}
