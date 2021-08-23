package libtgsconverter

import "image"

type imageWriter interface {
	init(w uint, h uint, options ConverterOptions)
	SupportsAnimation() bool
	AddFrame(image *image.RGBA, fps uint) error
	Result() []byte
}

func sameImage(a *image.RGBA, b *image.RGBA) bool {
	if len(a.Pix) != len(b.Pix) {
		return false
	}
	for i, v := range a.Pix {
		if v != b.Pix[i] {
			return false
		}
	}
	return true
}

func newImageWriter(extension string, w uint, h uint, options ConverterOptions) imageWriter {
	var writer imageWriter
	switch extension {
	case "apng":
		writer = &toapng{}
	case "gif":
		writer = &togif{}
	case "png":
		writer = &topng{}
	case "webp":
		writer = &towebp{}
	default:
		return nil
	}
	writer.init(w, h, options)
	return writer
}
