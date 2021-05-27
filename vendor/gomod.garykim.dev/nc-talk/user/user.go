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
	"crypto/tls"
	"encoding/json"
	"errors"
	"strings"

	"github.com/monaco-io/request"

	"gomod.garykim.dev/nc-talk/constants"
	"gomod.garykim.dev/nc-talk/ocs"
)

const (
	ocsCapabilitiesEndpoint = "/ocs/v2.php/cloud/capabilities"
	ocsRoomsv2Endpoint      = "/ocs/v2.php/apps/spreed/api/v2/room"
	ocsRoomsv4Endpoint      = "/ocs/v2.php/apps/spreed/api/v4/room"
)

var (
	// ErrUserIsNil is returned when a function is called with an nil user.
	ErrUserIsNil = errors.New("user is nil")

	// ErrCannotDownloadFile is returned when a function cannot download the requested file
	ErrCannotDownloadFile = errors.New("cannot download file")
)

// TalkUser represents a user of Nextcloud Talk
type TalkUser struct {
	User         string
	Pass         string
	NextcloudURL string
	Config       *TalkUserConfig
	capabilities *Capabilities
}

// TalkUserConfig is configuration options for TalkUsers
type TalkUserConfig struct {
	TLSConfig *tls.Config
}

// Capabilities describes the capabilities that the Nextcloud Talk instance is capable of. Visit https://nextcloud-talk.readthedocs.io/en/latest/capabilities/ for more info.
type Capabilities struct {
	AttachmentsFolder      string `ocscapability:"config => attachments => folder"`
	Audio                  bool   `ocscapability:"audio"`
	Video                  bool   `ocscapability:"video"`
	Chat                   bool   `ocscapability:"chat"`
	GuestSignaling         bool   `ocscapability:"guest-signaling"`
	EmptyGroupRoom         bool   `ocscapability:"empty-group-room"`
	GuestDisplayNames      bool   `ocscapability:"guest-display-names"`
	MultiRoomUsers         bool   `ocscapability:"multi-room-users"`
	ChatV2                 bool   `ocscapability:"chat-v2"`
	Favorites              bool   `ocscapability:"favorites"`
	LastRoomActivity       bool   `ocscapability:"last-room-activity"`
	NoPing                 bool   `ocscapability:"no-ping"`
	SystemMessages         bool   `ocscapability:"system-messages"`
	MentionFlag            bool   `ocscapability:"mention-flag"`
	InCallFlags            bool   `ocscapability:"in-call-flags"`
	InviteByMail           bool   `ocscapability:"invite-by-mail"`
	NotificationLevels     bool   `ocscapability:"notification-levels"`
	InviteGroupsAndMails   bool   `ocscapability:"invite-groups-and-mails"`
	LockedOneToOneRooms    bool   `ocscapability:"locked-one-to-one-rooms"`
	ReadOnlyRooms          bool   `ocscapability:"read-only-rooms"`
	ChatReadMarker         bool   `ocscapability:"chat-read-marker"`
	WebinaryLobby          bool   `ocscapability:"webinary-lobby"`
	StartCallFlag          bool   `ocscapability:"start-call-flag"`
	ChatReplies            bool   `ocscapability:"chat-replies"`
	CirclesSupport         bool   `ocscapability:"circles-support"`
	AttachmentsAllowed     bool   `ocscapability:"config => attachments => allowed"`
	ConversationsCanCreate bool   `ocscapability:"config => conversations => can-create"`
	ForceMute              bool   `ocscapability:"force-mute"`
	ConversationV2         bool   `ocscapability:"conversation-v2"`
	ChatReferenceID        bool   `ocscapability:"chat-reference-id"`
	ConversationV3         bool   `ocscapability:"conversation-v3"`
	ConversationV4         bool   `ocscapability:"conversation-v4"`
	SIPSupport             bool   `ocscapability:"sip-support"`
	ChatReadStatus         bool   `ocscapability:"chat-read-status"`
	ListableRooms          bool   `ocscapability:"listable-rooms"`
	PhonebookSearch        bool   `ocscapability:"phonebook-search"`
	RaiseHand              bool   `ocscapability:"raise-hand"`
	RoomDescription        bool   `ocscapability:"room-description"`
	DeleteMessages         bool   `ocscapability:"delete-messages"`
	RichObjectSharing      bool   `ocscapability:"rich-object-sharing"`
	ConversationCallFlags  bool   `ocscapability:"conversation-call-flags"`
	GeoLocationSharing     bool   `ocscapability:"geo-location-sharing"`
	ReadPrivacyConfig      bool   `ocscapability:"config => chat => read-privacy"`
	SignalingV3            bool   `ocscapability:"signaling-v3"`
	TempUserAvatarAPI      bool   `ocscapability:"temp-user-avatar-api"`
	MaxGifSizeConfig       int    `ocscapability:"config => previews => max-gif-size"`
	ChatMaxLength          int    `ocscapability:"config => chat => max-length"`
}

