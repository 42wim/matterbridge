package ring

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"

	"github.com/fogleman/gg"

	"github.com/status-im/status-go/multiaccounts"
)

type Theme int

const (
	LightTheme Theme = 1
	DarkTheme  Theme = 2
)

var (
	lightThemeIdenticonRingColors = []string{
		"#000000", "#726F6F", "#C4C4C4", "#E7E7E7", "#FFFFFF", "#00FF00",
		"#009800", "#B8FFBB", "#FFC413", "#9F5947", "#FFFF00", "#A8AC00",
		"#FFFFB0", "#FF5733", "#FF0000", "#9A0000", "#FF9D9D", "#FF0099",
		"#C80078", "#FF00FF", "#900090", "#FFB0FF", "#9E00FF", "#0000FF",
		"#000086", "#9B81FF", "#3FAEF9", "#9A6600", "#00FFFF", "#008694",
		"#C2FFFF", "#00F0B6"}
	darkThemeIdenticonRingColors = []string{
		"#000000", "#726F6F", "#C4C4C4", "#E7E7E7", "#FFFFFF", "#00FF00",
		"#009800", "#B8FFBB", "#FFC413", "#9F5947", "#FFFF00", "#A8AC00",
		"#FFFFB0", "#FF5733", "#FF0000", "#9A0000", "#FF9D9D", "#FF0099",
		"#C80078", "#FF00FF", "#900090", "#FFB0FF", "#9E00FF", "#0000FF",
		"#000086", "#9B81FF", "#3FAEF9", "#9A6600", "#00FFFF", "#008694",
		"#C2FFFF", "#00F0B6"}
)

type DrawRingParam struct {
	Theme      Theme                   `json:"theme"`
	ColorHash  multiaccounts.ColorHash `json:"colorHash"`
	ImageBytes []byte                  `json:"imageBytes"`
	Height     int                     `json:"height"`
	Width      int                     `json:"width"`
	RingWidth  float64                 `json:"ringWidth"`
}

func DrawRing(param *DrawRingParam) ([]byte, error) {
	var colors []string
	switch param.Theme {
	case LightTheme:
		colors = lightThemeIdenticonRingColors
	case DarkTheme:
		colors = darkThemeIdenticonRingColors
	default:
		return nil, fmt.Errorf("unknown theme")
	}

	dc := gg.NewContext(param.Width, param.Height)
	img, _, err := image.Decode(bytes.NewReader(param.ImageBytes))
	if err != nil {
		return nil, err
	}
	dc.DrawImage(img, 0, 0)

	radius := (float64(param.Height) - param.RingWidth) / 2
	arcPos := 0.0

	totalRingUnits := 0
	for i := 0; i < len(param.ColorHash); i++ {
		totalRingUnits += param.ColorHash[i][0]
	}
	unitRadLen := 2 * math.Pi / float64(totalRingUnits)

	for i := 0; i < len(param.ColorHash); i++ {
		dc.SetHexColor(colors[param.ColorHash[i][1]])
		dc.DrawArc(float64(param.Width/2), float64(param.Height/2), radius, arcPos, arcPos+unitRadLen*float64(param.ColorHash[i][0]))
		dc.SetLineWidth(param.RingWidth)
		dc.SetLineCapButt()
		dc.Stroke()
		arcPos += unitRadLen * float64(param.ColorHash[i][0])
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, dc.Image())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
