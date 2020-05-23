// Package stripmd strips Markdown from text
package stripmd

import (
	"regexp"
)

var (
	listLeadersReg = regexp.MustCompile(`(?m)^([\s\t]*)([\*\-\+]|\d\.)\s+`)

	headerReg = regexp.MustCompile(`\n={2,}`)
	strikeReg = regexp.MustCompile(`~~`)
	codeReg   = regexp.MustCompile("`{3}" + `.*\n`)

	htmlReg         = regexp.MustCompile("<(.*?)>")
	emphReg         = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	emphReg2        = regexp.MustCompile(`\*([^*]+)\*`)
	emphReg3        = regexp.MustCompile(`__([^_]+)__`)
	emphReg4        = regexp.MustCompile(`_([^_]+)_`)
	setextHeaderReg = regexp.MustCompile(`^[=\-]{2,}\s*$`)
	footnotesReg    = regexp.MustCompile(`\[\^.+?\](\: .*?$)?`)
	footnotes2Reg   = regexp.MustCompile(`\s{0,2}\[.*?\]: .*?$`)
	imagesReg       = regexp.MustCompile(`\!\[(.*?)\]\s?[\[\(].*?[\]\)]`)
	linksReg        = regexp.MustCompile(`\[(.*?)\][\[\(].*?[\]\)]`)
	blockquoteReg   = regexp.MustCompile(`>\s*`)
	refLinkReg      = regexp.MustCompile(`^\s{1,2}\[(.*?)\]: (\S+)( ".*?")?\s*$`)
	atxHeaderReg    = regexp.MustCompile(`(?m)^\#{1,6}\s*([^#]+)\s*(\#{1,6})?$`)
	atxHeaderReg2   = regexp.MustCompile(`([\*_]{1,3})(\S.*?\S)?P1`)
	atxHeaderReg3   = regexp.MustCompile("(?m)(`{3,})" + `(.*?)?P1`)
	atxHeaderReg4   = regexp.MustCompile(`^-{3,}\s*$`)
	atxHeaderReg5   = regexp.MustCompile("`(.+?)`")
	atxHeaderReg6   = regexp.MustCompile(`\n{2,}`)
)

// Strip returns the given string sans any Markdown.
// Where necessary, elements are replaced with their best textual forms, so
// for example, hyperlinks are stripped of their URL and become only the link
// text, and images lose their URL and become only the alt text.
func Strip(s string) string {
	res := s
	res = listLeadersReg.ReplaceAllString(res, "$1")

	res = headerReg.ReplaceAllString(res, "\n")
	res = strikeReg.ReplaceAllString(res, "")
	res = codeReg.ReplaceAllString(res, "")

	res = emphReg.ReplaceAllString(res, "$1")
	res = emphReg2.ReplaceAllString(res, "$1")
	res = emphReg3.ReplaceAllString(res, "$1")
	res = emphReg4.ReplaceAllString(res, "$1")
	res = htmlReg.ReplaceAllString(res, "$1")
	res = setextHeaderReg.ReplaceAllString(res, "")
	res = footnotesReg.ReplaceAllString(res, "")
	res = footnotes2Reg.ReplaceAllString(res, "")
	res = imagesReg.ReplaceAllString(res, "$1")
	res = linksReg.ReplaceAllString(res, "$1")
	res = blockquoteReg.ReplaceAllString(res, "  ")
	res = refLinkReg.ReplaceAllString(res, "")
	res = atxHeaderReg.ReplaceAllString(res, "$1")
	res = atxHeaderReg2.ReplaceAllString(res, "$2")
	res = atxHeaderReg3.ReplaceAllString(res, "$2")
	res = atxHeaderReg4.ReplaceAllString(res, "")
	res = atxHeaderReg5.ReplaceAllString(res, "$1")
	res = atxHeaderReg6.ReplaceAllString(res, "\n\n")
	return res
}
