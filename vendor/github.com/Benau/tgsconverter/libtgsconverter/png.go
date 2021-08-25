package libtgsconverter

import "bytes"
import "image"
import "image/png"

type topng struct {
	result []byte
}

func(to_png *topng) init(w uint, h uint, options ConverterOptions) {
}

func(to_png *topng) SupportsAnimation() bool {
	return false
}

func (to_png *topng) AddFrame(image *image.RGBA, fps uint) error {
	var data []byte
	w := bytes.NewBuffer(data)
	if err := png.Encode(w, image); err != nil {
		return err
	}
	to_png.result = w.Bytes()
	return nil
}

func (to_png *topng) Result() []byte {
	return to_png.result
}
