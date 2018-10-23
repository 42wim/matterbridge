package gateway

import (
	"bytes"
	"strings"

	"github.com/russross/blackfriday"
)

type renderer struct {
	*blackfriday.Html
}

func doubleSpace(out *bytes.Buffer) {
	if out.Len() > 0 {
		out.WriteByte('\n')
	}
}

func escapeSingleChar(char byte) (string, bool) {
	if char == '"' {
		return "&quot;", true
	}
	if char == '&' {
		return "&amp;", true
	}
	if char == '<' {
		return "&lt;", true
	}
	if char == '>' {
		return "&gt;", true
	}
	return "", false
}

func attrEscape(out *bytes.Buffer, src []byte) {
	org := 0
	for i, ch := range src {
		if entity, ok := escapeSingleChar(ch); ok {
			if i > org {
				// copy all the normal characters since the last escape
				out.Write(src[org:i])
			}
			org = i + 1
			out.WriteString(entity)
		}
	}
	if org < len(src) {
		out.Write(src[org:])
	}
}

// Using <code> rather than <pre> keeps Google Translate from trying to process it.
// BUT it collapses code into one line for some reason, and <pre> preserves newlines.
// #TODO Investigating the <pre><code> combo might work.
func (*renderer) BlockCode(out *bytes.Buffer, text []byte, info string) {
	doubleSpace(out)

	endOfLang := strings.IndexAny(info, "\t ")
	if endOfLang < 0 {
		endOfLang = len(info)
	}
	lang := info[:endOfLang]
	if len(lang) == 0 || lang == "." {
		out.WriteString("<pre translate='no'>")
	}
	attrEscape(out, text)
	out.WriteString("</pre>\n")
}
