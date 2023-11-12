package protocol

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
)

// PairMessage contains all message details.
type PairMessage struct {
	InstallationID string `json:"installationId"`
	// The type of the device
	DeviceType string `json:"deviceType"`
	// Name the user set name
	Name string `json:"name"`
	// The FCMToken for mobile platforms
	FCMToken string `json:"fcmToken"`

	// not protocol defined fields
	ID []byte `json:"-"`
}

func (m *PairMessage) MarshalJSON() ([]byte, error) {
	type PairMessageAlias PairMessage
	item := struct {
		*PairMessageAlias
		ID string `json:"id"`
	}{
		PairMessageAlias: (*PairMessageAlias)(m),
		ID:               "0x" + hex.EncodeToString(m.ID),
	}

	return json.Marshal(item)
}

// CreatePairMessage creates a PairMessage which is used
// to pair devices.
func CreatePairMessage(installationID string, name string, deviceType string, fcmToken string) PairMessage {
	return PairMessage{
		InstallationID: installationID,
		Name:           name,
		DeviceType:     deviceType,
		FCMToken:       fcmToken,
	}
}

// DecodePairMessage decodes a raw payload to Message struct.
func DecodePairMessage(data []byte) (message PairMessage, err error) {
	buf := bytes.NewBuffer(data)
	decoder := NewMessageDecoder(buf)
	value, err := decoder.Decode()
	if err != nil {
		return
	}

	message, ok := value.(PairMessage)
	if !ok {
		return message, ErrInvalidDecodedValue
	}
	return
}

// EncodePairMessage encodes a PairMessage using Transit serialization.
func EncodePairMessage(value PairMessage) ([]byte, error) {
	var buf bytes.Buffer
	encoder := NewMessageEncoder(&buf)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
