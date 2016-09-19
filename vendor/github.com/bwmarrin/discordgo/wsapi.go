// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains low level functions for interacting with the Discord
// data websocket interface.

package discordgo

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
)

type resumePacket struct {
	Op   int `json:"op"`
	Data struct {
		Token     string `json:"token"`
		SessionID string `json:"session_id"`
		Sequence  int    `json:"seq"`
	} `json:"d"`
}

// Open opens a websocket connection to Discord.
func (s *Session) Open() (err error) {

	s.log(LogInformational, "called")

	s.Lock()
	defer func() {
		if err != nil {
			s.Unlock()
		}
	}()

	if s.wsConn != nil {
		err = errors.New("Web socket already opened.")
		return
	}

	if s.VoiceConnections == nil {
		s.log(LogInformational, "creating new VoiceConnections map")
		s.VoiceConnections = make(map[string]*VoiceConnection)
	}

	// Get the gateway to use for the Websocket connection
	if s.gateway == "" {
		s.gateway, err = s.Gateway()
		if err != nil {
			return
		}

		// Add the version and encoding to the URL
		s.gateway = fmt.Sprintf("%s?v=4&encoding=json", s.gateway)
	}

	header := http.Header{}
	header.Add("accept-encoding", "zlib")

	s.log(LogInformational, "connecting to gateway %s", s.gateway)
	s.wsConn, _, err = websocket.DefaultDialer.Dial(s.gateway, header)
	if err != nil {
		s.log(LogWarning, "error connecting to gateway %s, %s", s.gateway, err)
		s.gateway = "" // clear cached gateway
		// TODO: should we add a retry block here?
		return
	}

	if s.sessionID != "" && s.sequence > 0 {

		p := resumePacket{}
		p.Op = 6
		p.Data.Token = s.Token
		p.Data.SessionID = s.sessionID
		p.Data.Sequence = s.sequence

		s.log(LogInformational, "sending resume packet to gateway")
		err = s.wsConn.WriteJSON(p)
		if err != nil {
			s.log(LogWarning, "error sending gateway resume packet, %s, %s", s.gateway, err)
			return
		}

	} else {

		err = s.identify()
		if err != nil {
			s.log(LogWarning, "error sending gateway identify packet, %s, %s", s.gateway, err)
			return
		}
	}

	// Create listening outside of listen, as it needs to happen inside the mutex
	// lock.
	s.listening = make(chan interface{})
	go s.listen(s.wsConn, s.listening)

	s.Unlock()

	s.initialize()
	s.log(LogInformational, "emit connect event")
	s.handle(&Connect{})

	s.log(LogInformational, "exiting")
	return
}

// listen polls the websocket connection for events, it will stop when the
// listening channel is closed, or an error occurs.
func (s *Session) listen(wsConn *websocket.Conn, listening <-chan interface{}) {

	s.log(LogInformational, "called")

	for {

		messageType, message, err := wsConn.ReadMessage()

		if err != nil {

			// Detect if we have been closed manually. If a Close() has already
			// happened, the websocket we are listening on will be different to
			// the current session.
			s.RLock()
			sameConnection := s.wsConn == wsConn
			s.RUnlock()

			if sameConnection {

				s.log(LogWarning, "error reading from gateway %s websocket, %s", s.gateway, err)
				// There has been an error reading, close the websocket so that
				// OnDisconnect event is emitted.
				err := s.Close()
				if err != nil {
					s.log(LogWarning, "error closing session connection, %s", err)
				}

				s.log(LogInformational, "calling reconnect() now")
				s.reconnect()
			}

			return
		}

		select {

		case <-listening:
			return

		default:
			s.onEvent(messageType, message)

		}
	}
}

type heartbeatOp struct {
	Op   int `json:"op"`
	Data int `json:"d"`
}

// heartbeat sends regular heartbeats to Discord so it knows the client
// is still connected.  If you do not send these heartbeats Discord will
// disconnect the websocket connection after a few seconds.
func (s *Session) heartbeat(wsConn *websocket.Conn, listening <-chan interface{}, i time.Duration) {

	s.log(LogInformational, "called")

	if listening == nil || wsConn == nil {
		return
	}

	var err error
	ticker := time.NewTicker(i * time.Millisecond)

	for {

		s.log(LogInformational, "sending gateway websocket heartbeat seq %d", s.sequence)
		s.wsMutex.Lock()
		err = wsConn.WriteJSON(heartbeatOp{1, s.sequence})
		s.wsMutex.Unlock()
		if err != nil {
			s.log(LogError, "error sending heartbeat to gateway %s, %s", s.gateway, err)
			s.Lock()
			s.DataReady = false
			s.Unlock()
			return
		}
		s.Lock()
		s.DataReady = true
		s.Unlock()

		select {
		case <-ticker.C:
			// continue loop and send heartbeat
		case <-listening:
			return
		}
	}
}

