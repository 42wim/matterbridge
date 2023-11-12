package images

import (
	"errors"

	"github.com/status-im/status-go/protocol/protobuf"
)

func GetProtobufImageFormat(buf []byte) protobuf.ImageFormat {
	switch GetType(buf) {
	case JPEG:
		return protobuf.ImageFormat_JPEG
	case PNG:
		return protobuf.ImageFormat_PNG
	case GIF:
		return protobuf.ImageFormat_GIF
	case WEBP:
		return protobuf.ImageFormat_WEBP
	default:
		return protobuf.ImageFormat_UNKNOWN_IMAGE_FORMAT
	}
}

func GetProtobufImageMime(buf []byte) (string, error) {
	switch GetType(buf) {
	case JPEG:
		return "image/jpeg", nil
	case PNG:
		return "image/png", nil
	case GIF:
		return "image/gif", nil
	case WEBP:
		return "image/webp", nil
	default:
		return "", errors.New("mime type not found")
	}
}
