package md

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/JohannesKaufmann/html-to-markdown/escape"
	"github.com/PuerkitoBio/goquery"
)

var multipleSpacesR = regexp.MustCompile(`  +`)

var commonmark = []Rule{
	{
		Filter: []string{"ul", "ol"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			parent := selec.Parent()

			// we have a nested list, were the ul/ol is inside a list item
			// -> based on work done by @requilence from @anytypeio
			if parent.Is("li") && parent.Children().Last().IsSelection(selec) {
				// add a line break prefix if the parent's text node doesn't have it.
				// that makes sure that every list item is on its on line
				lastContentTextNode := strings.TrimRight(parent.Nodes[0].FirstChild.Data, " \t")
				if !strings.HasSuffix(lastContentTextNode, "\n") {
					content = "\n" + content
				}

				// remove empty lines between lists
				trimmedSpaceContent := strings.TrimRight(content, " \t")
				if strings.HasSuffix(trimmedSpaceContent, "\n") {
					content = strings.TrimRightFunc(content, unicode.IsSpace)
				}
			} else {
				content = "\n\n" + content + "\n\n"
			}
			return &content
		},
	},
	{
		Filter: []string{"li"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			if strings.TrimSpace(content) == "" {
				return nil
			}

			parent := selec.Parent()
			index := selec.Index()

			var prefix string
			if parent.Is("ol") {
				prefix = strconv.Itoa(index+1) + ". "
			} else {
				prefix = opt.BulletListMarker + " "
			}
			// remove leading newlines
			content = leadingNewlinesR.ReplaceAllString(content, "")
			// replace trailing newlines with just a single one
			content = trailingNewlinesR.ReplaceAllString(content, "\n")
			// indent
			content = indentR.ReplaceAllString(content, "\n    ")

			return String(prefix + content + "\n")
		},
	},
	{
		Filter: []string{"#text"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			text := selec.Text()
			if trimmed := strings.TrimSpace(text); trimmed == "" {
				return String("")
			}
			text = tabR.ReplaceAllString(text, " ")

			// replace multiple spaces by one space: dont accidentally make
			// normal text be indented and thus be a code block.
			text = multipleSpacesR.ReplaceAllString(text, " ")

			text = escape.MarkdownCharacters(text)
			return &text
		},
	},
	{
		Filter: []string{"p", "div"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			parent := goquery.NodeName(selec.Parent())
			if IsInlineElement(parent) || parent == "li" {
				content = "\n" + content + "\n"
				return &content
			}

			// remove unnecessary spaces to have clean markdown
			content = TrimpLeadingSpaces(content)

			content = "\n\n" + content + "\n\n"
			return &content
		},
	},
	{
		Filter: []string{"h1", "h2", "h3", "h4", "h5", "h6"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			if strings.TrimSpace(content) == "" {
				return nil
			}

			content = strings.Replace(content, "\n", " ", -1)
			content = strings.Replace(content, "\r", " ", -1)
			content = strings.Replace(content, `#`, `\#`, -1)

			insideLink := selec.ParentsFiltered("a").Length() > 0
			if insideLink {
				text := opt.StrongDelimiter + content + opt.StrongDelimiter
				return &text
			}

			node := goquery.NodeName(selec)
			level, err := strconv.Atoi(node[1:])
			if err != nil {
				return nil
			}

			if opt.HeadingStyle == "setext" && level < 3 {
				line := "-"
				if level == 1 {
					line = "="
				}

				underline := strings.Repeat(line, len(content))
				return String("\n\n" + content + "\n" + underline + "\n\n")
			}

			prefix := strings.Repeat("#", level)
			content = strings.TrimSpace(content)
			text := "\n\n" + prefix + " " + content + "\n\n"
			return &text
		},
	},
	{
		Filter: []string{"strong", "b"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			// only use one bold tag if they are nested
			parent := selec.Parent()
			if parent.Is("strong") || parent.Is("b") {
				return &content
			}

			trimmed := strings.TrimSpace(content)
			if trimmed == "" {
				return &trimmed
			}
			trimmed = opt.StrongDelimiter + trimmed + opt.StrongDelimiter

			// always have a space to the side to recognize the delimiter
			trimmed = AddSpaceIfNessesary(selec, trimmed)

			return &trimmed
		},
	},
	{
		Filter: []string{"i", "em"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			// only use one italic tag if they are nested
			parent := selec.Parent()
			if parent.Is("i") || parent.Is("em") {
				return &content
			}

			trimmed := strings.TrimSpace(content)
			if trimmed == "" {
				return &trimmed
			}
			trimmed = opt.EmDelimiter + trimmed + opt.EmDelimiter

			// always have a space to the side to recognize the delimiter
			trimmed = AddSpaceIfNessesary(selec, trimmed)

			return &trimmed
		},
	},
	{
		Filter: []string{"img"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			src := selec.AttrOr("src", "")
			src = strings.TrimSpace(src)
			if src == "" {
				return String("")
			}

			src = opt.GetAbsoluteURL(selec, src, opt.domain)

			alt := selec.AttrOr("alt", "")
			alt = strings.Replace(alt, "\n", " ", -1)

			text := fmt.Sprintf("![%s](%s)", alt, src)
			return &text
		},
	},
	{
		Filter: []string{"a"},
		AdvancedReplacement: func(content string, selec *goquery.Selection, opt *Options) (AdvancedResult, bool) {
			// if there is no href, no link is used. So just return the content inside the link
			href, ok := selec.Attr("href")
			if !ok || strings.TrimSpace(href) == "" || strings.TrimSpace(href) == "#" {
				return AdvancedResult{
					Markdown: content,
				}, false
			}

			href = opt.GetAbsoluteURL(selec, href, opt.domain)

			// having multiline content inside a link is a bit tricky
			content = EscapeMultiLine(content)

			var title string
			if t, ok := selec.Attr("title"); ok {
				t = strings.Replace(t, "\n", " ", -1)
				title = fmt.Sprintf(` "%s"`, t)
			}

			// if there is no link content (for example because it contains an svg)
			// the 'title' or 'aria-label' attribute is used instead.
			if strings.TrimSpace(content) == "" {
				content = selec.AttrOr("title", selec.AttrOr("aria-label", ""))
			}

			// a link without text won't de displayed anyway
			if content == "" {
				return AdvancedResult{}, true
			}

			if opt.LinkStyle == "inlined" {
				md := fmt.Sprintf("[%s](%s%s)", content, href, title)
				md = AddSpaceIfNessesary(selec, md)

				return AdvancedResult{
					Markdown: md,
				}, false
			}

			var replacement string
			var reference string

			switch opt.LinkReferenceStyle {
			case "collapsed":

				replacement = "[" + content + "][]"
				reference = "[" + content + "]: " + href + title
			case "shortcut":
				replacement = "[" + content + "]"
				reference = "[" + content + "]: " + href + title

			default:
				id := selec.AttrOr("data-index", "")
				replacement = "[" + content + "][" + id + "]"
				reference = "[" + id + "]: " + href + title
			}

			replacement = AddSpaceIfNessesary(selec, replacement)
			return AdvancedResult{Markdown: replacement, Footer: reference}, false
		},
	},
	{
		Filter: []string{"code"},
		Replacement: func(_ string, selec *goquery.Selection, opt *Options) *string {
			code := selec.Text()

			fenceChar := '`'
			maxCount := calculateCodeFenceOccurrences(fenceChar, code)
			maxCount++

			fence := strings.Repeat(string(fenceChar), maxCount)

			// code block contains a backtick as first character
			if strings.HasPrefix(code, "`") {
				code = " " + code
			}
			// code block contains a backtick as last character
			if strings.HasSuffix(code, "`") {
				code = code + " "
			}

			// TODO: configure delimeter in options?
			text := fence + code + fence
			return &text
		},
	},
	{
		Filter: []string{"pre"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			codeElement := selec.Find("code")
			language := codeElement.AttrOr("class", "")
			language = strings.Replace(language, "language-", "", 1)

			code := codeElement.Text()
			if codeElement.Length() == 0 {
				code = selec.Text()
			}

			fenceChar, _ := utf8.DecodeRuneInString(opt.Fence)
			fence := CalculateCodeFence(fenceChar, code)

			text := "\n\n" + fence + language + "\n" +
				code +
				"\n" + fence + "\n\n"
			return &text
		},
	},
	{
		Filter: []string{"hr"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			text := "\n\n" + opt.HorizontalRule + "\n\n"
			return &text
		},
	},
	{
		Filter: []string{"br"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			return String("\n\n")
		},
	},
	{
		Filter: []string{"blockquote"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			content = strings.TrimSpace(content)
			if content == "" {
				return nil
			}

			content = multipleNewLinesRegex.ReplaceAllString(content, "\n\n")

			var beginningR = regexp.MustCompile(`(?m)^`)
			content = beginningR.ReplaceAllString(content, "> ")

			text := "\n\n" + content + "\n\n"
			return &text
		},
	},
	{
		Filter: []string{"noscript"},
		Replacement: func(content string, selec *goquery.Selection, opt *Options) *string {
			// for now remove the contents of noscript. But in the future we could
			// tell goquery to parse the contents of the tag.
			// -> https://github.com/PuerkitoBio/goquery/issues/139#issuecomment-517526070
			return nil
		},
	},
}
