package images

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
	"unicode/utf8"

	"golang.org/x/image/webp"

	"github.com/ethereum/go-ethereum/log"
)

var (
	htmlCommentRegex = regexp.MustCompile(`(?i)<!--([\\s\\S]*?)-->`)
	svgRegex         = regexp.MustCompile(`(?i)^\s*(?:<\?xml[^>]*>\s*)?(?:<!doctype svg[^>]*>\s*)?<svg[^>]*>[^*]*<\/svg>\s*$`)
)

// IsSVG returns true if the given buffer is a valid SVG image.
func IsSVG(buf []byte) bool {
	var isBinary bool
	if len(buf) < 24 {
		isBinary = false
	}
	for i := 0; i < 14; i++ {
		charCode, _ := utf8.DecodeRuneInString(string(buf[i]))
		if charCode == 65533 || charCode <= 8 {
			isBinary = true
		}
	}
	return !isBinary && svgRegex.Match(htmlCommentRegex.ReplaceAll(buf, []byte{}))
}

func Decode(fileName string) (image.Image, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fb, err := prepareFileForDecode(file)
	if err != nil {
		return nil, err
	}

	return DecodeImageData(fb, file)
}

func DecodeFromURL(path string) (image.Image, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("failed to close profile pic http request body", "err", err)
		}
	}()

	if res.StatusCode >= 400 {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return DecodeImageData(bodyBytes, bytes.NewReader(bodyBytes))
}

func prepareFileForDecode(file *os.File) ([]byte, error) {
	// Read the first 14 bytes, used for performing image type checks before parsing the image data
	fb := make([]byte, 14)
	_, err := file.Read(fb)
	if err != nil {
		return nil, err
	}

	// Reset the read cursor
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	return fb, nil
}

func DecodeImageData(buf []byte, r io.Reader) (img image.Image, err error) {
	switch GetType(buf) {
	case JPEG:
		img, err = jpeg.Decode(r)
	case PNG:
		img, err = png.Decode(r)
	case GIF:
		img, err = gif.Decode(r)
	case WEBP:
		img, err = webp.Decode(r)
	case UNKNOWN:
		fallthrough
	default:
		return nil, errors.New("unsupported file type")
	}
	if err != nil {
		return nil, err
	}

	return img, nil
}

func GetType(buf []byte) ImageType {
	switch {
	case IsJpeg(buf):
		return JPEG
	case IsPng(buf):
		return PNG
	case IsGif(buf):
		return GIF
	case IsWebp(buf):
		return WEBP
	case IsIco(buf):
		return ICO
	default:
		return UNKNOWN
	}
}

func GetMimeType(buf []byte) (string, error) {
	switch {
	case IsJpeg(buf):
		return "jpeg", nil
	case IsPng(buf):
		return "png", nil
	case IsGif(buf):
		return "gif", nil
	case IsWebp(buf):
		return "webp", nil
	case IsIco(buf):
		return "ico", nil
	case IsSVG(buf):
		return "svg", nil
	default:
		return "", errors.New("image format not supported")
	}
}

func IsJpeg(buf []byte) bool {
	return len(buf) > 2 &&
		buf[0] == 0xFF &&
		buf[1] == 0xD8 &&
		buf[2] == 0xFF
}

func IsPng(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x89 && buf[1] == 0x50 &&
		buf[2] == 0x4E && buf[3] == 0x47
}

func IsGif(buf []byte) bool {
	return len(buf) > 2 &&
		buf[0] == 0x47 && buf[1] == 0x49 && buf[2] == 0x46
}

func IsWebp(buf []byte) bool {
	return len(buf) > 11 &&
		buf[8] == 0x57 && buf[9] == 0x45 &&
		buf[10] == 0x42 && buf[11] == 0x50
}

func IsIco(buf []byte) bool {
	return len(buf) > 4 &&
		buf[0] == 0 && buf[1] == 0 && buf[2] == 1 || buf[2] == 2 &&
		buf[4] > 0
}

func GetImageDimensions(imgBytes []byte) (int, int, error) {
	// Decode image bytes
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return 0, 0, err
	}
	// Get the image dimensions
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	return width, height, nil
}
