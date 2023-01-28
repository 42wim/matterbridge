//go:build cgolottie

package helper

import (
	"fmt"

	"github.com/Benau/tgsconverter/libtgsconverter"
	"github.com/sirupsen/logrus"
)

func CanConvertTgsToX() error {
	return nil
}

// ConvertTgsToX convert input data (which should be tgs format) to any format supported by libtgsconverter
func ConvertTgsToX(data *[]byte, outputFormat string, logger *logrus.Entry) error {
	options := libtgsconverter.NewConverterOptions()
	options.SetExtension(outputFormat)
	blob, err := libtgsconverter.ImportFromData(*data, options)
	if err != nil {
		return fmt.Errorf("failed to run libtgsconverter.ImportFromData: %s", err.Error())
	}

	*data = blob
	return nil
}

func SupportsFormat(format string) bool {
	return libtgsconverter.SupportsExtension(format)
}

func LottieBackend() string {
	return "libtgsconverter"
}
