package mautrix

import (
	"encoding/json"
	"strconv"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/pushrules"
)

type AuthType string

const (
	AuthTypePassword   AuthType = "m.login.password"
	AuthTypeReCAPTCHA  AuthType = "m.login.recaptcha"
	AuthTypeOAuth2     AuthType = "m.login.oauth2"
	AuthTypeSSO        AuthType = "m.login.sso"
	AuthTypeEmail      AuthType = "m.login.email.identity"
	AuthTypeMSISDN     AuthType = "m.login.msisdn"
	AuthTypeToken      AuthType = "m.login.token"
	AuthTypeDummy      AuthType = "m.login.dummy"
	AuthTypeAppservice AuthType = "m.login.application_service"
)

type IdentifierType string

const (
	IdentifierTypeUser       = "m.id.user"
	IdentifierTypeThirdParty = "m.id.thirdparty"
	IdentifierTypePhone      = "m.id.phone"
)

type Direction rune

const (
	DirectionForward  Direction = 'f'
	DirectionBackward Direction = 'b'
)

// ReqRegister is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3register
type ReqRegister struct {
	Username                 string      `json:"username,omitempty"`
	Password                 string      `json:"password,omitempty"`
	DeviceID                 id.DeviceID `json:"device_id,omitempty"`
	InitialDeviceDisplayName string      `json:"initial_device_display_name,omitempty"`
	InhibitLogin             bool        `json:"inhibit_login,omitempty"`
	RefreshToken             bool        `json:"refresh_token,omitempty"`
	Auth                     interface{} `json:"auth,omitempty"`

	// Type for registration, only used for appservice user registrations
	// https://spec.matrix.org/v1.2/application-service-api/#server-admin-style-permissions
	Type AuthType `json:"type,omitempty"`
}

type BaseAuthData struct {
	Type    AuthType `json:"type"`
	Session string   `json:"session,omitempty"`
}

type UserIdentifier struct {
	Type IdentifierType `json:"type"`

	User string `json:"user,omitempty"`

	Medium  string `json:"medium,omitempty"`
	Address string `json:"address,omitempty"`

	Country string `json:"country,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// ReqLogin is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3login
type ReqLogin struct {
	Type                     AuthType       `json:"type"`
	Identifier               UserIdentifier `json:"identifier"`
	Password                 string         `json:"password,omitempty"`
	Token                    string         `json:"token,omitempty"`
	DeviceID                 id.DeviceID    `json:"device_id,omitempty"`
	InitialDeviceDisplayName string         `json:"initial_device_display_name,omitempty"`

	// Whether or not the returned credentials should be stored in the Client
	StoreCredentials bool `json:"-"`
	// Whether or not the returned .well-known data should update the homeserver URL in the Client
	StoreHomeserverURL bool `json:"-"`
}

type ReqUIAuthFallback struct {
	Session string `json:"session"`
	User    string `json:"user"`
}

type ReqUIAuthLogin struct {
	BaseAuthData
	User     string `json:"user"`
	Password string `json:"password"`
}

// ReqCreateRoom is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3createroom
type ReqCreateRoom struct {
	Visibility      string                 `json:"visibility,omitempty"`
	RoomAliasName   string                 `json:"room_alias_name,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Topic           string                 `json:"topic,omitempty"`
	Invite          []id.UserID            `json:"invite,omitempty"`
	Invite3PID      []ReqInvite3PID        `json:"invite_3pid,omitempty"`
	CreationContent map[string]interface{} `json:"creation_content,omitempty"`
	InitialState    []*event.Event         `json:"initial_state,omitempty"`
	Preset          string                 `json:"preset,omitempty"`
	IsDirect        bool                   `json:"is_direct,omitempty"`
	RoomVersion     string                 `json:"room_version,omitempty"`

	PowerLevelOverride *event.PowerLevelsEventContent `json:"power_level_content_override,omitempty"`

	MeowRoomID            id.RoomID `json:"fi.mau.room_id,omitempty"`
	BeeperAutoJoinInvites bool      `json:"com.beeper.auto_join_invites,omitempty"`
}

// ReqRedact is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#put_matrixclientv3roomsroomidredacteventidtxnid
type ReqRedact struct {
	Reason string
	TxnID  string
	Extra  map[string]interface{}
}

type ReqMembers struct {
	At            string           `json:"at"`
	Membership    event.Membership `json:"membership,omitempty"`
	NotMembership event.Membership `json:"not_membership,omitempty"`
}

// ReqInvite3PID is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidinvite-1
// It is also a JSON object used in https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3createroom
type ReqInvite3PID struct {
	IDServer string `json:"id_server"`
	Medium   string `json:"medium"`
	Address  string `json:"address"`
}

type ReqLeave struct {
	Reason string `json:"reason,omitempty"`
}

