package protocol

import (
	"errors"
	"io"
	"reflect"

	"github.com/russolsen/transit"
)

var (
	pairMessageType = reflect.TypeOf(PairMessage{})

	defaultMessageValueEncoder = &messageValueEncoder{}
)

// NewMessageEncoder returns a new Transit encoder
// that can encode Message values.
// More about Transit: https://github.com/cognitect/transit-format
func NewMessageEncoder(w io.Writer) *transit.Encoder {
	encoder := transit.NewEncoder(w, false)
	encoder.AddHandler(pairMessageType, defaultMessageValueEncoder)
	return encoder
}

type messageValueEncoder struct{}

func (messageValueEncoder) IsStringable(reflect.Value) bool {
	return false
}

func (messageValueEncoder) Encode(e transit.Encoder, value reflect.Value, asString bool) error {
	switch message := value.Interface().(type) {
	case PairMessage:
		taggedValue := transit.TaggedValue{
			Tag: pairMessageTag,
			Value: []interface{}{
				message.InstallationID,
				message.DeviceType,
				message.Name,
				message.FCMToken,
			},
		}
		return e.EncodeInterface(taggedValue, false)
	}

	return errors.New("unknown message type to encode")
}
