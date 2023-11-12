package audio

import (
	"github.com/status-im/status-go/protocol/protobuf"
)

func aac(buf []byte) bool {
	return len(buf) > 1 &&
		((buf[0] == 0xFF && buf[1] == 0xF1) ||
			(buf[0] == 0xFF && buf[1] == 0xF9))
}

func amr(buf []byte) bool {
	return len(buf) > 11 &&
		buf[0] == 0x23 && buf[1] == 0x21 &&
		buf[2] == 0x41 && buf[3] == 0x4D &&
		buf[4] == 0x52 && buf[5] == 0x0A
}

func Type(buf []byte) protobuf.AudioMessage_AudioType {
	switch {
	case aac(buf):
		return protobuf.AudioMessage_AAC
	case amr(buf):
		return protobuf.AudioMessage_AMR
	default:
		return protobuf.AudioMessage_UNKNOWN_AUDIO_TYPE
	}
}
