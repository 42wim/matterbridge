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

package user

import (
	"encoding/json"
	"strings"

	"github.com/monaco-io/request"

	"gomod.garykim.dev/nc-talk/ocs"
)

const (
	ocsCapabilitiesEndpoint = "/ocs/v2.php/cloud/capabilities"
)

// TalkUser represents a user of Nextcloud Talk
type TalkUser struct {
	User         string
	Pass         string
	NextcloudURL string
	capabilities *Capabilities
}

// Capabilities describes the capabilities that the Nextcloud Talk instance is capable of. Visit https://nextcloud-talk.readthedocs.io/en/latest/capabilities/ for more info.
type Capabilities struct {
	AttachmentsFolder      string `ocscapability:"config => attachments => folder"`
	ChatMaxLength          int
	Audio                  bool `ocscapability:"audio"`
	Video                  bool `ocscapability:"video"`
	Chat                   bool `ocscapability:"chat"`
	GuestSignaling         bool `ocscapability:"guest-signaling"`
	EmptyGroupRoom         bool `ocscapability:"empty-group-room"`
	GuestDisplayNames      bool `ocscapability:"guest-display-names"`
	MultiRoomUsers         bool `ocscapability:"multi-room-users"`
	ChatV2                 bool `ocscapability:"chat-v2"`
	Favorites              bool `ocscapability:"favorites"`
	LastRoomActivity       bool `ocscapability:"last-room-activity"`
	NoPing                 bool `ocscapability:"no-ping"`
	SystemMessages         bool `ocscapability:"system-messages"`
	MentionFlag            bool `ocscapability:"mention-flag"`
	InCallFlags            bool `ocscapability:"in-call-flags"`
	InviteByMail           bool `ocscapability:"invite-by-mail"`
	NotificationLevels     bool `ocscapability:"notification-levels"`
	InviteGroupsAndMails   bool `ocscapability:"invite-groups-and-mails"`
	LockedOneToOneRooms    bool `ocscapability:"locked-one-to-one-rooms"`
	ReadOnlyRooms          bool `ocscapability:"read-only-rooms"`
	ChatReadMarker         bool `ocscapability:"chat-read-marker"`
	WebinaryLobby          bool `ocscapability:"webinary-lobby"`
	StartCallFlag          bool `ocscapability:"start-call-flag"`
	ChatReplies            bool `ocscapability:"chat-replies"`
	CirclesSupport         bool `ocscapability:"circles-support"`
	AttachmentsAllowed     bool `ocscapability:"config => attachments => allowed"`
	ConversationsCanCreate bool `ocscapability:"config => conversations => can-create"`
	ForceMute              bool `ocscapability:"force-mute"`
	ConversationV2         bool `ocscapability:"conversation-v2"`
	ChatReferenceID        bool `ocscapability:"chat-reference-id"`
}

// RequestClient returns a monaco-io that is preconfigured to make OCS API calls
func (t *TalkUser) RequestClient(client request.Client) *request.Client {
	if client.Header == nil {
		client.Header = make(map[string]string)
	}
	if client.Header["OCS-APIRequest"] == "" {
		client.Header["OCS-APIRequest"] = "true"
	}
	if client.Header["Accept"] == "" {
		client.Header["Accept"] = "application/json"
	}
	client.BasicAuth = request.BasicAuth{
		Username: t.User,
		Password: t.Pass,
	}

	// Set Nextcloud URL if there is no host
	if !strings.HasPrefix(client.URL, t.NextcloudURL) {
		client.URL = t.NextcloudURL + "/" + client.URL
	}

	return &client
}

// Capabilities returns an instance of Capabilities that describes what the Nextcloud Talk instance supports
func (t *TalkUser) Capabilities() (*Capabilities, error) {
	if t.capabilities != nil {
		return t.capabilities, nil
	}

	client := t.RequestClient(request.Client{
		URL: ocsCapabilitiesEndpoint,
	})
	res, err := client.Do()
	if err != nil {
		return nil, err
	}

	capabilitiesRequest := &struct {
		Ocs ocs.Capabilities `json:"ocs"`
	}{}

	err = json.Unmarshal(res.Data, capabilitiesRequest)
	if err != nil {
		return nil, err
	}

	sc := capabilitiesRequest.Ocs.Data.Capabilities.SpreedCapabilities

	tr := &Capabilities{
		Audio:                  sliceContains(sc.Features, "audio"),
		Video:                  sliceContains(sc.Features, "video"),
		Chat:                   sliceContains(sc.Features, "chat"),
		GuestSignaling:         sliceContains(sc.Features, "guest-signaling"),
		EmptyGroupRoom:         sliceContains(sc.Features, "empty-group-room"),
		GuestDisplayNames:      sliceContains(sc.Features, "guest-display-names"),
		MultiRoomUsers:         sliceContains(sc.Features, "multi-room-users"),
		ChatV2:                 sliceContains(sc.Features, "chat-v2"),
		Favorites:              sliceContains(sc.Features, "favorites"),
		LastRoomActivity:       sliceContains(sc.Features, "last-room-activity"),
		NoPing:                 sliceContains(sc.Features, "no-ping"),
		SystemMessages:         sliceContains(sc.Features, "system-messages"),
		MentionFlag:            sliceContains(sc.Features, "mention-flag"),
		InCallFlags:            sliceContains(sc.Features, "in-call-flags"),
		InviteByMail:           sliceContains(sc.Features, "invite-by-mail"),
		NotificationLevels:     sliceContains(sc.Features, "notification-levels"),
		InviteGroupsAndMails:   sliceContains(sc.Features, "invite-groups-and-mails"),
		LockedOneToOneRooms:    sliceContains(sc.Features, "locked-one-to-one-rooms"),
		ReadOnlyRooms:          sliceContains(sc.Features, "read-only-rooms"),
		ChatReadMarker:         sliceContains(sc.Features, "chat-read-marker"),
		WebinaryLobby:          sliceContains(sc.Features, "webinary-lobby"),
		StartCallFlag:          sliceContains(sc.Features, "start-call-flag"),
		ChatReplies:            sliceContains(sc.Features, "chat-replies"),
		CirclesSupport:         sliceContains(sc.Features, "circles-support"),
		AttachmentsAllowed:     sc.Config.Attachments.Allowed,
		AttachmentsFolder:      sc.Config.Attachments.Folder,
		ConversationsCanCreate: sc.Config.Conversations.CanCreate,
		ForceMute:              sliceContains(sc.Features, "force-mute"),
		ConversationV2:         sliceContains(sc.Features, "conversation-v2"),
		ChatReferenceID:        sliceContains(sc.Features, "chat-reference-id"),
		ChatMaxLength:          sc.Config.Chat.MaxLength,
	}

	t.capabilities = tr
	return tr, nil
}

// sliceContains does the slice contain the string
func sliceContains(s []string, search string) bool {
	for _, n := range s {
		if n == search {
			return true
		}
	}
	return false
}
