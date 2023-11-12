package colorfinder

// colorfinder takes an image and tries to find its main color.
// It is a liberal port of
// http://pieroxy.net/blog/pages/color-finder/demo.html

import (
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"image/color"

	// Load supported image formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mat/besticon/ico"
)

func main() {
	arg := os.Args[1]

	var imageReader io.ReadCloser
	if strings.HasPrefix(arg, "http") {
		var err error
		response, err := http.Get(arg)
		if err != nil {
			log.Fatal(err)
		}
		imageReader = response.Body
	} else {
		var err error
		fmt.Fprintln(os.Stderr, "Reading "+arg+"...")
		imageReader, err = os.Open(arg)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer imageReader.Close()

	img, _, err := image.Decode(imageReader)
	if err != nil {
		log.Fatal(err)
	}

	cf := ColorFinder{}
	c, err := cf.FindMainColor(img)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("#" + ColorToHex(c))
}

type ColorFinder struct {
	img image.Image
}

// FindMainColor tries to identify the most important color in the given logo.
func (cf *ColorFinder) FindMainColor(img image.Image) (color.RGBA, error) {
	cf.img = img

	colorMap := cf.buildColorMap()

	sRGB := cf.findMainColor(colorMap, 6, nil)
	sRGB = cf.findMainColor(colorMap, 4, &sRGB)
	sRGB = cf.findMainColor(colorMap, 2, &sRGB)
	sRGB = cf.findMainColor(colorMap, 0, &sRGB)

	return sRGB.rgb, nil
}

const sampleThreshold = 160 * 160

func (cf *ColorFinder) buildColorMap() *map[color.RGBA]colorStats {
	colorMap := make(map[color.RGBA]colorStats)
	bounds := cf.img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := cf.img.At(x, y).RGBA()
			rgb := color.RGBA{}
			rgb.R = uint8(r >> shiftRGB)
			rgb.G = uint8(g >> shiftRGB)
			rgb.B = uint8(b >> shiftRGB)
			rgb.A = uint8(a >> shiftRGB)

			colrStats, exist := colorMap[rgb]
			if exist {
				colrStats.count++
			} else {
				colrStats := colorStats{count: 1, weight: weight(&rgb)}
				if colrStats.weight <= 0 {
					colrStats.weight = 1e-10
				}
				colorMap[rgb] = colrStats
			}
		}
	}
	return &colorMap
}

// Turns out using this is faster than using
// RGBAModel.Convert(img.At(x, y))).(color.RGBA)
const shiftRGB = uint8(8)

func (cf *ColorFinder) findMainColor(colorMap *map[color.RGBA]colorStats, shift uint, targetColor *shiftedRGBA) shiftedRGBA {
	colorWeights := make(map[shiftedRGBA]float64)

	bounds := cf.img.Bounds()
	stepLength := stepLength(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepLength {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepLength {
			r, g, b, a := cf.img.At(x, y).RGBA()
			color := color.RGBA{}
			color.R = uint8(r >> shiftRGB)
			color.G = uint8(g >> shiftRGB)
			color.B = uint8(b >> shiftRGB)
			color.A = uint8(a >> shiftRGB)

			if rgbMatchesTargetColor(targetColor, &color) {
				increaseColorWeight(&colorWeights, colorMap, &color, shift)
			}
		}
	}

	maxColor := shiftedRGBA{}
	maxWeight := 0.0
	for sRGB, weight := range colorWeights {
		if weight > maxWeight {
			maxColor = sRGB
			maxWeight = weight
		}
	}

	return maxColor
}

func increaseColorWeight(weightedColors *map[shiftedRGBA]float64, colorMap *map[color.RGBA]colorStats, rgb *color.RGBA, shift uint) {
	shiftedColor := color.RGBA{R: rgb.R >> shift, G: rgb.G >> shift, B: rgb.B >> shift}
	pixelGroup := shiftedRGBA{rgb: shiftedColor, shift: shift}
	colorStats := (*colorMap)[*rgb]
	(*weightedColors)[pixelGroup] += colorStats.weight * float64(colorStats.count)
}

type shiftedRGBA struct {
	rgb   color.RGBA
	shift uint
}

func rgbMatchesTargetColor(targetCol *shiftedRGBA, rgb *color.RGBA) bool {
	if targetCol == nil {
		return true
	}

	return targetCol.rgb.R == (rgb.R>>targetCol.shift) &&
		targetCol.rgb.G == (rgb.G>>targetCol.shift) &&
		targetCol.rgb.B == (rgb.B>>targetCol.shift)
}

type colorStats struct {
	weight float64
	count  int64
}

func stepLength(bounds image.Rectangle) int {
	width := bounds.Dx()
	height := bounds.Dy()
	pixelCount := width * height

	var stepLength int
	if pixelCount > sampleThreshold {
		stepLength = 2
	} else {
		stepLength = 1
	}

	return stepLength
}

func weight(rgb *color.RGBA) float64 {
	rr := float64(rgb.R)
	gg := float64(rgb.G)
	bb := float64(rgb.B)
	return (abs(rr-gg)*abs(rr-gg)+abs(rr-bb)*abs(rr-bb)+abs(gg-bb)*abs(gg-bb))/65535.0*1000.0 + 1
}

func abs(n float64) float64 {
	return math.Abs(float64(n))
}

func ColorToHex(c color.RGBA) string {
	return fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
}
