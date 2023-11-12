package identicon

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

const (
	Width  = 50
	Height = 50
)

func renderBase64(id Identicon) (string, error) {
	img, err := render(id)
	if err != nil {
		return "", err
	}
	encodedString := base64.StdEncoding.EncodeToString(img)
	image := "data:image/png;base64," + encodedString
	return image, nil
}

func setBackgroundTransparent(img *image.RGBA) {
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Transparent}, image.Point{}, draw.Src)
}

func drawRect(rgba *image.RGBA, i int, c color.Color) {
	sizeSquare := 6
	maxRow := 5

	r := image.Rect(
		10+(i%maxRow)*sizeSquare,
		10+(i/maxRow)*sizeSquare,
		10+(i%maxRow)*sizeSquare+sizeSquare,
		10+(i/maxRow)*sizeSquare+sizeSquare,
	)

	draw.Draw(rgba, r, &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func render(id Identicon) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, Width, Height))
	var buff bytes.Buffer

	setBackgroundTransparent(img)

	for i, v := range id.bitmap {
		if v == 1 {
			drawRect(img, i, id.color)
		}
	}

	if err := png.Encode(&buff, img); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// GenerateBase64 generates an identicon in base64 png format given a string
func GenerateBase64(id string) (string, error) {
	i := generate(id)
	return renderBase64(i)
}

func Generate(id string) ([]byte, error) {
	i := generate(id)
	return render(i)
}
