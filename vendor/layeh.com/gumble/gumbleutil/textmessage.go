package gumbleutil

import (
	"bytes"
	"encoding/xml"
	"strings"

	"layeh.com/gumble/gumble"
)

// PlainText returns the Message string without HTML tags or entities.
func PlainText(tm *gumble.TextMessage) string {
	d := xml.NewDecoder(strings.NewReader(tm.Message))
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	var b bytes.Buffer
	newline := false
	for {
		t, _ := d.Token()
		if t == nil {
			break
		}
		switch node := t.(type) {
		case xml.CharData:
			if len(node) > 0 {
				b.Write(node)
				newline = false
			}
		case xml.StartElement:
			switch node.Name.Local {
			case "address", "article", "aside", "audio", "blockquote", "canvas", "dd", "div", "dl", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hgroup", "hr", "noscript", "ol", "output", "p", "pre", "section", "table", "tfoot", "ul", "video":
				if !newline {
					b.WriteByte('\n')
					newline = true
				}
			case "br":
				b.WriteByte('\n')
				newline = true
			}
		}
	}
	return b.String()
}
