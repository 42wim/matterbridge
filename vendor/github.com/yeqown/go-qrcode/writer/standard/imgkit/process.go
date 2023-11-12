package imgkit

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// Binaryzation process image with threshold value (0-255) and return new image.
func Binaryzation(src image.Image, threshold uint8) image.Image {
	if threshold < 0 || threshold > 255 {
		threshold = 128
	}

	gray := Gray(src)
	bounds := src.Bounds()
	height, width := bounds.Max.Y-bounds.Min.Y, bounds.Max.X-bounds.Min.X

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			// var rgb int = int(gray[i][j][0]) + int(gray[i][j][1]) + int(gray[i][j][2])
			if gray.At(j, i).(color.Gray).Y > threshold {
				gray.Set(j, i, color.White)
			} else {
				gray.Set(j, i, color.Black)
			}
		}
	}

	return gray
}

func Gray(src image.Image) *image.Gray {
	bounds := src.Bounds()
	height, width := bounds.Max.Y-bounds.Min.Y, bounds.Max.X-bounds.Min.X
	gray := image.NewGray(bounds)

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			c := color.GrayModel.Convert(src.At(j, i))
			gray.SetGray(j, i, c.(color.Gray))
		}
	}

	return gray
}

func Scale(src image.Image, rect image.Rectangle, scale draw.Scaler) image.Image {
	if scale == nil {
		scale = draw.ApproxBiLinear
	}

	dst := image.NewRGBA(rect)
	scale.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}