// RoomInfo contains information about a room
type RoomInfo struct {
	Token                 string                   `json:"token"`
	Name                  string                   `json:"name"`
	DisplayName           string                   `json:"displayName"`
	SessionID             string                   `json:"sessionId"`
	ObjectType            string                   `json:"objectType"`
	ObjectID              string                   `json:"objectId"`
	Type                  int                      `json:"type"`
	ParticipantType       int                      `json:"participantType"`
	ParticipantFlags      int                      `json:"participantFlags"`
	ReadOnly              int                      `json:"readOnly"`
	LastPing              int                      `json:"lastPing"`
	LastActivity          int                      `json:"lastActivity"`
	NotificationLevel     int                      `json:"notificationLevel"`
	LobbyState            int                      `json:"lobbyState"`
	LobbyTimer            int                      `json:"lobbyTimer"`
	UnreadMessages        int                      `json:"unreadMessages"`
	LastReadMessage       int                      `json:"lastReadMessage"`
	HasPassword           bool                     `json:"hasPassword"`
	HasCall               bool                     `json:"hasCall"`
	CanStartCall          bool                     `json:"canStartCall"`
	CanDeleteConversation bool                     `json:"canDeleteConversation"`
	CanLeaveConversation  bool                     `json:"canLeaveConversation"`
	IsFavorite            bool                     `json:"isFavorite"`
	UnreadMention         bool                     `json:"unreadMention"`
	LastMessage           *ocs.TalkRoomMessageData `json:"lastMessage"`
}

// NewUser returns a TalkUser instance
// The url should be the full URL of the Nextcloud instance (e.g. https://cloud.mydomain.me)
func NewUser(url string, username string, password string, config *TalkUserConfig) (*TalkUser, error) {
	return &TalkUser{
		NextcloudURL: strings.TrimSuffix(url, "/"),
		User:         username,
		Pass:         password,
		Config:       config,
	}, nil
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
		if strings.HasPrefix(client.URL, "/") {
			client.URL = t.NextcloudURL + client.URL
		} else {
			client.URL = t.NextcloudURL + "/" + client.URL
		}
	}

	// Set TLS Config
	if t.Config != nil {
		client.TLSConfig = t.Config.TLSConfig
	}

	return &client
}

// GetRooms returns a list of all rooms the user is in
func (t *TalkUser) GetRooms() (*[]RoomInfo, error) {
	endpoint := ocsRoomsv2Endpoint
	capabilities, err := t.Capabilities()
	if err != nil {
		return nil, err
	}
	if capabilities.ConversationV4 {
		endpoint = ocsRoomsv4Endpoint
	}

	client := t.RequestClient(request.Client{
		URL: endpoint,
	})
	res, err := client.Do()
	if err != nil {
		return nil, err
	}

	var roomsRequest struct {
		OCS struct {
			Data []RoomInfo `json:"data"`
		} `json:"ocs"`
	}

	err = json.Unmarshal(res.Data, &roomsRequest)
	if err != nil {
		return nil, err
	}

	return &roomsRequest.OCS.Data, nil
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
		ConversationV3:         sliceContains(sc.Features, "conversation-v3"),
		ConversationV4:         sliceContains(sc.Features, "conversation-v4"),
		SIPSupport:             sliceContains(sc.Features, "sip-support"),
		ChatReadStatus:         sliceContains(sc.Features, "chat-read-status"),
		ListableRooms:          sliceContains(sc.Features, "listable-rooms"),
		PhonebookSearch:        sliceContains(sc.Features, "phonebook-search"),
		RaiseHand:              sliceContains(sc.Features, "raise-hand"),
		RoomDescription:        sliceContains(sc.Features, "room-description"),
		ReadPrivacyConfig:      sc.Config.Chat.ReadPrivacy != 0,
		MaxGifSizeConfig:       sc.Config.Previews.MaxGifSize,
		DeleteMessages:         sliceContains(sc.Features, "delete-messages"),
		RichObjectSharing:      sliceContains(sc.Features, "rich-object-sharing"),
		ConversationCallFlags:  sliceContains(sc.Features, "conversation-call-flags"),
		GeoLocationSharing:     sliceContains(sc.Features, "geo-location-sharing"),
		SignalingV3:            sliceContains(sc.Features, "signaling-v3"),
		TempUserAvatarAPI:      sliceContains(sc.Features, "temp-user-avatar-api"),
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

// DownloadFile downloads the file at the given path
//
// Meant to be used with rich object string's path.
func (t *TalkUser) DownloadFile(path string) (data *[]byte, err error) {
	url := t.NextcloudURL + constants.RemoteDavEndpoint(t.User, "files") + path
	c := t.RequestClient(request.Client{
		URL: url,
	})
	res, err := c.Do()
	if err != nil {
		return
	}
	if res.StatusCode() != 200 {
		err = ErrCannotDownloadFile
		return
	}
	data = &res.Data
	return
}
