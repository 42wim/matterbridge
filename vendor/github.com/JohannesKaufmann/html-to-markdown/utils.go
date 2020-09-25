package md

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

/*
WARNING: The functions from this file can be used externally
but there is no garanty that they will stay exported.
*/

// CollectText returns the text of the node and all its children
func CollectText(n *html.Node) string {
	text := &bytes.Buffer{}
	collectText(n, text)
	return text.String()
}
func collectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf)
	}
}

func getName(node *html.Node) string {
	selec := &goquery.Selection{Nodes: []*html.Node{node}}
	return goquery.NodeName(selec)
}

// What elements automatically trim their content?
// Don't add another space if the other element is going to add a
// space already.
func isTrimmedElement(name string) bool {
	nodes := []string{
		"a",
		"strong", "b",
		"i", "em",
		"del", "s", "strike",
	}

	for _, node := range nodes {
		if name == node {
			return true
		}
	}
	return false
}

func getPrevNodeText(node *html.Node) (string, bool) {
	if node == nil {
		return "", false
	}

	for ; node != nil; node = node.PrevSibling {
		text := CollectText(node)

		name := getName(node)
		if name == "br" {
			return "\n", true
		}

		// if the content is empty, try our luck with the next node
		if strings.TrimSpace(text) == "" {
			continue
		}

		if isTrimmedElement(name) {
			text = strings.TrimSpace(text)
		}

		return text, true
	}
	return "", false
}
func getNextNodeText(node *html.Node) (string, bool) {
	if node == nil {
		return "", false
	}

	for ; node != nil; node = node.NextSibling {
		text := CollectText(node)

		name := getName(node)
		if name == "br" {
			return "\n", true
		}

		// if the content is empty, try our luck with the next node
		if strings.TrimSpace(text) == "" {
			continue
		}

		// if you have "a a a", three elements that are trimmed, then only add
		// a space to one side, since the other's are also adding a space.
		if isTrimmedElement(name) {
			text = " "
		}

		return text, true
	}
	return "", false
}

// AddSpaceIfNessesary adds spaces to the text based on the neighbors.
// That makes sure that there is always a space to the side, to recognize the delimiter.
func AddSpaceIfNessesary(selec *goquery.Selection, markdown string) string {
	if len(selec.Nodes) == 0 {
		return markdown
	}
	rootNode := selec.Nodes[0]

	prev, hasPrev := getPrevNodeText(rootNode.PrevSibling)
	if hasPrev {
		lastChar, size := utf8.DecodeLastRuneInString(prev)
		if size > 0 && !unicode.IsSpace(lastChar) {
			markdown = " " + markdown
		}
	}

	next, hasNext := getNextNodeText(rootNode.NextSibling)
	if hasNext {
		firstChar, size := utf8.DecodeRuneInString(next)
		if size > 0 && !unicode.IsSpace(firstChar) && !unicode.IsPunct(firstChar) {
			markdown = markdown + " "
		}
	}

	return markdown
}

// TrimpLeadingSpaces removes spaces from the beginning of a line
// but makes sure that list items and code blocks are not affected.
func TrimpLeadingSpaces(text string) string {
	parts := strings.Split(text, "\n")
	for i := range parts {
		b := []byte(parts[i])

		var spaces int
		for i := 0; i < len(b); i++ {
			if unicode.IsSpace(rune(b[i])) {
				if b[i] == '	' {
					spaces = spaces + 4
				} else {
					spaces++
				}
				continue
			}

			// this seems to be a list item
			if b[i] == '-' {
				break
			}

			// this seems to be a code block
			if spaces >= 4 {
				break
			}

			// remove the space characters from the string
			b = b[i:]
			break
		}
		parts[i] = string(b)
	}

	return strings.Join(parts, "\n")
}

// TrimTrailingSpaces removes unnecessary spaces from the end of lines.
func TrimTrailingSpaces(text string) string {
	parts := strings.Split(text, "\n")
	for i := range parts {
		parts[i] = strings.TrimRightFunc(parts[i], func(r rune) bool {
			return unicode.IsSpace(r)
		})

	}

	return strings.Join(parts, "\n")
}

// The same as `multipleNewLinesRegex`, but applies to escaped new lines inside a link `\n\`
var multipleNewLinesInLinkRegex = regexp.MustCompile(`(\n\\){1,}`) // `([\n\r\s]\\)`

// EscapeMultiLine deals with multiline content inside a link
func EscapeMultiLine(content string) string {
	content = strings.TrimSpace(content)
	content = strings.Replace(content, "\n", `\`+"\n", -1)

	content = multipleNewLinesInLinkRegex.ReplaceAllString(content, "\n\\\n\\")

	return content
}

func calculateCodeFenceOccurrences(fenceChar rune, content string) int {
	var occurrences []int

	var charsTogether int
	for _, char := range content {
		// we encountered a fence character, now count how many
		// are directly afterwards
		if char == fenceChar {
			charsTogether++
		} else if charsTogether != 0 {
			occurrences = append(occurrences, charsTogether)
			charsTogether = 0
		}
	}

	// if the last element in the content was a fenceChar
	if charsTogether != 0 {
		occurrences = append(occurrences, charsTogether)
	}

	return findMax(occurrences)
}

// CalculateCodeFence can be passed the content of a code block and it returns
// how many fence characters (` or ~) should be used.
//
// This is useful if the html content includes the same fence characters
// for example ```
// -> https://stackoverflow.com/a/49268657
func CalculateCodeFence(fenceChar rune, content string) string {
	repeat := calculateCodeFenceOccurrences(fenceChar, content)

	// the outer fence block always has to have
	// at least one character more than any content inside
	repeat++

	// you have to have at least three fence characters
	// to be recognized as a code block
	if repeat < 3 {
		repeat = 3
	}

	return strings.Repeat(string(fenceChar), repeat)
}

func findMax(a []int) (max int) {
	for i, value := range a {
		if i == 0 {
			max = a[i]
		}

		if value > max {
			max = value
		}
	}
	return max
}
