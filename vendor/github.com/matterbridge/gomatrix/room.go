package gomatrix

// Room represents a single Matrix room.
type Room struct {
	ID    string
	State map[string]map[string]*Event
}

// PublicRoom represents the information about a public room obtainable from the room directory
type PublicRoom struct {
	CanonicalAlias   string   `json:"canonical_alias"`
	Name             string   `json:"name"`
	WorldReadable    bool     `json:"world_readable"`
	Topic            string   `json:"topic"`
	NumJoinedMembers int      `json:"num_joined_members"`
	AvatarURL        string   `json:"avatar_url"`
	RoomID           string   `json:"room_id"`
	GuestCanJoin     bool     `json:"guest_can_join"`
	Aliases          []string `json:"aliases"`
}

// UpdateState updates the room's current state with the given Event. This will clobber events based
// on the type/state_key combination.
func (room Room) UpdateState(event *Event) {
	_, exists := room.State[event.Type]
	if !exists {
		room.State[event.Type] = make(map[string]*Event)
	}
	room.State[event.Type][*event.StateKey] = event
}

// GetStateEvent returns the state event for the given type/state_key combo, or nil.
func (room Room) GetStateEvent(eventType string, stateKey string) *Event {
	stateEventMap := room.State[eventType]
	event := stateEventMap[stateKey]
	return event
}

// GetMembershipState returns the membership state of the given user ID in this room. If there is
// no entry for this member, 'leave' is returned for consistency with left users.
func (room Room) GetMembershipState(userID string) string {
	state := "leave"
	event := room.GetStateEvent("m.room.member", userID)
	if event != nil {
		membershipState, found := event.Content["membership"]
		if found {
			mState, isString := membershipState.(string)
			if isString {
				state = mState
			}
		}
	}
	return state
}

// NewRoom creates a new Room with the given ID
func NewRoom(roomID string) *Room {
	// Init the State map and return a pointer to the Room
	return &Room{
		ID:    roomID,
		State: make(map[string]map[string]*Event),
	}
}
