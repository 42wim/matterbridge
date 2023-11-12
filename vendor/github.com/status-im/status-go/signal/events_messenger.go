package signal

const (
	// EventMediaServerStarted triggers when the media server successfully binds a new port
	EventMediaServerStarted = "mediaserver.started"

	// EventMesssageDelivered triggered when we got acknowledge from datasync level, that means peer got message
	EventMesssageDelivered = "message.delivered"

	// EventCommunityInfoFound triggered when user requested info about some community and messenger successfully
	// retrieved it from mailserver
	EventCommunityInfoFound = "community.found"

	// EventStatusUpdatesTimedOut Event Automatic Status Updates Timed out
	EventStatusUpdatesTimedOut = "status.updates.timedout"

	// EventCuratedCommunitiesUpdate triggered when it is time to refresh the list of curated communities
	EventCuratedCommunitiesUpdate = "curated.communities.update"
)

// MessageDeliveredSignal specifies chat and message that was delivered
type MessageDeliveredSignal struct {
	ChatID    string `json:"chatID"`
	MessageID string `json:"messageID"`
}

// MediaServerStarted specifies chat and message that was delivered
type MediaServerStarted struct {
	Port int `json:"port"`
}

// MessageDeliveredSignal specifies chat and message that was delivered
type CommunityInfoFoundSignal struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	MembersCount int    `json:"membersCount"`
	Verified     bool   `json:"verified"`
}

// SendMessageDelivered notifies about delivered message
func SendMessageDelivered(chatID string, messageID string) {
	send(EventMesssageDelivered, MessageDeliveredSignal{ChatID: chatID, MessageID: messageID})
}

// SendMediaServerStarted notifies about restarts of the media server
func SendMediaServerStarted(port int) {
	send(EventMediaServerStarted, MediaServerStarted{Port: port})
}

// SendMessageDelivered notifies about delivered message
func SendCommunityInfoFound(community interface{}) {
	send(EventCommunityInfoFound, community)
}

func SendStatusUpdatesTimedOut(statusUpdates interface{}) {
	send(EventStatusUpdatesTimedOut, statusUpdates)
}

func SendCuratedCommunitiesUpdate(curatedCommunitiesUpdate interface{}) {
	send(EventCuratedCommunitiesUpdate, curatedCommunitiesUpdate)
}
