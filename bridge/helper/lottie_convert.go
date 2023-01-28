//go:build !cgolottie

package helper

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// CanConvertTgsToX Checks whether the external command necessary for ConvertTgsToX works.
func CanConvertTgsToX() error {
	// We depend on the fact that `lottie_convert.py --help` has exit status 0.
	// Hyrum's Law predicted this, and Murphy's Law predicts that this will break eventually.
	// However, there is no alternative like `lottie_convert.py --is-properly-installed`
	cmd := exec.Command("lottie_convert.py", "--help")
	return cmd.Run()
}

// ConvertTgsToWebP convert input data (which should be tgs format) to WebP format
// This relies on an external command, which is ugly, but works.
func ConvertTgsToX(data *[]byte, outputFormat string, logger *logrus.Entry) error {
	// lottie can't handle input from a pipe, so write to a temporary file:
	tmpInFile, err := ioutil.TempFile(os.TempDir(), "matterbridge-lottie-input-*.tgs")
	if err != nil {
		return err
	}
	tmpInFileName := tmpInFile.Name()
	defer func() {
		if removeErr := os.Remove(tmpInFileName); removeErr != nil {
			logger.Errorf("Could not delete temporary (input) file %s: %v", tmpInFileName, removeErr)
		}
	}()
	// lottie can handle writing to a pipe, but there is no way to do that platform-independently.
	// "/dev/stdout" won't work on Windows, and "-" upsets Cairo for some reason. So we need another file:
	tmpOutFile, err := ioutil.TempFile(os.TempDir(), "matterbridge-lottie-output-*.data")
	if err != nil {
		return err
	}
	tmpOutFileName := tmpOutFile.Name()
	defer func() {
		if removeErr := os.Remove(tmpOutFileName); removeErr != nil {
			logger.Errorf("Could not delete temporary (output) file %s: %v", tmpOutFileName, removeErr)
		}
	}()

	if _, writeErr := tmpInFile.Write(*data); writeErr != nil {
		return writeErr
	}
	// Must close before calling lottie to avoid data races:
	if closeErr := tmpInFile.Close(); closeErr != nil {
		return closeErr
	}

	// Call lottie to transform:
	cmd := exec.Command("lottie_convert.py", "--input-format", "lottie", "--output-format", outputFormat, tmpInFileName, tmpOutFileName)
	cmd.Stdout = nil
	cmd.Stderr = nil
	// NB: lottie writes progress into to stderr in all cases.
	_, stderr := cmd.Output()
	if stderr != nil {
		// 'stderr' already contains some parts of Stderr, because it was set to 'nil'.
		return stderr
	}
	dataContents, err := ioutil.ReadFile(tmpOutFileName)
	if err != nil {
		return err
	}

	*data = dataContents
	return nil
}

func SupportsFormat(format string) bool {
	switch format {
	case "png":
		fallthrough
	case "webp":
		return true
	default:
		return false
	}
	return false
}

func LottieBackend() string {
	return "lottie_convert.py"
}
