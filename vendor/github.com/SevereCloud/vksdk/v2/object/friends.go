package object // import "github.com/SevereCloud/vksdk/v2/object"

// FriendsFriendStatus FriendStatus type.
const (
	FriendsStatusNotFriend        = iota // not a friend
	FriendsStatusOutComingRequest        // out coming request
	FriendsStatusInComingRequest         // incoming request
	FriendsStatusIsFriend                // is friend
)

// FriendsFriendStatus struct.
type FriendsFriendStatus struct {
	FriendStatus   int         `json:"friend_status"`
	ReadState      BaseBoolInt `json:"read_state"`      // Information whether request is unviewed
	RequestMessage string      `json:"request_message"` // Message sent with request
	Sign           string      `json:"sign"`            // MD5 hash for the result validation
	UserID         int         `json:"user_id"`         // User ID
}

// FriendsFriendsList struct.
type FriendsFriendsList struct {
	ID   int    `json:"id"`   // List ID
	Name string `json:"name"` // List title
}

// type friendsMutualFriend struct {
// 	CommonCount   int   `json:"common_count"` // Total mutual friends number
// 	CommonFriends []int `json:"common_friends"`
// 	ID            int   `json:"id"` // User ID
// }

// FriendsRequests struct.
type FriendsRequests struct {
	UsersUser
	From      string                `json:"from"` // ID of the user by whom friend has been suggested
	Mutual    FriendsRequestsMutual `json:"mutual"`
	UserID    int                   `json:"user_id"` // User ID
	TrackCode string                `json:"track_code"`
}

// FriendsRequestsMutual struct.
type FriendsRequestsMutual struct {
	Count int   `json:"count"` // Total mutual friends number
	Users []int `json:"users"`
}

// FriendsRequestsXtrMessage struct.
type FriendsRequestsXtrMessage struct {
	FriendsRequests
	Message string `json:"message"` // Message sent with a request
}

// FriendsUserXtrLists struct.
type FriendsUserXtrLists struct {
	UsersUser
	Lists []int `json:"lists"` // IDs of friend lists with user
}

// FriendsUserXtrPhone struct.
type FriendsUserXtrPhone struct {
	UsersUser
	Phone string `json:"phone"` // User phone
}
