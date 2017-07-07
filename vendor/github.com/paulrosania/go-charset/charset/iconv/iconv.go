// The iconv package provides an interface to the GNU iconv character set
// conversion library (see http://www.gnu.org/software/libiconv/).
// It automatically registers all the character sets with the charset package,
// so it is usually used simply for the side effects of importing it.
// Example:
//   import (
//		"go-charset.googlecode.com/hg/charset"
//		_ "go-charset.googlecode.com/hg/charset/iconv"
//   )
package iconv

//#cgo darwin LDFLAGS: -liconv
//#include <stdlib.h>
//#include <iconv.h>
//#include <errno.h>
//iconv_t iconv_open_error = (iconv_t)-1;
//size_t iconv_error = (size_t)-1;
import "C"
import (
	"errors"
	"fmt"
	"github.com/paulrosania/go-charset/charset"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

type iconvTranslator struct {
	cd      C.iconv_t
	invalid rune
	scratch []byte
}

func canonicalChar(c rune) rune {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}

func canonicalName(s string) string {
	return strings.Map(canonicalChar, s)
}

func init() {
	charset.Register(iconvFactory{})
}

type iconvFactory struct {
}

func (iconvFactory) TranslatorFrom(name string) (charset.Translator, error) {
	return Translator("UTF-8", name, utf8.RuneError)
}

func (iconvFactory) TranslatorTo(name string) (charset.Translator, error) {
	// BUG This is wrong.  The target character set may not be ASCII
	// compatible.  There's no easy solution to this other than
	// removing the offending code point.
	return Translator(name, "UTF-8", '?')
}

// Translator returns a Translator that translates between
// the named character sets. When an invalid multibyte
// character is found, the bytes in invalid are substituted instead.
func Translator(toCharset, fromCharset string, invalid rune) (charset.Translator, error) {
	cto, cfrom := C.CString(toCharset), C.CString(fromCharset)
	cd, err := C.iconv_open(cto, cfrom)

	C.free(unsafe.Pointer(cfrom))
	C.free(unsafe.Pointer(cto))

	if cd == C.iconv_open_error {
		if err == syscall.EINVAL {
			return nil, errors.New("iconv: conversion not supported")
		}
		return nil, err
	}
	t := &iconvTranslator{cd: cd, invalid: invalid}
	runtime.SetFinalizer(t, func(*iconvTranslator) {
		C.iconv_close(cd)
	})
	return t, nil
}

func (iconvFactory) Names() []string {
	all := aliases()
	names := make([]string, 0, len(all))
	for name, aliases := range all {
		if aliases[0] == name {
			names = append(names, name)
		}
	}
	return names
}

func (iconvFactory) Info(name string) *charset.Charset {
	name = strings.ToLower(name)
	all := aliases()
	a, ok := all[name]
	if !ok {
		return nil
	}
	return &charset.Charset{
		Name:    name,
		Aliases: a,
	}
}

func (p *iconvTranslator) Translate(data []byte, eof bool) (rn int, rd []byte, rerr error) {
	n := 0
	p.scratch = p.scratch[:0]
	for len(data) > 0 {
		p.scratch = ensureCap(p.scratch, len(p.scratch)+len(data)*utf8.UTFMax)
		cData := (*C.char)(unsafe.Pointer(&data[:1][0]))
		nData := C.size_t(len(data))

		ns := len(p.scratch)
		cScratch := (*C.char)(unsafe.Pointer(&p.scratch[ns : ns+1][0]))
		nScratch := C.size_t(cap(p.scratch) - ns)
		r, err := C.iconv(p.cd, &cData, &nData, &cScratch, &nScratch)

		p.scratch = p.scratch[0 : cap(p.scratch)-int(nScratch)]
		n += len(data) - int(nData)
		data = data[len(data)-int(nData):]

		if r != C.iconv_error || err == nil {
			return n, p.scratch, nil
		}
		switch err := err.(syscall.Errno); err {
		case C.EILSEQ:
			// invalid multibyte sequence - skip one byte and continue
			p.scratch = appendRune(p.scratch, p.invalid)
			n++
			data = data[1:]
		case C.EINVAL:
			// incomplete multibyte sequence
			return n, p.scratch, nil
		case C.E2BIG:
			// output buffer not large enough; try again with larger buffer.
			p.scratch = ensureCap(p.scratch, cap(p.scratch)+utf8.UTFMax)
		default:
			panic(fmt.Sprintf("unexpected error code: %v", err))
		}
	}
	return n, p.scratch, nil
}

// ensureCap returns s with a capacity of at least n bytes.
// If cap(s) < n, then it returns a new copy of s with the
// required capacity.
func ensureCap(s []byte, n int) []byte {
	if n <= cap(s) {
		return s
	}
	// logic adapted from appendslice1 in runtime
	m := cap(s)
	if m == 0 {
		m = n
	} else {
		for {
			if m < 1024 {
				m += m
			} else {
				m += m / 4
			}
			if m >= n {
				break
			}
		}
	}
	t := make([]byte, len(s), m)
	copy(t, s)
	return t
}

func appendRune(buf []byte, r rune) []byte {
	n := len(buf)
	buf = ensureCap(buf, n+utf8.UTFMax)
	nu := utf8.EncodeRune(buf[n:n+utf8.UTFMax], r)
	return buf[0 : n+nu]
}
