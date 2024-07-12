package bsoulseek

import (
	"fmt"
	"net"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
)

type Bsoulseek struct {
	conn                                      net.Conn
	messagesToSend                            chan soulseekMessage
	local                                     chan config.Message
	loginResponse                             chan soulseekMessageResponse
	joinRoomResponse                          chan joinRoomMessageResponse
	fatalErrors                               chan error
	disconnect                                chan bool
	firstConnectResponse                      chan error

	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bsoulseek{}
	b.Config = cfg
	b.messagesToSend = make(chan soulseekMessage, 256)
	b.local = make(chan config.Message, 256)
	b.loginResponse = make(chan soulseekMessageResponse)
	b.joinRoomResponse = make(chan joinRoomMessageResponse)
	b.fatalErrors = make(chan error)
	b.disconnect = make(chan bool)
	b.firstConnectResponse = make(chan error)
	return b
}

func (b *Bsoulseek) receiveMessages() {
	for {
		msg, err := readMessage(b.conn)
		if err != nil {
			b.fatalErrors <- fmt.Errorf("Reading message failed: %s", err)
			return
		}
		b.handleMessage(msg)
	}
}

func sliceEqual(s []string) bool {
	// Return true if every element in s is equal to each other
	if len(s) <= 1 {
		return true
	}
	for _, x := range(s) {
		if x != s[0] {
			return false
		}
	}
	return true
}

func (b *Bsoulseek) sendMessages() {
	lastFourChatMessages := []string {"1", "2", "3", ""}
	for {
		message, more := <-b.messagesToSend
		if !more {
			return
		}
		msg, is_say := message.(sayChatroomMessage)
		if is_say {
			// can't send 5 of the same message in a row or we get banned
			if (sliceEqual(append(lastFourChatMessages, msg.Message))) {
				b.Log.Warnf("Dropping message: %s", msg.Message)
				continue
			}
		}
		data, err := packMessage(message)
		if err != nil {
			b.fatalErrors <- fmt.Errorf("Packing message failed: %s", err)
			return
		}
		_, err = b.conn.Write(data)
		if err != nil {
			b.fatalErrors <- fmt.Errorf("Sending message failed: %s", err)
			return
		}
		b.Log.Debugf("Sent message: %v", message)
		if is_say {
			lastFourChatMessages = append(lastFourChatMessages[1:], msg.Message)
			time.Sleep(3500 * time.Millisecond) // rate limit so less than 20 can be sent per min
		}
	}
}

func (b *Bsoulseek) sendLocalToRemote() {
	for {
		message, more := <-b.local
		if !more {
			return
		}
		b.Remote <- message
	}
}

func (b *Bsoulseek) loginLoop() {
	firstConnect := true
	for {
		if !firstConnect {
			// Cleanup as we are making new sender/receiver routines
			b.fatalErrors = make(chan error)
			close(b.messagesToSend)
		}
		// Connect to slsk server
		server := b.GetString("Server")
		b.Log.Infof("Connecting %s", server)
		conn, err := net.Dial("tcp", server)
		b.conn = conn
		if err != nil {
			if firstConnect {
				b.firstConnectResponse <- err
				return
			}
		}

		// Init sender and receiver
		go b.receiveMessages()
		go b.sendMessages()
		go b.sendLocalToRemote()

		// Attempt login
		b.messagesToSend <- makeLoginMessage(b.GetString("Nick"), b.GetString("Password"))
		var msg soulseekMessageResponse
		connected := false
		select {
		case msg = <-b.loginResponse:
			switch msg := msg.(type) {
			case loginMessageResponseSuccess:
				if firstConnect {
					b.firstConnectResponse <- nil
				}
				connected = true
			case loginMessageResponseFailure:
				if firstConnect {
					b.firstConnectResponse <- fmt.Errorf("Login failed: %s", msg.Reason)
					return
				}
				b.Log.Errorf("Login failed: %s", msg.Reason)
			default:
				panic("Unreachable")
			}
		case err := <-b.fatalErrors:
			// error
			if firstConnect {
				b.firstConnectResponse <- fmt.Errorf("Login failed: %s", err)
				return
			}
			b.Log.Errorf("Login failed: %s", err)
		case <-time.After(30 * time.Second):
			// timeout
			if firstConnect {
				b.firstConnectResponse <- fmt.Errorf("Login failed: timeout")
				return
			}
			b.Log.Errorf("Login failed: timeout")
		}

		if !connected {
			// If we reach here, we are not logged in and
			// it is not the first connect, so we should try again
			b.Log.Info("Retrying in 30s")
			time.Sleep(30 * time.Second)
			continue
		}

		// Now we are connected
		select {
		case err = <- b.fatalErrors:
			b.Log.Errorf("%s", err)
			// Retry connect
			continue
		case <- b.disconnect:
			// We are done
			return
		}
	}
}

func (b *Bsoulseek) Connect() error {
	go b.loginLoop()
	err := <-b.firstConnectResponse
	return err
}


func (b *Bsoulseek) JoinChannel(channel config.ChannelInfo) error {
	b.messagesToSend <- makeJoinRoomMessage(channel.Name)
	select {
	case <-b.joinRoomResponse:
		b.Log.Infof("Joined room: '%s'", channel.Name)
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("Could not join room '%s': timeout", channel.Name)
	}
}


func (b *Bsoulseek) Send(msg config.Message) (string, error) {
	// Only process text messages
	b.Log.Debugf("=> Received local message %v", msg)
	if msg.Event != "" && msg.Event != config.EventUserAction && msg.Event != config.EventJoinLeave {
		return "", nil
	}
	b.messagesToSend <- makeSayChatroomMessage(msg.Channel, msg.Username + msg.Text)
	return "", nil
}


func (b *Bsoulseek) doDisconnect() error {
	b.disconnect <- true
	close(b.messagesToSend)
	close(b.joinRoomResponse)
	close(b.loginResponse)
	close(b.local)
	return nil
}


func (b *Bsoulseek) Disconnect() error {
	b.doDisconnect()
	return nil
}