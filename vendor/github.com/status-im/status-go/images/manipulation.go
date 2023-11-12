package images

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"go.uber.org/zap"
	xdraw "golang.org/x/image/draw"

	"github.com/ethereum/go-ethereum/log"
)

type Circle struct {
	X, Y, R int
}

func (c *Circle) ColorModel() color.Model {
	return color.AlphaModel
}
func (c *Circle) Bounds() image.Rectangle {
	return image.Rect(c.X-c.R, c.Y-c.R, c.X+c.R, c.Y+c.R)
}
func (c *Circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.X)+0.5, float64(y-c.Y)+0.5, float64(c.R)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

func Resize(size ResizeDimension, img image.Image) image.Image {
	var width, height uint

	switch {
	case img.Bounds().Max.X == img.Bounds().Max.Y:
		width, height = uint(size), uint(size)
	case img.Bounds().Max.X > img.Bounds().Max.Y:
		width, height = 0, uint(size)
	default:
		width, height = uint(size), 0
	}

	log.Info("resizing", "size", size, "width", width, "height", height)

	return resize.Resize(width, height, img, resize.Bilinear)
}

func ResizeTo(percent int, img image.Image) image.Image {
	width := uint(img.Bounds().Max.X * percent / 100)
	height := uint(img.Bounds().Max.Y * percent / 100)

	return resize.Resize(width, height, img, resize.Bilinear)
}

func ShrinkOnly(size ResizeDimension, img image.Image) image.Image {
	finalSize := int(math.Min(float64(size), math.Min(float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))))
	return Resize(ResizeDimension(finalSize), img)
}

func Crop(img image.Image, rect image.Rectangle) (image.Image, error) {

	if img.Bounds().Max.X < rect.Max.X || img.Bounds().Max.Y < rect.Max.Y {
		return nil, fmt.Errorf(
			"crop dimensions out of bounds of image, image width '%dpx' & height '%dpx'; crop bottom right coordinate at X '%dpx' Y '%dpx'",
			img.Bounds().Max.X, img.Bounds().Max.Y,
			rect.Max.X, rect.Max.Y,
		)
	}

	return cutter.Crop(img, cutter.Config{
		Width:  rect.Dx(),
		Height: rect.Dy(),
		Anchor: rect.Min,
	})
}

// CropCenter takes an image, usually downloaded from a URL
// If the image is square, the full image is returned
// If the image is rectangular, the largest central square is returned
// calculations at _docs/image-center-crop-calculations.png
func CropCenter(img image.Image) (image.Image, error) {
	var cropRect image.Rectangle
	maxBounds := img.Bounds().Max

	if maxBounds.X == maxBounds.Y {
		return img, nil
	}

	if maxBounds.X > maxBounds.Y {
		// the final output should be YxY
		cropRect = image.Rectangle{
			Min: image.Point{X: maxBounds.X/2 - maxBounds.Y/2, Y: 0},
			Max: image.Point{X: maxBounds.X/2 + maxBounds.Y/2, Y: maxBounds.Y},
		}
	} else {
		// the final output should be XxX
		cropRect = image.Rectangle{
			Min: image.Point{X: 0, Y: maxBounds.Y/2 - maxBounds.X/2},
			Max: image.Point{X: maxBounds.X, Y: maxBounds.Y/2 + maxBounds.X/2},
		}
	}
	return Crop(img, cropRect)
}

func ImageToBytes(imagePath string) ([]byte, error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Create a new buffer to hold the image data
	var imgBuffer bytes.Buffer

	// Encode the image to the desired format and save it in the buffer
	err = png.Encode(&imgBuffer, img)
	if err != nil {
		return nil, err
	}

	// Return the image data as a byte slice
	return imgBuffer.Bytes(), nil
}

func ImageToBytesAndImage(imagePath string) ([]byte, image.Image, error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, nil, err
	}

	// Create a new buffer to hold the image data
	var imgBuffer bytes.Buffer

	// Encode the image to the desired format and save it in the buffer
	err = png.Encode(&imgBuffer, img)
	if err != nil {
		return nil, nil, err
	}

	// Return the image data as a byte slice
	return imgBuffer.Bytes(), img, nil
}