type updateStatusData struct {
	IdleSince *int  `json:"idle_since"`
	Game      *Game `json:"game"`
}

type updateStatusOp struct {
	Op   int              `json:"op"`
	Data updateStatusData `json:"d"`
}

// UpdateStreamingStatus is used to update the user's streaming status.
// If idle>0 then set status to idle.
// If game!="" then set game.
// If game!="" and url!="" then set the status type to streaming with the URL set.
// if otherwise, set status to active, and no game.
func (s *Session) UpdateStreamingStatus(idle int, game string, url string) (err error) {

	s.log(LogInformational, "called")

	s.RLock()
	defer s.RUnlock()
	if s.wsConn == nil {
		return errors.New("no websocket connection exists")
	}

	var usd updateStatusData
	if idle > 0 {
		usd.IdleSince = &idle
	}

	if game != "" {
		gameType := 0
		if url != "" {
			gameType = 1
		}
		usd.Game = &Game{
			Name: game,
			Type: gameType,
			URL:  url,
		}
	}

	s.wsMutex.Lock()
	err = s.wsConn.WriteJSON(updateStatusOp{3, usd})
	s.wsMutex.Unlock()

	return
}

// UpdateStatus is used to update the user's status.
// If idle>0 then set status to idle.
// If game!="" then set game.
// if otherwise, set status to active, and no game.
func (s *Session) UpdateStatus(idle int, game string) (err error) {
	return s.UpdateStreamingStatus(idle, game, "")
}

// onEvent is the "event handler" for all messages received on the
// Discord Gateway API websocket connection.
//
// If you use the AddHandler() function to register a handler for a
// specific event this function will pass the event along to that handler.
//
// If you use the AddHandler() function to register a handler for the
// "OnEvent" event then all events will be passed to that handler.
//
// TODO: You may also register a custom event handler entirely using...
func (s *Session) onEvent(messageType int, message []byte) {

	var err error
	var reader io.Reader
	reader = bytes.NewBuffer(message)

	// If this is a compressed message, uncompress it.
	if messageType == websocket.BinaryMessage {

		z, err2 := zlib.NewReader(reader)
		if err2 != nil {
			s.log(LogError, "error uncompressing websocket message, %s", err)
			return
		}

		defer func() {
			err3 := z.Close()
			if err3 != nil {
				s.log(LogWarning, "error closing zlib, %s", err)
			}
		}()

		reader = z
	}

	// Decode the event into an Event struct.
	var e *Event
	decoder := json.NewDecoder(reader)
	if err = decoder.Decode(&e); err != nil {
		s.log(LogError, "error decoding websocket message, %s", err)
		return
	}

	s.log(LogDebug, "Op: %d, Seq: %d, Type: %s, Data: %s\n\n", e.Operation, e.Sequence, e.Type, string(e.RawData))

	// Ping request.
	// Must respond with a heartbeat packet within 5 seconds
	if e.Operation == 1 {
		s.log(LogInformational, "sending heartbeat in response to Op1")
		s.wsMutex.Lock()
		err = s.wsConn.WriteJSON(heartbeatOp{1, s.sequence})
		s.wsMutex.Unlock()
		if err != nil {
			s.log(LogError, "error sending heartbeat in response to Op1")
			return
		}

		return
	}

	// Reconnect
	// Must immediately disconnect from gateway and reconnect to new gateway.
	if e.Operation == 7 {
		// TODO
	}

	// Invalid Session
	// Must respond with a Identify packet.
	if e.Operation == 9 {

		s.log(LogInformational, "sending identify packet to gateway in response to Op9")

		err = s.identify()
		if err != nil {
			s.log(LogWarning, "error sending gateway identify packet, %s, %s", s.gateway, err)
			return
		}

		return
	}

	// Do not try to Dispatch a non-Dispatch Message
	if e.Operation != 0 {
		// But we probably should be doing something with them.
		// TEMP
		s.log(LogWarning, "unknown Op: %d, Seq: %d, Type: %s, Data: %s, message: %s", e.Operation, e.Sequence, e.Type, string(e.RawData), string(message))
		return
	}

	// Store the message sequence
	s.sequence = e.Sequence

	// Map event to registered event handlers and pass it along
	// to any registered functions
	i := eventToInterface[e.Type]
	if i != nil {

		// Create a new instance of the event type.
		i = reflect.New(reflect.TypeOf(i)).Interface()

		// Attempt to unmarshal our event.
		if err = json.Unmarshal(e.RawData, i); err != nil {
			s.log(LogError, "error unmarshalling %s event, %s", e.Type, err)
		}

		// Send event to any registered event handlers for it's type.
		// Because the above doesn't cancel this, in case of an error
		// the struct could be partially populated or at default values.
		// However, most errors are due to a single field and I feel
		// it's better to pass along what we received than nothing at all.
		// TODO: Think about that decision :)
		// Either way, READY events must fire, even with errors.
		go s.handle(i)

	} else {
		s.log(LogWarning, "unknown event: Op: %d, Seq: %d, Type: %s, Data: %s", e.Operation, e.Sequence, e.Type, string(e.RawData))
	}

	// Emit event to the OnEvent handler
	e.Struct = i
	go s.handle(e)
}

