// inner_events.go provides EventsAPI particular inner events

package slackevents

import (
	"github.com/slack-go/slack"
)

// EventsAPIInnerEvent the inner event of a EventsAPI event_callback Event.
type EventsAPIInnerEvent struct {
	Type string `json:"type"`
	Data interface{}
}

// AppMentionEvent is an (inner) EventsAPI subscribable event.
type AppMentionEvent struct {
	Type            string `json:"type"`
	User            string `json:"user"`
	Text            string `json:"text"`
	TimeStamp       string `json:"ts"`
	ThreadTimeStamp string `json:"thread_ts"`
	Channel         string `json:"channel"`
	EventTimeStamp  string `json:"event_ts"`

	// When Message comes from a channel that is shared between workspaces
	UserTeam   string `json:"user_team,omitempty"`
	SourceTeam string `json:"source_team,omitempty"`

	// BotID is filled out when a bot triggers the app_mention event
	BotID string `json:"bot_id,omitempty"`

	// When the app is mentioned in the edited message
	Edited *Edited `json:"edited,omitempty"`
}

// AppHomeOpenedEvent Your Slack app home was opened.
type AppHomeOpenedEvent struct {
	Type           string     `json:"type"`
	User           string     `json:"user"`
	Channel        string     `json:"channel"`
	EventTimeStamp string     `json:"event_ts"`
	Tab            string     `json:"tab"`
	View           slack.View `json:"view"`
}

// AppUninstalledEvent Your Slack app was uninstalled.
type AppUninstalledEvent struct {
	Type string `json:"type"`
}

// ChannelCreatedEvent represents the Channel created event
type ChannelCreatedEvent struct {
	Type           string             `json:"type"`
	Channel        ChannelCreatedInfo `json:"channel"`
	EventTimestamp string             `json:"event_ts"`
}

// ChannelDeletedEvent represents the Channel deleted event
type ChannelDeletedEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	EventTimestamp string `json:"event_ts"`
}

// ChannelArchiveEvent represents the Channel archive event
type ChannelArchiveEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	User           string `json:"user"`
	EventTimestamp string `json:"event_ts"`
}

// ChannelUnarchiveEvent represents the Channel unarchive event
type ChannelUnarchiveEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	User           string `json:"user"`
	EventTimestamp string `json:"event_ts"`
}

// ChannelLeftEvent represents the Channel left event
type ChannelLeftEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	EventTimestamp string `json:"event_ts"`
}

// ChannelRenameEvent represents the Channel rename event
type ChannelRenameEvent struct {
	Type           string            `json:"type"`
	Channel        ChannelRenameInfo `json:"channel"`
	EventTimestamp string            `json:"event_ts"`
}

// ChannelIDChangedEvent represents the Channel identifier changed event
type ChannelIDChangedEvent struct {
	Type           string `json:"type"`
	OldChannelID   string `json:"old_channel_id"`
	NewChannelID   string `json:"new_channel_id"`
	EventTimestamp string `json:"event_ts"`
}

// ChannelCreatedInfo represents the information associated with the Channel created event
type ChannelCreatedInfo struct {
	ID        string `json:"id"`
	IsChannel bool   `json:"is_channel"`
	Name      string `json:"name"`
	Created   int    `json:"created"`
	Creator   string `json:"creator"`
}

// ChannelRenameInfo represents the information associated with the Channel rename event
type ChannelRenameInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Created int    `json:"created"`
}

// GroupDeletedEvent represents the Group deleted event
type GroupDeletedEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	EventTimestamp string `json:"event_ts"`
}

// GroupArchiveEvent represents the Group archive event
type GroupArchiveEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	EventTimestamp string `json:"event_ts"`
}

// GroupUnarchiveEvent represents the Group unarchive event
type GroupUnarchiveEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	EventTimestamp string `json:"event_ts"`
}

// GroupLeftEvent represents the Group left event
type GroupLeftEvent struct {
	Type           string `json:"type"`
	Channel        string `json:"channel"`
	EventTimestamp string `json:"event_ts"`
}

