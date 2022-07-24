package mautrix

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/util"
	"maunium.net/go/mautrix/util/jsontime"
)

// RespWhoami is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3accountwhoami
type RespWhoami struct {
	UserID   id.UserID   `json:"user_id"`
	DeviceID id.DeviceID `json:"device_id"`
}

// RespCreateFilter is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3useruseridfilter
type RespCreateFilter struct {
	FilterID string `json:"filter_id"`
}

// RespJoinRoom is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidjoin
type RespJoinRoom struct {
	RoomID id.RoomID `json:"room_id"`
}

// RespLeaveRoom is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidleave
type RespLeaveRoom struct{}

// RespForgetRoom is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidforget
type RespForgetRoom struct{}

// RespInviteUser is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidinvite
type RespInviteUser struct{}

// RespKickUser is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidkick
type RespKickUser struct{}

// RespBanUser is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidban
type RespBanUser struct{}

// RespUnbanUser is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidunban
type RespUnbanUser struct{}

// RespTyping is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#put_matrixclientv3roomsroomidtypinguserid
type RespTyping struct{}

// RespPresence is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3presenceuseridstatus
type RespPresence struct {
	Presence        event.Presence `json:"presence"`
	LastActiveAgo   int            `json:"last_active_ago"`
	StatusMsg       string         `json:"status_msg"`
	CurrentlyActive bool           `json:"currently_active"`
}

// RespJoinedRooms is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3joined_rooms
type RespJoinedRooms struct {
	JoinedRooms []id.RoomID `json:"joined_rooms"`
}

// RespJoinedMembers is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3roomsroomidjoined_members
type RespJoinedMembers struct {
	Joined map[id.UserID]JoinedMember `json:"joined"`
}

