// a go port of html2md javascript version

package html2md

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func P() *Rule {
	return &Rule{
		Patterns: []string{"p"},
		Replacement: func(innerHTML string, attrs []string) string {
			if len(attrs) > 1 {
				return "\n\n" + attrs[1] + "\n"
			}
			return ""
		},
	}
}

func Br() *Rule {
	return &Rule{
		Patterns: []string{"br"},
		Tp:       Void,
		Replacement: func(innerHTML string, attrs []string) string {
			return "  \n"
		},
	}
}

func H() *Rule {
	return &Rule{
		Patterns: []string{"h([1-6])"},
		Replacement: func(innerHTML string, attrs []string) string {
			if len(attrs) < 4 || attrs[0] != attrs[len(attrs)-1] {
				return ""
			}

			hLevel, err := strconv.Atoi(attrs[0])
			if err != nil {
				fmt.Println(err)
				return ""
			}

			return "\n\n" + strings.Repeat("#", hLevel) +
				" " + attrs[2] + "\n"
		},
	}
}

func Hr() *Rule {
	return &Rule{
		Patterns: []string{"hr"},
		Tp:       Void,
		Replacement: func(innerHTML string, attrs []string) string {
			return "\n\n* * *\n"
		},
	}
}

func B() *Rule {
	return &Rule{
		Patterns: []string{"b", "strong"},
		Replacement: func(innerHTML string, attrs []string) string {
			if len(attrs) > 1 {
				return wrapInlineTag(attrs[1], "**", "**")
			}
			return ""
		},
	}
}

func I() *Rule {
	return &Rule{
		Patterns: []string{"i", "em"},
		Replacement: func(innerHTML string, attrs []string) string {
			if len(attrs) > 1 {
				return wrapInlineTag(attrs[1], "_", "_")
			}
			return ""
		},
	}
}

func Code() *Rule {
	return &Rule{
		Patterns: []string{"code", "tt", "pre"},
		Replacement: func(innerHTML string, attrs []string) string {
			if len(attrs) > 1 {
				return "`" + attrs[1] + "`"
			}
			return ""
		},
	}
}

func A() *Rule {
	return &Rule{
		Patterns: []string{"a"},
		Replacement: func(innerHTML string, attrs []string) string {
			var href string
			hrefR := AttrRegExp("href")
			matches := hrefR.FindStringSubmatch(attrs[0])
			if len(matches) > 1 {
				href = matches[1]
			}

			/*targetR := AttrRegExp("target")
			matches = targetR.FindStringSubmatch(attrs[0])
			if len(matches) > 1 {
				target = matches[1]
			}*/

			//if len(target) > 0 {
			//	return "[" + alt + "]" + "(" + src + " \"" + title + "\")"
			//}
			return wrapInlineTag(attrs[1], "[", "]") + "(" + href + ")"
		},
	}
}

func SameRule(tag string, tp int) *Rule {
	return &Rule{Patterns: []string{tag},
		Tp: tp,
		Replacement: func(innerHTML string, attrs []string) string {
			return innerHTML
		},
	}
}

func Img() *Rule {
	return &Rule{
		Patterns: []string{"img"},
		Tp:       Void,
		Replacement: func(innerHTML string, attrs []string) string {
			var src, alt, title string
			srcR := AttrRegExp("src")
			matches := srcR.FindStringSubmatch(attrs[0])
			if len(matches) > 1 {
				src = matches[1]
			}

			altR := AttrRegExp("alt")
			matches = altR.FindStringSubmatch(attrs[0])
			if len(matches) > 1 {
				alt = matches[1]
			}

			titleR := AttrRegExp("title")
			matches = titleR.FindStringSubmatch(attrs[0])
			if len(matches) > 1 {
				title = matches[1]
			}

			if len(title) > 0 {
				if len(alt) == 0 {
					alt = title
				}
				return "![" + alt + "]" + "(" + src + " \"" + title + "\")"
			}
			if len(alt) == 0 {
				alt = "image"
			}
			return "![" + alt + "]" + "(" + src + ")"
		},
	}
}

func replaceEls(html, tag string, tp int, replacement ReplaceFunc) string {
	var pattern string
	if tp == Void {
		pattern = "<" + tag + "\\b([^>]*)\\/?>"
	} else {
		pattern = "<" + tag + "\\b([^>]*)>([\\s\\S]*?)<\\/" + tag + ">"
	}

	re := regexp.MustCompile(pattern)
	return re.ReplaceAllStringFunc(html, func(subHtml string) string {
		matches := re.FindStringSubmatch(subHtml)
		//fmt.Println("xx", subHtml, matches)
		return replacement(subHtml, matches[1:])
	})
}

func replaceLists(tag, html string) string {
	re := regexp.MustCompile(`<(` + tag + `)\b[^>]*>([\s\S]*?)</` + tag + `>`)
	html = re.ReplaceAllStringFunc(html, func(innerHTML string) string {
		var lis = strings.Split(innerHTML, "</li>")
		var newLis = make([]string, 0)
		var prefix string = "*   "

		for i, li := range lis[:len(lis)-1] {
			if tag == "ol" {
				prefix = fmt.Sprintf("%d.  ", i+1)
			}

			re := regexp.MustCompile(`([\s\S]*)<li[^>]*>([\s\S]*)`)
			newLis = append(newLis, re.ReplaceAllStringFunc(li, func(innerHTML string) string {
				matches := re.FindStringSubmatch(innerHTML)
				innerHTML = regexp.MustCompile(`/^\s+/`).ReplaceAllString(matches[2], "")
				innerHTML = regexp.MustCompile(`/\n\n/g`).ReplaceAllString(innerHTML, "\n\n    ")
				// indent nested lists
				innerHTML = regexp.MustCompile(`/\n([ ]*)+(\*|\d+\.) /g`).ReplaceAllString(innerHTML, "\n$1    $2 ")
				return prefix + innerHTML
			}))
		}

		return strings.Join(newLis, "\n")
	})

	return "\n\n" + regexp.MustCompile(`[ \t]+\n|\s+$`).ReplaceAllString(html, "")
}

