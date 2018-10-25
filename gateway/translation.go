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

var (
	uriRE = regexp.MustCompile(`(((http(s)?(\:\/\/))+(www\.)?([\w\-\.\/])*(\.[a-zA-Z]{2,3}\/?))[^\s\n|]*[^.,;:\?\!\@\^\$ -])`)
	usernameRE = regexp.MustCompile(`(@[a-zA-Z0-9-]+)`)
	channelsRE = regexp.MustCompile(`(#[a-zA-Z0-9-]+)`)
	bugfixRE = regexp.MustCompile(`(:)([ $])`)
	bugfixUndoRE = regexp.MustCompile(`(ː)([ $])`)
)

func (gw *Gateway) translationEnabled() bool {
	return gw.Router.GTClient != nil
}

func protectUrls(text string) string {
	return uriRE.ReplaceAllString(text, "<span translate='no'>$0</span>")
}

func protectBullets(text string) string {
	// Get rid of these wierdo bullets that Slack uses, which confuse translation
	return strings.Replace(text, "•", "-", -1)
}

func protectUsernames(text string) string {
	// @usernames
	results := usernameRE.FindAllStringSubmatch(text, -1)

	// Sort so that longest channel names are acted on first
	sort.SliceStable(results, func(i, j int) bool {
		return len(results[i][1]) > len(results[j][1])
	})
	for _, r := range results {
		text = regexp.MustCompile(fmt.Sprintf(`([^>])(%s)`, r[1])).ReplaceAllString(text, "$1<span translate='no'>$2</span>")
	}
	flog.Debugf("Post cleanup:usernames: " + text)

	return text
}

func protectChannels(text string) string {
	// #channels
	results := channelsRE.FindAllStringSubmatch(text, -1)
	// Sort so that longest channel names are acted on first
	sort.SliceStable(results, func(i, j int) bool {
		return len(results[i][1]) > len(results[j][1])
	})
	for _, r := range results {
		// If a channel that's a substring of another channel (processed earlier)  matches, it will abort due to the <tag> in front
		text = regexp.MustCompile(fmt.Sprintf(`([^>])(%s)`, r[1])).ReplaceAllString(text, "$1<span translate='no'>$2</span>")
	}
	flog.Debugf("Post cleanup:channels: " + text)

	return text

}

func protectEmoji(text string) string {
	// :emoji:
	text = regexp.MustCompile(`:[a-z0-9-_]+?:`).ReplaceAllString(text, "<span translate='no'>$0</span>")

	// :emoji: codepoints, ie. 💎
	text = emoji.NewEmojiParser().ReplaceAllString(text, "<span translate='no'>$0</span>")

	flog.Debugf("post cleanup:emojis:"+text)
	return text
}

func convertMarkdown2Html(text string) string {
	// Make sure we use closed <br/> tags
	const htmlFlags = blackfriday.UseXHTML
	const extensions = blackfriday.HardLineBreak |
		blackfriday.Strikethrough |
		blackfriday.FencedCode
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: htmlFlags,
	})
	optList := []blackfriday.Option{
		blackfriday.WithNoExtensions(),
		blackfriday.WithExtensions(extensions),
		blackfriday.WithRenderer(renderer),
	}
	text = string(blackfriday.Run([]byte(text), optList...))
	flog.Debugf("Post-md2html: " + text)

	return text
}

func (gw *Gateway) translateText(msg *config.Message, locale string) string {
	ctx := context.Background()
	text := msg.Text

	channelLang, err := language.Parse(locale)
	if err != nil {
		flog.Error(err)
	}

	client := gw.Router.GTClient
	defer client.Close()

	resp, _ := client.Translate(ctx, []string{text}, channelLang, &translate.Options{
		Format: "html",
		Model: "nmt",
	})

	text = resp[0].Text
	flog.Debugf("Post-translation: " + text)

	if resp[0].Source == channelLang {
		msg.IsTranslation = false
	}

	return text
}

func guardBugfix(text string) string {
	// colons: add temp token
	// This is an ugly hack to work around what seems to be a bug in the Google Translate API.
	// See: https://github.com/42wim/matterbridge/pull/512#issuecomment-428910199
	return bugfixRE.ReplaceAllString(text, "<span translate='no'>ː$2</span>")
}

func stripHtml(text string) string {
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
	flog.Debugf("Post-strip: " + text)

	return text
}

func convertHtml2Markdown(text string) string {
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
	flog.Debugf("Post-html2md: " + text)

	return text
}

func unguardBugfix(text string) string {
	// colons: revert temp token
	// See: previous comment on colons
	return bugfixUndoRE.ReplaceAllString(text, ":$2")
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

	text := msg.Text
	text = guardBugfix(text)
	text = protectUrls(text)
	text = protectBullets(text)
	text = convertMarkdown2Html(text)
	text = protectUsernames(text)
	text = protectChannels(text)
	text = protectEmoji(text)
	text = gw.translateText(msg, channel.Options.Locale)

	// Add space buffer after html <span> before stripping, or characters after tags get merged into urls or usernames
	text = regexp.MustCompile(`<span translate='no'>.+?</span>`).ReplaceAllString(text, " $0 ")

	text = stripHtml(text)
	text = convertHtml2Markdown(text)
	text = unguardBugfix(text)

	text = html.UnescapeString(text)
	flog.Debugf("post-unescaped:"+text)

	// Don't show translation if only whitespace/caps different.
	// eg. messages with only emoji, links, or untranslatable gibberish
	if strings.ToLower(strings.Replace(text, " ", "", -1)) == strings.ToLower(strings.Replace(msg.Text, " ", "", -1)) {
		msg.IsTranslation = false
	}

	if msg.IsTranslation == false {
		return
	}

	if dest.Protocol == "slack" {
		// Attribution will be in attachment for Slack
	} else {
		text = text + gw.Router.General.TranslationAttribution
	}

	msg.Text = text
}