// ------------------------------------------------------------------------------------------------
// Code related to voice connections that initiate over the data websocket
// ------------------------------------------------------------------------------------------------

// A VoiceServerUpdate stores the data received during the Voice Server Update
// data websocket event. This data is used during the initial Voice Channel
// join handshaking.
type VoiceServerUpdate struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

type voiceChannelJoinData struct {
	GuildID   *string `json:"guild_id"`
	ChannelID *string `json:"channel_id"`
	SelfMute  bool    `json:"self_mute"`
	SelfDeaf  bool    `json:"self_deaf"`
}

type voiceChannelJoinOp struct {
	Op   int                  `json:"op"`
	Data voiceChannelJoinData `json:"d"`
}

// ChannelVoiceJoin joins the session user to a voice channel.
//
//    gID     : Guild ID of the channel to join.
//    cID     : Channel ID of the channel to join.
//    mute    : If true, you will be set to muted upon joining.
//    deaf    : If true, you will be set to deafened upon joining.
func (s *Session) ChannelVoiceJoin(gID, cID string, mute, deaf bool) (voice *VoiceConnection, err error) {

	s.log(LogInformational, "called")

	voice, _ = s.VoiceConnections[gID]

	if voice == nil {
		voice = &VoiceConnection{}
		s.VoiceConnections[gID] = voice
	}

	voice.GuildID = gID
	voice.ChannelID = cID
	voice.deaf = deaf
	voice.mute = mute
	voice.session = s

	// Send the request to Discord that we want to join the voice channel
	data := voiceChannelJoinOp{4, voiceChannelJoinData{&gID, &cID, mute, deaf}}
	s.wsMutex.Lock()
	err = s.wsConn.WriteJSON(data)
	s.wsMutex.Unlock()
	if err != nil {
		return
	}

	// doesn't exactly work perfect yet.. TODO
	err = voice.waitUntilConnected()
	if err != nil {
		s.log(LogWarning, "error waiting for voice to connect, %s", err)
		voice.Close()
		return
	}

	return
}

// onVoiceStateUpdate handles Voice State Update events on the data websocket.
func (s *Session) onVoiceStateUpdate(se *Session, st *VoiceStateUpdate) {

	// If we don't have a connection for the channel, don't bother
	if st.ChannelID == "" {
		return
	}

	// Check if we have a voice connection to update
	voice, exists := s.VoiceConnections[st.GuildID]
	if !exists {
		return
	}

	// Need to have this happen at login and store it in the Session
	// TODO : This should be done upon connecting to Discord, or
	// be moved to a small helper function
	self, err := s.User("@me") // TODO: move to Login/New
	if err != nil {
		log.Println(err)
		return
	}

	// We only care about events that are about us
	if st.UserID != self.ID {
		return
	}

	// Store the SessionID for later use.
	voice.UserID = self.ID // TODO: Review
	voice.sessionID = st.SessionID
}

