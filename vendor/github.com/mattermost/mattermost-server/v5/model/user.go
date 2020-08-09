// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mattermost/mattermost-server/v5/services/timezones"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/language"
)

const (
	ME                                 = "me"
	USER_NOTIFY_ALL                    = "all"
	USER_NOTIFY_HERE                   = "here"
	USER_NOTIFY_MENTION                = "mention"
	USER_NOTIFY_NONE                   = "none"
	DESKTOP_NOTIFY_PROP                = "desktop"
	DESKTOP_SOUND_NOTIFY_PROP          = "desktop_sound"
	MARK_UNREAD_NOTIFY_PROP            = "mark_unread"
	PUSH_NOTIFY_PROP                   = "push"
	PUSH_STATUS_NOTIFY_PROP            = "push_status"
	EMAIL_NOTIFY_PROP                  = "email"
	CHANNEL_MENTIONS_NOTIFY_PROP       = "channel"
	COMMENTS_NOTIFY_PROP               = "comments"
	MENTION_KEYS_NOTIFY_PROP           = "mention_keys"
	COMMENTS_NOTIFY_NEVER              = "never"
	COMMENTS_NOTIFY_ROOT               = "root"
	COMMENTS_NOTIFY_ANY                = "any"
	FIRST_NAME_NOTIFY_PROP             = "first_name"
	AUTO_RESPONDER_ACTIVE_NOTIFY_PROP  = "auto_responder_active"
	AUTO_RESPONDER_MESSAGE_NOTIFY_PROP = "auto_responder_message"

	DEFAULT_LOCALE          = "en"
	USER_AUTH_SERVICE_EMAIL = "email"

	USER_EMAIL_MAX_LENGTH     = 128
	USER_NICKNAME_MAX_RUNES   = 64
	USER_POSITION_MAX_RUNES   = 128
	USER_FIRST_NAME_MAX_RUNES = 64
	USER_LAST_NAME_MAX_RUNES  = 64
	USER_AUTH_DATA_MAX_LENGTH = 128
	USER_NAME_MAX_LENGTH      = 64
	USER_NAME_MIN_LENGTH      = 1
	USER_PASSWORD_MAX_LENGTH  = 72
	USER_LOCALE_MAX_LENGTH    = 5
)

type User struct {
	Id                     string    `json:"id"`
	CreateAt               int64     `json:"create_at,omitempty"`
	UpdateAt               int64     `json:"update_at,omitempty"`
	DeleteAt               int64     `json:"delete_at"`
	Username               string    `json:"username"`
	Password               string    `json:"password,omitempty"`
	AuthData               *string   `json:"auth_data,omitempty"`
	AuthService            string    `json:"auth_service"`
	Email                  string    `json:"email"`
	EmailVerified          bool      `json:"email_verified,omitempty"`
	Nickname               string    `json:"nickname"`
	FirstName              string    `json:"first_name"`
	LastName               string    `json:"last_name"`
	Position               string    `json:"position"`
	Roles                  string    `json:"roles"`
	AllowMarketing         bool      `json:"allow_marketing,omitempty"`
	Props                  StringMap `json:"props,omitempty"`
	NotifyProps            StringMap `json:"notify_props,omitempty"`
	LastPasswordUpdate     int64     `json:"last_password_update,omitempty"`
	LastPictureUpdate      int64     `json:"last_picture_update,omitempty"`
	FailedAttempts         int       `json:"failed_attempts,omitempty"`
	Locale                 string    `json:"locale"`
	Timezone               StringMap `json:"timezone"`
	MfaActive              bool      `json:"mfa_active,omitempty"`
	MfaSecret              string    `json:"mfa_secret,omitempty"`
	LastActivityAt         int64     `db:"-" json:"last_activity_at,omitempty"`
	IsBot                  bool      `db:"-" json:"is_bot,omitempty"`
	BotDescription         string    `db:"-" json:"bot_description,omitempty"`
	BotLastIconUpdate      int64     `db:"-" json:"bot_last_icon_update,omitempty"`
	TermsOfServiceId       string    `db:"-" json:"terms_of_service_id,omitempty"`
	TermsOfServiceCreateAt int64     `db:"-" json:"terms_of_service_create_at,omitempty"`
}

