package mautrix

import (
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// RespWhoami is the JSON response for https://matrix.org/docs/spec/client_server/r0.6.1#get-matrix-client-r0-account-whoami
type RespWhoami struct {
	UserID id.UserID `json:"user_id"`
	// N.B. This field is not in the spec yet, it's expected to land in r0.6.2 or r0.7.0
	DeviceID id.DeviceID `json:"device_id"`
}

// RespCreateFilter is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-user-userid-filter
type RespCreateFilter struct {
	FilterID string `json:"filter_id"`
}

// RespVersions is the JSON response for http://matrix.org/docs/spec/client_server/r0.6.1.html#get-matrix-client-versions
type RespVersions struct {
	Versions         []string        `json:"versions"`
	UnstableFeatures map[string]bool `json:"unstable_features"`
}

// RespJoinRoom is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-join
type RespJoinRoom struct {
	RoomID id.RoomID `json:"room_id"`
}

// RespLeaveRoom is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-leave
type RespLeaveRoom struct{}

// RespForgetRoom is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-forget
type RespForgetRoom struct{}

// RespInviteUser is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-invite
type RespInviteUser struct{}

// RespKickUser is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-kick
type RespKickUser struct{}

// RespBanUser is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-ban
type RespBanUser struct{}

// RespUnbanUser is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-unban
type RespUnbanUser struct{}

// RespTyping is the JSON response for https://matrix.org/docs/spec/client_server/r0.2.0.html#put-matrix-client-r0-rooms-roomid-typing-userid
type RespTyping struct{}

// RespPresence is the JSON response for https://matrix.org/docs/spec/client_server/r0.6.1#get-matrix-client-r0-presence-userid-status
type RespPresence struct {
	Presence        event.Presence `json:"presence"`
	LastActiveAgo   int            `json:"last_active_ago"`
	StatusMsg       string         `json:"status_msg"`
	CurrentlyActive bool           `json:"currently_active"`
}

// RespJoinedRooms is the JSON response for https://matrix.org/docs/spec/client_server/r0.4.0.html#get-matrix-client-r0-joined-rooms
type RespJoinedRooms struct {
	JoinedRooms []id.RoomID `json:"joined_rooms"`
}

// RespJoinedMembers is the JSON response for https://matrix.org/docs/spec/client_server/r0.4.0.html#get-matrix-client-r0-joined-rooms
type RespJoinedMembers struct {
	Joined map[id.UserID]struct {
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
	} `json:"joined"`
}

// RespMessages is the JSON response for https://matrix.org/docs/spec/client_server/r0.2.0.html#get-matrix-client-r0-rooms-roomid-messages
type RespMessages struct {
	Start string         `json:"start"`
	Chunk []*event.Event `json:"chunk"`
	State []*event.Event `json:"state"`
	End   string         `json:"end"`
}

// RespSendEvent is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#put-matrix-client-r0-rooms-roomid-send-eventtype-txnid
type RespSendEvent struct {
	EventID id.EventID `json:"event_id"`
}

// RespMediaUpload is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-media-r0-upload
type RespMediaUpload struct {
	ContentURI id.ContentURI `json:"content_uri"`
}

// RespUserInteractive is the JSON response for https://matrix.org/docs/spec/client_server/r0.2.0.html#user-interactive-authentication-api
type RespUserInteractive struct {
	Flows []struct {
		Stages []AuthType `json:"stages"`
	} `json:"flows"`
	Params    map[AuthType]interface{} `json:"params"`
	Session   string                   `json:"session"`
	Completed []string                 `json:"completed"`

	ErrCode string `json:"errcode"`
	Error   string `json:"error"`
}

// HasSingleStageFlow returns true if there exists at least 1 Flow with a single stage of stageName.
func (r RespUserInteractive) HasSingleStageFlow(stageName AuthType) bool {
	for _, f := range r.Flows {
		if len(f.Stages) == 1 && f.Stages[0] == stageName {
			return true
		}
	}
	return false
}

// RespUserDisplayName is the JSON response for https://matrix.org/docs/spec/client_server/r0.2.0.html#get-matrix-client-r0-profile-userid-displayname
type RespUserDisplayName struct {
	DisplayName string `json:"displayname"`
}

// RespRegister is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-register
type RespRegister struct {
	AccessToken  string      `json:"access_token"`
	DeviceID     id.DeviceID `json:"device_id"`
	HomeServer   string      `json:"home_server"`
	RefreshToken string      `json:"refresh_token"`
	UserID       id.UserID   `json:"user_id"`
}

type RespLoginFlows struct {
	Flows []struct {
		Type AuthType `json:"type"`
	} `json:"flows"`
}

func (rlf *RespLoginFlows) HasFlow(flowType AuthType) bool {
	for _, flow := range rlf.Flows {
		if flow.Type == flowType {
			return true
		}
	}
	return false
}

// RespLogin is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-login
type RespLogin struct {
	AccessToken string      `json:"access_token"`
	DeviceID    id.DeviceID `json:"device_id"`
	UserID      id.UserID   `json:"user_id"`
	// TODO add .well-known field here
}

// RespLogout is the JSON response for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-logout
type RespLogout struct{}

// RespCreateRoom is the JSON response for https://matrix.org/docs/spec/client_server/r0.6.0.html#post-matrix-client-r0-createroom
type RespCreateRoom struct {
	RoomID id.RoomID `json:"room_id"`
}

type RespMembers struct {
	Chunk []*event.Event `json:"chunk"`
}

type LazyLoadSummary struct {
	Heroes             []id.UserID `json:"m.heroes,omitempty"`
	JoinedMemberCount  *int        `json:"m.joined_member_count,omitempty"`
	InvitedMemberCount *int        `json:"m.invited_member_count,omitempty"`
}

// RespSync is the JSON response for http://matrix.org/docs/spec/client_server/r0.6.0.html#get-matrix-client-r0-sync
type RespSync struct {
	NextBatch string `json:"next_batch"`

	AccountData struct {
		Events []*event.Event `json:"events"`
	} `json:"account_data"`
	Presence struct {
		Events []*event.Event `json:"events"`
	} `json:"presence"`
	ToDevice struct {
		Events []*event.Event `json:"events"`
	} `json:"to_device"`

	DeviceLists struct {
		Changed []id.UserID `json:"changed"`
		Left    []id.UserID `json:"left"`
	} `json:"device_lists"`
	DeviceOneTimeKeysCount OneTimeKeysCount `json:"device_one_time_keys_count"`

	Rooms struct {
		Leave  map[id.RoomID]SyncLeftRoom    `json:"leave"`
		Join   map[id.RoomID]SyncJoinedRoom  `json:"join"`
		Invite map[id.RoomID]SyncInvitedRoom `json:"invite"`
	} `json:"rooms"`
}

type OneTimeKeysCount struct {
	Curve25519       int `json:"curve25519"`
	SignedCurve25519 int `json:"signed_curve25519"`
}

type SyncLeftRoom struct {
	Summary LazyLoadSummary `json:"summary"`
	State   struct {
		Events []*event.Event `json:"events"`
	} `json:"state"`
	Timeline struct {
		Events    []*event.Event `json:"events"`
		Limited   bool           `json:"limited"`
		PrevBatch string         `json:"prev_batch"`
	} `json:"timeline"`
}

type SyncJoinedRoom struct {
	Summary LazyLoadSummary `json:"summary"`
	State   struct {
		Events []*event.Event `json:"events"`
	} `json:"state"`
	Timeline struct {
		Events    []*event.Event `json:"events"`
		Limited   bool           `json:"limited"`
		PrevBatch string         `json:"prev_batch"`
	} `json:"timeline"`
	Ephemeral struct {
		Events []*event.Event `json:"events"`
	} `json:"ephemeral"`
	AccountData struct {
		Events []*event.Event `json:"events"`
	} `json:"account_data"`
}

type SyncInvitedRoom struct {
	Summary LazyLoadSummary `json:"summary"`
	State   struct {
		Events []*event.Event `json:"events"`
	} `json:"invite_state"`
}

type RespTurnServer struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	TTL      int      `json:"ttl"`
	URIs     []string `json:"uris"`
}