// onVoiceServerUpdate handles the Voice Server Update data websocket event.
//
// This is also fired if the Guild's voice region changes while connected
// to a voice channel.  In that case, need to re-establish connection to
// the new region endpoint.
func (s *Session) onVoiceServerUpdate(se *Session, st *VoiceServerUpdate) {

	s.log(LogInformational, "called")

	voice, exists := s.VoiceConnections[st.GuildID]

	// If no VoiceConnection exists, just skip this
	if !exists {
		return
	}

	// If currently connected to voice ws/udp, then disconnect.
	// Has no effect if not connected.
	voice.Close()

	// Store values for later use
	voice.token = st.Token
	voice.endpoint = st.Endpoint
	voice.GuildID = st.GuildID

	// Open a conenction to the voice server
	err := voice.open()
	if err != nil {
		s.log(LogError, "onVoiceServerUpdate voice.open, %s", err)
	}
}

type identifyProperties struct {
	OS              string `json:"$os"`
	Browser         string `json:"$browser"`
	Device          string `json:"$device"`
	Referer         string `json:"$referer"`
	ReferringDomain string `json:"$referring_domain"`
}

type identifyData struct {
	Token          string             `json:"token"`
	Properties     identifyProperties `json:"properties"`
	LargeThreshold int                `json:"large_threshold"`
	Compress       bool               `json:"compress"`
	Shard          *[2]int            `json:"shard,omitempty"`
}

type identifyOp struct {
	Op   int          `json:"op"`
	Data identifyData `json:"d"`
}

// identify sends the identify packet to the gateway
func (s *Session) identify() error {

	properties := identifyProperties{runtime.GOOS,
		"Discordgo v" + VERSION,
		"",
		"",
		"",
	}

	data := identifyData{s.Token,
		properties,
		250,
		s.Compress,
		nil,
	}

	if s.ShardCount > 1 {

		if s.ShardID >= s.ShardCount {
			return errors.New("ShardID must be less than ShardCount")
		}

		data.Shard = &[2]int{s.ShardID, s.ShardCount}
	}

	op := identifyOp{2, data}

	s.wsMutex.Lock()
	err := s.wsConn.WriteJSON(op)
	s.wsMutex.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) reconnect() {

	s.log(LogInformational, "called")

	var err error

	if s.ShouldReconnectOnError {

		wait := time.Duration(1)

		for {
			s.log(LogInformational, "trying to reconnect to gateway")

			err = s.Open()
			if err == nil {
				s.log(LogInformational, "successfully reconnected to gateway")

				// I'm not sure if this is actually needed.
				// if the gw reconnect works properly, voice should stay alive
				// However, there seems to be cases where something "weird"
				// happens.  So we're doing this for now just to improve
				// stability in those edge cases.
				for _, v := range s.VoiceConnections {

					s.log(LogInformational, "reconnecting voice connection to guild %s", v.GuildID)
					go v.reconnect()

					// This is here just to prevent violently spamming the
					// voice reconnects
					time.Sleep(1 * time.Second)

				}
				return
			}

			s.log(LogError, "error reconnecting to gateway, %s", err)

			<-time.After(wait * time.Second)
			wait *= 2
			if wait > 600 {
				wait = 600
			}
		}
	}
}

// Close closes a websocket and stops all listening/heartbeat goroutines.
// TODO: Add support for Voice WS/UDP connections
func (s *Session) Close() (err error) {

	s.log(LogInformational, "called")
	s.Lock()

	s.DataReady = false

	if s.listening != nil {
		s.log(LogInformational, "closing listening channel")
		close(s.listening)
		s.listening = nil
	}

	// TODO: Close all active Voice Connections too
	// this should force stop any reconnecting voice channels too

	if s.wsConn != nil {

		s.log(LogInformational, "sending close frame")
		// To cleanly close a connection, a client should send a close
		// frame and wait for the server to close the connection.
		err := s.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			s.log(LogError, "error closing websocket, %s", err)
		}

		// TODO: Wait for Discord to actually close the connection.
		time.Sleep(1 * time.Second)

		s.log(LogInformational, "closing gateway websocket")
		err = s.wsConn.Close()
		if err != nil {
			s.log(LogError, "error closing websocket, %s", err)
		}

		s.wsConn = nil
	}

	s.Unlock()

	s.log(LogInformational, "emit disconnect event")
	s.handle(&Disconnect{})

	return
}
