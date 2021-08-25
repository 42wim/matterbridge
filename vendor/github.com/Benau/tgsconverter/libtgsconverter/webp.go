package libtgsconverter

import "bytes"
import "image"

import "github.com/sizeofint/webpanimation"

type towebp struct {
	timestamp int
	webpanim *webpanimation.WebpAnimation
	config webpanimation.WebPConfig
}

func(to_webp *towebp) init(w uint, h uint, options ConverterOptions) {
	to_webp.timestamp = 0
	to_webp.webpanim = webpanimation.NewWebpAnimation(int(w), int(h), 0)
	to_webp.config = webpanimation.NewWebpConfig()
	to_webp.config.SetQuality(options.GetWebpQuality())
}

func(to_webp *towebp) SupportsAnimation() bool {
	return true
}

func (to_webp *towebp) AddFrame(image *image.RGBA, fps uint) error {
	err := to_webp.webpanim.AddFrame(image, to_webp.timestamp, to_webp.config)
	to_webp.timestamp += int((1.0 / float32(fps)) * 1000.)
	return err
}

func (to_webp *towebp) Result() []byte {
	var buf bytes.Buffer
	err := to_webp.webpanim.Encode(&buf)
	if err != nil {
		return nil
	}
	to_webp.webpanim.ReleaseMemory()
	return buf.Bytes()
}
