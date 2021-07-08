// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"maunium.net/go/mautrix/id"
)

var HTMLReplyFallbackRegex = regexp.MustCompile(`^<mx-reply>[\s\S]+?</mx-reply>`)

func TrimReplyFallbackHTML(html string) string {
	return HTMLReplyFallbackRegex.ReplaceAllString(html, "")
}

func TrimReplyFallbackText(text string) string {
	if !strings.HasPrefix(text, "> ") || !strings.Contains(text, "\n") {
		return text
	}

	lines := strings.Split(text, "\n")
	for len(lines) > 0 && strings.HasPrefix(lines[0], "> ") {
		lines = lines[1:]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func (content *MessageEventContent) RemoveReplyFallback() {
	if len(content.GetReplyTo()) > 0 && !content.replyFallbackRemoved {
		if content.Format == FormatHTML {
			content.FormattedBody = TrimReplyFallbackHTML(content.FormattedBody)
		}
		content.Body = TrimReplyFallbackText(content.Body)
		content.replyFallbackRemoved = true
	}
}

func (content *MessageEventContent) GetReplyTo() id.EventID {
	if content.RelatesTo != nil && content.RelatesTo.Type == RelReply {
		return content.RelatesTo.EventID
	}
	return ""
}

const ReplyFormat = `<mx-reply><blockquote><a href="https://matrix.to/#/%s/%s">In reply to</a> <a href="https://matrix.to/#/%s">%s</a><br>%s</blockquote></mx-reply>`

func (evt *Event) GenerateReplyFallbackHTML() string {
	parsedContent, ok := evt.Content.Parsed.(*MessageEventContent)
	if !ok {
		return ""
	}
	parsedContent.RemoveReplyFallback()
	body := parsedContent.FormattedBody
	if len(body) == 0 {
		body = html.EscapeString(parsedContent.Body)
	}

	senderDisplayName := evt.Sender

	return fmt.Sprintf(ReplyFormat, evt.RoomID, evt.ID, evt.Sender, senderDisplayName, body)
}

func (evt *Event) GenerateReplyFallbackText() string {
	parsedContent, ok := evt.Content.Parsed.(*MessageEventContent)
	if !ok {
		return ""
	}
	parsedContent.RemoveReplyFallback()
	body := parsedContent.Body
	lines := strings.Split(strings.TrimSpace(body), "\n")
	firstLine, lines := lines[0], lines[1:]

	senderDisplayName := evt.Sender

	var fallbackText strings.Builder
	_, _ = fmt.Fprintf(&fallbackText, "> <%s> %s", senderDisplayName, firstLine)
	for _, line := range lines {
		_, _ = fmt.Fprintf(&fallbackText, "\n> %s", line)
	}
	fallbackText.WriteString("\n\n")
	return fallbackText.String()
}

func (content *MessageEventContent) SetReply(inReplyTo *Event) {
	content.RelatesTo = &RelatesTo{
		EventID: inReplyTo.ID,
		Type:    RelReply,
	}

	if content.MsgType == MsgText || content.MsgType == MsgNotice {
		if len(content.FormattedBody) == 0 || content.Format != FormatHTML {
			content.FormattedBody = html.EscapeString(content.Body)
			content.Format = FormatHTML
		}
		content.FormattedBody = inReplyTo.GenerateReplyFallbackHTML() + content.FormattedBody
		content.Body = inReplyTo.GenerateReplyFallbackText() + content.Body
		content.replyFallbackRemoved = false
	}
}
