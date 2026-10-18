package birc

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testChannel = "#bridge"

var allCaps = []string{capMessageTags, capEchoMessage}

func TestClientTagDenied(t *testing.T) {
	for _, tc := range []struct {
		deny   string
		denied bool
	}{
		{"", false},
		{"typing", false},
		{"typing,draft/reply", true},
		{"*", true},
		{"*,-draft/reply", false},
		{"*,-draft/typing,-typing,-draft/reply,-reply", false}, // UnrealIRCd 6 default
	} {
		assert.Equal(t, tc.denied, clientTagDenied(tc.deny, "draft/reply"), "CLIENTTAGDENY=%q", tc.deny)
	}
}

func TestThreadsEchoMatching(t *testing.T) {
	threads := newIRCThreads()
	threads.expect(testChannel, "<a> one", "mbirc-1")
	threads.expect("#other", "<c> elsewhere", "mbirc-3")
	threads.expect(testChannel, "<b> two", "mbirc-2")

	// lines nobody is waiting on are ignored
	_, ok := threads.echoed(testChannel, "<z> unrelated", "x")
	assert.False(t, ok)

	// "<a> one" was never echoed; matching "<b> two" drops it but keeps other channels
	localID, ok := threads.echoed(testChannel, "<b> two", "msgid2")
	assert.True(t, ok)
	assert.Equal(t, "mbirc-2", localID)
	_, ok = threads.echoed(testChannel, "<a> one", "msgid1")
	assert.False(t, ok)
	localID, ok = threads.echoed("#OTHER", "<c> elsewhere", "msgid3")
	assert.True(t, ok)
	assert.Equal(t, "mbirc-3", localID)

	assert.Equal(t, "msgid2", threads.replyTarget("mbirc-2"))
	assert.Equal(t, "mbirc-2", threads.gatewayID("msgid2"))
}

func TestThreadsEchoOfSplitLine(t *testing.T) {
	threads := newIRCThreads()
	threads.expect(testChannel, "<a> a long line that the client library split", "mbirc-1")

	localID, ok := threads.echoed(testChannel, "<a> a long line", "msgid1")
	assert.True(t, ok)
	assert.Equal(t, "mbirc-1", localID)
}

func TestThreadsEchoExpires(t *testing.T) {
	threads := newIRCThreads()
	now := time.Now()
	threads.now = func() time.Time { return now }
	threads.expect(testChannel, "<a> slow", "mbirc-1")

	now = now.Add(2 * pendingEchoExpiry)
	_, ok := threads.echoed(testChannel, "<a> slow", "msgid1")
	assert.False(t, ok)
}

func TestThreadsUnknownIDs(t *testing.T) {
	threads := newIRCThreads()
	assert.Equal(t, "", threads.replyTarget("mbirc-unknown"), "local IDs never echoed can't be replied to")
	assert.Equal(t, "msgidX", threads.replyTarget("msgidX"), "other parents are msgids from IRC")
	assert.Equal(t, "msgidX", threads.gatewayID("msgidX"))
}

// testBridge connects a real IRC bridge to a fake server and joins testChannel.
type testBridge struct {
	*Birc
	server *fakeIRCServer
	remote chan config.Message
}

func newTestBridge(t *testing.T, caps []string, clientTagDeny string, extraConfig string) *testBridge {
	server := newFakeIRCServer(t, caps, clientTagDeny)

	cfgText := fmt.Sprintf(`
[irc.test]
Server=%q
Nick="bridge"
MessageDelay=1
RemoteNickFormat="<{NICK}> "
%s
`, server.addr(), extraConfig)

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	br := bridge.New(&config.Bridge{Account: "irc.test"})
	br.Config = config.NewConfigFromString(logger, []byte(cfgText))
	br.Log = logrus.NewEntry(logger)
	remote := make(chan config.Message, 100)

	b := New(&bridge.Config{Bridge: br, Remote: remote}).(*Birc)
	require.NoError(t, b.Connect())
	t.Cleanup(func() { b.i.Close() })
	require.NoError(t, b.JoinChannel(config.ChannelInfo{Name: testChannel}))

	select {
	case <-server.joined:
	case <-time.After(10 * time.Second):
		t.Fatal("bridge never joined")
	}
	return &testBridge{b, server, remote}
}

func (tb *testBridge) send(t *testing.T, text, parentID string) string {
	id, err := tb.Send(config.Message{Text: text, Username: "<WhaTheSht> ", Channel: testChannel, ParentID: parentID})
	require.NoError(t, err)
	return id
}

