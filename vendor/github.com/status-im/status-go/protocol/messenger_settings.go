package protocol

import (
	"context"

	"github.com/status-im/status-go/nodecfg"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/timesource"
)

func (m *Messenger) SetLightClient(request *requests.SetLightClient) error {
	return nodecfg.SetLightClient(m.database, request.Enabled)
}

func (m *Messenger) SetLogLevel(request *requests.SetLogLevel) error {
	if err := request.Validate(); err != nil {
		return err
	}

	return nodecfg.SetLogLevel(m.database, request.LogLevel)
}

func (m *Messenger) SetCustomNodes(request *requests.SetCustomNodes) error {
	return nodecfg.SetWakuV2CustomNodes(m.database, request.CustomNodes)
}

func (m *Messenger) SetCustomizationColor(ctx context.Context, request *requests.SetCustomizationColor) error {
	if err := request.Validate(); err != nil {
		return err
	}

	acc, err := m.multiAccounts.GetAccount(request.KeyUID)
	if err != nil {
		return err
	}

	acc.CustomizationColor = request.CustomizationColor

	//Use a combination of wall clock + lamport timestamp, just like Chat#NextClockAndTimestamp
	tNow := timesource.GetCurrentTimeInMillis()
	if acc.CustomizationColorClock >= tNow {
		acc.CustomizationColorClock++
	} else {
		acc.CustomizationColorClock = tNow
	}

	affected, err := m.multiAccounts.UpdateAccountCustomizationColor(request.KeyUID, string(acc.CustomizationColor), acc.CustomizationColorClock)
	if err != nil {
		return err
	}

	if affected > 0 {
		err = m.syncAccountCustomizationColor(ctx, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Messenger) TogglePeerSyncing(request *requests.TogglePeerSyncingRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}

	err := m.settings.SetPeerSyncingEnabled(request.Enabled)
	if err != nil {
		return err
	}
	return nil
}
