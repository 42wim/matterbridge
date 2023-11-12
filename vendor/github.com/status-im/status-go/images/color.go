package images

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

func ParseColor(colorStr string) (color.RGBA, error) {
	var c color.RGBA

	if strings.HasPrefix(colorStr, "#") {
		// Parse hex color string
		// Remove "#" prefix
		colorStr = colorStr[1:]

		// Convert to RGBA
		val, err := strconv.ParseUint(colorStr, 16, 32)
		if err != nil {
			return c, err
		}
		c.R = uint8(val >> 16)
		c.G = uint8(val >> 8)
		c.B = uint8(val)
		c.A = 255
	} else if strings.HasPrefix(colorStr, "rgb(") {
		// Parse RGB color string
		// Remove prefix and suffix
		colorStr = strings.TrimSuffix(strings.TrimPrefix(colorStr, "rgb("), ")")

		// Split the string into comma separated parts
		parts := strings.Split(colorStr, ",")
		if len(parts) != 3 {
			return c, fmt.Errorf("invalid RGB color string")
		}

		// Convert to RGBA
		r, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return c, err
		}
		g, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return c, err
		}
		b, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			return c, err
		}
		c.R = uint8(r)
		c.G = uint8(g)
		c.B = uint8(b)
		c.A = 255
	} else if strings.HasPrefix(colorStr, "rgba(") {
		// Parse RGBA color string
		// Remove prefix and suffix
		colorStr = strings.TrimSuffix(strings.TrimPrefix(colorStr, "rgba("), ")")

		// Split the string into comma separated parts
		parts := strings.Split(colorStr, ",")
		if len(parts) != 4 {
			return c, fmt.Errorf("invalid RGBA color string")
		}

		// Convert to RGBA
		r, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return c, err
		}
		g, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return c, err
		}
		b, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			return c, err
		}
		a, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil {
			return c, err
		}
		if a < 0 || a > 1 {
			return c, fmt.Errorf("invalid RGBA alpha value")
		}
		c.R = uint8(r)
		c.G = uint8(g)
		c.B = uint8(b)
		c.A = uint8(a * 255)
	} else {
		return c, fmt.Errorf("invalid color string format")
	}

	return c, nil
}
