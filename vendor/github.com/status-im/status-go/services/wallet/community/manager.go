package community

import (
	"database/sql"
	"encoding/json"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

// These events are used to notify the UI of state changes
const (
	EventCommmunityDataUpdated walletevent.EventType = "wallet-community-data-updated"
)

type Manager struct {
	db                    *DataDB
	communityInfoProvider thirdparty.CommunityInfoProvider
	mediaServer           *server.MediaServer
	feed                  *event.Feed
}

type Data struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Image string `json:"image,omitempty"`
}

func NewManager(db *sql.DB, mediaServer *server.MediaServer, feed *event.Feed) *Manager {
	return &Manager{
		db:          NewDataDB(db),
		mediaServer: mediaServer,
		feed:        feed,
	}
}

// Used to break circular dependency, call once as soon as possible after initialization
func (cm *Manager) SetCommunityInfoProvider(communityInfoProvider thirdparty.CommunityInfoProvider) {
	cm.communityInfoProvider = communityInfoProvider
}

func (cm *Manager) GetCommunityInfo(id string) (*thirdparty.CommunityInfo, *InfoState, error) {
	communityInfo, state, err := cm.db.GetCommunityInfo(id)
	if err != nil {
		return nil, nil, err
	}
	if cm.mediaServer != nil && communityInfo != nil && len(communityInfo.CommunityImagePayload) > 0 {
		communityInfo.CommunityImage = cm.GetCommunityImageURL(id)
	}
	return communityInfo, state, err
}

func (cm *Manager) GetCommunityID(tokenURI string) string {
	return cm.communityInfoProvider.GetCommunityID(tokenURI)
}

func (cm *Manager) FillCollectibleMetadata(c *thirdparty.FullCollectibleData) error {
	return cm.communityInfoProvider.FillCollectibleMetadata(c)
}

func (cm *Manager) setCommunityInfo(id string, c *thirdparty.CommunityInfo) (err error) {
	return cm.db.SetCommunityInfo(id, c)
}

func (cm *Manager) FetchCommunityInfo(communityID string) (*thirdparty.CommunityInfo, error) {
	communityInfo, err := cm.communityInfoProvider.FetchCommunityInfo(communityID)
	if err != nil {
		dbErr := cm.setCommunityInfo(communityID, nil)
		if dbErr != nil {
			log.Error("SetCommunityInfo failed", "communityID", communityID, "err", dbErr)
		}
		return nil, err
	}
	err = cm.setCommunityInfo(communityID, communityInfo)
	return communityInfo, err
}

func (cm *Manager) FetchCommunityMetadataAsync(communityID string) {
	go func() {
		communityInfo, err := cm.FetchCommunityMetadata(communityID)
		if err != nil {
			log.Error("FetchCommunityInfo failed", "communityID", communityID, "err", err)
		}
		cm.signalUpdatedCommunityMetadata(communityID, communityInfo)
	}()
}

func (cm *Manager) FetchCommunityMetadata(communityID string) (*thirdparty.CommunityInfo, error) {
	communityInfo, err := cm.FetchCommunityInfo(communityID)
	if err != nil {
		return nil, err
	}
	_ = cm.setCommunityInfo(communityID, communityInfo)
	return communityInfo, err
}

func (cm *Manager) GetCommunityImageURL(communityID string) string {
	if cm.mediaServer != nil {
		return cm.mediaServer.MakeWalletCommunityImagesURL(communityID)
	}
	return ""
}

func (cm *Manager) signalUpdatedCommunityMetadata(communityID string, communityInfo *thirdparty.CommunityInfo) {
	if communityInfo == nil {
		return
	}
	data := Data{
		ID:    communityID,
		Name:  communityInfo.CommunityName,
		Color: communityInfo.CommunityColor,
		Image: cm.GetCommunityImageURL(communityID),
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Error("Error marshaling response: %v", err)
		return
	}

	event := walletevent.Event{
		Type:    EventCommmunityDataUpdated,
		Message: string(payload),
	}

	cm.feed.Send(event)
}