// ReqInviteUser is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidinvite
type ReqInviteUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqKickUser is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidkick
type ReqKickUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqBanUser is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidban
type ReqBanUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqUnbanUser is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3roomsroomidunban
type ReqUnbanUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqTyping is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#put_matrixclientv3roomsroomidtypinguserid
type ReqTyping struct {
	Typing  bool  `json:"typing"`
	Timeout int64 `json:"timeout,omitempty"`
}

type ReqPresence struct {
	Presence event.Presence `json:"presence"`
}

type ReqAliasCreate struct {
	RoomID id.RoomID `json:"room_id"`
}

type OneTimeKey struct {
	Key        id.Curve25519  `json:"key"`
	Fallback   bool           `json:"fallback,omitempty"`
	Signatures Signatures     `json:"signatures,omitempty"`
	Unsigned   map[string]any `json:"unsigned,omitempty"`
	IsSigned   bool           `json:"-"`

	// Raw data in the one-time key. This must be used for signature verification to ensure unrecognized fields
	// aren't thrown away (because that would invalidate the signature).
	RawData json.RawMessage `json:"-"`
}

type serializableOTK OneTimeKey

func (otk *OneTimeKey) UnmarshalJSON(data []byte) (err error) {
	if len(data) > 0 && data[0] == '"' && data[len(data)-1] == '"' {
		err = json.Unmarshal(data, &otk.Key)
		otk.Signatures = nil
		otk.Unsigned = nil
		otk.IsSigned = false
	} else {
		err = json.Unmarshal(data, (*serializableOTK)(otk))
		otk.RawData = data
		otk.IsSigned = true
	}
	return err
}

func (otk *OneTimeKey) MarshalJSON() ([]byte, error) {
	if !otk.IsSigned {
		return json.Marshal(otk.Key)
	} else {
		return json.Marshal((*serializableOTK)(otk))
	}
}

type ReqUploadKeys struct {
	DeviceKeys  *DeviceKeys             `json:"device_keys,omitempty"`
	OneTimeKeys map[id.KeyID]OneTimeKey `json:"one_time_keys"`
}

type ReqKeysSignatures struct {
	UserID     id.UserID              `json:"user_id"`
	DeviceID   id.DeviceID            `json:"device_id,omitempty"`
	Algorithms []id.Algorithm         `json:"algorithms,omitempty"`
	Usage      []id.CrossSigningUsage `json:"usage,omitempty"`
	Keys       map[id.KeyID]string    `json:"keys"`
	Signatures Signatures             `json:"signatures"`
}

type ReqUploadSignatures map[id.UserID]map[string]ReqKeysSignatures

type DeviceKeys struct {
	UserID     id.UserID              `json:"user_id"`
	DeviceID   id.DeviceID            `json:"device_id"`
	Algorithms []id.Algorithm         `json:"algorithms"`
	Keys       KeyMap                 `json:"keys"`
	Signatures Signatures             `json:"signatures"`
	Unsigned   map[string]interface{} `json:"unsigned,omitempty"`
}

type CrossSigningKeys struct {
	UserID     id.UserID                         `json:"user_id"`
	Usage      []id.CrossSigningUsage            `json:"usage"`
	Keys       map[id.KeyID]id.Ed25519           `json:"keys"`
	Signatures map[id.UserID]map[id.KeyID]string `json:"signatures,omitempty"`
}

func (csk *CrossSigningKeys) FirstKey() id.Ed25519 {
	for _, key := range csk.Keys {
		return key
	}
	return ""
}

type UploadCrossSigningKeysReq struct {
	Master      CrossSigningKeys `json:"master_key"`
	SelfSigning CrossSigningKeys `json:"self_signing_key"`
	UserSigning CrossSigningKeys `json:"user_signing_key"`
	Auth        interface{}      `json:"auth,omitempty"`
}

type KeyMap map[id.DeviceKeyID]string

func (km KeyMap) GetEd25519(deviceID id.DeviceID) id.Ed25519 {
	val, ok := km[id.NewDeviceKeyID(id.KeyAlgorithmEd25519, deviceID)]
	if !ok {
		return ""
	}
	return id.Ed25519(val)
}

func (km KeyMap) GetCurve25519(deviceID id.DeviceID) id.Curve25519 {
	val, ok := km[id.NewDeviceKeyID(id.KeyAlgorithmCurve25519, deviceID)]
	if !ok {
		return ""
	}
	return id.Curve25519(val)
}

type Signatures map[id.UserID]map[id.KeyID]string

type ReqQueryKeys struct {
	DeviceKeys DeviceKeysRequest `json:"device_keys"`

	Timeout int64  `json:"timeout,omitempty"`
	Token   string `json:"token,omitempty"`
}

type DeviceKeysRequest map[id.UserID]DeviceIDList

type DeviceIDList []id.DeviceID

type ReqClaimKeys struct {
	OneTimeKeys OneTimeKeysRequest `json:"one_time_keys"`

	Timeout int64 `json:"timeout,omitempty"`
}

type OneTimeKeysRequest map[id.UserID]map[id.DeviceID]id.KeyAlgorithm

