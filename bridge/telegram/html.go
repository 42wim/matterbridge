package btelegram

import (
	"bytes"
	"html"
	"io"

	"github.com/russross/blackfriday"
)

type customHTML struct {
	blackfriday.Renderer
}

func (options *customHTML) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()

	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("\n")
}

func (options *customHTML) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString("<pre>")

	out.WriteString(html.EscapeString(string(text)))
	out.WriteString("</pre>\n")
}

func (options *customHTML) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	options.Paragraph(out, text)
}

func (options *customHTML) HRule(out io.ByteWriter) {
	out.WriteByte('\n')
}

func (options *customHTML) BlockQuote(out *bytes.Buffer, text []byte) {
	out.WriteString("> ")
	out.Write(text)
	out.WriteByte('\n')
}

func (options *customHTML) List(out *bytes.Buffer, text func() bool, flags int) {
	options.Paragraph(out, text)
}

func (options *customHTML) ListItem(out *bytes.Buffer, text []byte, flags int) {
	out.WriteString("- ")
	out.Write(text)
	out.WriteByte('\n')
}

func makeHTML(input string) string {
	extensions := blackfriday.NoIntraEmphasis |
		blackfriday.FencedCode |
		blackfriday.Autolink |
		blackfriday.SpaceHeadings |
		blackfriday.HeadingIDs |
		blackfriday.BackslashLineBreak |
		blackfriday.DefinitionLists

	renderer := &customHTML{blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.UseXHTML | blackfriday.SkipImages,
	})}
	return string(blackfriday.Run([]byte(input), blackfriday.WithExtensions(extensions), blackfriday.WithRenderer(renderer)))
}
