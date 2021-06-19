// Copyright (c) 2020 Gary Kim <gary@garykim.dev>, All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package room

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/monaco-io/request"

	"gomod.garykim.dev/nc-talk/constants"
	"gomod.garykim.dev/nc-talk/ocs"
	"gomod.garykim.dev/nc-talk/user"
)

var (
	// ErrEmptyToken is returned when the room token is empty
	ErrEmptyToken = errors.New("given an empty token")
	// ErrRoomNotFound is returned when a room with the given token could not be found
	ErrRoomNotFound = errors.New("room could not be found")
	// ErrUnauthorized is returned when the room could not be accessed due to being unauthorized
	ErrUnauthorized = errors.New("unauthorized error when accessing room")
	// ErrNotModeratorInLobby is returned when the room is in lobby mode but the user is not a moderator
	ErrNotModeratorInLobby = errors.New("room is in lobby mode but user is not a moderator")
	// ErrUnexpectedReturnCode is returned when the server did not respond with an expected return code
	ErrUnexpectedReturnCode = errors.New("unexpected return code")
	// ErrTooManyRequests is returned if the server returns a 429
	ErrTooManyRequests = errors.New("too many requests")
	// ErrLackingCapabilities is returned if the server lacks the required capability for the given function
	ErrLackingCapabilities = errors.New("lacking required capabilities")
	// ErrForbidden is returned if the user is forbidden from accessing the requested resource
	ErrForbidden = errors.New("request forbidden")
	// ErrUnexpectedResponse is returned if the response from the Nextcloud Talk server is not formatted as expected
	ErrUnexpectedResponse = errors.New("unexpected response")
)

// TalkRoom represents a room in Nextcloud Talk
type TalkRoom struct {
	User  *user.TalkUser
	Token string
}

// Message represents a message to be sent
type Message struct {
	Message          string
	ActorDisplayName string
	ReplyTo          int
}

func (t *Message) toParameters() map[string]string {
	return map[string]string{
		"message":          t.Message,
		"actorDisplayName": t.ActorDisplayName,
		"replyTo":          strconv.Itoa(t.ReplyTo),
	}
}

// NewTalkRoom returns a new TalkRoom instance
// Token should be the Nextcloud Room Token (e.g. "d6zoa2zs" if the room URL is https://cloud.mydomain.me/call/d6zoa2zs)
func NewTalkRoom(tuser *user.TalkUser, token string) (*TalkRoom, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}
	if tuser == nil {
		return nil, user.ErrUserIsNil
	}
	return &TalkRoom{
		User:  tuser,
		Token: token,
	}, nil
}

// SendMessage sends a string message in the Talk room
func (t *TalkRoom) SendMessage(msg string) (*ocs.TalkRoomMessageData, error) {
	return t.SendComplexMessage(&Message{Message: msg})
}

// SendComplexMessage sends a Message type message in the talk room
func (t *TalkRoom) SendComplexMessage(msg *Message) (*ocs.TalkRoomMessageData, error) {
	url := t.User.NextcloudURL + constants.BaseEndpoint + "chat/" + t.Token

	client := t.User.RequestClient(request.Client{
		URL:    url,
		Method: "POST",
		Params: msg.toParameters(),
	})
	res, err := client.Do()
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != 201 {
		return nil, ErrUnexpectedReturnCode
	}
	msgInfo, err := ocs.TalkRoomSentResponseUnmarshal(&res.Data)
	if err != nil {
		return nil, err
	}
	return &msgInfo.OCS.TalkRoomMessage, err
}

