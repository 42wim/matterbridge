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

// MessageType describes what kind of message a returned Nextcloud Talk message is
type MessageType string

const (
	// MessageComment is a Nextcloud Talk message that is a comment
	MessageComment MessageType = "comment"

	// MessageSystem is a Nextcloud Talk message that is a system
	MessageSystem MessageType = "system"

	// MessageCommand is a Nextcloud Talk message that is a command
	MessageCommand MessageType = "command"
)

// TalkRoomMessageData describes the data part of a ocs response for a Talk room message
type TalkRoomMessageData struct {
	Message          string      `json:"message"`
	ID               int         `json:"id"`
	ActorID          string      `json:"actorId"`
	ActorDisplayName string      `json:"actorDisplayName"`
	SystemMessage    string      `json:"systemMessage"`
	Timestamp        int         `json:"timestamp"`
	MessageType      MessageType `json:"messageType"`
}

// TalkRoomMessage describes an ocs response for a Talk room message
type TalkRoomMessage struct {
	ocs
	TalkRoomMessage []TalkRoomMessageData `json:"data"`
}

// TalkRoomSentResponse describes an ocs response for what is returned when a message is sent
type TalkRoomSentResponse struct {
	ocs
	TalkRoomMessage TalkRoomMessageData `json:"data"`
}
