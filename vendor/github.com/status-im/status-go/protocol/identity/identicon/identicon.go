package identicon

import (
	"crypto/md5" // nolint: gosec
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
)

const (
	defaultSaturation = 0.5
	defaultLightness  = 0.7
)

type Identicon struct {
	bitmap []byte
	color  color.Color
}

func generate(key string) Identicon {
	hash := md5.Sum([]byte(key)) // nolint: gosec
	return Identicon{
		convertPatternToBinarySwitch(generatePatternFromHash(hash)),
		getColorFromHash(hash),
	}
}

func getColorFromHash(h [16]byte) color.Color {
	// Take the last 3 relevant bytes, and convert to a float between [0..360]
	sum := float64(h[13]) + float64(h[14]) + float64(h[15])
	t := (sum / 765) * 360
	return colorful.Hsl(t, defaultSaturation, defaultLightness)
}

func generatePatternFromHash(sum [16]byte) []byte {
	p := make([]byte, 25)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			jCount := j

			if j > 2 {
				jCount = 4 - j
			}

			p[5*i+j] = sum[3*i+jCount]
		}
	}
	return p
}

func convertPatternToBinarySwitch(pattern []byte) []byte {
	b := make([]byte, 25)
	for i, v := range pattern {
		if v%2 == 0 {
			b[i] = 1
		} else {
			b[i] = 0
		}
	}
	return b
}
