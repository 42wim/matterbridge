package mautrix

import (
	"encoding/json"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/pushrules"
)

type AuthType string

const (
	AuthTypePassword  = "m.login.password"
	AuthTypeReCAPTCHA = "m.login.recaptcha"
	AuthTypeOAuth2    = "m.login.oauth2"
	AuthTypeSSO       = "m.login.sso"
	AuthTypeEmail     = "m.login.email.identity"
	AuthTypeMSISDN    = "m.login.msisdn"
	AuthTypeToken     = "m.login.token"
	AuthTypeDummy     = "m.login.dummy"

	AuthTypeAppservice      = "m.login.application_service"
	AuthTypeHalfyAppservice = "uk.half-shot.msc2778.login.application_service"
)

type IdentifierType string

const (
	IdentifierTypeUser       = "m.id.user"
	IdentifierTypeThirdParty = "m.id.thirdparty"
	IdentifierTypePhone      = "m.id.phone"
)

// ReqRegister is the JSON request for https://matrix.org/docs/spec/client_server/r0.6.1#post-matrix-client-r0-register
type ReqRegister struct {
	Username                 string      `json:"username,omitempty"`
	Password                 string      `json:"password,omitempty"`
	DeviceID                 id.DeviceID `json:"device_id,omitempty"`
	InitialDeviceDisplayName string      `json:"initial_device_display_name,omitempty"`
	InhibitLogin             bool        `json:"inhibit_login,omitempty"`
	Auth                     interface{} `json:"auth,omitempty"`

	// Type for registration, only used for appservice user registrations
	// https://matrix.org/docs/spec/application_service/r0.1.2#server-admin-style-permissions
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

// ReqLogin is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-login
type ReqLogin struct {
	Type                     AuthType       `json:"type"`
	Identifier               UserIdentifier `json:"identifier"`
	Password                 string         `json:"password,omitempty"`
	Token                    string         `json:"token,omitempty"`
	DeviceID                 id.DeviceID    `json:"device_id,omitempty"`
	InitialDeviceDisplayName string         `json:"initial_device_display_name,omitempty"`

	// Whether or not the returned credentials should be stored in the Client
	StoreCredentials bool `json:"-"`
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

// ReqCreateRoom is the JSON request for https://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-createroom
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
}

// ReqRedact is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#put-matrix-client-r0-rooms-roomid-redact-eventid-txnid
type ReqRedact struct {
	Reason string `json:"reason,omitempty"`
	TxnID  string `json:"-"`
}

type ReqMembers struct {
	At            string           `json:"at"`
	Membership    event.Membership `json:"membership,omitempty"`
	NotMembership event.Membership `json:"not_membership,omitempty"`
}

// ReqInvite3PID is the JSON request for https://matrix.org/docs/spec/client_server/r0.2.0.html#id57
// It is also a JSON object used in https://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-createroom
type ReqInvite3PID struct {
	IDServer string `json:"id_server"`
	Medium   string `json:"medium"`
	Address  string `json:"address"`
}

// ReqInviteUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-invite
type ReqInviteUser struct {
	UserID id.UserID `json:"user_id"`
}

// ReqKickUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-kick
type ReqKickUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqBanUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-ban
type ReqBanUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqUnbanUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-unban
type ReqUnbanUser struct {
	UserID id.UserID `json:"user_id"`
}

// ReqTyping is the JSON request for https://matrix.org/docs/spec/client_server/r0.2.0.html#put-matrix-client-r0-rooms-roomid-typing-userid
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
	Key        id.Curve25519          `json:"key"`
	IsSigned   bool                   `json:"-"`
	Signatures Signatures             `json:"signatures,omitempty"`
	Unsigned   map[string]interface{} `json:"unsigned,omitempty"`
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

// ReqDeviceInfo is the JSON request for https://matrix.org/docs/spec/client_server/r0.6.1#put-matrix-client-r0-devices-deviceid
type ReqDeviceInfo struct {
	DisplayName string `json:"display_name,omitempty"`
}

// ReqDeleteDevice is the JSON request for https://matrix.org/docs/spec/client_server/r0.6.1#delete-matrix-client-r0-devices-deviceid
type ReqDeleteDevice struct {
	Auth interface{} `json:"auth,omitempty"`
}

// ReqDeleteDevices is the JSON request for https://matrix.org/docs/spec/client_server/r0.6.1#post-matrix-client-r0-delete-devices
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