// GroupRenameEvent represents the Group rename event
type GroupRenameEvent struct {
	Type           string          `json:"type"`
	Channel        GroupRenameInfo `json:"channel"`
	EventTimestamp string          `json:"event_ts"`
}

// GroupRenameInfo represents the information associated with the Group rename event
type GroupRenameInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Created int    `json:"created"`
}

// FileChangeEvent represents the information associated with the File change
// event.
type FileChangeEvent struct {
	Type   string        `json:"type"`
	FileID string        `json:"file_id"`
	File   FileEventFile `json:"file"`
}

// FileDeletedEvent represents the information associated with the File deleted
// event.
type FileDeletedEvent struct {
	Type           string `json:"type"`
	FileID         string `json:"file_id"`
	EventTimestamp string `json:"event_ts"`
}

// FileSharedEvent represents the information associated with the File shared
// event.
type FileSharedEvent struct {
	Type           string        `json:"type"`
	ChannelID      string        `json:"channel_id"`
	FileID         string        `json:"file_id"`
	UserID         string        `json:"user_id"`
	File           FileEventFile `json:"file"`
	EventTimestamp string        `json:"event_ts"`
}

// FileUnsharedEvent represents the information associated with the File
// unshared event.
type FileUnsharedEvent struct {
	Type   string        `json:"type"`
	FileID string        `json:"file_id"`
	File   FileEventFile `json:"file"`
}

// FileEventFile represents information on the specific file being shared in a
// file-related Slack event.
type FileEventFile struct {
	ID string `json:"id"`
}

// GridMigrationFinishedEvent An enterprise grid migration has finished on this workspace.
type GridMigrationFinishedEvent struct {
	Type         string `json:"type"`
	EnterpriseID string `json:"enterprise_id"`
}

// GridMigrationStartedEvent An enterprise grid migration has started on this workspace.
type GridMigrationStartedEvent struct {
	Type         string `json:"type"`
	EnterpriseID string `json:"enterprise_id"`
}

// LinkSharedEvent A message was posted containing one or more links relevant to your application
type LinkSharedEvent struct {
	Type      string `json:"type"`
	User      string `json:"user"`
	TimeStamp string `json:"ts"`
	Channel   string `json:"channel"`
	// MessageTimeStamp can be both a numeric timestamp if the LinkSharedEvent corresponds to a sent
	// message and (contrary to the field name) a uuid if the LinkSharedEvent is generated in the
	// compose text area.
	MessageTimeStamp string        `json:"message_ts"`
	ThreadTimeStamp  string        `json:"thread_ts"`
	Links            []SharedLinks `json:"links"`
	EventTimestamp   string        `json:"event_ts"`
}

type SharedLinks struct {
	Domain string `json:"domain"`
	URL    string `json:"url"`
}

// MessageEvent occurs when a variety of types of messages has been posted.
// Parse ChannelType to see which
// if ChannelType = "group", this is a private channel message
// if ChannelType = "channel", this message was sent to a channel
// if ChannelType = "im", this is a private message
// if ChannelType = "mim", A message was posted in a multiparty direct message channel
// TODO: Improve this so that it is not required to manually parse ChannelType
type MessageEvent struct {
	// Basic Message Event - https://api.slack.com/events/message
	ClientMsgID     string `json:"client_msg_id"`
	Type            string `json:"type"`
	User            string `json:"user"`
	Text            string `json:"text"`
	ThreadTimeStamp string `json:"thread_ts"`
	TimeStamp       string `json:"ts"`
	Channel         string `json:"channel"`
	ChannelType     string `json:"channel_type"`
	EventTimeStamp  string `json:"event_ts"`

	// When Message comes from a channel that is shared between workspaces
	UserTeam   string `json:"user_team,omitempty"`
	SourceTeam string `json:"source_team,omitempty"`

	// Edited Message
	Message         *MessageEvent `json:"message,omitempty"`
	PreviousMessage *MessageEvent `json:"previous_message,omitempty"`
	Edited          *Edited       `json:"edited,omitempty"`

	// Deleted Message
	DeletedTimeStamp string `json:"deleted_ts,omitempty"`

	// Message Subtypes
	SubType string `json:"subtype,omitempty"`

	// bot_message (https://api.slack.com/events/message/bot_message)
	BotID    string `json:"bot_id,omitempty"`
	Username string `json:"username,omitempty"`
	Icons    *Icon  `json:"icons,omitempty"`

	Upload bool   `json:"upload"`
	Files  []File `json:"files"`

	Blocks      slack.Blocks       `json:"blocks,omitempty"`
	Attachments []slack.Attachment `json:"attachments,omitempty"`

	// Root is the message that was broadcast to the channel when the SubType is
	// thread_broadcast. If this is not a thread_broadcast message event, this
	// value is nil.
	Root *MessageEvent `json:"root"`
}

