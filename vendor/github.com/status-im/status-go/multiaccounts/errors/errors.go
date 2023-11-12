package errors

import (
	"github.com/pkg/errors"
)

var (
	// ErrWalletNotUnique returned if another account has `wallet` field set to true.
	ErrWalletNotUnique = errors.New("another account is set to be default wallet. disable it before using new")
	// ErrChatNotUnique returned if another account has `chat` field set to true.
	ErrChatNotUnique = errors.New("another account is set to be default chat. disable it before using new")
	// ErrInvalidConfig returned if config isn't allowed
	ErrInvalidConfig = errors.New("configuration value not allowed")
	// ErrNewClockOlderThanCurrent returned if a given clock is older than the current clock
	ErrNewClockOlderThanCurrent = errors.New("the new clock value is older than the current clock value")
	// ErrUnrecognisedSyncSettingProtobufType returned if there is no handler or record of a given protobuf.SyncSetting_Type
	ErrUnrecognisedSyncSettingProtobufType = errors.New("unrecognised protobuf.SyncSetting_Type")
	ErrDbTransactionIsNil                  = errors.New("database transaction is nil")
)