type RespAliasCreate struct{}
type RespAliasDelete struct{}
type RespAliasResolve struct {
	RoomID  id.RoomID `json:"room_id"`
	Servers []string  `json:"servers"`
}

type RespUploadKeys struct {
	OneTimeKeyCounts OneTimeKeysCount `json:"one_time_key_counts"`
}

type RespQueryKeys struct {
	Failures        map[string]interface{}                   `json:"failures"`
	DeviceKeys      map[id.UserID]map[id.DeviceID]DeviceKeys `json:"device_keys"`
	MasterKeys      map[id.UserID]CrossSigningKeys           `json:"master_keys"`
	SelfSigningKeys map[id.UserID]CrossSigningKeys           `json:"self_signing_keys"`
	UserSigningKeys map[id.UserID]CrossSigningKeys           `json:"user_signing_keys"`
}

type RespClaimKeys struct {
	Failures    map[string]interface{}                                `json:"failures"`
	OneTimeKeys map[id.UserID]map[id.DeviceID]map[id.KeyID]OneTimeKey `json:"one_time_keys"`
}

type RespUploadSignatures struct {
	Failures map[string]interface{} `json:"failures"`
}

type RespKeyChanges struct {
	Changed []id.UserID `json:"changed"`
	Left    []id.UserID `json:"left"`
}

type RespSendToDevice struct{}

// RespDevicesInfo is the JSON response for https://matrix.org/docs/spec/client_server/r0.6.1#get-matrix-client-r0-devices
type RespDevicesInfo struct {
	Devices []RespDeviceInfo `json:"devices"`
}

// RespDeviceInfo is the JSON response for https://matrix.org/docs/spec/client_server/r0.6.1#get-matrix-client-r0-devices-deviceid
type RespDeviceInfo struct {
	DeviceID    id.DeviceID `json:"device_id"`
	DisplayName string      `json:"display_name"`
	LastSeenIP  string      `json:"last_seen_ip"`
	LastSeenTS  int64       `json:"last_seen_ts"`
}