// MemberJoinedChannelEvent A member joined a public or private channel
type MemberJoinedChannelEvent struct {
	Type           string `json:"type"`
	User           string `json:"user"`
	Channel        string `json:"channel"`
	ChannelType    string `json:"channel_type"`
	Team           string `json:"team"`
	Inviter        string `json:"inviter"`
	EventTimestamp string `json:"event_ts"`
}

// MemberLeftChannelEvent A member left a public or private channel
type MemberLeftChannelEvent struct {
	Type           string `json:"type"`
	User           string `json:"user"`
	Channel        string `json:"channel"`
	ChannelType    string `json:"channel_type"`
	Team           string `json:"team"`
	EventTimestamp string `json:"event_ts"`
}

type pinEvent struct {
	Type           string `json:"type"`
	User           string `json:"user"`
	Item           Item   `json:"item"`
	Channel        string `json:"channel_id"`
	EventTimestamp string `json:"event_ts"`
	HasPins        bool   `json:"has_pins,omitempty"`
}

type reactionEvent struct {
	Type           string `json:"type"`
	User           string `json:"user"`
	Reaction       string `json:"reaction"`
	ItemUser       string `json:"item_user"`
	Item           Item   `json:"item"`
	EventTimestamp string `json:"event_ts"`
}

// ReactionAddedEvent An reaction was added to a message - https://api.slack.com/events/reaction_added
type ReactionAddedEvent reactionEvent

// ReactionRemovedEvent An reaction was removed from a message - https://api.slack.com/events/reaction_removed
type ReactionRemovedEvent reactionEvent

// PinAddedEvent An item was pinned to a channel - https://api.slack.com/events/pin_added
type PinAddedEvent pinEvent

// PinRemovedEvent An item was unpinned from a channel - https://api.slack.com/events/pin_removed
type PinRemovedEvent pinEvent

type tokens struct {
	Oauth []string `json:"oauth"`
	Bot   []string `json:"bot"`
}

// TeamJoinEvent A new member joined a workspace -  https://api.slack.com/events/team_join
type TeamJoinEvent struct {
	Type           string      `json:"type"`
	User           *slack.User `json:"user"`
	EventTimestamp string      `json:"event_ts"`
}

// TokensRevokedEvent APP's API tokens are revoked - https://api.slack.com/events/tokens_revoked
type TokensRevokedEvent struct {
	Type           string `json:"type"`
	Tokens         tokens `json:"tokens"`
	EventTimestamp string `json:"event_ts"`
}

// EmojiChangedEvent is the event of custom emoji has been added or changed
type EmojiChangedEvent struct {
	Type           string `json:"type"`
	Subtype        string `json:"subtype"`
	EventTimeStamp string `json:"event_ts"`

	// filled out when custom emoji added
	Name string `json:"name,omitempty"`

	// filled out when custom emoji removed
	Names []string `json:"names,omitempty"`

	// filled out when custom emoji renamed
	OldName string `json:"old_name,omitempty"`
	NewName string `json:"new_name,omitempty"`

	// filled out when custom emoji added or renamed
	Value string `json:"value,omitempty"`
}

