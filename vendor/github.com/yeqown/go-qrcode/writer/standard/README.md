## Standard Writer

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/yeqown/go-qrcode/writer/standard)

Standard Writer is a writer that is used to draw QR Code image into `io.Writer`, normally a file.

### Usage

```go
options := []ImageOption{
	WithBgColorRGBHex("#ffffff"),
	WithFgColorRGBHex("#000000"),
	// more ...
}

// New will create file automatically.
writer, err := standard.New("filename", options...)

// or use io.WriteCloser
var w io.WriterCloser
writer2, err := standard.NewWith(w, options...)
```

### Options

```go
// WithBgTransparent makes the background transparent.
func WithBgTransparent() ImageOption {}

// WithBgColor background color
func WithBgColor(c color.Color) ImageOption {}

// WithBgColorRGBHex background color
func WithBgColorRGBHex(hex string) ImageOption {}

// WithFgColor QR color
func WithFgColor(c color.Color) ImageOption {}

// WithFgColorRGBHex Hex string to set QR Color
func WithFgColorRGBHex(hex string) ImageOption {}

// WithLogoImage .
func WithLogoImage(img image.Image) ImageOption {}

// WithLogoImageFilePNG load image from file, PNG is required
func WithLogoImageFilePNG(f string) ImageOption {}

// WithLogoImageFileJPEG load image from file, JPEG is required
func WithLogoImageFileJPEG(f string) ImageOption {}

// WithQRWidth specify width of each qr block
func WithQRWidth(width uint8) ImageOption {}

// WithCircleShape use circle shape as rectangle(default)
func WithCircleShape() ImageOption {}

// WithCustomShape use custom shape as rectangle(default)
func WithCustomShape(shape IShape) ImageOption {}

// WithBuiltinImageEncoder option includes: JPEG_FORMAT as default, PNG_FORMAT.
// This works like WithBuiltinImageEncoder, the different between them is
// formatTyp is enumerated in (JPEG_FORMAT, PNG_FORMAT)
func WithBuiltinImageEncoder(format formatTyp) ImageOption

// WithCustomImageEncoder to use custom image encoder to encode image.Image into
// io.Writer
func WithCustomImageEncoder(encoder ImageEncoder) ImageOption

// WithBorderWidth specify the both 4 sides' border width. Notice that
// WithBorderWidth(a) means all border width use this variable `a`,
// WithBorderWidth(a, b) mean top/bottom equal to `a`, left/right equal to `b`.
// WithBorderWidth(a, b, c, d) mean top, right, bottom, left.
func WithBorderWidth(widths ...int) ImageOption

// WithHalftone ...
func WithHalftone(path string) ImageOption
```

### extension

- [How to customize QR Code shape](./how-to-use-custom-shape.md)
- [How to customize ImageEncoder](./how-to-use-image-encoder.md)