package settings

import (
	"encoding/json"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/errors"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/sqlite"
)

func StringFromSyncProtobuf(ss *protobuf.SyncSetting) interface{} {
	return ss.GetValueString()
}

func BoolFromSyncProtobuf(ss *protobuf.SyncSetting) interface{} {
	return ss.GetValueBool()
}

func BytesFromSyncProtobuf(ss *protobuf.SyncSetting) interface{} {
	return ss.GetValueBytes()
}

func Int64FromSyncProtobuf(ss *protobuf.SyncSetting) interface{} {
	return ss.GetValueInt64()
}

func BoolHandler(value interface{}) (interface{}, error) {
	_, ok := value.(bool)
	if !ok {
		return value, errors.ErrInvalidConfig
	}

	return value, nil
}

func Int64Handler(value interface{}) (interface{}, error) {
	_, ok := value.(int64)
	if !ok {
		return value, errors.ErrInvalidConfig
	}

	return value, nil
}

func JSONBlobHandler(value interface{}) (interface{}, error) {
	return &sqlite.JSONBlob{Data: value}, nil
}

func AddressHandler(value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if ok {
		value = types.HexToAddress(str)
	} else {
		return value, errors.ErrInvalidConfig
	}
	return value, nil
}

func NodeConfigHandler(value interface{}) (interface{}, error) {
	jsonString, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	nodeConfig := new(params.NodeConfig)
	err = json.Unmarshal(jsonString, nodeConfig)
	if err != nil {
		return nil, err
	}

	return nodeConfig, nil
}
