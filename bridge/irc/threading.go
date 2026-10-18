package birc

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/lrstanley/girc"
)

// IRCv3 support for PreserveThreading.
//
// IRC has no message IDs of its own, but servers with the message-tags capability give
// each message a msgid tag and relay the +draft/reply client tag, which clients such as
// IRCCloud use to show replies as threads. See
// https://ircv3.net/specs/extensions/message-ids and
// https://ircv3.net/specs/client-tags/reply
//
// Messages are sent through a throttled queue, so Send() can't wait for the server to
// assign a msgid. Instead Send() returns a local ID straight away, and when the server
// echoes the line back (echo-message) the local ID is mapped to the real msgid.

const (
	capEchoMessage = "echo-message"
	capMessageTags = "message-tags"
	tagMsgID       = "msgid"
	tagReply       = "+draft/reply"
	localIDPrefix  = "mbirc-"

	threadCacheSize   = 5000
	maxPendingEchoes  = 256
	pendingEchoExpiry = time.Minute
)

type pendingEcho struct {
	channel string
	text    string
	localID string
	queued  time.Time
}

// ircThreads tracks which IRC msgid belongs to each message ID handed to the gateway.
type ircThreads struct {
	sync.Mutex
	pending      []pendingEcho
	localToMsgID *lru.Cache
	msgIDToLocal *lru.Cache
	now          func() time.Time
}

func newIRCThreads() *ircThreads {
	localToMsgID, _ := lru.New(threadCacheSize)
	msgIDToLocal, _ := lru.New(threadCacheSize)
	return &ircThreads{
		localToMsgID: localToMsgID,
		msgIDToLocal: msgIDToLocal,
		now:          time.Now,
	}
}

func newLocalID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return localIDPrefix + hex.EncodeToString(b)
}

// expect records that we're about to send text to channel, so its echo can be matched to localID.
func (t *ircThreads) expect(channel, text, localID string) {
	t.Lock()
	defer t.Unlock()
	if len(t.pending) >= maxPendingEchoes {
		t.pending = t.pending[1:]
	}
	t.pending = append(t.pending, pendingEcho{channel, text, localID, t.now()})
}

// echoed handles the server echoing a line we sent, mapping its msgid to the local ID it was sent under.
// Lines are echoed in the order they were sent, so anything still pending before the match was lost.
func (t *ircThreads) echoed(channel, text, msgID string) (string, bool) {
	t.Lock()
	defer t.Unlock()

	cutoff := t.now().Add(-pendingEchoExpiry)
	match := -1
	for i, p := range t.pending {
		if p.queued.After(cutoff) && strings.EqualFold(p.channel, channel) && strings.HasPrefix(p.text, text) && text != "" {
			match = i
			break
		}
	}
	if match == -1 {
		return "", false
	}

	localID := t.pending[match].localID
	kept := t.pending[:0]
	for i, p := range t.pending {
		if i > match || (!strings.EqualFold(p.channel, channel) && p.queued.After(cutoff)) {
			kept = append(kept, p)
		}
	}
	t.pending = kept

	t.localToMsgID.Add(localID, msgID)
	t.msgIDToLocal.Add(msgID, localID)
	return localID, true
}

// replyTarget returns the msgid to put in a +draft/reply tag for a gateway parent ID, or "" if unknown.
// The parent is either a local ID we returned from Send(), or the msgid of a message that came from IRC.
func (t *ircThreads) replyTarget(parentID string) string {
	if !strings.HasPrefix(parentID, localIDPrefix) {
		return parentID
	}
	if msgID, ok := t.localToMsgID.Get(parentID); ok {
		return msgID.(string)
	}
	return ""
}

// gatewayID returns the ID the gateway knows an IRC message by: the local ID if we sent it, otherwise its msgid.
func (t *ircThreads) gatewayID(msgID string) string {
	if localID, ok := t.msgIDToLocal.Get(msgID); ok {
		return localID.(string)
	}
	return msgID
}

// clientTagDenied checks a CLIENTTAGDENY ISUPPORT value for the given client tag (without the + prefix).
// See https://ircv3.net/specs/extensions/message-tags#rpl_isupport-tokens
func clientTagDenied(clientTagDeny, tag string) bool {
	if clientTagDeny == "" {
		return false
	}
	entries := strings.Split(clientTagDeny, ",")
	denyAll := false
	for _, e := range entries {
		switch e {
		case "*":
			denyAll = true
		case tag:
			return true
		}
	}
	if denyAll {
		for _, e := range entries {
			if e == "-"+tag {
				return false
			}
		}
		return true
	}
	return false
}

// threadingEnabled reports whether PreserveThreading is on for this IRC account.
func (b *Birc) threadingEnabled() bool {
	return b.GetBool("PreserveThreading")
}

// canTrackEchoes reports whether we can learn the msgid of messages we send.
func (b *Birc) canTrackEchoes() bool {
	return b.threadingEnabled() && b.i.HasCapability(capEchoMessage) && b.i.HasCapability(capMessageTags)
}

// canSendReplyTags reports whether the server will relay +draft/reply for us.
func (b *Birc) canSendReplyTags() bool {
	if !b.threadingEnabled() || !b.i.HasCapability(capMessageTags) {
		return false
	}
	deny, _ := b.i.GetServerOption("CLIENTTAGDENY")
	return !clientTagDenied(deny, strings.TrimPrefix(tagReply, "+"))
}

// replyTags returns the tags to send a message with, or nil to send it untagged.
func (b *Birc) replyTags(parentID string) girc.Tags {
	if parentID == "" || !b.canSendReplyTags() {
		return nil
	}
	target := b.threads.replyTarget(parentID)
	if target == "" {
		return nil
	}
	tags := girc.Tags{}
	if err := tags.Set(tagReply, target); err != nil {
		b.Log.Debugf("not threading reply, invalid msgid %q: %s", target, err)
		return nil
	}
	return tags
}

// handleEcho maps the server's echo of a message we sent to the local ID we returned for it.
func (b *Birc) handleEcho(client *girc.Client, event girc.Event) {
	if !event.Echo || (event.Command != girc.PRIVMSG && event.Command != girc.NOTICE) || len(event.Params) == 0 {
		return
	}
	msgID, ok := event.Tags.Get(tagMsgID)
	if !ok || msgID == "" {
		return
	}
	if localID, ok := b.threads.echoed(event.Params[0], event.Last(), msgID); ok {
		b.Log.Debugf("message %s was sent to %s as msgid %s", localID, event.Params[0], msgID)
	}
}
