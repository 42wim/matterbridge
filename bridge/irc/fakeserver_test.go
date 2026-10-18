package birc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeIRCServer is a minimal single-client IRCv3 server for testing the IRC bridge.
//
// It does not relay messages between users. It records every PRIVMSG/NOTICE the bridge
// sends, gives it a msgid, and echoes it back if echo-message was negotiated.
type fakeIRCServer struct {
	t             *testing.T
	ln            net.Listener
	caps          []string
	clientTagDeny string

	mu       sync.Mutex
	conn     net.Conn
	nick     string
	gotUser  bool
	capOpen  bool
	welcomed bool
	enabled  map[string]bool
	sent     []sentLine
	msgIDs   int
	joined   chan string
}

type sentLine struct {
	command string
	tags    map[string]string
	target  string
	text    string
	msgID   string
}

func newFakeIRCServer(t *testing.T, caps []string, clientTagDeny string) *fakeIRCServer {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %s", err)
	}
	s := &fakeIRCServer{
		t:             t,
		ln:            ln,
		caps:          caps,
		clientTagDeny: clientTagDeny,
		nick:          "*",
		enabled:       map[string]bool{},
		joined:        make(chan string, 10),
	}
	go s.serve()
	t.Cleanup(s.close)
	return s
}

func (s *fakeIRCServer) addr() string { return s.ln.Addr().String() }

func (s *fakeIRCServer) close() {
	s.ln.Close()
	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.mu.Unlock()
}

func (s *fakeIRCServer) serve() {
	conn, err := s.ln.Accept()
	if err != nil {
		return
	}
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		s.handle(scanner.Text())
	}
}

func (s *fakeIRCServer) write(format string, a ...interface{}) {
	line := fmt.Sprintf(format, a...)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		fmt.Fprintf(s.conn, "%s\r\n", line)
	}
}

func (s *fakeIRCServer) capEnabled(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enabled[name]
}

func (s *fakeIRCServer) nextMsgID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgIDs++
	return fmt.Sprintf("msgid%d", s.msgIDs)
}

func parseLine(line string) (tags map[string]string, command string, params []string) {
	tags = map[string]string{}
	if strings.HasPrefix(line, "@") {
		space := strings.IndexByte(line, ' ')
		for _, tag := range strings.Split(line[1:space], ";") {
			kv := strings.SplitN(tag, "=", 2)
			if len(kv) == 2 {
				tags[kv[0]] = kv[1]
			} else {
				tags[kv[0]] = ""
			}
		}
		line = line[space+1:]
	}
	if strings.HasPrefix(line, ":") {
		line = line[strings.IndexByte(line, ' ')+1:]
	}
	trailing := ""
	hasTrailing := false
	if i := strings.Index(line, " :"); i != -1 {
		trailing = line[i+2:]
		hasTrailing = true
		line = line[:i]
	}
	fields := strings.Fields(line)
	command = strings.ToUpper(fields[0])
	params = fields[1:]
	if hasTrailing {
		params = append(params, trailing)
	}
	return tags, command, params
}

func (s *fakeIRCServer) currentNick() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nick
}

func (s *fakeIRCServer) handle(line string) {
	tags, command, params := parseLine(line)
	nick := s.currentNick()
	switch command {
	case "CAP":
		s.handleCap(params)
	case "NICK":
		s.mu.Lock()
		s.nick = params[0]
		s.mu.Unlock()
		s.tryWelcome()
	case "USER":
		s.mu.Lock()
		s.gotUser = true
		s.mu.Unlock()
		s.tryWelcome()
	case "PING":
		s.write(":fake.irc PONG fake.irc :%s", strings.Join(params, " "))
	case "JOIN":
		for _, channel := range strings.Split(params[0], ",") {
			s.write(":%s!bridge@bridge.host JOIN %s", nick, channel)
			s.write(":fake.irc 353 %s = %s :%s", nick, channel, nick)
			s.write(":fake.irc 366 %s %s :End of /NAMES list.", nick, channel)
			s.joined <- channel
		}
	case "WHO":
		s.write(":fake.irc 315 %s %s :End of /WHO list.", nick, params[0])
	case "PRIVMSG", "NOTICE":
		msgID := s.nextMsgID()
		s.mu.Lock()
		s.sent = append(s.sent, sentLine{command, tags, params[0], params[1], msgID})
		s.mu.Unlock()
		if s.capEnabled("echo-message") {
			echoTags := []string{"msgid=" + msgID}
			for k, v := range tags {
				echoTags = append(echoTags, k+"="+v)
			}
			s.write("@%s :%s!bridge@bridge.host %s %s :%s", strings.Join(echoTags, ";"), nick, command, params[0], params[1])
		}
	}
}

func (s *fakeIRCServer) handleCap(params []string) {
	switch strings.ToUpper(params[0]) {
	case "LS":
		s.mu.Lock()
		s.capOpen = true
		s.mu.Unlock()
		s.write(":fake.irc CAP * LS :%s", strings.Join(s.caps, " "))
	case "REQ":
		requested := strings.Fields(params[1])
		for _, r := range requested {
			if !contains(s.caps, r) {
				s.write(":fake.irc CAP %s NAK :%s", s.currentNick(), params[1])
				return
			}
		}
		s.mu.Lock()
		for _, r := range requested {
			s.enabled[r] = true
		}
		s.mu.Unlock()
		s.write(":fake.irc CAP %s ACK :%s", s.currentNick(), strings.Join(requested, " "))
	case "END":
		s.mu.Lock()
		s.capOpen = false
		s.mu.Unlock()
		s.tryWelcome()
	}
}

// tryWelcome completes registration once NICK and USER are in and CAP negotiation is over, like a real server.
func (s *fakeIRCServer) tryWelcome() {
	s.mu.Lock()
	if s.welcomed || s.nick == "*" || !s.gotUser || s.capOpen {
		s.mu.Unlock()
		return
	}
	s.welcomed = true
	nick := s.nick
	s.mu.Unlock()

	isupport := "NETWORK=Fake"
	if s.clientTagDeny != "" {
		isupport += " CLIENTTAGDENY=" + s.clientTagDeny
	}
	s.write(":fake.irc 001 %s :Welcome", nick)
	s.write(":fake.irc 005 %s %s :are supported by this server", nick, isupport)
	s.write(":fake.irc 422 %s :MOTD File is missing", nick)
}

// sendFromUser sends a channel message to the bridge from another user and returns its msgid.
func (s *fakeIRCServer) sendFromUser(nick, channel, text string, extraTags map[string]string) string {
	msgID := s.nextMsgID()
	prefix := ""
	if s.capEnabled("message-tags") {
		tags := []string{"msgid=" + msgID}
		for k, v := range extraTags {
			tags = append(tags, k+"="+v)
		}
		prefix = "@" + strings.Join(tags, ";") + " "
	}
	s.write("%s:%s!user@user.host PRIVMSG %s :%s", prefix, nick, channel, text)
	return msgID
}

// awaitSent waits until the bridge has sent at least n lines, and returns them.
func (s *fakeIRCServer) awaitSent(n int) []sentLine {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		if len(s.sent) >= n {
			out := append([]sentLine(nil), s.sent...)
			s.mu.Unlock()
			return out
		}
		s.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.t.Fatalf("timed out waiting for %d sent lines, got %d: %#v", n, len(s.sent), s.sent)
	return nil
}

func (s *fakeIRCServer) sentCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

func contains(list []string, item string) bool {
	for _, l := range list {
		if l == item {
			return true
		}
	}
	return false
}