// DeleteMessage deletes the message with the given messageID on the server.
//
// Requires "delete-messages" capability on the Nextcloud Talk server
func (t *TalkRoom) DeleteMessage(messageID int) (*ocs.TalkRoomMessageData, error) {
	// Check for required capability
	capable, err := t.User.Capabilities()
	if err != nil {
		return nil, err
	}
	if !capable.DeleteMessages {
		return nil, ErrLackingCapabilities
	}

	url := t.User.NextcloudURL + constants.BaseEndpoint + "/chat/" + t.Token + "/" + strconv.Itoa(messageID)

	client := t.User.RequestClient(request.Client{
		URL:    url,
		Method: "DELETE",
	})
	res, err := client.Do()
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusAccepted {
		return nil, ErrUnexpectedReturnCode
	}
	msgInfo, err := ocs.TalkRoomSentResponseUnmarshal(&res.Data)
	if err != nil {
		return nil, err
	}
	return &msgInfo.OCS.TalkRoomMessage, nil
}

// ReceiveMessages starts watching for new messages
func (t *TalkRoom) ReceiveMessages(ctx context.Context) (chan ocs.TalkRoomMessageData, error) {
	c := make(chan ocs.TalkRoomMessageData)
	err := t.TestConnection()
	if err != nil {
		return nil, err
	}
	url := t.User.NextcloudURL + constants.BaseEndpoint + "chat/" + t.Token
	requestParam := map[string]string{
		"lookIntoFuture":   "1",
		"includeLastKnown": "0",
	}
	lastKnown := ""
	res, err := t.User.GetRooms()
	if err != nil {
		return nil, err
	}
	for _, r := range *res {
		if r.Token == t.Token {
			lastKnown = strconv.Itoa(r.LastReadMessage)
			break
		}
	}
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			if lastKnown != "" {
				requestParam["lastKnownMessageId"] = lastKnown
			}
			client := t.User.RequestClient(request.Client{
				URL:     url,
				Params:  requestParam,
				Timeout: time.Second * 60,
			})

			res, err := client.Resp()
			if err != nil {
				continue
			}

			// If it seems that we no longer have access to the chat for one reason or another, stop the goroutine and set error in the next return.
			if res.StatusCode == http.StatusNotFound {
				_ = res.Body.Close()
				c <- ocs.TalkRoomMessageData{Error: ErrRoomNotFound}
				return
			}
			if res.StatusCode == http.StatusUnauthorized {
				_ = res.Body.Close()
				c <- ocs.TalkRoomMessageData{Error: ErrUnauthorized}
				return
			}
			if res.StatusCode == http.StatusTooManyRequests {
				_ = res.Body.Close()
				c <- ocs.TalkRoomMessageData{Error: ErrTooManyRequests}
				return
			}
			if res.StatusCode == http.StatusForbidden {
				_ = res.Body.Close()
				c <- ocs.TalkRoomMessageData{Error: ErrForbidden}
				return
			}

			if res.StatusCode == http.StatusOK {
				lastKnown = res.Header.Get("X-Chat-Last-Given")
				data, err := ioutil.ReadAll(res.Body)
				_ = res.Body.Close()
				if err != nil {
					continue
				}
				message, err := ocs.TalkRoomMessageDataUnmarshal(&data)
				if err != nil {
					continue
				}
				for _, msg := range message.OCS.TalkRoomMessage {
					c <- msg
				}
				continue
			}
			_ = res.Body.Close()
		}
	}()
	return c, nil
}

// TestConnection tests the connection with the Nextcloud Talk instance and returns an error if it could not connect
func (t *TalkRoom) TestConnection() error {
	if t.Token == "" {
		return ErrEmptyToken
	}
	url := t.User.NextcloudURL + constants.BaseEndpoint + "chat/" + t.Token
	requestParam := map[string]string{
		"lookIntoFuture":   "0",
		"includeLastKnown": "0",
	}
	client := t.User.RequestClient(request.Client{
		URL:     url,
		Params:  requestParam,
		Timeout: time.Second * 30,
	})

	res, err := client.Do()
	if err != nil {
		return err
	}
	switch res.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotModified:
		return nil
	case http.StatusNotFound:
		return ErrRoomNotFound
	case http.StatusPreconditionFailed:
		return ErrNotModeratorInLobby
	}
	return ErrUnexpectedReturnCode
}
