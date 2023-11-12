package images

import (
	"bytes"
	"errors"
	"image"
	"io/ioutil"
	"os"
)

const (
	maxChatMessageImageSize = 400000
	resizeTargetImageSize   = 350000
	idealTargetImageSize    = 50000
)

var DefaultBounds = FileSizeLimits{Ideal: idealTargetImageSize, Max: resizeTargetImageSize}

func OpenAndAdjustImage(inputImage CroppedImage, crop bool) ([]byte, error) {
	file, err := os.Open(inputImage.ImagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	payload, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	img, err := Decode(inputImage.ImagePath)
	if err != nil {
		return nil, err
	}

	if crop {
		cropRect := image.Rectangle{
			Min: image.Point{X: inputImage.X, Y: inputImage.Y},
			Max: image.Point{X: inputImage.X + inputImage.Width, Y: inputImage.Y + inputImage.Height},
		}
		img, err = Crop(img, cropRect)
		if err != nil {
			return nil, err
		}
	}

	bb := bytes.NewBuffer([]byte{})
	err = CompressToFileLimits(bb, img, DefaultBounds)
	if err != nil {
		return nil, err
	}

	// We keep the smallest one
	if len(payload) > len(bb.Bytes()) {
		payload = bb.Bytes()
	}

	if len(payload) > maxChatMessageImageSize {
		return nil, errors.New("image too large")
	}

	return payload, nil
}