type JoinedMember struct {
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// RespMessages is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3roomsroomidmessages
type RespMessages struct {
	Start string         `json:"start"`
	Chunk []*event.Event `json:"chunk"`
	State []*event.Event `json:"state"`
	End   string         `json:"end,omitempty"`
}

// RespContext is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3roomsroomidcontexteventid
type RespContext struct {
	End          string         `json:"end"`
	Event        *event.Event   `json:"event"`
	EventsAfter  []*event.Event `json:"events_after"`
	EventsBefore []*event.Event `json:"events_before"`
	Start        string         `json:"start"`
	State        []*event.Event `json:"state"`
}

// RespSendEvent is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#put_matrixclientv3roomsroomidsendeventtypetxnid
type RespSendEvent struct {
	EventID id.EventID `json:"event_id"`
}

// RespMediaConfig is the JSON response for https://spec.matrix.org/v1.4/client-server-api/#get_matrixmediav3config
type RespMediaConfig struct {
	UploadSize int64 `json:"m.upload.size,omitempty"`
}

// RespMediaUpload is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixmediav3upload
type RespMediaUpload struct {
	ContentURI id.ContentURI `json:"content_uri"`
}

// RespCreateMXC is the JSON response for /_matrix/media/v3/create as specified in https://github.com/matrix-org/matrix-spec-proposals/pull/2246
type RespCreateMXC struct {
	ContentURI      id.ContentURI `json:"content_uri"`
	UnusedExpiresAt int           `json:"unused_expires_at,omitempty"`
	UploadURL       string        `json:"upload_url,omitempty"`
}

// RespPreviewURL is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixmediav3preview_url
type RespPreviewURL struct {
	CanonicalURL string `json:"og:url,omitempty"`
	Title        string `json:"og:title,omitempty"`
	Type         string `json:"og:type,omitempty"`
	Description  string `json:"og:description,omitempty"`

	ImageURL id.ContentURIString `json:"og:image,omitempty"`

	ImageSize   int    `json:"matrix:image:size,omitempty"`
	ImageWidth  int    `json:"og:image:width,omitempty"`
	ImageHeight int    `json:"og:image:height,omitempty"`
	ImageType   string `json:"og:image:type,omitempty"`
}

// RespUserInteractive is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#user-interactive-authentication-api
type RespUserInteractive struct {
	Flows     []UIAFlow                `json:"flows,omitempty"`
	Params    map[AuthType]interface{} `json:"params,omitempty"`
	Session   string                   `json:"session,omitempty"`
	Completed []string                 `json:"completed,omitempty"`

	ErrCode string `json:"errcode,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UIAFlow struct {
	Stages []AuthType `json:"stages,omitempty"`
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

// RespUserDisplayName is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3profileuseriddisplayname
type RespUserDisplayName struct {
	DisplayName string `json:"displayname"`
}

type RespUserProfile struct {
	DisplayName string        `json:"displayname"`
	AvatarURL   id.ContentURI `json:"avatar_url"`
}

// RespRegisterAvailable is the JSON response for https://spec.matrix.org/v1.4/client-server-api/#get_matrixclientv3registeravailable
type RespRegisterAvailable struct {
	Available bool `json:"available"`
}

// RespRegister is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3register
type RespRegister struct {
	AccessToken string      `json:"access_token,omitempty"`
	DeviceID    id.DeviceID `json:"device_id,omitempty"`
	UserID      id.UserID   `json:"user_id"`

	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresInMS  int64  `json:"expires_in_ms,omitempty"`

	// Deprecated: homeserver should be parsed from the user ID
	HomeServer string `json:"home_server,omitempty"`
}

type LoginFlow struct {
	Type AuthType `json:"type"`
}

// RespLoginFlows is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3login
type RespLoginFlows struct {
	Flows []LoginFlow `json:"flows"`
}

func (rlf *RespLoginFlows) FirstFlowOfType(flowTypes ...AuthType) *LoginFlow {
	for _, flow := range rlf.Flows {
		for _, flowType := range flowTypes {
			if flow.Type == flowType {
				return &flow
			}
		}
	}
	return nil
}

func (rlf *RespLoginFlows) HasFlow(flowType ...AuthType) bool {
	return rlf.FirstFlowOfType(flowType...) != nil
}

// RespLogin is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3login
type RespLogin struct {
	AccessToken string           `json:"access_token"`
	DeviceID    id.DeviceID      `json:"device_id"`
	UserID      id.UserID        `json:"user_id"`
	WellKnown   *ClientWellKnown `json:"well_known,omitempty"`
}

// RespLogout is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3logout
type RespLogout struct{}

// RespCreateRoom is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3createroom
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

type SyncEventsList struct {
	Events []*event.Event `json:"events,omitempty"`
}

type SyncTimeline struct {
	SyncEventsList
	Limited   bool   `json:"limited,omitempty"`
	PrevBatch string `json:"prev_batch,omitempty"`
}

// RespSync is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3sync
type RespSync struct {
	NextBatch string `json:"next_batch"`

	AccountData SyncEventsList `json:"account_data"`
	Presence    SyncEventsList `json:"presence"`
	ToDevice    SyncEventsList `json:"to_device"`

	DeviceLists    DeviceLists       `json:"device_lists"`
	DeviceOTKCount OTKCount          `json:"device_one_time_keys_count"`
	FallbackKeys   []id.KeyAlgorithm `json:"device_unused_fallback_key_types"`

	Rooms RespSyncRooms `json:"rooms"`
}

type RespSyncRooms struct {
	Leave  map[id.RoomID]*SyncLeftRoom    `json:"leave,omitempty"`
	Join   map[id.RoomID]*SyncJoinedRoom  `json:"join,omitempty"`
	Invite map[id.RoomID]*SyncInvitedRoom `json:"invite,omitempty"`
	Knock  map[id.RoomID]*SyncKnockedRoom `json:"knock,omitempty"`
}

type marshalableRespSync RespSync

var syncPathsToDelete = []string{"account_data", "presence", "to_device", "device_lists", "device_one_time_keys_count", "rooms"}

func (rs *RespSync) MarshalJSON() ([]byte, error) {
	return util.MarshalAndDeleteEmpty((*marshalableRespSync)(rs), syncPathsToDelete)
}

type DeviceLists struct {
	Changed []id.UserID `json:"changed,omitempty"`
	Left    []id.UserID `json:"left,omitempty"`
}

type OTKCount struct {
	Curve25519       int `json:"curve25519,omitempty"`
	SignedCurve25519 int `json:"signed_curve25519,omitempty"`

	// For appservice OTK counts only: the user ID in question
	UserID   id.UserID   `json:"-"`
	DeviceID id.DeviceID `json:"-"`
}

type SyncLeftRoom struct {
	Summary  LazyLoadSummary `json:"summary"`
	State    SyncEventsList  `json:"state"`
	Timeline SyncTimeline    `json:"timeline"`
}

type marshalableSyncLeftRoom SyncLeftRoom

var syncLeftRoomPathsToDelete = []string{"summary", "state", "timeline"}

func (slr SyncLeftRoom) MarshalJSON() ([]byte, error) {
	return util.MarshalAndDeleteEmpty((marshalableSyncLeftRoom)(slr), syncLeftRoomPathsToDelete)
}

type SyncJoinedRoom struct {
	Summary     LazyLoadSummary `json:"summary"`
	State       SyncEventsList  `json:"state"`
	Timeline    SyncTimeline    `json:"timeline"`
	Ephemeral   SyncEventsList  `json:"ephemeral"`
	AccountData SyncEventsList  `json:"account_data"`

	UnreadNotifications *UnreadNotificationCounts `json:"unread_notifications,omitempty"`
	// https://github.com/matrix-org/matrix-spec-proposals/pull/2654
	MSC2654UnreadCount *int `json:"org.matrix.msc2654.unread_count,omitempty"`
}

type UnreadNotificationCounts struct {
	HighlightCount    int `json:"highlight_count"`
	NotificationCount int `json:"notification_count"`
}

type marshalableSyncJoinedRoom SyncJoinedRoom

var syncJoinedRoomPathsToDelete = []string{"summary", "state", "timeline", "ephemeral", "account_data"}

func (sjr SyncJoinedRoom) MarshalJSON() ([]byte, error) {
	return util.MarshalAndDeleteEmpty((marshalableSyncJoinedRoom)(sjr), syncJoinedRoomPathsToDelete)
}

type SyncInvitedRoom struct {
	Summary LazyLoadSummary `json:"summary"`
	State   SyncEventsList  `json:"invite_state"`
}

type marshalableSyncInvitedRoom SyncInvitedRoom

var syncInvitedRoomPathsToDelete = []string{"summary"}

func (sir SyncInvitedRoom) MarshalJSON() ([]byte, error) {
	return util.MarshalAndDeleteEmpty((marshalableSyncInvitedRoom)(sir), syncInvitedRoomPathsToDelete)
}

type SyncKnockedRoom struct {
	State SyncEventsList `json:"knock_state"`
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
type RespAliasList struct {
	Aliases []id.RoomAlias `json:"aliases"`
}

type RespUploadKeys struct {
	OneTimeKeyCounts OTKCount `json:"one_time_key_counts"`
}

type RespQueryKeys struct {
	Failures        map[string]interface{}                   `json:"failures,omitempty"`
	DeviceKeys      map[id.UserID]map[id.DeviceID]DeviceKeys `json:"device_keys"`
	MasterKeys      map[id.UserID]CrossSigningKeys           `json:"master_keys"`
	SelfSigningKeys map[id.UserID]CrossSigningKeys           `json:"self_signing_keys"`
	UserSigningKeys map[id.UserID]CrossSigningKeys           `json:"user_signing_keys"`
}

type RespClaimKeys struct {
	Failures    map[string]interface{}                                `json:"failures,omitempty"`
	OneTimeKeys map[id.UserID]map[id.DeviceID]map[id.KeyID]OneTimeKey `json:"one_time_keys"`
}

type RespUploadSignatures struct {
	Failures map[string]interface{} `json:"failures,omitempty"`
}

type RespKeyChanges struct {
	Changed []id.UserID `json:"changed"`
	Left    []id.UserID `json:"left"`
}

type RespSendToDevice struct{}

// RespDevicesInfo is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3devices
type RespDevicesInfo struct {
	Devices []RespDeviceInfo `json:"devices"`
}

// RespDeviceInfo is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientv3devicesdeviceid
type RespDeviceInfo struct {
	DeviceID    id.DeviceID `json:"device_id"`
	DisplayName string      `json:"display_name"`
	LastSeenIP  string      `json:"last_seen_ip"`
	LastSeenTS  int64       `json:"last_seen_ts"`
}

type RespBatchSend struct {
	StateEventIDs []id.EventID `json:"state_event_ids"`
	EventIDs      []id.EventID `json:"event_ids"`

	InsertionEventID     id.EventID `json:"insertion_event_id"`
	BatchEventID         id.EventID `json:"batch_event_id"`
	BaseInsertionEventID id.EventID `json:"base_insertion_event_id"`

	NextBatchID id.BatchID `json:"next_batch_id"`
}

// RespCapabilities is the JSON response for https://spec.matrix.org/v1.3/client-server-api/#get_matrixclientv3capabilities
type RespCapabilities struct {
	RoomVersions    *CapRoomVersions `json:"m.room_versions,omitempty"`
	ChangePassword  *CapBooleanTrue  `json:"m.change_password,omitempty"`
	SetDisplayname  *CapBooleanTrue  `json:"m.set_displayname,omitempty"`
	SetAvatarURL    *CapBooleanTrue  `json:"m.set_avatar_url,omitempty"`
	ThreePIDChanges *CapBooleanTrue  `json:"m.3pid_changes,omitempty"`

	Custom map[string]interface{} `json:"-"`
}

type serializableRespCapabilities RespCapabilities

func (rc *RespCapabilities) UnmarshalJSON(data []byte) error {
	res := gjson.GetBytes(data, "capabilities")
	if !res.Exists() || !res.IsObject() {
		return nil
	}
	if res.Index > 0 {
		data = data[res.Index : res.Index+len(res.Raw)]
	} else {
		data = []byte(res.Raw)
	}
	err := json.Unmarshal(data, (*serializableRespCapabilities)(rc))
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &rc.Custom)
	if err != nil {
		return err
	}
	// Remove non-custom capabilities from the custom map so that they don't get overridden when serializing back
	for _, field := range reflect.VisibleFields(reflect.TypeOf(rc).Elem()) {
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "-" && jsonTag != "" {
			delete(rc.Custom, jsonTag)
		}
	}
	return nil
}

func (rc *RespCapabilities) MarshalJSON() ([]byte, error) {
	marshalableCopy := make(map[string]interface{}, len(rc.Custom))
	val := reflect.ValueOf(rc).Elem()
	for _, field := range reflect.VisibleFields(val.Type()) {
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "-" && jsonTag != "" {
			fieldVal := val.FieldByIndex(field.Index)
			if !fieldVal.IsNil() {
				marshalableCopy[jsonTag] = fieldVal.Interface()
			}
		}
	}
	if rc.Custom != nil {
		for key, value := range rc.Custom {
			marshalableCopy[key] = value
		}
	}
	var buf bytes.Buffer
	buf.WriteString(`{"capabilities":`)
	err := json.NewEncoder(&buf).Encode(marshalableCopy)
	if err != nil {
		return nil, err
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type CapBoolean struct {
	Enabled bool `json:"enabled"`
}

type CapBooleanTrue CapBoolean

// IsEnabled returns true if the capability is either enabled explicitly or not specified (nil)
func (cb *CapBooleanTrue) IsEnabled() bool {
	// Default to true when
	return cb == nil || cb.Enabled
}

type CapBooleanFalse CapBoolean

// IsEnabled returns true if the capability is enabled explicitly. If it's not specified, this returns false.
func (cb *CapBooleanFalse) IsEnabled() bool {
	return cb != nil && cb.Enabled
}

type CapRoomVersionStability string

const (
	CapRoomVersionStable   CapRoomVersionStability = "stable"
	CapRoomVersionUnstable CapRoomVersionStability = "unstable"
)

type CapRoomVersions struct {
	Default   string                             `json:"default"`
	Available map[string]CapRoomVersionStability `json:"available"`
}

func (vers *CapRoomVersions) IsStable(version string) bool {
	if vers == nil || vers.Available == nil {
		val, err := strconv.Atoi(version)
		return err == nil && val > 0
	}
	return vers.Available[version] == CapRoomVersionStable
}

func (vers *CapRoomVersions) IsAvailable(version string) bool {
	if vers == nil || vers.Available == nil {
		return false
	}
	_, available := vers.Available[version]
	return available
}

// RespHierarchy is the JSON response for https://spec.matrix.org/v1.4/client-server-api/#get_matrixclientv1roomsroomidhierarchy
type RespHierarchy struct {
	NextBatch string            `json:"next_batch,omitempty"`
	Rooms     []ChildRoomsChunk `json:"rooms"`
}

type ChildRoomsChunk struct {
	AvatarURL        id.ContentURI           `json:"avatar_url,omitempty"`
	CanonicalAlias   id.RoomAlias            `json:"canonical_alias,omitempty"`
	ChildrenState    []StrippedStateWithTime `json:"children_state"`
	GuestCanJoin     bool                    `json:"guest_can_join"`
	JoinRule         event.JoinRule          `json:"join_rule,omitempty"`
	Name             string                  `json:"name,omitempty"`
	NumJoinedMembers int                     `json:"num_joined_members"`
	RoomID           id.RoomID               `json:"room_id"`
	RoomType         event.RoomType          `json:"room_type"`
	Topic            string                  `json:"topic,omitempty"`
	WorldReadble     bool                    `json:"world_readable"`
}

type StrippedStateWithTime struct {
	event.StrippedState
	Timestamp jsontime.UnixMilli `json:"origin_server_ts"`
}

type RespAppservicePing struct {
	DurationMS int64 `json:"duration"`
}

type RespBeeperMergeRoom RespCreateRoom

type RespBeeperSplitRoom struct {
	RoomIDs map[string]id.RoomID `json:"room_ids"`
}

type RespTimestampToEvent struct {
	EventID   id.EventID         `json:"event_id"`
	Timestamp jsontime.UnixMilli `json:"origin_server_ts"`
}

type RespRoomKeysVersionCreate struct {
	Version string `json:"version"`
}

type RespRoomKeysVersion struct {
	Algorithm string          `json:"algorithm"`
	AuthData  json.RawMessage `json:"auth_data"`
	Count     int             `json:"count"`
	ETag      string          `json:"etag"`
	Version   string          `json:"version"`
}

type RespRoomKeys struct {
	Rooms map[id.RoomID]RespRoomKeysRoom `json:"rooms"`
}

type RespRoomKeysRoom struct {
	Sessions map[id.SessionID]RespRoomKeysSession `json:"sessions"`
}

type RespRoomKeysSession struct {
	FirstMessageIndex int             `json:"first_message_index"`
	ForwardedCount    int             `json:"forwarded_count"`
	IsVerified        bool            `json:"is_verified"`
	SessionData       json.RawMessage `json:"session_data"`
}

type RespRoomKeysUpdate struct {
	Count int    `json:"count"`
	ETag  string `json:"etag"`
}
