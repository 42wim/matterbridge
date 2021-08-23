package libtgsconverter

import "bytes"
import "image"

import "github.com/kettek/apng"
import "github.com/av-elier/go-decimal-to-rational"

type toapng struct {
	apng apng.APNG
	prev_frame *image.RGBA
}

func(to_apng *toapng) init(w uint, h uint, options ConverterOptions) {
}

func(to_apng *toapng) SupportsAnimation() bool {
	return true
}

func (to_apng *toapng) AddFrame(image *image.RGBA, fps uint) error {
	if to_apng.prev_frame != nil && sameImage(to_apng.prev_frame, image) {
		var idx = len(to_apng.apng.Frames) - 1
		var prev_fps = float64(to_apng.apng.Frames[idx].DelayNumerator) / float64(to_apng.apng.Frames[idx].DelayDenominator)
		prev_fps += 1.0 / float64(fps)
		rat := dectofrac.NewRatP(prev_fps, 0.001)
		to_apng.apng.Frames[idx].DelayNumerator = uint16(rat.Num().Int64())
		to_apng.apng.Frames[idx].DelayDenominator = uint16(rat.Denom().Int64())
		return nil
	}
	f := apng.Frame{}
	f.Image = image
	f.DelayNumerator = 1
	f.DelayDenominator = uint16(fps)
	f.DisposeOp = apng.DISPOSE_OP_BACKGROUND
	f.BlendOp = apng.BLEND_OP_SOURCE
	f.IsDefault = false
	to_apng.apng.Frames = append(to_apng.apng.Frames, f)
	to_apng.prev_frame = image
	return nil
}

func (to_apng *toapng) Result() []byte {
	var data []byte
	w := bytes.NewBuffer(data)
	err := apng.Encode(w, to_apng.apng)
	if err != nil {
		return nil
	}
	return w.Bytes()
}
