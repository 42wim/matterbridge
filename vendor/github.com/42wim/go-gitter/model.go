package gitter

import "time"

// A Room in Gitter can represent a GitHub Organization, a GitHub Repository, a Gitter Channel or a One-to-one conversation.
// In the case of the Organizations and Repositories, the access control policies are inherited from GitHub.
type Room struct {

	// Room ID
	ID string `json:"id"`

	// Room name
	Name string `json:"name"`

	// Room topic. (default: GitHub repo description)
	Topic string `json:"topic"`

	// Room URI on Gitter
	URI string `json:"uri"`

	// Indicates if the room is a one-to-one chat
	OneToOne bool `json:"oneToOne"`

	// Count of users in the room
	UserCount int `json:"userCount"`

	// Number of unread messages for the current user
	UnreadItems int `json:"unreadItems"`

	// Number of unread mentions for the current user
	Mentions int `json:"mentions"`

	// Last time the current user accessed the room in ISO format
	LastAccessTime time.Time `json:"lastAccessTime"`

	// Indicates if the current user has disabled notifications
	Lurk bool `json:"lurk"`

	// Path to the room on gitter
	URL string `json:"url"`

	// Type of the room
	// - ORG: A room that represents a GitHub Organization.
	// - REPO: A room that represents a GitHub Repository.
	// - ONETOONE: A one-to-one chat.
	// - ORG_CHANNEL: A Gitter channel nested under a GitHub Organization.
	// - REPO_CHANNEL A Gitter channel nested under a GitHub Repository.
	// - USER_CHANNEL A Gitter channel nested under a GitHub User.
	GithubType string `json:"githubType"`

	// Tags that define the room
	Tags []string `json:"tags"`

	RoomMember bool `json:"roomMember"`

	// Room version.
	Version int `json:"v"`
}

type User struct {

	// Gitter User ID
	ID string `json:"id"`

	// Gitter/GitHub username
	Username string `json:"username"`

	// Gitter/GitHub user real name
	DisplayName string `json:"displayName"`

	// Path to the user on Gitter
	URL string `json:"url"`

	// User avatar URI (small)
	AvatarURLSmall string `json:"avatarUrlSmall"`

	// User avatar URI (medium)
	AvatarURLMedium string `json:"avatarUrlMedium"`
}

type Message struct {

	// ID of the message
	ID string `json:"id"`

	// Original message in plain-text/markdown
	Text string `json:"text"`

	// HTML formatted message
	HTML string `json:"html"`

	// ISO formatted date of the message
	Sent time.Time `json:"sent"`

	// ISO formatted date of the message if edited
	EditedAt time.Time `json:"editedAt"`

	// User that sent the message
	From User `json:"fromUser"`

	// Boolean that indicates if the current user has read the message.
	Unread bool `json:"unread"`

	// Number of users that have read the message
	ReadBy int `json:"readBy"`

	// List of URLs present in the message
	Urls []URL `json:"urls"`

	// List of @Mentions in the message
	Mentions []Mention `json:"mentions"`

	// List of #Issues referenced in the message
	Issues []Issue `json:"issues"`

	// Version
	Version int `json:"v"`
}

// Mention holds data about mentioned user in the message
type Mention struct {

	// User's username
	ScreenName string `json:"screenName"`

	// Gitter User ID
	UserID string `json:"userID"`
}

// Issue references issue in the message
type Issue struct {

	// Issue number
	Number string `json:"number"`
}

// URL presented in the message
type URL struct {

	// URL
	URL string `json:"url"`
}
