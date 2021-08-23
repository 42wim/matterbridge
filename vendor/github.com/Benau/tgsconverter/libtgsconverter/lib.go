package libtgsconverter

import "bytes"
import "errors"
import "compress/gzip"
import "image"
import "io/ioutil"

import "github.com/Benau/go_rlottie"

type ConverterOptions interface {
	SetExtension(ext string)
	SetFPS(fps uint)
	SetScale(scale float32)
	SetWebpQuality(webp_quality float32)
	GetExtension() string
	GetFPS() uint
	GetScale() float32
	GetWebpQuality() float32
}

type converter_options struct {
	// apng, gif, png or webp
	extension string
	// Frame per second of output image (if you specify apng, gif or webp)
	fps uint
	// Scale of image result
	scale float32
	// Webp encoder quality (0 to 100)
	webpQuality float32
}

func(opt *converter_options) SetExtension(ext string) {
	opt.extension = ext
}

func(opt *converter_options) SetFPS(fps uint) {
	opt.fps = fps
}

func(opt *converter_options) SetScale(scale float32) {
	opt.scale = scale
}

func(opt *converter_options) SetWebpQuality(webp_quality float32) {
	opt.webpQuality = webp_quality
}

func(opt *converter_options) GetExtension() string {
	return opt.extension
}

func(opt *converter_options) GetFPS() uint {
	return opt.fps
}

func(opt *converter_options) GetScale() float32 {
	return opt.scale
}

func(opt *converter_options) GetWebpQuality() float32 {
	return opt.webpQuality
}

func NewConverterOptions() ConverterOptions {
	return &converter_options{"png", 30, 1.0, 75}
}

func imageFromBuffer(p []byte, w uint, h uint) *image.RGBA {
	// rlottie use ARGB32_Premultiplied
	for i := 0; i < len(p); i += 4 {
		p[i + 0], p[i + 2] = p[i + 2], p[i + 0]
	}
	m := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	m.Pix = p
	m.Stride = int(w) * 4
	return m
}

var disabled_cache = false
func ImportFromData(data []byte, options ConverterOptions) ([]byte, error) {
	if !disabled_cache {
		disabled_cache = true
		go_rlottie.LottieConfigureModelCacheSize(0)
	}
	z, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("Failed to create gzip reader:" + err.Error())
	}
	uncompressed, err := ioutil.ReadAll(z)
	if err != nil {
		return nil, errors.New("Failed to read gzip archive")
	}
	z.Close()

	animation := go_rlottie.LottieAnimationFromData(string(uncompressed[:]), "", "")
	if animation == nil {
		return nil, errors.New("Failed to import lottie animation data")
	}

	w, h := go_rlottie.LottieAnimationGetSize(animation)
	w = uint(float32(w) * options.GetScale())
	h = uint(float32(h) * options.GetScale())

	frame_rate := go_rlottie.LottieAnimationGetFramerate(animation)
	frame_count := go_rlottie.LottieAnimationGetTotalframe(animation)
	duration := float32(frame_count) / float32(frame_rate)
	var desired_framerate = float32(options.GetFPS())
	// Most (Gif) player doesn't support ~60fps (found in most tgs)
	if desired_framerate > 50. {
		desired_framerate = 50.
	}
	step := 1.0 / desired_framerate

	writer := newImageWriter(options.GetExtension(), w, h, options)
	if writer == nil {
		return nil, errors.New("Failed create imagewriter")
	}

	var i float32
	for i = 0.; i < duration; i += step {
		frame := go_rlottie.LottieAnimationGetFrameAtPos(animation, i / duration)
		buf := make([]byte, w * h * 4)
		go_rlottie.LottieAnimationRender(animation, frame, buf, w, h, w * 4)
		m := imageFromBuffer(buf, w, h)
		err := writer.AddFrame(m, uint(desired_framerate))
		if err != nil {
			return nil, errors.New("Failed to add frame:" + err.Error())
		}
		if !writer.SupportsAnimation() {
			break
		}
	}
	go_rlottie.LottieAnimationDestroy(animation)
	return writer.Result(), nil
}

func ImportFromFile(path string, options ConverterOptions) ([]byte, error) {
	tgs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("Error when opening file:" + err.Error())
	}
	return ImportFromData(tgs, options)
}

func SupportsExtension(extension string) (bool) {
	switch extension {
	case "apng":
		fallthrough
	case "gif":
		fallthrough
	case "png":
		fallthrough
	case "webp":
		return true
	default:
		return false
	}
	return false
}
