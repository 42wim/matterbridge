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

package ocs

import (
	"encoding/json"
	"strings"
)

// MessageType describes what kind of message a returned Nextcloud Talk message is
type MessageType string

// ActorType describes what kind of actor a returned Nextcloud Talk message is from
type ActorType string

const (
	// MessageComment is a Nextcloud Talk message that is a comment
	MessageComment MessageType = "comment"

	// MessageSystem is a Nextcloud Talk message that is a system
	MessageSystem MessageType = "system"

	// MessageCommand is a Nextcloud Talk message that is a command
	MessageCommand MessageType = "command"

	// MessageDelete is a Nextcloud Talk message indicating a message that was deleted
	//
	// If a message has been deleted, a message of MessageType MessageSystem is
	// sent through the channel for which the parent message's MessageType is MessageDelete.
	// So, in order to check if a new message is a message deletion request, a check
	// like this can be used:
	// msg.MessageType == ocs.MessageSystem && msg.Parent != nil && msg.Parent.MessageType == ocs.MessageDelete
	MessageDelete MessageType = "comment_deleted"

	// ActorUser is a Nextcloud Talk message sent by a user
	ActorUser ActorType = "users"

	// ActorGuest is a Nextcloud Talk message sent by a guest
	ActorGuest ActorType = "guests"
)

// TalkRoomMessageData describes the data part of a ocs response for a Talk room message
//
// Error will be set if a message request ran into an error.
type TalkRoomMessageData struct {
	Error             error                       `json:"-"`
	Message           string                      `json:"message"`
	ID                int                         `json:"id"`
	ActorType         ActorType                   `json:"actorType"`
	ActorID           string                      `json:"actorId"`
	ActorDisplayName  string                      `json:"actorDisplayName"`
	SystemMessage     string                      `json:"systemMessage"`
	Timestamp         int                         `json:"timestamp"`
	MessageType       MessageType                 `json:"messageType"`
	Deleted           bool                        `json:"deleted"`
	Parent            *TalkRoomMessageData        `json:"parent"`
	MessageParameters map[string]RichObjectString `json:"-"`
}

// talkRoomMessageParameters is used to unmarshal only MessageParameters
type talkRoomMessageParameters struct {
	MessageParameters map[string]RichObjectString `json:"messageParameters"`
}

// PlainMessage returns the message string with placeholders replaced
//
// * User and group placeholders will be replaced with the name of the user or group respectively.
//
// * File placeholders will be replaced with the name of the file.
func (m *TalkRoomMessageData) PlainMessage() string {
	tr := m.Message
	for key, value := range m.MessageParameters {
		tr = strings.ReplaceAll(tr, "{"+key+"}", value.Name)
	}
	return tr
}

// DisplayName returns the display name for the sender of the message (" (Guest)" is appended if sent by a guest user)
func (m *TalkRoomMessageData) DisplayName() string {
	if m.ActorType == ActorGuest {
		if m.ActorDisplayName == "" {
			return "Guest"
		}
		return m.ActorDisplayName + " (Guest)"
	}
	return m.ActorDisplayName
}

// TalkRoomMessage describes an ocs response for a Talk room message
type TalkRoomMessage struct {
	OCS talkRoomMessage `json:"ocs"`
}

type talkRoomMessage struct {
	ocs
	TalkRoomMessage []TalkRoomMessageData `json:"data"`
}

// TalkRoomMessageDataUnmarshal unmarshals given ocs request data and returns a TalkRoomMessageData
func TalkRoomMessageDataUnmarshal(data *[]byte) (*TalkRoomMessage, error) {
	message := &TalkRoomMessage{}
	err := json.Unmarshal(*data, message)
	if err != nil {
		return nil, err
	}

	// Get RCS
	var rcs struct {
		OCS struct {
			ocs
			TalkRoomMessage []talkRoomMessageParameters `json:"data"`
		} `json:"ocs"`
	}
	err = json.Unmarshal(*data, &rcs)
	// There is no RCS data
	if err != nil {
		for i := range message.OCS.TalkRoomMessage {
			message.OCS.TalkRoomMessage[i].MessageParameters = map[string]RichObjectString{}
		}
		return message, nil
	}

	// There is RCS data
	for i := range message.OCS.TalkRoomMessage {
		message.OCS.TalkRoomMessage[i].MessageParameters = rcs.OCS.TalkRoomMessage[i].MessageParameters
	}
	return message, nil
}

// TalkRoomSentResponse describes an ocs response for what is returned when a message is sent
type TalkRoomSentResponse struct {
	OCS talkRoomSentResponse `json:"ocs"`
}

type talkRoomSentResponse struct {
	ocs
	TalkRoomMessage TalkRoomMessageData `json:"data"`
}

// TalkRoomSentResponseUnmarshal unmarshals given ocs request data and returns a TalkRoomMessageData
func TalkRoomSentResponseUnmarshal(data *[]byte) (*TalkRoomSentResponse, error) {
	message := &TalkRoomSentResponse{}
	err := json.Unmarshal(*data, message)
	if err != nil {
		return nil, err
	}

	// Get RCS
	var rcs struct {
		OCS struct {
			ocs
			TalkRoomMessage talkRoomMessageParameters `json:"data"`
		} `json:"ocs"`
	}
	err = json.Unmarshal(*data, &rcs)
	// There is no RCS data
	if err != nil {
		message.OCS.TalkRoomMessage.MessageParameters = map[string]RichObjectString{}
		return message, nil
	}

	// There is RCS data
	message.OCS.TalkRoomMessage.MessageParameters = rcs.OCS.TalkRoomMessage.MessageParameters
	return message, nil
}
