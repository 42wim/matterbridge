package imgkit

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Read reads an image from a file. only support PNG and JPEG yet.
func Read(path string) (img image.Image, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer fd.Close()

	img, _, err = image.Decode(fd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode image")
	}

	return img, nil
}

// Save saves the image to the given path.
func Save(img image.Image, filename string) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	switch filepath.Ext(filename) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(fd, img, nil)
	case ".png":
		err = png.Encode(fd, img)
	default:
		err = errors.New("unsupported image format, jpg or png only")
	}

	return err
}
