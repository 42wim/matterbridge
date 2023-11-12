package standard

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/yeqown/go-qrcode/writer/standard/imgkit"
)

// funcOption wraps a function that modifies outputImageOptions into an
// implementation of the ImageOption interface.
type funcOption struct {
	f func(oo *outputImageOptions)
}

func (fo *funcOption) apply(oo *outputImageOptions) {
	fo.f(oo)
}

func newFuncOption(f func(oo *outputImageOptions)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithBgTransparent makes the background transparent.
func WithBgTransparent() ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		oo.bgTransparent = true
	})
}

// WithBgColor background color
func WithBgColor(c color.Color) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		if c == nil {
			return
		}

		oo.bgColor = parseFromColor(c)
	})
}

// WithBgColorRGBHex background color
func WithBgColorRGBHex(hex string) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		if hex == "" {
			return
		}

		oo.bgColor = parseFromHex(hex)
	})
}

// WithFgColor QR color
func WithFgColor(c color.Color) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		if c == nil {
			return
		}

		oo.qrColor = parseFromColor(c)
	})
}

// WithFgColorRGBHex Hex string to set QR Color
func WithFgColorRGBHex(hex string) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		oo.qrColor = parseFromHex(hex)
	})
}

// WithLogoImage image should only has 1/5 width of QRCode at most
func WithLogoImage(img image.Image) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		if img == nil {
			return
		}

		oo.logo = img
	})
}

// WithLogoImageFileJPEG load image from file, jpeg is required.
// image should only have 1/5 width of QRCode at most
func WithLogoImageFileJPEG(f string) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		fd, err := os.Open(f)
		if err != nil {
			fmt.Printf("could not open file(%s), error=%v\n", f, err)
			return
		}

		img, err := jpeg.Decode(fd)
		if err != nil {
			fmt.Printf("could not open file(%s), error=%v\n", f, err)
			return
		}

		oo.logo = img
	})
}

// WithLogoImageFilePNG load image from file, PNG is required.
// image should only have 1/5 width of QRCode at most
func WithLogoImageFilePNG(f string) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		fd, err := os.Open(f)
		if err != nil {
			fmt.Printf("Open file(%s) failed: %v\n", f, err)
			return
		}

		img, err := png.Decode(fd)
		if err != nil {
			fmt.Printf("Decode file(%s) as PNG failed: %v\n", f, err)
			return
		}

		oo.logo = img
	})
}

// WithQRWidth specify width of each qr block
func WithQRWidth(width uint8) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		oo.qrWidth = int(width)
	})
}

// WithCircleShape use circle shape as rectangle(default)
func WithCircleShape() ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		oo.shape = _shapeCircle
	})
}

// WithCustomShape use custom shape as rectangle(default)
func WithCustomShape(shape IShape) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		oo.shape = shape
	})
}

// WithBuiltinImageEncoder option includes: JPEG_FORMAT as default, PNG_FORMAT.
// This works like WithBuiltinImageEncoder, the different between them is
// formatTyp is enumerated in (JPEG_FORMAT, PNG_FORMAT)
func WithBuiltinImageEncoder(format formatTyp) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		var encoder ImageEncoder
		switch format {
		case JPEG_FORMAT:
			encoder = jpegEncoder{}
		case PNG_FORMAT:
			encoder = pngEncoder{}
		default:
			panic("Not supported file format")
		}

		oo.imageEncoder = encoder
	})
}

// WithCustomImageEncoder to use custom image encoder to encode image.Image into
// io.Writer
func WithCustomImageEncoder(encoder ImageEncoder) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		if encoder == nil {
			return
		}

		oo.imageEncoder = encoder
	})
}

// WithBorderWidth specify the both 4 sides' border width. Notice that
// WithBorderWidth(a) means all border width use this variable `a`,
// WithBorderWidth(a, b) mean top/bottom equal to `a`, left/right equal to `b`.
// WithBorderWidth(a, b, c, d) mean top, right, bottom, left.
func WithBorderWidth(widths ...int) ImageOption {
	apply := func(arr *[4]int, top, right, bottom, left int) {
		arr[0] = top
		arr[1] = right
		arr[2] = bottom
		arr[3] = left
	}

	return newFuncOption(func(oo *outputImageOptions) {
		n := len(widths)
		switch n {
		case 0:
			apply(&oo.borderWidths, _defaultPadding, _defaultPadding, _defaultPadding, _defaultPadding)
		case 1:
			apply(&oo.borderWidths, widths[0], widths[0], widths[0], widths[0])
		case 2, 3:
			apply(&oo.borderWidths, widths[0], widths[1], widths[0], widths[1])
		default:
			// 4+
			apply(&oo.borderWidths, widths[0], widths[1], widths[2], widths[3])
		}
	})
}

// WithHalftone ...
func WithHalftone(path string) ImageOption {
	return newFuncOption(func(oo *outputImageOptions) {
		srcImg, err := imgkit.Read(path)
		if err != nil {
			fmt.Println("Read halftone image failed: ", err)
			return
		}

		oo.halftoneImg = srcImg
	})
}
