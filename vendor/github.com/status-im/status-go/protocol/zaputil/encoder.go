package zaputil

import (
	"encoding/hex"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type jsonHexEncoder struct {
	zapcore.Encoder
}

// NewJSONHexEncoder creates a JSON logger based on zapcore.NewJSONEncoder
// but overwrites encoding of byte slices. Instead encoding them with base64,
// jsonHexEncoder uses hex-encoding.
// Each hex-encoded value is prefixed with 0x so that it's clear it's a hex string.
func NewJSONHexEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	jsonEncoder := zapcore.NewJSONEncoder(cfg)
	return &jsonHexEncoder{
		Encoder: jsonEncoder,
	}
}

func (enc *jsonHexEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, "0x"+hex.EncodeToString(val))
}

func (enc *jsonHexEncoder) Clone() zapcore.Encoder {
	encoderClone := enc.Encoder.Clone()
	return &jsonHexEncoder{Encoder: encoderClone}
}

// RegisterJSONHexEncoder registers a jsonHexEncoder under "json-hex" name.
// Later, this name can be used as a value for zap.Config.Encoding to enable
// jsonHexEncoder.
func RegisterJSONHexEncoder() error {
	return zap.RegisterEncoder("json-hex", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewJSONHexEncoder(cfg), nil
	})
}

type consoleHexEncoder struct {
	zapcore.Encoder
}

func NewConsoleHexEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	consoleEncoder := zapcore.NewConsoleEncoder(cfg)
	return &consoleHexEncoder{
		Encoder: consoleEncoder,
	}
}

func (enc *consoleHexEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, "0x"+hex.EncodeToString(val))
}

func (enc *consoleHexEncoder) Clone() zapcore.Encoder {
	encoderClone := enc.Encoder.Clone()
	return &consoleHexEncoder{Encoder: encoderClone}
}

func RegisterConsoleHexEncoder() error {
	return zap.RegisterEncoder("console-hex", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewConsoleHexEncoder(cfg), nil
	})
}