// WorkflowStepExecuteEvent is fired, if a workflow step of your app is invoked
type WorkflowStepExecuteEvent struct {
	Type           string            `json:"type"`
	CallbackID     string            `json:"callback_id"`
	WorkflowStep   EventWorkflowStep `json:"workflow_step"`
	EventTimestamp string            `json:"event_ts"`
}

// MessageMetadataPostedEvent is sent, if a message with metadata is posted
type MessageMetadataPostedEvent struct {
	Type             string               `json:"type"`
	AppId            string               `json:"app_id"`
	BotId            string               `json:"bot_id"`
	UserId           string               `json:"user_id"`
	TeamId           string               `json:"team_id"`
	ChannelId        string               `json:"channel_id"`
	Metadata         *slack.SlackMetadata `json:"metadata"`
	MessageTimestamp string               `json:"message_ts"`
	EventTimestamp   string               `json:"event_ts"`
}

// MessageMetadataUpdatedEvent is sent, if a message with metadata is deleted
type MessageMetadataUpdatedEvent struct {
	Type             string               `json:"type"`
	ChannelId        string               `json:"channel_id"`
	EventTimestamp   string               `json:"event_ts"`
	PreviousMetadata *slack.SlackMetadata `json:"previous_metadata"`
	AppId            string               `json:"app_id"`
	BotId            string               `json:"bot_id"`
	UserId           string               `json:"user_id"`
	TeamId           string               `json:"team_id"`
	MessageTimestamp string               `json:"message_ts"`
	Metadata         *slack.SlackMetadata `json:"metadata"`
}

// MessageMetadataDeletedEvent is sent, if a message with metadata is deleted
type MessageMetadataDeletedEvent struct {
	Type             string               `json:"type"`
	ChannelId        string               `json:"channel_id"`
	EventTimestamp   string               `json:"event_ts"`
	PreviousMetadata *slack.SlackMetadata `json:"previous_metadata"`
	AppId            string               `json:"app_id"`
	BotId            string               `json:"bot_id"`
	UserId           string               `json:"user_id"`
	TeamId           string               `json:"team_id"`
	MessageTimestamp string               `json:"message_ts"`
	DeletedTimestamp string               `json:"deleted_ts"`
}

type EventWorkflowStep struct {
	WorkflowStepExecuteID string                      `json:"workflow_step_execute_id"`
	WorkflowID            string                      `json:"workflow_id"`
	WorkflowInstanceID    string                      `json:"workflow_instance_id"`
	StepID                string                      `json:"step_id"`
	Inputs                *slack.WorkflowStepInputs   `json:"inputs,omitempty"`
	Outputs               *[]slack.WorkflowStepOutput `json:"outputs,omitempty"`
}

// JSONTime exists so that we can have a String method converting the date
type JSONTime int64

// Comment contains all the information relative to a comment
type Comment struct {
	ID        string   `json:"id,omitempty"`
	Created   JSONTime `json:"created,omitempty"`
	Timestamp JSONTime `json:"timestamp,omitempty"`
	User      string   `json:"user,omitempty"`
	Comment   string   `json:"comment,omitempty"`
}