// awaitMsgID waits for the echo of a message we sent to be mapped to its msgid.
func (tb *testBridge) awaitMsgID(t *testing.T, localID string) string {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if msgID := tb.threads.replyTarget(localID); msgID != "" {
			return msgID
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("echo of %s was never mapped", localID)
	return ""
}

func (tb *testBridge) receive(t *testing.T) config.Message {
	select {
	case msg := <-tb.remote:
		return msg
	case <-time.After(10 * time.Second):
		t.Fatal("nothing relayed to the gateway")
	}
	return config.Message{}
}

func TestThreadingWithFullSupport(t *testing.T) {
	tb := newTestBridge(t, allCaps, "", "PreserveThreading=true")

	t.Run("reply to a message from discord is threaded", func(t *testing.T) {
		before := tb.server.sentCount()
		localID := tb.send(t, "kick attempt is dumb", "")
		require.True(t, strings.HasPrefix(localID, localIDPrefix), "Send should return an ID when echo-message is on")
		msgID := tb.awaitMsgID(t, localID)
		assert.Equal(t, tb.server.awaitSent(before + 1)[before].msgID, msgID)

		tb.send(t, "hilarious hearing this from a pats fan", localID)
		reply := tb.server.awaitSent(before + 2)[before+1]
		assert.Equal(t, "<WhaTheSht> hilarious hearing this from a pats fan", reply.text)
		assert.Equal(t, msgID, reply.tags[tagReply])
	})

	t.Run("reply to a message from irc is threaded", func(t *testing.T) {
		msgID := tb.server.sendFromUser("skud", testChannel, "lmao broadcast camera follows the fake", nil)
		received := tb.receive(t)
		assert.Equal(t, msgID, received.ID, "IRC messages should carry their msgid to the gateway")

		before := tb.server.sentCount()
		tb.send(t, "roasted", msgID)
		assert.Equal(t, msgID, tb.server.awaitSent(before + 1)[before].tags[tagReply])
	})

	t.Run("irc replies carry their parent to the gateway", func(t *testing.T) {
		// parent is a message that came from IRC: the gateway knows it by msgid
		parentMsgID := tb.server.sendFromUser("skud", testChannel, "first", nil)
		tb.receive(t)
		tb.server.sendFromUser("RustyCloud", testChannel, "reply to first", map[string]string{tagReply: parentMsgID})
		assert.Equal(t, parentMsgID, tb.receive(t).ParentID)

		// parent is a message we sent: the gateway knows it by the ID Send() returned
		localID := tb.send(t, "from discord", "")
		ourMsgID := tb.awaitMsgID(t, localID)
		tb.server.sendFromUser("RustyCloud", testChannel, "reply to discord", map[string]string{tagReply: ourMsgID})
		assert.Equal(t, localID, tb.receive(t).ParentID)
	})

	t.Run("unknown parents are sent unthreaded", func(t *testing.T) {
		before := tb.server.sentCount()
		tb.send(t, "parent not in the gateway cache", config.ParentIDNotFound)
		tb.send(t, "parent never echoed", localIDPrefix+"neverseen")
		for _, line := range tb.server.awaitSent(before + 2)[before:] {
			assert.NotContains(t, line.tags, tagReply, line.text)
		}
	})

	t.Run("every line of a multi-line reply is threaded", func(t *testing.T) {
		before := tb.server.sentCount()
		localID := tb.send(t, "first", "")
		msgID := tb.awaitMsgID(t, localID)

		tb.send(t, "line one\nline two\nline three", localID)
		lines := tb.server.awaitSent(before + 4)[before+1:]
		assert.Equal(t, []string{"<WhaTheSht> line one", "<WhaTheSht> line two", "<WhaTheSht> line three"},
			[]string{lines[0].text, lines[1].text, lines[2].text})
		for _, line := range lines {
			assert.Equal(t, msgID, line.tags[tagReply], line.text)
		}
	})

	t.Run("actions are threaded", func(t *testing.T) {
		before := tb.server.sentCount()
		localID := tb.send(t, "first", "")
		msgID := tb.awaitMsgID(t, localID)

		_, err := tb.Send(config.Message{Text: "facepalms", Username: "WhaTheSht ", Channel: testChannel, ParentID: localID, Event: config.EventUserAction})
		require.NoError(t, err)
		action := tb.server.awaitSent(before + 2)[before+1]
		assert.Equal(t, "\x01ACTION WhaTheSht facepalms\x01", action.text)
		assert.Equal(t, msgID, action.tags[tagReply])
	})
}

func TestThreadingFallsBack(t *testing.T) {
	for _, tc := range []struct {
		name          string
		caps          []string
		clientTagDeny string
		config        string
		tracksEchoes  bool
		sendsTags     bool
	}{
		{"PreserveThreading off", allCaps, "", "", false, false},
		{"no message-tags", []string{capEchoMessage}, "", "PreserveThreading=true", false, false},
		{"no echo-message", []string{capMessageTags}, "", "PreserveThreading=true", false, true},
		{"reply tag denied", allCaps, "*", "PreserveThreading=true", true, false},
		{"reply tag allowed by exception", allCaps, "*,-draft/reply", "PreserveThreading=true", true, true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tb := newTestBridge(t, tc.caps, tc.clientTagDeny, tc.config)

			id := tb.send(t, "first", "")
			assert.Equal(t, tc.tracksEchoes, id != "", "Send() returned %q", id)

			// replying to a message from IRC only needs the reply tag
			msgID := tb.server.sendFromUser("skud", testChannel, "from irc", nil)
			received := tb.receive(t)
			if tc.config == "" {
				assert.Empty(t, received.ID, "msgids should only be passed on with PreserveThreading")
				msgID = "msgid-from-irc"
			}

			tb.send(t, "reply", msgID)
			lines := tb.server.awaitSent(2)
			assert.Equal(t, "<WhaTheSht> first", lines[0].text)
			assert.Equal(t, "<WhaTheSht> reply", lines[1].text)
			assert.Empty(t, lines[0].tags)
			if tc.sendsTags {
				assert.Equal(t, msgID, lines[1].tags[tagReply])
			} else {
				assert.Empty(t, lines[1].tags)
			}
		})
	}
}
