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
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"github.com/monaco-io/request"

	"gomod.garykim.dev/nc-talk/constants"
	"gomod.garykim.dev/nc-talk/ocs"
	"gomod.garykim.dev/nc-talk/user"
)

// TalkRoom represents a room in Nextcloud Talk
type TalkRoom struct {
	User  *user.TalkUser
	Token string
}

// SendMessage sends a message in the Talk room
func (t *TalkRoom) SendMessage(msg string) (*ocs.TalkRoomMessageData, error) {
	url := t.User.NextcloudURL + constants.BaseEndpoint + "/chat/" + t.Token
	requestParams := map[string]string{
		"message": msg,
	}

	client := t.User.RequestClient(request.Client{
		URL:    url,
		Method: "POST",
		Params: requestParams,
	})
	res, err := client.Do()
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != 201 {
		return nil, errors.New("unexpected return code")
	}
	var msgInfo struct {
		OCS ocs.TalkRoomSentResponse `json:"ocs"`
	}
	err = json.Unmarshal(res.Data, &msgInfo)
	return &msgInfo.OCS.TalkRoomMessage, err
}

// ReceiveMessages starts watching for new messages
func (t *TalkRoom) ReceiveMessages(ctx context.Context) (chan ocs.TalkRoomMessageData, error) {
	c := make(chan ocs.TalkRoomMessageData)
	url := t.User.NextcloudURL + constants.BaseEndpoint + "/chat/" + t.Token
	requestParam := map[string]string{
		"lookIntoFuture":   "1",
		"includeLastKnown": "0",
	}
	lastKnown := ""
	client := t.User.RequestClient(request.Client{
		URL:     url,
		Params:  requestParam,
		Timeout: time.Second * 60,
	})
	res, err := client.Resp()
	if err != nil {
		return nil, err
	}
	lastKnown = res.Header.Get("X-Chat-Last-Given")
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
			if res.StatusCode == 200 {
				lastKnown = res.Header.Get("X-Chat-Last-Given")
				var message struct {
					OCS ocs.TalkRoomMessage `json:"ocs"`
				}
				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					continue
				}
				err = json.Unmarshal(data, &message)
				if err != nil {
					continue
				}
				for _, msg := range message.OCS.TalkRoomMessage {
					c <- msg
				}
			}
		}
	}()
	return c, nil
}

// TestConnection tests the connection with the Nextcloud Talk instance and returns an error if it could not connect
func (t *TalkRoom) TestConnection() error {
	url := t.User.NextcloudURL + constants.BaseEndpoint + "/chat/" + t.Token
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
	case 200:
		return nil
	case 304:
		return nil
	case 404:
		return errors.New("room could not be found")
	case 412:
		return errors.New("room is in lobby mode but user is not a moderator")
	}
	return errors.New("unknown return code")
}