// File is a file upload
type File struct {
	ID                 string `json:"id"`
	Created            int    `json:"created"`
	Timestamp          int    `json:"timestamp"`
	Name               string `json:"name"`
	Title              string `json:"title"`
	Mimetype           string `json:"mimetype"`
	Filetype           string `json:"filetype"`
	PrettyType         string `json:"pretty_type"`
	User               string `json:"user"`
	Editable           bool   `json:"editable"`
	Size               int    `json:"size"`
	Mode               string `json:"mode"`
	IsExternal         bool   `json:"is_external"`
	ExternalType       string `json:"external_type"`
	IsPublic           bool   `json:"is_public"`
	PublicURLShared    bool   `json:"public_url_shared"`
	DisplayAsBot       bool   `json:"display_as_bot"`
	Username           string `json:"username"`
	URLPrivate         string `json:"url_private"`
	URLPrivateDownload string `json:"url_private_download"`
	Thumb64            string `json:"thumb_64"`
	Thumb80            string `json:"thumb_80"`
	Thumb360           string `json:"thumb_360"`
	Thumb360W          int    `json:"thumb_360_w"`
	Thumb360H          int    `json:"thumb_360_h"`
	Thumb480           string `json:"thumb_480"`
	Thumb480W          int    `json:"thumb_480_w"`
	Thumb480H          int    `json:"thumb_480_h"`
	Thumb160           string `json:"thumb_160"`
	Thumb720           string `json:"thumb_720"`
	Thumb720W          int    `json:"thumb_720_w"`
	Thumb720H          int    `json:"thumb_720_h"`
	Thumb800           string `json:"thumb_800"`
	Thumb800W          int    `json:"thumb_800_w"`
	Thumb800H          int    `json:"thumb_800_h"`
	Thumb960           string `json:"thumb_960"`
	Thumb960W          int    `json:"thumb_960_w"`
	Thumb960H          int    `json:"thumb_960_h"`
	Thumb1024          string `json:"thumb_1024"`
	Thumb1024W         int    `json:"thumb_1024_w"`
	Thumb1024H         int    `json:"thumb_1024_h"`
	ImageExifRotation  int    `json:"image_exif_rotation"`
	OriginalW          int    `json:"original_w"`
	OriginalH          int    `json:"original_h"`
	Permalink          string `json:"permalink"`
	PermalinkPublic    string `json:"permalink_public"`
}

// Edited is included when a Message is edited
type Edited struct {
	User      string `json:"user"`
	TimeStamp string `json:"ts"`
}

