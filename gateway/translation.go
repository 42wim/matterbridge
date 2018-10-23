package gateway

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"

	"github.com/russross/blackfriday"
	"github.com/urakozz/go-emoji"
	"github.com/lunny/html2md"
	"github.com/darkoatanasovski/htmltags"
	"golang.org/x/text/language"
	"cloud.google.com/go/translate"
)

func (gw *Gateway) translationEnabled() bool {
	return gw.Router.GTClient != nil
}

func (gw *Gateway) handleTranslation(msg *config.Message, dest *bridge.Bridge, channel config.ChannelInfo) {
	// Skip if channel locale not set
	if channel.Options.Locale == "" {
		return
	}

	// Don't try to translate empty messages
	if msg.OrigMsg.Text == "" {
		return
	}

	msg.IsTranslation = true
	ctx := context.Background()

	client := gw.Router.GTClient
	defer client.Close()

	text := msg.Text
	var results [][]string

	// colons: add temp token
	// This is an ugly hack to work around what seems to be a bug in the Google Translate API.
	// See: https://github.com/42wim/matterbridge/pull/512#issuecomment-428910199
	text = regexp.MustCompile(`(:)([ $])`).ReplaceAllString(text, "<span translate='no'>ː$2</span>")

	// url
	url_re := regexp.MustCompile(`(((http(s)?(\:\/\/))+(www\.)?([\w\-\.\/])*(\.[a-zA-Z]{2,3}\/?))[^\s\n|]*[^.,;:\?\!\@\^\$ -])`)
	text = url_re.ReplaceAllString(text, "<span translate='no'>$0</span>")

	flog.Debugf("pre-parseMD:"+text)

	// Get rid of these wierdo bullets that Slack uses, which confuse translation
	text = strings.Replace(text, "•", "-", -1)

	// Make sure we use closed <br/> tags
	const htmlFlags = blackfriday.HTML_USE_XHTML
	renderer := &renderer{Html: blackfriday.HtmlRenderer(htmlFlags, "", "").(*blackfriday.Html)}
	const extensions = blackfriday.LINK_TYPE_NOT_AUTOLINK |
		blackfriday.EXTENSION_HARD_LINE_BREAK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_HARD_LINE_BREAK
	output := blackfriday.Markdown([]byte(text), renderer, extensions)
	text = string(output)
	flog.Debugf("post-parseMD:"+string(output))

	// @usernames
	results = regexp.MustCompile(`(@[a-zA-Z0-9-]+)`).FindAllStringSubmatch(text, -1)
	// Sort so that longest channel names are acted on first
	sort.SliceStable(results, func(i, j int) bool {
		return len(results[i][1]) > len(results[j][1])
	})
	for _, r := range results {
		text = regexp.MustCompile(fmt.Sprintf(`([^>])(%s)`, r[1])).ReplaceAllString(text, "$1<span translate='no'>$2</span>")
	}

	flog.Debugf("post cleanup:usernames:"+text)

	// #channels
	results = regexp.MustCompile(`(#[a-zA-Z0-9-]+)`).FindAllStringSubmatch(text, -1)
	// Sort so that longest channel names are acted on first
	sort.SliceStable(results, func(i, j int) bool {
		return len(results[i][1]) > len(results[j][1])
	})
	for _, r := range results {
		// If a channel that's a substring of another channel (processed earlier)  matches, it will abort due to the <tag> in front
		text = regexp.MustCompile(fmt.Sprintf(`([^>])(%s)`, r[1])).ReplaceAllString(text, "$1<span translate='no'>$2</span>")
	}

	flog.Debugf("post cleanup:channels:"+text)

	// :emoji:
	text = regexp.MustCompile(`:[a-z0-9-_]+?:`).ReplaceAllString(text, "<span translate='no'>$0</span>")

	// :emoji: codepoints, ie. 💎
	text = emoji.NewEmojiParser().ReplaceAllString(text, "<span translate='no'>$0</span>")

	flog.Debugf("post cleanup:emojis:"+text)

	channelLang, err := language.Parse(channel.Options.Locale)
	if err != nil {
		flog.Error(err)
	}

	resp, _ := client.Translate(ctx, []string{text}, channelLang, &translate.Options{
		Format: "html",
	})

	text = resp[0].Text
	flog.Debugf("post-translate:"+text)

	if resp[0].Source == channelLang {
		msg.IsTranslation = false
		return
	}

	// Add space buffer after html <span> before stripping, or characters after tags get merged into urls or usernames
	text = regexp.MustCompile(`<span translate='no'>.+?</span>`).ReplaceAllString(text, " $0 ")

	allowableTags := []string{
		"p",
		"i",
		"b",
		"em",
		"strong",
		"br",
		"del",
		"blockquote",
		"pre",
		"code",
		"li",
		"ul",
		"ol",
	}

	stripped, _ := htmltags.Strip(text, allowableTags, false)
	text = stripped.ToString()
	flog.Debugf("post-strip:"+text)
	html2md.AddRule("del", &html2md.Rule{
		Patterns: []string{"del"},
		Replacement: func(innerHTML string, attrs []string) string {
			if len(attrs) > 1 {
				// Extra spaces so that Slack will process, even though Chinese characters don't get spaces
				return html2md.WrapInlineTag(attrs[1], " ~", "~ ")
			}
			return ""
		},
	})
	// Custom override for slackdown
	html2md.AddRule("b", &html2md.Rule{
		Patterns: []string{"b", "strong"},
		Replacement: func(innerHTML string, attrs []string) string {
			if len(attrs) > 1 {
				// trailing whitespace due to Mandarin issues
				return html2md.WrapInlineTag(attrs[1], "*", "* ")
			}
			return ""
		},
	})
	// Custom override of default code rule:
	// This converts multiline code tags to codeblocks
	html2md.AddRule("code", &html2md.Rule{
		Patterns: []string{"code", "tt", "pre"},
		Replacement: func(innerHTML string, attrs []string) string {
			contents := attrs[1]
			if strings.Contains(contents, "\n") {
				r := regexp.MustCompile(`/^\t+`)
				innerHTML = r.ReplaceAllString(contents, "  ")
				return "\n\n```\n" + innerHTML + "```\n"
			}
			if len(attrs) > 1 {
				return "`" + attrs[1] + "`"
			}
			return ""
		},
	})
	text = html2md.Convert(text)

	// colons: revert temp token
	// See: previous comment on colons
	text = regexp.MustCompile(`(ː)([ $])`).ReplaceAllString(text, ":$2")

	flog.Debugf("post-MDconvert:"+text)
	text = html.UnescapeString(text)
	flog.Debugf("post-unescaped:"+text)

	if dest.Protocol == "slack" {
		// Attribution will be in attachment for Slack
	} else {
		text = text + gw.Router.General.TranslationAttribution
	}

	msg.Text = text
}
