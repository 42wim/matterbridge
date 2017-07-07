package charset

import (
	"fmt"
	"unicode/utf8"
)

func init() {
	registerClass("cp932", fromCP932, nil)
}

// encoding details
// (Traditional) Shift-JIS
//
// 00..1f	control characters
// 20		space
// 21..7f	JIS X 0201:1976/1997 roman (see notes)
// 80		undefined
// 81..9f	lead byte of JIS X 0208-1983 or JIS X 0202:1990/1997
// a0		undefined
// a1..df	JIS X 0201:1976/1997 katakana
// e0..ea	lead byte of JIS X 0208-1983 or JIS X 0202:1990/1997
// eb..ff	undefined
//
// CP932 (windows-31J)
//
// this encoding scheme extends Shift-JIS in the following way
//
// eb..ec	undefined (marked as lead bytes - see notes below)
// ed..ee	lead byte of NEC-selected IBM extended characters
// ef		undefined (marked as lead byte - see notes below)
// f0..f9	lead byte of User defined GAIJI (see note below)
// fa..fc	lead byte of IBM extended characters
// fd..ff	undefined
//
//
// Notes
//
// JISX 0201:1976/1997 roman
//	this is the same as ASCII but with 0x5c (ASCII code for '\')
//	representing the Yen currency symbol '¥' (U+00a5)
//	This mapping is contentious, some conversion packages implent it
//	others do not.
//	The mapping files from The Unicode Consortium show cp932 mapping
//	plain ascii in the range 00..7f whereas shift-jis maps 0x5c ('\') to the yen
//	symbol (¥) and 0x7e ('~') to overline (¯)
//
// CP932 double-byte character codes:
//
// eb-ec, ef, f0-f9:
// 	Marked as DBCS LEAD BYTEs in the unicode mapping data
//	obtained from:
//		https://www.unicode.org/Public/MAPPINGS/VENDORS/MICSFT/WINDOWS/CP932.TXT
//
// 	but there are no defined mappings for codes in this range.
// 	It is not clear whether or not an implementation should
// 	consume one or two bytes before emitting an error char.

const (
	kanaPages    = 1
	kanaPageSize = 63
	kanaChar0    = 0xa1

	cp932Pages    = 45  // 81..84, 87..9f, e0..ea, ed..ee, fa..fc
	cp932PageSize = 189 // 40..fc (including 7f)
	cp932Char0    = 0x40
)

type jisTables struct {
	page0   [256]rune
	dbcsoff [256]int
	cp932   []rune
}

type translateFromCP932 struct {
	tables  *jisTables
	scratch []byte
}

func (p *translateFromCP932) Translate(data []byte, eof bool) (int, []byte, error) {
	tables := p.tables
	p.scratch = p.scratch[:0]
	n := 0
	for i := 0; i < len(data); i++ {
		b := data[i]
		r := tables.page0[b]
		if r != -1 {
			p.scratch = appendRune(p.scratch, r)
			n++
			continue
		}
		// DBCS
		i++
		if i >= len(data) {
			break
		}
		pnum := tables.dbcsoff[b]
		ix := int(data[i]) - cp932Char0
		if pnum == -1 || ix < 0 || ix >= cp932PageSize {
			r = utf8.RuneError
		} else {
			r = tables.cp932[pnum*cp932PageSize+ix]
		}
		p.scratch = appendRune(p.scratch, r)
		n += 2
	}
	return n, p.scratch, nil
}

type cp932Key bool

func fromCP932(arg string) (Translator, error) {
	shiftJIS := arg == "shiftjis"
	tables, err := cache(cp932Key(shiftJIS), func() (interface{}, error) {
		tables := new(jisTables)
		kana, err := jisGetMap("jisx0201kana.dat", kanaPageSize, kanaPages)
		if err != nil {
			return nil, err
		}
		tables.cp932, err = jisGetMap("cp932.dat", cp932PageSize, cp932Pages)
		if err != nil {
			return nil, err
		}

		// jisx0201kana is mapped into 0xA1..0xDF
		for i := 0; i < kanaPageSize; i++ {
			tables.page0[i+kanaChar0] = kana[i]
		}

		// 00..7f same as ascii in cp932
		for i := rune(0); i < 0x7f; i++ {
			tables.page0[i] = i
		}

		if shiftJIS {
			// shift-jis uses JIS X 0201 for the ASCII range
			// this is the same as ASCII apart from
			// 0x5c ('\') maps to yen symbol (¥) and 0x7e ('~') maps to overline (¯)
			tables.page0['\\'] = '¥'
			tables.page0['~'] = '¯'
		}

		// pre-calculate DBCS page numbers to mapping file page numbers
		// and mark codes in page0 that are DBCS lead bytes
		pnum := 0
		for i := 0x81; i <= 0x84; i++ {
			tables.page0[i] = -1
			tables.dbcsoff[i] = pnum
			pnum++
		}
		for i := 0x87; i <= 0x9f; i++ {
			tables.page0[i] = -1
			tables.dbcsoff[i] = pnum
			pnum++
		}
		for i := 0xe0; i <= 0xea; i++ {
			tables.page0[i] = -1
			tables.dbcsoff[i] = pnum
			pnum++
		}
		if shiftJIS {
			return tables, nil
		}
		// add in cp932 extensions
		for i := 0xed; i <= 0xee; i++ {
			tables.page0[i] = -1
			tables.dbcsoff[i] = pnum
			pnum++
		}
		for i := 0xfa; i <= 0xfc; i++ {
			tables.page0[i] = -1
			tables.dbcsoff[i] = pnum
			pnum++
		}
		return tables, nil
	})

	if err != nil {
		return nil, err
	}

	return &translateFromCP932{tables: tables.(*jisTables)}, nil
}

func jisGetMap(name string, pgsize, npages int) ([]rune, error) {
	data, err := readFile(name)
	if err != nil {
		return nil, err
	}
	m := []rune(string(data))
	if len(m) != pgsize*npages {
		return nil, fmt.Errorf("%q: incorrect length data", name)
	}
	return m, nil
}
