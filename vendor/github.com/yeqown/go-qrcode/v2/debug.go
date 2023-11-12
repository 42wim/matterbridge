package qrcode

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"os"
	"sync"
)

var (
	// _debug mode switch, true means enable debug mode, false means disable.
	_debug     = false
	_debugOnce sync.Once
)

func debugEnabled() bool {
	// load debug switch from environment only once.
	_debugOnce.Do(func() {
		switch os.Getenv("QRCODE_DEBUG") {
		case "1", "true", "TRUE", "enabled", "ENABLED":
			_debug = true
		}
	})

	return _debug
}

// SetDebugMode open debug switch, you can also enable debug by runtime
// environments variables: QRCODE_DEBUG=1 [1, true, TRUE, enabled, ENABLED] which is recommended.
func SetDebugMode() {
	_debug = true
}

func debugLogf(format string, v ...interface{}) {
	if !debugEnabled() {
		return
	}
	log.Printf("[qrcode] DEBUG: "+format, v...)
}

func debugDraw(filename string, mat Matrix) error {
	if !debugEnabled() {
		return nil
	}

	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("debugDraw open file %s failed: %w", filename, err)
	}
	defer func(fd *os.File) {
		_ = fd.Close()
	}(fd)

	return debugDrawTo(fd, mat)
}

func debugDrawTo(w io.Writer, mat Matrix) error {
	if !debugEnabled() {
		return nil
	}

	// width as image width, height as image height
	padding := 10
	blockWidth := 10
	width := mat.Width()*blockWidth + 2*padding
	height := width
	img := image.NewGray16(image.Rect(0, 0, width, height))

	rectangle := func(x1, y1 int, x2, y2 int, img *image.Gray16, c color.Gray16) {
		for x := x1; x < x2; x++ {
			for y := y1; y < y2; y++ {
				img.SetGray16(x, y, c)
			}
		}
	}

	// background
	rectangle(0, 0, width, height, img, color.Gray16{Y: 0xff12})

	mat.iter(IterDirection_COLUMN, func(x int, y int, v qrvalue) {
		sx := x*blockWidth + padding
		sy := y*blockWidth + padding
		es := (x+1)*blockWidth + padding
		ey := (y+1)*blockWidth + padding

		// choose color, false use black, others use black on white background
		var gray color.Gray16
		switch v.qrbool() {
		case false:
			gray = color.White
		default:
			gray = color.Black
		}

		rectangle(sx, sy, es, ey, img, gray)
	})

	// save to writer
	err := jpeg.Encode(w, img, nil)
	if err != nil {
		return fmt.Errorf("debugDrawTo: encode image in JPEG failed: %v", err)
	}

	return nil
}
