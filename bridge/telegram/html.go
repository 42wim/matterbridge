package btelegram

import (
	"bytes"
	"github.com/russross/blackfriday"
	"html"
)

type customHtml struct {
	blackfriday.Renderer
}

func (options *customHtml) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()

	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("\n")
}

func (options *customHtml) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString("<pre>")

	out.WriteString(html.EscapeString(string(text)))
	out.WriteString("</pre>\n")
}

func (options *customHtml) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	options.Paragraph(out, text)
}

func (options *customHtml) HRule(out *bytes.Buffer) {
	out.WriteByte('\n')
}

func (options *customHtml) BlockQuote(out *bytes.Buffer, text []byte) {
	out.WriteString("> ")
	out.Write(text)
	out.WriteByte('\n')
}

func (options *customHtml) List(out *bytes.Buffer, text func() bool, flags int) {
	options.Paragraph(out, text)
}

func (options *customHtml) ListItem(out *bytes.Buffer, text []byte, flags int) {
	out.WriteString("- ")
	out.Write(text)
	out.WriteByte('\n')
}

func makeHTML(input string) string {
	return string(blackfriday.Markdown([]byte(input),
		&customHtml{blackfriday.HtmlRenderer(blackfriday.HTML_USE_XHTML|blackfriday.HTML_SKIP_IMAGES, "", "")},
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS|
			blackfriday.EXTENSION_FENCED_CODE|
			blackfriday.EXTENSION_AUTOLINK|
			blackfriday.EXTENSION_SPACE_HEADERS|
			blackfriday.EXTENSION_HEADER_IDS|
			blackfriday.EXTENSION_BACKSLASH_LINE_BREAK|
			blackfriday.EXTENSION_DEFINITION_LISTS))
}
