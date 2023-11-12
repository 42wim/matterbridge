package standard

import (
	"fmt"
	"image"
	"image/color"

	"github.com/yeqown/go-qrcode/v2"
)

type ImageOption interface {
	apply(o *outputImageOptions)
}

// defaultOutputImageOption default output image background color and etc options
func defaultOutputImageOption() *outputImageOptions {
	return &outputImageOptions{
		bgColor:       color_WHITE,     // white
		bgTransparent: false,           // not transparent
		qrColor:       color_BLACK,     // black
		logo:          nil,             //
		qrWidth:       20,              //
		shape:         _shapeRectangle, //
		imageEncoder:  jpegEncoder{},
		borderWidths:  [4]int{_defaultPadding, _defaultPadding, _defaultPadding, _defaultPadding},
	}
}

// outputImageOptions to output QR code image
type outputImageOptions struct {
	// bgColor is the background color of the QR code image.
	bgColor color.RGBA
	// bgTransparent only affects on PNG_FORMAT
	bgTransparent bool

	// qrColor is the foreground color of the QR code.
	qrColor color.RGBA

	// logo this icon image would be put the center of QR Code image
	// NOTE: logo only should have 1/5 size of QRCode image
	logo image.Image

	// qrWidth width of each qr block
	qrWidth int

	// shape means how to draw the shape of each cell.
	shape IShape

	// imageEncoder specify which file format would be encoded the QR image.
	imageEncoder ImageEncoder

	// borderWidths indicates the border width of the output image. the order is
	// top, right, bottom, left same as the WithBorder
	borderWidths [4]int

	// halftoneImg is the halftone image for the output image.
	halftoneImg image.Image
}

func (oo *outputImageOptions) backgroundColor() color.RGBA {
	if oo == nil {
		return color_WHITE
	}

	if oo.bgTransparent {
		(&oo.bgColor).A = 0x00
	}

	return oo.bgColor
}

func (oo *outputImageOptions) logoImage() image.Image {
	if oo == nil || oo.logo == nil {
		return nil
	}

	return oo.logo
}

func (oo *outputImageOptions) qrBlockWidth() int {
	if oo == nil || (oo.qrWidth <= 0 || oo.qrWidth > 255) {
		return 20
	}

	return oo.qrWidth
}

func (oo *outputImageOptions) getShape() IShape {
	if oo == nil || oo.shape == nil {
		return _shapeRectangle
	}

	return oo.shape
}

// preCalculateAttribute this function must reference to draw function.
func (oo *outputImageOptions) preCalculateAttribute(dimension int) *Attribute {
	if oo == nil {
		return nil
	}

	top, right, bottom, left := oo.borderWidths[0], oo.borderWidths[1], oo.borderWidths[2], oo.borderWidths[3]
	return &Attribute{
		W:          dimension*oo.qrBlockWidth() + right + left,
		H:          dimension*oo.qrBlockWidth() + top + bottom,
		Borders:    oo.borderWidths,
		BlockWidth: oo.qrBlockWidth(),
	}
}

var (
	color_WHITE = parseFromHex("#ffffff")
	color_BLACK = parseFromHex("#000000")
)

var (
	// _STATE_MAPPING mapping matrix.State to color.RGBA in debug mode.
	_STATE_MAPPING = map[qrcode.QRType]color.RGBA{
		qrcode.QRType_INIT:     parseFromHex("#ffffff"), // [bg]
		qrcode.QRType_DATA:     parseFromHex("#cdc9c3"), // [bg]
		qrcode.QRType_VERSION:  parseFromHex("#000000"), // [fg]
		qrcode.QRType_FORMAT:   parseFromHex("#444444"), // [fg]
		qrcode.QRType_FINDER:   parseFromHex("#555555"), // [fg]
		qrcode.QRType_DARK:     parseFromHex("#2BA859"), // [fg]
		qrcode.QRType_SPLITTER: parseFromHex("#2BA859"), // [fg]
		qrcode.QRType_TIMING:   parseFromHex("#000000"), // [fg]
	}
)

// translateToRGBA get color.RGBA by value State, if not found, return outputImageOptions.qrColor.
// NOTE: this function decides the state should use qrColor or bgColor.
func (oo *outputImageOptions) translateToRGBA(v qrcode.QRValue) (rgba color.RGBA) {
	// TODO(@yeqown): use _STATE_MAPPING to replace this function while in debug mode
	// or some special flag.
	if v.IsSet() {
		rgba = oo.qrColor
		return rgba
	}

	if oo.bgTransparent {
		(&oo.bgColor).A = 0x00
	}
	rgba = oo.bgColor

	return rgba
}

// parseFromHex convert hex string into color.RGBA
func parseFromHex(s string) color.RGBA {
	c := color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 0xff,
	}

	var err error
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	if err != nil {
		panic(err)
	}

	return c
}

func parseFromColor(c color.Color) color.RGBA {
	rgba, ok := c.(color.RGBA)
	if ok {
		return rgba
	}

	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}
}
