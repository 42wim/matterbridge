package wakusync

import (
	"encoding/json"

	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/protocol/protobuf"
)

type WakuBackedUpDataResponse struct {
	Clock                uint64
	FetchingDataProgress map[string]*protobuf.FetchingBackedUpDataDetails // key represents the data/section backup details refer to
	Profile              *BackedUpProfile
	Setting              *settings.SyncSettingField
	Keypair              *accounts.Keypair
	WatchOnlyAccount     *accounts.Account
}

func (sfwr *WakuBackedUpDataResponse) MarshalJSON() ([]byte, error) {
	responseItem := struct {
		Clock                uint64                                 `json:"clock,omitempty"`
		FetchingDataProgress map[string]FetchingBackupedDataDetails `json:"fetchingBackedUpDataProgress,omitempty"`
		Profile              *BackedUpProfile                       `json:"backedUpProfile,omitempty"`
		Setting              *settings.SyncSettingField             `json:"backedUpSettings,omitempty"`
		Keypair              *accounts.Keypair                      `json:"backedUpKeypair,omitempty"`
		WatchOnlyAccount     *accounts.Account                      `json:"backedUpWatchOnlyAccount,omitempty"`
	}{
		Clock:            sfwr.Clock,
		Profile:          sfwr.Profile,
		Setting:          sfwr.Setting,
		Keypair:          sfwr.Keypair,
		WatchOnlyAccount: sfwr.WatchOnlyAccount,
	}

	responseItem.FetchingDataProgress = sfwr.FetchingBackedUpDataDetails()

	return json.Marshal(responseItem)
}