type ReqSendToDevice struct {
	Messages map[id.UserID]map[id.DeviceID]*event.Content `json:"messages"`
}

// ReqDeviceInfo is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#put_matrixclientv3devicesdeviceid
type ReqDeviceInfo struct {
	DisplayName string `json:"display_name,omitempty"`
}

// ReqDeleteDevice is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#delete_matrixclientv3devicesdeviceid
type ReqDeleteDevice struct {
	Auth interface{} `json:"auth,omitempty"`
}

// ReqDeleteDevices is the JSON request for https://spec.matrix.org/v1.2/client-server-api/#post_matrixclientv3delete_devices
type ReqDeleteDevices struct {
	Devices []id.DeviceID `json:"devices"`
	Auth    interface{}   `json:"auth,omitempty"`
}

type ReqPutPushRule struct {
	Before string `json:"-"`
	After  string `json:"-"`

	Actions    []pushrules.PushActionType `json:"actions"`
	Conditions []pushrules.PushCondition  `json:"conditions"`
	Pattern    string                     `json:"pattern"`
}

type ReqBatchSend struct {
	PrevEventID id.EventID `json:"-"`
	BatchID     id.BatchID `json:"-"`

	BeeperNewMessages bool      `json:"-"`
	BeeperMarkReadBy  id.UserID `json:"-"`

	StateEventsAtStart []*event.Event `json:"state_events_at_start"`
	Events             []*event.Event `json:"events"`
}

type ReqSetReadMarkers struct {
	Read        id.EventID `json:"m.read,omitempty"`
	ReadPrivate id.EventID `json:"m.read.private,omitempty"`
	FullyRead   id.EventID `json:"m.fully_read,omitempty"`

	BeeperReadExtra        interface{} `json:"com.beeper.read.extra,omitempty"`
	BeeperReadPrivateExtra interface{} `json:"com.beeper.read.private.extra,omitempty"`
	BeeperFullyReadExtra   interface{} `json:"com.beeper.fully_read.extra,omitempty"`
}

type ReqSendReceipt struct {
	ThreadID string `json:"thread_id,omitempty"`
}

// ReqHierarchy contains the parameters for https://spec.matrix.org/v1.4/client-server-api/#get_matrixclientv1roomsroomidhierarchy
//
// As it's a GET method, there is no JSON body, so this is only query parameters.
type ReqHierarchy struct {
	// A pagination token from a previous Hierarchy call.
	// If specified, max_depth and suggested_only cannot be changed from the first request.
	From string
	// Limit for the maximum number of rooms to include per response.
	// The server will apply a default value if a limit isn't provided.
	Limit int
	// Limit for how far to go into the space. When reached, no further child rooms will be returned.
	// The server will apply a default value if a max depth isn't provided.
	MaxDepth *int
	// Flag to indicate whether the server should only consider suggested rooms.
	// Suggested rooms are annotated in their m.space.child event contents.
	SuggestedOnly bool
}

func (req *ReqHierarchy) Query() map[string]string {
	query := map[string]string{}
	if req == nil {
		return query
	}
	if req.From != "" {
		query["from"] = req.From
	}
	if req.Limit > 0 {
		query["limit"] = strconv.Itoa(req.Limit)
	}
	if req.MaxDepth != nil {
		query["max_depth"] = strconv.Itoa(*req.MaxDepth)
	}
	if req.SuggestedOnly {
		query["suggested_only"] = "true"
	}
	return query
}

type ReqAppservicePing struct {
	TxnID string `json:"transaction_id,omitempty"`
}

type ReqBeeperMergeRoom struct {
	NewRoom ReqCreateRoom `json:"create"`
	Key     string        `json:"key"`
	Rooms   []id.RoomID   `json:"rooms"`
	User    id.UserID     `json:"user_id"`
}

type BeeperSplitRoomPart struct {
	UserID  id.UserID     `json:"user_id"`
	Values  []string      `json:"values"`
	NewRoom ReqCreateRoom `json:"create"`
}

type ReqBeeperSplitRoom struct {
	RoomID id.RoomID `json:"-"`

	Key   string                `json:"key"`
	Parts []BeeperSplitRoomPart `json:"parts"`
}

type ReqRoomKeysVersionCreate struct {
	Algorithm string          `json:"algorithm"`
	AuthData  json.RawMessage `json:"auth_data"`
}

type ReqRoomKeysUpdate struct {
	Rooms map[id.RoomID]ReqRoomKeysRoomUpdate `json:"rooms"`
}

type ReqRoomKeysRoomUpdate struct {
	Sessions map[id.SessionID]ReqRoomKeysSessionUpdate `json:"sessions"`
}

type ReqRoomKeysSessionUpdate struct {
	FirstMessageIndex int             `json:"first_message_index"`
	ForwardedCount    int             `json:"forwarded_count"`
	IsVerified        bool            `json:"is_verified"`
	SessionData       json.RawMessage `json:"session_data"`
}
