// The charset package implements translation between character sets.
// It uses Unicode as the intermediate representation.
// Because it can be large, the character set data is separated
// from the charset package. It can be embedded in the Go
// executable by importing the data package:
//
//	import _ "github.com/paulrosania/go-charset/data"
//
// It can also made available in a data directory (by setting CharsetDir).
package charset

import (
	"io"
	"strings"
	"unicode/utf8"
)

// Charset holds information about a given character set.
type Charset struct {
	Name    string   // Canonical name of character set.
	Aliases []string // Known aliases.
	Desc    string   // Description.
	NoFrom  bool     // Not possible to translate from this charset.
	NoTo    bool     // Not possible to translate to this charset.
}

// Translator represents a character set converter.
// The Translate method translates the given data,
// and returns the number of bytes of data consumed,
// a slice containing the converted data (which may be
// overwritten on the next call to Translate), and any
// conversion error. If eof is true, the data represents
// the final bytes of the input.
type Translator interface {
	Translate(data []byte, eof bool) (n int, cdata []byte, err error)
}

// A Factory can be used to make character set translators.
type Factory interface {
	// TranslatorFrom creates a translator that will translate from the named character
	// set to UTF-8.
	TranslatorFrom(name string) (Translator, error) // Create a Translator from this character set to.

	// TranslatorTo creates a translator that will translate from UTF-8 to the named character set.
	TranslatorTo(name string) (Translator, error) // Create a Translator To this character set.

	// Names returns all the character set names accessibile through the factory.
	Names() []string

	// Info returns information on the named character set. It returns nil if the
	// factory doesn't recognise the given name.
	Info(name string) *Charset
}

var factories = []Factory{localFactory{}}

// Register registers a new Factory which will be consulted when NewReader
// or NewWriter needs a character set translator for a given name.
func Register(factory Factory) {
	factories = append(factories, factory)
}

// NewReader returns a new Reader that translates from the named
// character set to UTF-8 as it reads r.
func NewReader(charset string, r io.Reader) (io.Reader, error) {
	tr, err := TranslatorFrom(charset)
	if err != nil {
		return nil, err
	}
	return NewTranslatingReader(r, tr), nil
}

// NewWriter returns a new WriteCloser writing to w.  It converts writes
// of UTF-8 text into writes on w of text in the named character set.
// The Close is necessary to flush any remaining partially translated
// characters to the output.
func NewWriter(charset string, w io.Writer) (io.WriteCloser, error) {
	tr, err := TranslatorTo(charset)
	if err != nil {
		return nil, err
	}
	return NewTranslatingWriter(w, tr), nil
}

// Info returns information about a character set, or nil
// if the character set is not found.
func Info(name string) *Charset {
	for _, f := range factories {
		if info := f.Info(name); info != nil {
			return info
		}
	}
	return nil
}

// Names returns the canonical names of all supported character sets, in alphabetical order.
func Names() []string {
	// TODO eliminate duplicates
	var names []string
	for _, f := range factories {
		names = append(names, f.Names()...)
	}
	return names
}

// TranslatorFrom returns a translator that will translate from
// the named character set to UTF-8.
func TranslatorFrom(charset string) (Translator, error) {
	var err error
	var tr Translator
	for _, f := range factories {
		tr, err = f.TranslatorFrom(charset)
		if err == nil {
			break
		}
	}
	if tr == nil {
		return nil, err
	}
	return tr, nil
}

// TranslatorTo returns a translator that will translate from UTF-8
// to the named character set.
func TranslatorTo(charset string) (Translator, error) {
	var err error
	var tr Translator
	for _, f := range factories {
		tr, err = f.TranslatorTo(charset)
		if err == nil {
			break
		}
	}
	if tr == nil {
		return nil, err
	}
	return tr, nil
}

func normalizedChar(c rune) rune {
	switch {
	case c >= 'A' && c <= 'Z':
		c = c - 'A' + 'a'
	case c == '_':
		c = '-'
	}
	return c
}

// NormalisedName returns s with all Roman capitals
// mapped to lower case, and '_' mapped to '-'
func NormalizedName(s string) string {
	return strings.Map(normalizedChar, s)
}

type translatingWriter struct {
	w   io.Writer
	tr  Translator
	buf []byte // unconsumed data from writer.
}

// NewTranslatingWriter returns a new WriteCloser writing to w.
// It passes the written bytes through the given Translator.
func NewTranslatingWriter(w io.Writer, tr Translator) io.WriteCloser {
	return &translatingWriter{w: w, tr: tr}
}

func (w *translatingWriter) Write(data []byte) (rn int, rerr error) {
	wdata := data
	if len(w.buf) > 0 {
		w.buf = append(w.buf, data...)
		wdata = w.buf
	}
	n, cdata, err := w.tr.Translate(wdata, false)
	if err != nil {
		// TODO
	}
	if n > 0 {
		_, err = w.w.Write(cdata)
		if err != nil {
			return 0, err
		}
	}
	w.buf = w.buf[:0]
	if n < len(wdata) {
		w.buf = append(w.buf, wdata[n:]...)
	}
	return len(data), nil
}

func (p *translatingWriter) Close() error {
	for {
		n, data, err := p.tr.Translate(p.buf, true)
		p.buf = p.buf[n:]
		if err != nil {
			// TODO
		}
		// If the Translator produces no data
		// at EOF, then assume that it never will.
		if len(data) == 0 {
			break
		}
		n, err = p.w.Write(data)
		if err != nil {
			return err
		}
		if n < len(data) {
			return io.ErrShortWrite
		}
		if len(p.buf) == 0 {
			break
		}
	}
	return nil
}

type translatingReader struct {
	r     io.Reader
	tr    Translator
	cdata []byte // unconsumed data from converter.
	rdata []byte // unconverted data from reader.
	err   error  // final error from reader.
}

// NewTranslatingReader returns a new Reader that
// translates data using the given Translator as it reads r.
func NewTranslatingReader(r io.Reader, tr Translator) io.Reader {
	return &translatingReader{r: r, tr: tr}
}

func (r *translatingReader) Read(buf []byte) (int, error) {
	for {
		if len(r.cdata) > 0 {
			n := copy(buf, r.cdata)
			r.cdata = r.cdata[n:]
			return n, nil
		}
		if r.err == nil {
			r.rdata = ensureCap(r.rdata, len(r.rdata)+len(buf))
			n, err := r.r.Read(r.rdata[len(r.rdata):cap(r.rdata)])
			// Guard against non-compliant Readers.
			if n == 0 && err == nil {
				err = io.EOF
			}
			r.rdata = r.rdata[0 : len(r.rdata)+n]
			r.err = err
		} else if len(r.rdata) == 0 {
			break
		}
		nc, cdata, cvterr := r.tr.Translate(r.rdata, r.err != nil)
		if cvterr != nil {
			// TODO
		}
		r.cdata = cdata

		// Ensure that we consume all bytes at eof
		// if the converter refuses them.
		if nc == 0 && r.err != nil {
			nc = len(r.rdata)
		}

		// Copy unconsumed data to the start of the rdata buffer.
		r.rdata = r.rdata[0:copy(r.rdata, r.rdata[nc:])]
	}
	return 0, r.err
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