// Icon is used for bot messages
type Icon struct {
	IconURL   string `json:"icon_url,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// Item is any type of slack message - message, file, or file comment.
type Item struct {
	Type      string       `json:"type"`
	Channel   string       `json:"channel,omitempty"`
	Message   *ItemMessage `json:"message,omitempty"`
	File      *File        `json:"file,omitempty"`
	Comment   *Comment     `json:"comment,omitempty"`
	Timestamp string       `json:"ts,omitempty"`
}

// ItemMessage is the event message
type ItemMessage struct {
	Type            string   `json:"type"`
	User            string   `json:"user"`
	Text            string   `json:"text"`
	Timestamp       string   `json:"ts"`
	PinnedTo        []string `json:"pinned_to"`
	ReplaceOriginal bool     `json:"replace_original"`
	DeleteOriginal  bool     `json:"delete_original"`
}

// IsEdited checks if the MessageEvent is caused by an edit
func (e MessageEvent) IsEdited() bool {
	return e.Message != nil &&
		e.Message.Edited != nil
}

// TeamAccessGrantedEvent is sent if access to teams was granted for your org-wide app.
type TeamAccessGrantedEvent struct {
	Type    string   `json:"type"`
	TeamIDs []string `json:"team_ids"`
}

// TeamAccessRevokedEvent is sent if access to teams was revoked for your org-wide app.
type TeamAccessRevokedEvent struct {
	Type    string   `json:"type"`
	TeamIDs []string `json:"team_ids"`
}

// UserProfileChangedEvent is sent if access to teams was revoked for your org-wide app.
type UserProfileChangedEvent struct {
	User    *slack.User `json:"user"`
	CacheTs int         `json:"cache_ts"`
	Type    string      `json:"type"`
	EventTs string      `json:"event_ts"`
}

// SharedChannelInviteApprovedEvent is sent if your invitation has been approved
type SharedChannelInviteApprovedEvent struct {
	Type            string              `json:"type"`
	Invite          *SharedInvite       `json:"invite"`
	Channel         *slack.Conversation `json:"channel"`
	ApprovingTeamID string              `json:"approving_team_id"`
	TeamsInChannel  []*SlackEventTeam   `json:"teams_in_channel"`
	ApprovingUser   *SlackEventUser     `json:"approving_user"`
	EventTs         string              `json:"event_ts"`
}

// SharedChannelInviteAcceptedEvent is sent if external org accepts a Slack Connect channel invite
type SharedChannelInviteAcceptedEvent struct {
	Type                string            `json:"type"`
	ApprovalRequired    bool              `json:"approval_required"`
	Invite              *SharedInvite     `json:"invite"`
	Channel             *SharedChannel    `json:"channel"`
	TeamsInChannel      []*SlackEventTeam `json:"teams_in_channel"`
	AcceptingUser       *SlackEventUser   `json:"accepting_user"`
	EventTs             string            `json:"event_ts"`
	RequiresSponsorship bool              `json:"requires_sponsorship,omitempty"`
}

// SharedChannelInviteDeclinedEvent is sent if external or internal org declines the Slack Connect invite
type SharedChannelInviteDeclinedEvent struct {
	Type            string            `json:"type"`
	Invite          *SharedInvite     `json:"invite"`
	Channel         *SharedChannel    `json:"channel"`
	DecliningTeamID string            `json:"declining_team_id"`
	TeamsInChannel  []*SlackEventTeam `json:"teams_in_channel"`
	DecliningUser   *SlackEventUser   `json:"declining_user"`
	EventTs         string            `json:"event_ts"`
}

// SharedChannelInviteReceivedEvent is sent if a bot or app is invited to a Slack Connect channel
type SharedChannelInviteReceivedEvent struct {
	Type    string         `json:"type"`
	Invite  *SharedInvite  `json:"invite"`
	Channel *SharedChannel `json:"channel"`
	EventTs string         `json:"event_ts"`
}

// SlackEventTeam is a struct for teams in ShareChannel events
type SlackEventTeam struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Icon                *SlackEventIcon `json:"icon,omitempty"`
	AvatarBaseURL       string          `json:"avatar_base_url,omitempty"`
	IsVerified          bool            `json:"is_verified"`
	Domain              string          `json:"domain"`
	DateCreated         int             `json:"date_created"`
	RequiresSponsorship bool            `json:"requires_sponsorship,omitempty"`
	// TeamID              string          `json:"team_id,omitempty"`
}

// SlackEventIcon is a struct for icons in ShareChannel events
type SlackEventIcon struct {
	ImageDefault bool   `json:"image_default,omitempty"`
	Image34      string `json:"image_34,omitempty"`
	Image44      string `json:"image_44,omitempty"`
	Image68      string `json:"image_68,omitempty"`
	Image88      string `json:"image_88,omitempty"`
	Image102     string `json:"image_102,omitempty"`
	Image132     string `json:"image_132,omitempty"`
	Image230     string `json:"image_230,omitempty"`
}

// SlackEventUser is a struct for users in ShareChannel events
type SlackEventUser struct {
	ID                     string             `json:"id"`
	TeamID                 string             `json:"team_id"`
	Name                   string             `json:"name"`
	Updated                int                `json:"updated,omitempty"`
	Profile                *slack.UserProfile `json:"profile,omitempty"`
	WhoCanShareContactCard string             `json:"who_can_share_contact_card,omitempty"`
}

// SharedChannel is a struct for shared channels in ShareChannel events
type SharedChannel struct {
	ID        string `json:"id"`
	IsPrivate bool   `json:"is_private"`
	IsIm      bool   `json:"is_im"`
	Name      string `json:"name,omitempty"`
}

// SharedInvite is a struct for shared invites in ShareChannel events
type SharedInvite struct {
	ID                string          `json:"id"`
	DateCreated       int             `json:"date_created"`
	DateInvalid       int             `json:"date_invalid"`
	InvitingTeam      *SlackEventTeam `json:"inviting_team,omitempty"`
	InvitingUser      *SlackEventUser `json:"inviting_user,omitempty"`
	RecipientEmail    string          `json:"recipient_email,omitempty"`
	RecipientUserID   string          `json:"recipient_user_id,omitempty"`
	IsSponsored       bool            `json:"is_sponsored,omitempty"`
	IsExternalLimited bool            `json:"is_external_limited,omitempty"`
}

type EventsAPIType string

const (
	// AppMention is an Events API subscribable event
	AppMention = EventsAPIType("app_mention")
	// AppHomeOpened Your Slack app home was opened
	AppHomeOpened = EventsAPIType("app_home_opened")
	// AppUninstalled Your Slack app was uninstalled.
	AppUninstalled = EventsAPIType("app_uninstalled")
	// ChannelCreated is sent when a new channel is created.
	ChannelCreated = EventsAPIType("channel_created")
	// ChannelDeleted is sent when a channel is deleted.
	ChannelDeleted = EventsAPIType("channel_deleted")
	// ChannelArchive is sent when a channel is archived.
	ChannelArchive = EventsAPIType("channel_archive")
	// ChannelUnarchive is sent when a channel is unarchived.
	ChannelUnarchive = EventsAPIType("channel_unarchive")
	// ChannelLeft is sent when a channel is left.
	ChannelLeft = EventsAPIType("channel_left")
	// ChannelRename is sent when a channel is rename.
	ChannelRename = EventsAPIType("channel_rename")
	// ChannelIDChanged is sent when a channel identifier is changed.
	ChannelIDChanged = EventsAPIType("channel_id_changed")
	// GroupDeleted is sent when a group is deleted.
	GroupDeleted = EventsAPIType("group_deleted")
	// GroupArchive is sent when a group is archived.
	GroupArchive = EventsAPIType("group_archive")
	// GroupUnarchive is sent when a group is unarchived.
	GroupUnarchive = EventsAPIType("group_unarchive")
	// GroupLeft is sent when a group is left.
	GroupLeft = EventsAPIType("group_left")
	// GroupRename is sent when a group is renamed.
	GroupRename = EventsAPIType("group_rename")
	// FileChange is sent when a file is changed.
	FileChange = EventsAPIType("file_change")
	// FileDeleted is sent when a file is deleted.
	FileDeleted = EventsAPIType("file_deleted")
	// FileShared is sent when a file is shared.
	FileShared = EventsAPIType("file_shared")
	// FileUnshared is sent when a file is unshared.
	FileUnshared = EventsAPIType("file_unshared")
	// GridMigrationFinished An enterprise grid migration has finished on this workspace.
	GridMigrationFinished = EventsAPIType("grid_migration_finished")
	// GridMigrationStarted An enterprise grid migration has started on this workspace.
	GridMigrationStarted = EventsAPIType("grid_migration_started")
	// LinkShared A message was posted containing one or more links relevant to your application
	LinkShared = EventsAPIType("link_shared")
	// Message A message was posted to a channel, private channel (group), im, or mim
	Message = EventsAPIType("message")
	// MemberJoinedChannel is sent if a member joined a channel.
	MemberJoinedChannel = EventsAPIType("member_joined_channel")
	// MemberLeftChannel is sent if a member left a channel.
	MemberLeftChannel = EventsAPIType("member_left_channel")
	// PinAdded An item was pinned to a channel
	PinAdded = EventsAPIType("pin_added")
	// PinRemoved An item was unpinned from a channel
	PinRemoved = EventsAPIType("pin_removed")
	// ReactionAdded An reaction was added to a message
	ReactionAdded = EventsAPIType("reaction_added")
	// ReactionRemoved An reaction was removed from a message
	ReactionRemoved = EventsAPIType("reaction_removed")
	// TeamJoin A new user joined the workspace
	TeamJoin = EventsAPIType("team_join")
	// Slack connect app or bot invite received
	SharedChannelInviteReceived = EventsAPIType("shared_channel_invite_received")
	// Slack connect channel invite approved
	SharedChannelInviteApproved = EventsAPIType("shared_channel_invite_approved")
	// Slack connect channel invite declined
	SharedChannelInviteDeclined = EventsAPIType("shared_channel_invite_declined")
	// Slack connect channel invite accepted by an end user
	SharedChannelInviteAccepted = EventsAPIType("shared_channel_invite_accepted")
	// TokensRevoked APP's API tokes are revoked
	TokensRevoked = EventsAPIType("tokens_revoked")
	// EmojiChanged A custom emoji has been added or changed
	EmojiChanged = EventsAPIType("emoji_changed")
	// WorkflowStepExecute Happens, if a workflow step of your app is invoked
	WorkflowStepExecute = EventsAPIType("workflow_step_execute")
	// MessageMetadataPosted A message with metadata was posted
	MessageMetadataPosted = EventsAPIType("message_metadata_posted")
	// MessageMetadataUpdated A message with metadata was updated
	MessageMetadataUpdated = EventsAPIType("message_metadata_updated")
	// MessageMetadataDeleted A message with metadata was deleted
	MessageMetadataDeleted = EventsAPIType("message_metadata_deleted")
	// TeamAccessGranted is sent if access to teams was granted for your org-wide app.
	TeamAccessGranted = EventsAPIType("team_access_granted")
	// TeamAccessRevoked is sent if access to teams was revoked for your org-wide app.
	TeamAccessRevoked = EventsAPIType("team_access_revoked")
	// UserProfileChanged is sent if a user's profile information has changed.
	UserProfileChanged = EventsAPIType("user_profile_changed")
)

// EventsAPIInnerEventMapping maps INNER Event API events to their corresponding struct
// implementations. The structs should be instances of the unmarshalling
// target for the matching event type.
var EventsAPIInnerEventMapping = map[EventsAPIType]interface{}{
	AppMention:                  AppMentionEvent{},
	AppHomeOpened:               AppHomeOpenedEvent{},
	AppUninstalled:              AppUninstalledEvent{},
	ChannelCreated:              ChannelCreatedEvent{},
	ChannelDeleted:              ChannelDeletedEvent{},
	ChannelArchive:              ChannelArchiveEvent{},
	ChannelUnarchive:            ChannelUnarchiveEvent{},
	ChannelLeft:                 ChannelLeftEvent{},
	ChannelRename:               ChannelRenameEvent{},
	ChannelIDChanged:            ChannelIDChangedEvent{},
	FileChange:                  FileChangeEvent{},
	FileDeleted:                 FileDeletedEvent{},
	FileShared:                  FileSharedEvent{},
	FileUnshared:                FileUnsharedEvent{},
	GroupDeleted:                GroupDeletedEvent{},
	GroupArchive:                GroupArchiveEvent{},
	GroupUnarchive:              GroupUnarchiveEvent{},
	GroupLeft:                   GroupLeftEvent{},
	GroupRename:                 GroupRenameEvent{},
	GridMigrationFinished:       GridMigrationFinishedEvent{},
	GridMigrationStarted:        GridMigrationStartedEvent{},
	LinkShared:                  LinkSharedEvent{},
	Message:                     MessageEvent{},
	MemberJoinedChannel:         MemberJoinedChannelEvent{},
	MemberLeftChannel:           MemberLeftChannelEvent{},
	PinAdded:                    PinAddedEvent{},
	PinRemoved:                  PinRemovedEvent{},
	ReactionAdded:               ReactionAddedEvent{},
	ReactionRemoved:             ReactionRemovedEvent{},
	SharedChannelInviteApproved: SharedChannelInviteApprovedEvent{},
	SharedChannelInviteAccepted: SharedChannelInviteAcceptedEvent{},
	SharedChannelInviteDeclined: SharedChannelInviteDeclinedEvent{},
	SharedChannelInviteReceived: SharedChannelInviteReceivedEvent{},
	TeamJoin:                    TeamJoinEvent{},
	TokensRevoked:               TokensRevokedEvent{},
	EmojiChanged:                EmojiChangedEvent{},
	WorkflowStepExecute:         WorkflowStepExecuteEvent{},
	MessageMetadataPosted:       MessageMetadataPostedEvent{},
	MessageMetadataUpdated:      MessageMetadataUpdatedEvent{},
	MessageMetadataDeleted:      MessageMetadataDeletedEvent{},
	TeamAccessGranted:           TeamAccessGrantedEvent{},
	TeamAccessRevoked:           TeamAccessRevokedEvent{},
	UserProfileChanged:          UserProfileChangedEvent{},
}