func replaceBlockquotes(html string) string {
	re := regexp.MustCompile(`<blockquote\b[^>]*>([\s\S]*?)</blockquote>`)
	return re.ReplaceAllStringFunc(html, func(inner string) string {
		matches := re.FindStringSubmatch(inner)
		inner = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(matches[1], "")
		inner = cleanUp(inner)
		inner = regexp.MustCompile(`(?m)^`).ReplaceAllString(inner, "> ")
		inner = regexp.MustCompile(`^(>([ \t]{2,}>)+)`).ReplaceAllString(inner, "> >")
		return inner
	})
}

func blockQuote(content string) string {
	// Blockquotes
	//var deepest = `<blockquote\b[^>]*>((?:(?!<blockquote)[\s\S])*?)</blockquote>`
	var deepest = `<blockquote\b[^>]*>((?:[\s\S])*?)</blockquote>`

	re := regexp.MustCompile(deepest)
	content = re.ReplaceAllStringFunc(content, func(str string) string {
		return replaceBlockquotes(str)
	})

	return content
}

func Remove(ct, tag string) string {
	re := regexp.MustCompile("\\<" + tag + "[\\S\\s]+?\\</" + tag + "\\>")
	return re.ReplaceAllString(ct, "")
}

func cleanUp(ct string) string {
	// trim leading/trailing whitespace
	str := regexp.MustCompile("^[\t\r\n]+|[\t\r\n]+$").ReplaceAllString(ct, "")
	str = regexp.MustCompile(`\n\s+\n`).ReplaceAllString(str, "\n\n")
	// limit consecutive linebreaks to 2
	str = regexp.MustCompile(`\n{3,}`).ReplaceAllString(str, "\n\n")

	//去除STYLE
	str = Remove(str, "style")

	//去除SCRIPT
	str = Remove(str, "script")

	//去除所有尖括号内的HTML代码，并换成换行符
	re := regexp.MustCompile("\\<[\\S\\s]+?\\>")
	str = re.ReplaceAllString(str, "\n")

	//去除连续的换行符
	//re = regexp.MustCompile("\\s{2,}")
	//str = re.ReplaceAllString(str, "\n")
	return str
}

func pre(content string) string {
	// Pre code blocks
	re := regexp.MustCompile(`<pre\b[^>]*>([\s\S]*)</pre>`)
	content = re.ReplaceAllStringFunc(content, func(innerHTML string) string {
		matches := re.FindStringSubmatch(innerHTML)
		// convert tabs to spaces (you know it makes sense)
		r := regexp.MustCompile(`/^\t+`)
		innerHTML = r.ReplaceAllString(matches[1], "  ")
		r = regexp.MustCompile(`/\n`)
		innerHTML = r.ReplaceAllString(innerHTML, "\n    ")
		return "\n\n    " + innerHTML + "\n"
	})
	return content
}

func ul(content string) string {
	return ulol("ul", content)
}

func ol(content string) string {
	return ulol("ol", content)
}

func ulol(tag, content string) string {
	// Lists

	// Escape numbers that could trigger an ol
	// If there are more than three spaces before the code, it would be in a pre tag
	// Make sure we are escaping the period not matching any character

	//content = string.replace(^(\s{0,3}\d+)\. /g, "$1\\. ");

	// Converts lists that have no child lists (of same type) first, then works it"s way up
	//var noChildrenRegex = /<(ul|ol)\b[^>]*>(?:(?!<ul|<ol)[\s\S])*?<\/\1>/gi;
	var noChildrenRegex = `<(` + tag + `)\b[^>]*>(?:[\s\S])*?</` + tag + `>`
	re := regexp.MustCompile(noChildrenRegex)
	return re.ReplaceAllStringFunc(content, func(str string) string {
		return replaceLists(tag, str)
	})
}

func wrapInlineTag(content, openWrap, closeWrap string) string {
	wrappedStr := openWrap + strings.TrimSpace(content) + closeWrap
	if regexp.MustCompile(`^\s.*`).MatchString(content) {
		wrappedStr = " " + wrappedStr
	}
	if regexp.MustCompile(`.*\s$`).MatchString(content) {
		wrappedStr = wrappedStr + " "
	}
	return wrappedStr
}

func WrapInlineTag(content, openWrap, closeWrap string) string {
  return wrapInlineTag(content, openWrap, closeWrap)
}

func init() {
	AddRule("p", P())
	AddRule("i", I())
	AddRule("h", H())
	AddRule("hr", Hr())
	AddRule("img", Img())
	AddRule("b", B())
	AddRule("br", Br())
	AddRule("code", Code())
	AddRule("a", A())

	AddConvert(pre)
	AddConvert(ul)
	AddConvert(ol)
	AddConvert(blockQuote)
	AddConvert(cleanUp)
}

func Convert(content string) string {
	for _, rule := range rules {
		for _, pattern := range rule.Patterns {
			content = replaceEls(content, pattern, rule.Tp, rule.Replacement)
		}
	}

	for _, convert := range converts {
		content = convert(content)
	}

	return content
}
