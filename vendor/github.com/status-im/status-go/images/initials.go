package images

import (
	"bytes"
	"image/color"
	"image/png"
	"io/ioutil"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type RGBA struct {
	R, G, B, A float64
}

var parsedFont *opentype.Font = nil

func ExtractInitials(fullName string, amountInitials int) string {
	if fullName == "" {
		return ""
	}
	var initials strings.Builder
	namesList := strings.Fields(fullName)
	for _, name := range namesList {
		if len(initials.String()) >= amountInitials {
			break
		}
		if name != "" {
			initials.WriteString(strings.ToUpper(name[0:1]))
		}
	}
	return initials.String()
}

// GenerateInitialsImage uppercaseRatio is <height of any upper case> / dc.FontHeight() (line height)
// 0.60386123 for Inter-UI-Medium.otf
func GenerateInitialsImage(initials string, bgColor, fontColor color.Color, fontFile string, size int, fontSize float64, uppercaseRatio float64) ([]byte, error) {
	// Load otf file
	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}

	if parsedFont == nil {
		parsedFont, err = opentype.Parse(fontBytes)
		if err != nil {
			return nil, err
		}
	}

	halfSize := float64(size / 2)

	dc := gg.NewContext(size, size)
	dc.DrawCircle(halfSize, halfSize, halfSize)
	dc.SetColor(bgColor)
	dc.Fill()

	// Load font
	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, err
	}
	dc.SetFontFace(face)

	// Draw initials
	dc.SetColor(fontColor)

	dc.DrawStringAnchored(initials, halfSize, halfSize, 0.5, uppercaseRatio/2)

	img := dc.Image()
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, img)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