func AddPadding(img image.Image, padding int) *image.RGBA {
	bounds := img.Bounds()
	newBounds := image.Rect(bounds.Min.X-padding, bounds.Min.Y-padding, bounds.Max.X+padding, bounds.Max.Y+padding)
	paddedImg := image.NewRGBA(newBounds)
	draw.Draw(paddedImg, newBounds, &image.Uniform{C: color.White}, image.ZP, draw.Src)

	return paddedImg
}

func EncodePNG(img *image.RGBA) ([]byte, error) {
	resultImg := &bytes.Buffer{}
	err := png.Encode(resultImg, img)
	if err != nil {
		return nil, err
	}
	return resultImg.Bytes(), nil
}

func CreateCircleWithPadding(img image.Image, padding int) *image.RGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	// only relying on width as a metric here because we know that we
	// store profile images in a perfect circle
	radius := width / 2

	paddedWidth := width + 2*padding
	paddedRadius := paddedWidth / 2

	// Create a new circular image with padding
	newBounds := image.Rect(0, 0, paddedWidth, paddedWidth)
	circle := image.NewRGBA(newBounds)

	// Create a larger circular mask for the padding
	paddingMask := &Circle{
		X: paddedRadius,
		Y: paddedRadius,
		R: paddedRadius,
	}

	// Draw the white color onto the circle with padding mask
	draw.DrawMask(circle, circle.Bounds(), image.NewUniform(color.White), image.ZP, paddingMask, image.ZP, draw.Src)

	// Create a new circle mask with the original size
	circleMask := &Circle{
		X: radius,
		Y: radius,
		R: radius,
	}

	// Draw the original image onto the white circular image at the center (with padding offset)
	draw.DrawMask(circle, bounds.Add(image.Pt(padding, padding)), img, image.ZP, circleMask, image.ZP, draw.Over)

	return circle
}

func RoundCrop(inputImage []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(inputImage))
	if err != nil {
		return nil, err
	}
	result := CreateCircleWithPadding(img, 0)

	var outputImage bytes.Buffer
	err = png.Encode(&outputImage, result)
	if err != nil {
		return nil, err
	}
	return outputImage.Bytes(), nil
}

func PlaceCircleInCenter(paddedImg, circle *image.RGBA) *image.RGBA {
	bounds := circle.Bounds()
	centerX := (paddedImg.Bounds().Min.X + paddedImg.Bounds().Max.X) / 2
	centerY := (paddedImg.Bounds().Min.Y + paddedImg.Bounds().Max.Y) / 2
	draw.Draw(paddedImg, bounds.Add(image.Pt(centerX-bounds.Dx()/2, centerY-bounds.Dy()/2)), circle, image.ZP, draw.Over)
	return paddedImg
}

func ResizeImage(imgBytes []byte, width, height int) ([]byte, error) {
	// Decode image bytes
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}
	// Create a new image with the desired dimensions
	newImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	xdraw.BiLinear.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	// Encode the new image to bytes
	var newImgBytes bytes.Buffer
	if err = png.Encode(&newImgBytes, newImg); err != nil {
		return nil, err
	}
	return newImgBytes.Bytes(), nil
}

func SuperimposeLogoOnQRImage(imageBytes []byte, qrFilepath []byte) []byte {
	// Read the two images from bytes
	img1, _, err := image.Decode(bytes.NewReader(imageBytes))

	if err != nil {
		log.Error("error decoding logo Image", zap.Error(err))
		return nil
	}

	img2, _, err := image.Decode(bytes.NewReader(qrFilepath))

	if err != nil {
		log.Error("error decoding QR Image", zap.Error(err))
		return nil
	}
	// Create a new image with the dimensions of the first image
	result := image.NewRGBA(img1.Bounds())
	// Draw the first image on the new image
	draw.Draw(result, img1.Bounds(), img1, image.ZP, draw.Src)
	// Get the dimensions of the second image
	img2Bounds := img2.Bounds()
	// Calculate the x and y coordinates to center the second image
	x := (img1.Bounds().Dx() - img2Bounds.Dx()) / 2
	y := (img1.Bounds().Dy() - img2Bounds.Dy()) / 2
	// Draw the second image on top of the first image at the calculated coordinates
	draw.Draw(result, img2Bounds.Add(image.Pt(x, y)), img2, image.ZP, draw.Over)
	// Encode the final image to a desired format
	var b bytes.Buffer
	err = png.Encode(&b, result)

	if err != nil {
		log.Error("error encoding final result Image to Buffer", zap.Error(err))
		return nil
	}

	return b.Bytes()
}
