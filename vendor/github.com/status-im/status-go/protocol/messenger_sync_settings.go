package protocol

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/status-im/status-go/multiaccounts/errors"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

// syncSettings syncs all settings that are syncable
func (m *Messenger) prepareSyncSettingsMessages(currentClock uint64, prepareForBackup bool) (resultRaw []*common.RawMessage, resultSync []*protobuf.SyncSetting, errors []error) {
	s, err := m.settings.GetSettings()
	if err != nil {
		errors = append(errors, err)
		return
	}

	logger := m.logger.Named("prepareSyncSettings")
	// Do not use the network clock, use the db value
	_, chat := m.getLastClockWithRelatedChat()

	for _, sf := range settings.SettingFieldRegister {
		if sf.CanSync(settings.FromStruct) {
			// DisplayName is backed up via `protobuf.BackedUpProfile` message.
			if prepareForBackup && sf.SyncProtobufFactory().SyncSettingProtobufType() == protobuf.SyncSetting_DISPLAY_NAME {
				continue
			}

			// Pull clock from the db
			clock, err := m.settings.GetSettingLastSynced(sf)
			if err != nil {
				logger.Error("m.settings.GetSettingLastSynced", zap.Error(err), zap.Any("SettingField", sf))
				errors = append(errors, err)
				return
			}
			if clock == 0 {
				clock = currentClock
			}

			// Build protobuf
			rm, sm, err := sf.SyncProtobufFactory().FromStruct()(s, clock, chat.ID)
			if err != nil {
				// Collect errors to give other sync messages a chance to send
				logger.Error("SyncProtobufFactory.Struct", zap.Error(err))
				errors = append(errors, err)
			}

			resultRaw = append(resultRaw, rm)
			resultSync = append(resultSync, sm)
		}
	}
	return
}

func (m *Messenger) syncSettings(rawMessageHandler RawMessageHandler) error {
	logger := m.logger.Named("syncSettings")

	clock, _ := m.getLastClockWithRelatedChat()
	rawMessages, _, errors := m.prepareSyncSettingsMessages(clock, false)

	if len(errors) != 0 {
		// return just the first error, the others have been logged
		return errors[0]
	}

	for _, rm := range rawMessages {
		_, err := rawMessageHandler(context.Background(), *rm)
		if err != nil {
			logger.Error("dispatchMessage", zap.Error(err))
			return err
		}
		logger.Debug("dispatchMessage success", zap.Any("rm", rm))
	}

	return nil
}

// extractSyncSetting parses incoming *protobuf.SyncSetting and stores the setting data if needed
func (m *Messenger) extractAndSaveSyncSetting(syncSetting *protobuf.SyncSetting) (*settings.SyncSettingField, error) {
	sf, err := settings.GetFieldFromProtobufType(syncSetting.Type)
	if err != nil {
		m.logger.Error(
			"extractSyncSetting - settings.GetFieldFromProtobufType",
			zap.Error(err),
			zap.Any("syncSetting", syncSetting),
		)
		return nil, err
	}

	spf := sf.SyncProtobufFactory()
	if spf == nil {
		m.logger.Warn("extractSyncSetting - received protobuf for setting with no SyncProtobufFactory", zap.Any("SettingField", sf))
		return nil, nil
	}
	if spf.Inactive() {
		m.logger.Warn("extractSyncSetting - received protobuf for inactive sync setting", zap.Any("SettingField", sf))
		return nil, nil
	}

	value := spf.ExtractValueFromProtobuf()(syncSetting)

	err = m.settings.SaveSyncSetting(sf, value, syncSetting.Clock)
	if err == errors.ErrNewClockOlderThanCurrent {
		m.logger.Info("extractSyncSetting - SaveSyncSetting :", zap.Error(err))
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if v, ok := value.([]byte); ok {
		value = json.RawMessage(v)
	}

	return &settings.SyncSettingField{SettingField: sf, Value: value}, nil
}

// startSyncSettingsLoop watches the m.settings.SyncQueue and sends a sync message in response to a settings update
func (m *Messenger) startSyncSettingsLoop() {
	go func() {
		logger := m.logger.Named("SyncSettingsLoop")

		for {
			select {
			case s := <-m.settings.GetSyncQueue():
				if s.CanSync(settings.FromInterface) {
					logger.Debug("setting for sync received from settings.SyncQueue")

					clock, chat := m.getLastClockWithRelatedChat()

					// Only the messenger has access to the clock, so set the settings sync clock here.
					err := m.settings.SetSettingLastSynced(s.SettingField, clock)
					if err != nil {
						logger.Error("m.settings.SetSettingLastSynced", zap.Error(err))
						break
					}
					rm, _, err := s.SyncProtobufFactory().FromInterface()(s.Value, clock, chat.ID)
					if err != nil {
						logger.Error("SyncProtobufFactory().FromInterface", zap.Error(err), zap.Any("SyncSettingField", s))
						break
					}

					_, err = m.dispatchMessage(context.Background(), *rm)
					if err != nil {
						logger.Error("dispatchMessage", zap.Error(err))
						break
					}

					logger.Debug("message dispatched")
				}
			case <-m.quit:
				return
			}
		}
	}()
}

func (m *Messenger) startSettingsChangesLoop() {
	channel := m.settings.SubscribeToChanges()
	go func() {
		for {
			select {
			case s := <-channel:
				switch s.GetReactName() {
				case settings.DisplayName.GetReactName():
					m.selfContact.DisplayName = s.Value.(string)
					m.publishSelfContactSubscriptions(&SelfContactChangeEvent{DisplayNameChanged: true})
				case settings.PreferredName.GetReactName():
					m.selfContact.EnsName = s.Value.(string)
					m.publishSelfContactSubscriptions(&SelfContactChangeEvent{PreferredNameChanged: true})
				case settings.Bio.GetReactName():
					m.selfContact.Bio = s.Value.(string)
					m.publishSelfContactSubscriptions(&SelfContactChangeEvent{BioChanged: true})
				}
			case <-m.quit:
				return
			}
		}
	}()
}
