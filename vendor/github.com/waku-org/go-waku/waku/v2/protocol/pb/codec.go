package pb

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (m *WakuMessage) MarshalJSON() ([]byte, error) {
	return (protojson.MarshalOptions{}).Marshal(m)
}

func Unmarshal(data []byte) (*WakuMessage, error) {
	msg := &WakuMessage{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	err = msg.Validate()
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *WakuMessage) UnmarshalJSON(data []byte) error {
	return (protojson.UnmarshalOptions{}).Unmarshal(data, m)
}
