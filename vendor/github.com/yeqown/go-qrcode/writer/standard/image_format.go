package standard

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

type formatTyp uint8

const (
	// JPEG_FORMAT as default output file format.
	JPEG_FORMAT formatTyp = iota
	// PNG_FORMAT .
	PNG_FORMAT
)

// ImageEncoder is an interface which describes the rule how to encode image.Image into io.Writer
type ImageEncoder interface {
	// Encode specify which format to encode image into io.Writer.
	Encode(w io.Writer, img image.Image) error
}

type jpegEncoder struct{}

func (j jpegEncoder) Encode(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, nil)
}

type pngEncoder struct{}

func (j pngEncoder) Encode(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}