type UserUpdate struct {
	Old *User
	New *User
}

type UserPatch struct {
	Username    *string   `json:"username"`
	Password    *string   `json:"password,omitempty"`
	Nickname    *string   `json:"nickname"`
	FirstName   *string   `json:"first_name"`
	LastName    *string   `json:"last_name"`
	Position    *string   `json:"position"`
	Email       *string   `json:"email"`
	Props       StringMap `json:"props,omitempty"`
	NotifyProps StringMap `json:"notify_props,omitempty"`
	Locale      *string   `json:"locale"`
	Timezone    StringMap `json:"timezone"`
}

type UserAuth struct {
	Password    string  `json:"password,omitempty"`
	AuthData    *string `json:"auth_data,omitempty"`
	AuthService string  `json:"auth_service,omitempty"`
}

type UserForIndexing struct {
	Id          string   `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	CreateAt    int64    `json:"create_at"`
	DeleteAt    int64    `json:"delete_at"`
	TeamsIds    []string `json:"team_id"`
	ChannelsIds []string `json:"channel_id"`
}

type ViewUsersRestrictions struct {
	Teams    []string
	Channels []string
}

func (r *ViewUsersRestrictions) Hash() string {
	if r == nil {
		return ""
	}
	ids := append(r.Teams, r.Channels...)
	sort.Strings(ids)
	hash := sha256.New()
	hash.Write([]byte(strings.Join(ids, "")))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

type UserSlice []*User

func (u UserSlice) Usernames() []string {
	usernames := []string{}
	for _, user := range u {
		usernames = append(usernames, user.Username)
	}
	sort.Strings(usernames)
	return usernames
}

func (u UserSlice) IDs() []string {
	ids := []string{}
	for _, user := range u {
		ids = append(ids, user.Id)
	}
	return ids
}

func (u UserSlice) FilterWithoutBots() UserSlice {
	var matches []*User

	for _, user := range u {
		if !user.IsBot {
			matches = append(matches, user)
		}
	}
	return UserSlice(matches)
}

func (u UserSlice) FilterByActive(active bool) UserSlice {
	var matches []*User

	for _, user := range u {
		if user.DeleteAt == 0 && active {
			matches = append(matches, user)
		} else if user.DeleteAt != 0 && !active {
			matches = append(matches, user)
		}
	}
	return UserSlice(matches)
}

func (u UserSlice) FilterByID(ids []string) UserSlice {
	var matches []*User
	for _, user := range u {
		for _, id := range ids {
			if id == user.Id {
				matches = append(matches, user)
			}
		}
	}
	return UserSlice(matches)
}

func (u UserSlice) FilterWithoutID(ids []string) UserSlice {
	var keep []*User
	for _, user := range u {
		present := false
		for _, id := range ids {
			if id == user.Id {
				present = true
			}
		}
		if !present {
			keep = append(keep, user)
		}
	}
	return UserSlice(keep)
}

func (u *User) DeepCopy() *User {
	copyUser := *u
	if u.AuthData != nil {
		copyUser.AuthData = NewString(*u.AuthData)
	}
	if u.Props != nil {
		copyUser.Props = CopyStringMap(u.Props)
	}
	if u.NotifyProps != nil {
		copyUser.NotifyProps = CopyStringMap(u.NotifyProps)
	}
	if u.Timezone != nil {
		copyUser.Timezone = CopyStringMap(u.Timezone)
	}
	return &copyUser
}

// IsValid validates the user and returns an error if it isn't configured
// correctly.
func (u *User) IsValid() *AppError {

	if !IsValidId(u.Id) {
		return InvalidUserError("id", "")
	}

	if u.CreateAt == 0 {
		return InvalidUserError("create_at", u.Id)
	}

	if u.UpdateAt == 0 {
		return InvalidUserError("update_at", u.Id)
	}

	if !IsValidUsername(u.Username) {
		return InvalidUserError("username", u.Id)
	}

	if len(u.Email) > USER_EMAIL_MAX_LENGTH || len(u.Email) == 0 || !IsValidEmail(u.Email) {
		return InvalidUserError("email", u.Id)
	}

	if utf8.RuneCountInString(u.Nickname) > USER_NICKNAME_MAX_RUNES {
		return InvalidUserError("nickname", u.Id)
	}

	if utf8.RuneCountInString(u.Position) > USER_POSITION_MAX_RUNES {
		return InvalidUserError("position", u.Id)
	}

	if utf8.RuneCountInString(u.FirstName) > USER_FIRST_NAME_MAX_RUNES {
		return InvalidUserError("first_name", u.Id)
	}

	if utf8.RuneCountInString(u.LastName) > USER_LAST_NAME_MAX_RUNES {
		return InvalidUserError("last_name", u.Id)
	}

	if u.AuthData != nil && len(*u.AuthData) > USER_AUTH_DATA_MAX_LENGTH {
		return InvalidUserError("auth_data", u.Id)
	}

	if u.AuthData != nil && len(*u.AuthData) > 0 && len(u.AuthService) == 0 {
		return InvalidUserError("auth_data_type", u.Id)
	}

	if len(u.Password) > 0 && u.AuthData != nil && len(*u.AuthData) > 0 {
		return InvalidUserError("auth_data_pwd", u.Id)
	}

	if len(u.Password) > USER_PASSWORD_MAX_LENGTH {
		return InvalidUserError("password_limit", u.Id)
	}

	if !IsValidLocale(u.Locale) {
		return InvalidUserError("locale", u.Id)
	}

	return nil
}

func InvalidUserError(fieldName string, userId string) *AppError {
	id := fmt.Sprintf("model.user.is_valid.%s.app_error", fieldName)
	details := ""
	if userId != "" {
		details = "user_id=" + userId
	}
	return NewAppError("User.IsValid", id, nil, details, http.StatusBadRequest)
}

func NormalizeUsername(username string) string {
	return strings.ToLower(username)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(email)
}

// PreSave will set the Id and Username if missing.  It will also fill
// in the CreateAt, UpdateAt times.  It will also hash the password.  It should
// be run before saving the user to the db.
func (u *User) PreSave() {
	if u.Id == "" {
		u.Id = NewId()
	}

	if u.Username == "" {
		u.Username = NewId()
	}

	if u.AuthData != nil && *u.AuthData == "" {
		u.AuthData = nil
	}

	u.Username = SanitizeUnicode(u.Username)
	u.FirstName = SanitizeUnicode(u.FirstName)
	u.LastName = SanitizeUnicode(u.LastName)
	u.Nickname = SanitizeUnicode(u.Nickname)

	u.Username = NormalizeUsername(u.Username)
	u.Email = NormalizeEmail(u.Email)

	u.CreateAt = GetMillis()
	u.UpdateAt = u.CreateAt

	u.LastPasswordUpdate = u.CreateAt

	u.MfaActive = false

	if u.Locale == "" {
		u.Locale = DEFAULT_LOCALE
	}

	if u.Props == nil {
		u.Props = make(map[string]string)
	}

	if u.NotifyProps == nil || len(u.NotifyProps) == 0 {
		u.SetDefaultNotifications()
	}

	if u.Timezone == nil {
		u.Timezone = timezones.DefaultUserTimezone()
	}

	if len(u.Password) > 0 {
		u.Password = HashPassword(u.Password)
	}
}

// PreUpdate should be run before updating the user in the db.
func (u *User) PreUpdate() {
	u.Username = SanitizeUnicode(u.Username)
	u.FirstName = SanitizeUnicode(u.FirstName)
	u.LastName = SanitizeUnicode(u.LastName)
	u.Nickname = SanitizeUnicode(u.Nickname)
	u.BotDescription = SanitizeUnicode(u.BotDescription)

	u.Username = NormalizeUsername(u.Username)
	u.Email = NormalizeEmail(u.Email)
	u.UpdateAt = GetMillis()

	u.FirstName = SanitizeUnicode(u.FirstName)
	u.LastName = SanitizeUnicode(u.LastName)
	u.Nickname = SanitizeUnicode(u.Nickname)
	u.BotDescription = SanitizeUnicode(u.BotDescription)

	if u.AuthData != nil && *u.AuthData == "" {
		u.AuthData = nil
	}

	if u.NotifyProps == nil || len(u.NotifyProps) == 0 {
		u.SetDefaultNotifications()
	} else if _, ok := u.NotifyProps[MENTION_KEYS_NOTIFY_PROP]; ok {
		// Remove any blank mention keys
		splitKeys := strings.Split(u.NotifyProps[MENTION_KEYS_NOTIFY_PROP], ",")
		goodKeys := []string{}
		for _, key := range splitKeys {
			if len(key) > 0 {
				goodKeys = append(goodKeys, strings.ToLower(key))
			}
		}
		u.NotifyProps[MENTION_KEYS_NOTIFY_PROP] = strings.Join(goodKeys, ",")
	}
}

func (u *User) SetDefaultNotifications() {
	u.NotifyProps = make(map[string]string)
	u.NotifyProps[EMAIL_NOTIFY_PROP] = "true"
	u.NotifyProps[PUSH_NOTIFY_PROP] = USER_NOTIFY_MENTION
	u.NotifyProps[DESKTOP_NOTIFY_PROP] = USER_NOTIFY_MENTION
	u.NotifyProps[DESKTOP_SOUND_NOTIFY_PROP] = "true"
	u.NotifyProps[MENTION_KEYS_NOTIFY_PROP] = ""
	u.NotifyProps[CHANNEL_MENTIONS_NOTIFY_PROP] = "true"
	u.NotifyProps[PUSH_STATUS_NOTIFY_PROP] = STATUS_AWAY
	u.NotifyProps[COMMENTS_NOTIFY_PROP] = COMMENTS_NOTIFY_NEVER
	u.NotifyProps[FIRST_NAME_NOTIFY_PROP] = "false"
}

func (u *User) UpdateMentionKeysFromUsername(oldUsername string) {
	nonUsernameKeys := []string{}
	for _, key := range u.GetMentionKeys() {
		if key != oldUsername && key != "@"+oldUsername {
			nonUsernameKeys = append(nonUsernameKeys, key)
		}
	}

	u.NotifyProps[MENTION_KEYS_NOTIFY_PROP] = ""
	if len(nonUsernameKeys) > 0 {
		u.NotifyProps[MENTION_KEYS_NOTIFY_PROP] += "," + strings.Join(nonUsernameKeys, ",")
	}
}

func (u *User) GetMentionKeys() []string {
	var keys []string

	for _, key := range strings.Split(u.NotifyProps[MENTION_KEYS_NOTIFY_PROP], ",") {
		trimmedKey := strings.TrimSpace(key)

		if trimmedKey == "" {
			continue
		}

		keys = append(keys, trimmedKey)
	}

	return keys
}

func (u *User) Patch(patch *UserPatch) {
	if patch.Username != nil {
		u.Username = *patch.Username
	}

	if patch.Nickname != nil {
		u.Nickname = *patch.Nickname
	}

	if patch.FirstName != nil {
		u.FirstName = *patch.FirstName
	}

	if patch.LastName != nil {
		u.LastName = *patch.LastName
	}

	if patch.Position != nil {
		u.Position = *patch.Position
	}

	if patch.Email != nil {
		u.Email = *patch.Email
	}

	if patch.Props != nil {
		u.Props = patch.Props
	}

	if patch.NotifyProps != nil {
		u.NotifyProps = patch.NotifyProps
	}

	if patch.Locale != nil {
		u.Locale = *patch.Locale
	}

	if patch.Timezone != nil {
		u.Timezone = patch.Timezone
	}
}

// ToJson convert a User to a json string
func (u *User) ToJson() string {
	b, _ := json.Marshal(u)
	return string(b)
}

func (u *UserPatch) ToJson() string {
	b, _ := json.Marshal(u)
	return string(b)
}

func (u *UserAuth) ToJson() string {
	b, _ := json.Marshal(u)
	return string(b)
}

// Generate a valid strong etag so the browser can cache the results
func (u *User) Etag(showFullName, showEmail bool) string {
	return Etag(u.Id, u.UpdateAt, u.TermsOfServiceId, u.TermsOfServiceCreateAt, showFullName, showEmail, u.BotLastIconUpdate)
}

// Remove any private data from the user object
func (u *User) Sanitize(options map[string]bool) {
	u.Password = ""
	u.AuthData = NewString("")
	u.MfaSecret = ""

	if len(options) != 0 && !options["email"] {
		u.Email = ""
	}
	if len(options) != 0 && !options["fullname"] {
		u.FirstName = ""
		u.LastName = ""
	}
	if len(options) != 0 && !options["passwordupdate"] {
		u.LastPasswordUpdate = 0
	}
	if len(options) != 0 && !options["authservice"] {
		u.AuthService = ""
	}
}

// Remove any input data from the user object that is not user controlled
func (u *User) SanitizeInput(isAdmin bool) {
	if !isAdmin {
		u.AuthData = NewString("")
		u.AuthService = ""
	}
	u.LastPasswordUpdate = 0
	u.LastPictureUpdate = 0
	u.FailedAttempts = 0
	u.EmailVerified = false
	u.MfaActive = false
	u.MfaSecret = ""
}

func (u *User) ClearNonProfileFields() {
	u.Password = ""
	u.AuthData = NewString("")
	u.MfaSecret = ""
	u.EmailVerified = false
	u.AllowMarketing = false
	u.NotifyProps = StringMap{}
	u.LastPasswordUpdate = 0
	u.FailedAttempts = 0
}

func (u *User) SanitizeProfile(options map[string]bool) {
	u.ClearNonProfileFields()

	u.Sanitize(options)
}

func (u *User) MakeNonNil() {
	if u.Props == nil {
		u.Props = make(map[string]string)
	}

	if u.NotifyProps == nil {
		u.NotifyProps = make(map[string]string)
	}
}

func (u *User) AddNotifyProp(key string, value string) {
	u.MakeNonNil()

	u.NotifyProps[key] = value
}

func (u *User) GetFullName() string {
	if len(u.FirstName) > 0 && len(u.LastName) > 0 {
		return u.FirstName + " " + u.LastName
	} else if len(u.FirstName) > 0 {
		return u.FirstName
	} else if len(u.LastName) > 0 {
		return u.LastName
	} else {
		return ""
	}
}

func (u *User) getDisplayName(baseName, nameFormat string) string {
	displayName := baseName

	if nameFormat == SHOW_NICKNAME_FULLNAME {
		if len(u.Nickname) > 0 {
			displayName = u.Nickname
		} else if fullName := u.GetFullName(); len(fullName) > 0 {
			displayName = fullName
		}
	} else if nameFormat == SHOW_FULLNAME {
		if fullName := u.GetFullName(); len(fullName) > 0 {
			displayName = fullName
		}
	}

	return displayName
}

func (u *User) GetDisplayName(nameFormat string) string {
	displayName := u.Username

	return u.getDisplayName(displayName, nameFormat)
}

func (u *User) GetDisplayNameWithPrefix(nameFormat, prefix string) string {
	displayName := prefix + u.Username

	return u.getDisplayName(displayName, nameFormat)
}

func (u *User) GetRoles() []string {
	return strings.Fields(u.Roles)
}

func (u *User) GetRawRoles() string {
	return u.Roles
}

func IsValidUserRoles(userRoles string) bool {

	roles := strings.Fields(userRoles)

	for _, r := range roles {
		if !IsValidRoleName(r) {
			return false
		}
	}

	// Exclude just the system_admin role explicitly to prevent mistakes
	if len(roles) == 1 && roles[0] == "system_admin" {
		return false
	}

	return true
}

// Make sure you acually want to use this function. In context.go there are functions to check permissions
// This function should not be used to check permissions.
func (u *User) IsGuest() bool {
	return IsInRole(u.Roles, SYSTEM_GUEST_ROLE_ID)
}

func (u *User) IsSystemAdmin() bool {
	return IsInRole(u.Roles, SYSTEM_ADMIN_ROLE_ID)
}

// Make sure you acually want to use this function. In context.go there are functions to check permissions
// This function should not be used to check permissions.
func (u *User) IsInRole(inRole string) bool {
	return IsInRole(u.Roles, inRole)
}

// Make sure you acually want to use this function. In context.go there are functions to check permissions
// This function should not be used to check permissions.
func IsInRole(userRoles string, inRole string) bool {
	roles := strings.Split(userRoles, " ")

	for _, r := range roles {
		if r == inRole {
			return true
		}
	}

	return false
}

func (u *User) IsSSOUser() bool {
	return u.AuthService != "" && u.AuthService != USER_AUTH_SERVICE_EMAIL
}

func (u *User) IsOAuthUser() bool {
	return u.AuthService == USER_AUTH_SERVICE_GITLAB
}

func (u *User) IsLDAPUser() bool {
	return u.AuthService == USER_AUTH_SERVICE_LDAP
}

func (u *User) IsSAMLUser() bool {
	return u.AuthService == USER_AUTH_SERVICE_SAML
}

func (u *User) GetPreferredTimezone() string {
	return GetPreferredTimezone(u.Timezone)
}

// UserFromJson will decode the input and return a User
func UserFromJson(data io.Reader) *User {
	var user *User
	json.NewDecoder(data).Decode(&user)
	return user
}

func UserPatchFromJson(data io.Reader) *UserPatch {
	var user *UserPatch
	json.NewDecoder(data).Decode(&user)
	return user
}

func UserAuthFromJson(data io.Reader) *UserAuth {
	var user *UserAuth
	json.NewDecoder(data).Decode(&user)
	return user
}

func UserMapToJson(u map[string]*User) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func UserMapFromJson(data io.Reader) map[string]*User {
	var users map[string]*User
	json.NewDecoder(data).Decode(&users)
	return users
}

func UserListToJson(u []*User) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func UserListFromJson(data io.Reader) []*User {
	var users []*User
	json.NewDecoder(data).Decode(&users)
	return users
}

// HashPassword generates a hash using the bcrypt.GenerateFromPassword
func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}

	return string(hash)
}

// ComparePassword compares the hash
func ComparePassword(hash string, password string) bool {

	if len(password) == 0 || len(hash) == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

var validUsernameChars = regexp.MustCompile(`^[a-z0-9\.\-_]+$`)

var restrictedUsernames = []string{
	"all",
	"channel",
	"matterbot",
	"system",
}

func IsValidUsername(s string) bool {
	if len(s) < USER_NAME_MIN_LENGTH || len(s) > USER_NAME_MAX_LENGTH {
		return false
	}

	if !validUsernameChars.MatchString(s) {
		return false
	}

	for _, restrictedUsername := range restrictedUsernames {
		if s == restrictedUsername {
			return false
		}
	}

	return true
}

func CleanUsername(s string) string {
	s = NormalizeUsername(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if s == value {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validUsernameChars.MatchString(char) {
			s = strings.Replace(s, char, "-", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidUsername(s) {
		s = "a" + NewId()
	}

	return s
}

func IsValidUserNotifyLevel(notifyLevel string) bool {
	return notifyLevel == CHANNEL_NOTIFY_ALL ||
		notifyLevel == CHANNEL_NOTIFY_MENTION ||
		notifyLevel == CHANNEL_NOTIFY_NONE
}

func IsValidPushStatusNotifyLevel(notifyLevel string) bool {
	return notifyLevel == STATUS_ONLINE ||
		notifyLevel == STATUS_AWAY ||
		notifyLevel == STATUS_OFFLINE
}

func IsValidCommentsNotifyLevel(notifyLevel string) bool {
	return notifyLevel == COMMENTS_NOTIFY_ANY ||
		notifyLevel == COMMENTS_NOTIFY_ROOT ||
		notifyLevel == COMMENTS_NOTIFY_NEVER
}

func IsValidEmailBatchingInterval(emailInterval string) bool {
	return emailInterval == PREFERENCE_EMAIL_INTERVAL_IMMEDIATELY ||
		emailInterval == PREFERENCE_EMAIL_INTERVAL_FIFTEEN ||
		emailInterval == PREFERENCE_EMAIL_INTERVAL_HOUR
}

func IsValidLocale(locale string) bool {
	if locale != "" {
		if len(locale) > USER_LOCALE_MAX_LENGTH {
			return false
		} else if _, err := language.Parse(locale); err != nil {
			return false
		}
	}

	return true
}

type UserWithGroups struct {
	User
	GroupIDs    *string  `json:"-"`
	Groups      []*Group `json:"groups"`
	SchemeGuest bool     `json:"scheme_guest"`
	SchemeUser  bool     `json:"scheme_user"`
	SchemeAdmin bool     `json:"scheme_admin"`
}

func (u *UserWithGroups) GetGroupIDs() []string {
	if u.GroupIDs == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*u.GroupIDs)
	if len(trimmed) == 0 {
		return nil
	}
	return strings.Split(trimmed, ",")
}

type UsersWithGroupsAndCount struct {
	Users []*UserWithGroups `json:"users"`
	Count int64             `json:"total_count"`
}

func UsersWithGroupsAndCountFromJson(data io.Reader) *UsersWithGroupsAndCount {
	uwg := &UsersWithGroupsAndCount{}
	bodyBytes, _ := ioutil.ReadAll(data)
	json.Unmarshal(bodyBytes, uwg)
	return uwg
}

var passwordRandomSource = rand.NewSource(time.Now().Unix())
var passwordSpecialChars = "!$%^&*(),."
var passwordNumbers = "0123456789"
var passwordUpperCaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
var passwordLowerCaseLetters = "abcdefghijklmnopqrstuvwxyz"
var passwordAllChars = passwordSpecialChars + passwordNumbers + passwordUpperCaseLetters + passwordLowerCaseLetters

func GeneratePassword(minimumLength int) string {
	r := rand.New(passwordRandomSource)

	// Make sure we are guaranteed at least one of each type to meet any possible password complexity requirements.
	password := string([]rune(passwordUpperCaseLetters)[r.Intn(len(passwordUpperCaseLetters))]) +
		string([]rune(passwordNumbers)[r.Intn(len(passwordNumbers))]) +
		string([]rune(passwordLowerCaseLetters)[r.Intn(len(passwordLowerCaseLetters))]) +
		string([]rune(passwordSpecialChars)[r.Intn(len(passwordSpecialChars))])

	for len(password) < minimumLength {
		i := r.Intn(len(passwordAllChars))
		password = password + string([]rune(passwordAllChars)[i])
	}

	return password
}
