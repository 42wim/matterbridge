package server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"net/url"
	"strconv"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.uber.org/zap"

	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/multiaccounts"
)

type WriterCloserByteBuffer struct {
	*bytes.Buffer
}

func (wc WriterCloserByteBuffer) Close() error {
	return nil
}

func NewWriterCloserByteBuffer() *WriterCloserByteBuffer {
	return &WriterCloserByteBuffer{bytes.NewBuffer([]byte{})}
}

type QRConfig struct {
	DecodedQRURL    string
	WithLogo        bool
	CorrectionLevel qrcode.EncodeOption
	KeyUID          string
	ImageName       string
	Size            int
	Params          url.Values
}

func NewQRConfig(params url.Values, logger *zap.Logger) (*QRConfig, error) {
	config := &QRConfig{}
	config.Params = params
	err := config.setQrURL()

	if err != nil {
		logger.Error("[qrops-error] error in setting QRURL", zap.Error(err))
		return nil, err
	}

	config.setAllowProfileImage()
	config.setErrorCorrectionLevel()
	err = config.setSize()

	if err != nil {
		logger.Error("[qrops-error] could not convert string to int for size param ", zap.Error(err))
		return nil, err
	}

	if config.WithLogo {
		err = config.setKeyUID()

		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}

		config.setImageName()
	}

	return config, nil
}

func (q *QRConfig) setQrURL() error {
	qrURL, ok := q.Params["url"]

	if !ok || len(qrURL) == 0 {
		return errors.New("[qrops-error] no qr url provided")
	}

	decodedURL, err := base64.StdEncoding.DecodeString(qrURL[0])

	if err != nil {
		return err
	}

	q.DecodedQRURL = string(decodedURL)
	return nil
}

func (q *QRConfig) setAllowProfileImage() {
	allowProfileImage, ok := q.Params["allowProfileImage"]

	if !ok || len(allowProfileImage) == 0 {
		// we default to false when this flag was not provided
		// so someone does not want to allowProfileImage on their QR Image
		// fine then :)
		q.WithLogo = false
	}

	LogoOnImage, err := strconv.ParseBool(allowProfileImage[0])

	if err != nil {
		// maybe for fun someone tries to send non-boolean values to this flag
		// we also default to false in that case
		q.WithLogo = false
	}

	// if we reach here its most probably true
	q.WithLogo = LogoOnImage
}

func (q *QRConfig) setErrorCorrectionLevel() {
	level, ok := q.Params["level"]
	if !ok || len(level) == 0 {
		// we default to MediumLevel of error correction when the level flag
		// is not passed.
		q.CorrectionLevel = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}

	levelInt, err := strconv.Atoi(level[0])
	if err != nil || levelInt < 0 {
		// if there is any issue with string to int conversion
		// we still default to MediumLevel of error correction
		q.CorrectionLevel = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}

	switch levelInt {
	case 1:
		q.CorrectionLevel = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow)
	case 2:
		q.CorrectionLevel = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	case 3:
		q.CorrectionLevel = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart)
	case 4:
		q.CorrectionLevel = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest)
	default:
		q.CorrectionLevel = qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}
}

func (q *QRConfig) setSize() error {
	size, ok := q.Params["size"]

	if ok {
		imageToBeResized, err := strconv.Atoi(size[0])

		if err != nil {
			return err
		}

		if imageToBeResized <= 0 {
			return errors.New("[qrops-error] Got an invalid size parameter, it should be greater than zero")
		}

		q.Size = imageToBeResized

	}

	return nil
}

func (q *QRConfig) setKeyUID() error {
	keyUID, ok := q.Params["keyUid"]
	// the keyUID was not passed, which is a requirement to get the multiaccount image,
	// so we log this error
	if !ok || len(keyUID) == 0 {
		return errors.New("[qrops-error] A keyUID is required to put logo on image and it was not passed in the parameters")
	}

	q.KeyUID = keyUID[0]
	return nil
}

func (q *QRConfig) setImageName() {
	imageName, ok := q.Params["imageName"]
	//if the imageName was not passed, we default to const images.LargeDimName
	if !ok || len(imageName) == 0 {
		q.ImageName = images.LargeDimName
	}

	q.ImageName = imageName[0]
}

func ToLogoImageFromBytes(imageBytes []byte, padding int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("decoding image failed: %v", err)
	}
	circle := images.CreateCircleWithPadding(img, padding)
	resultBytes, err := images.EncodePNG(circle)
	if err != nil {
		return nil, fmt.Errorf("encoding PNG failed: %v", err)
	}
	return resultBytes, nil
}

func GetLogoImage(multiaccountsDB *multiaccounts.Database, keyUID string, imageName string) ([]byte, error) {
	var (
		padding   int
		LogoBytes []byte
	)

	staticImageData, err := images.Asset("_assets/tests/qr/status.png")
	if err != nil { // Asset was not found.
		return nil, err
	}
	identityImageObjectFromDB, err := multiaccountsDB.GetIdentityImage(keyUID, imageName)

	if err != nil {
		return nil, err
	}

	// default padding to 10 to make the QR with profile image look as per
	// the designs
	padding = 10

	if identityImageObjectFromDB == nil {
		LogoBytes, err = ToLogoImageFromBytes(staticImageData, padding)
	} else {
		LogoBytes, err = ToLogoImageFromBytes(identityImageObjectFromDB.Payload, padding)
	}

	return LogoBytes, err
}

func GetPadding(imgBytes []byte) int {
	const (
		defaultPadding = 20
	)
	size, _, err := images.GetImageDimensions(imgBytes)
	if err != nil {
		return defaultPadding
	}
	return size / 5
}

func generateQRBytes(params url.Values, logger *zap.Logger, multiaccountsDB *multiaccounts.Database) []byte {

	qrGenerationConfig, err := NewQRConfig(params, logger)

	if err != nil {
		logger.Error("could not generate QRConfig please rectify the errors with input parameters", zap.Error(err))
		return nil
	}

	qrc, err := qrcode.NewWith(qrGenerationConfig.DecodedQRURL,
		qrcode.WithEncodingMode(qrcode.EncModeAuto),
		qrGenerationConfig.CorrectionLevel,
	)

	if err != nil {
		logger.Error("could not generate QRCode with provided options", zap.Error(err))
		return nil
	}

	buf := NewWriterCloserByteBuffer()
	nw := standard.NewWithWriter(buf)
	err = qrc.Save(nw)

	if err != nil {
		logger.Error("could not save image", zap.Error(err))
		return nil
	}

	payload := buf.Bytes()

	if qrGenerationConfig.WithLogo {
		logo, err := GetLogoImage(multiaccountsDB, qrGenerationConfig.KeyUID, qrGenerationConfig.ImageName)

		if err != nil {
			logger.Error("could not get logo image from multiaccountsDB", zap.Error(err))
			return nil
		}

		qrWidth, qrHeight, err := images.GetImageDimensions(payload)

		if err != nil {
			logger.Error("could not get image dimensions from payload", zap.Error(err))
			return nil
		}

		logo, err = images.ResizeImage(logo, qrWidth/5, qrHeight/5)

		if err != nil {
			logger.Error("could not resize logo image ", zap.Error(err))
			return nil
		}

		payload = images.SuperimposeLogoOnQRImage(payload, logo)
	}

	if qrGenerationConfig.Size > 0 {

		payload, err = images.ResizeImage(payload, qrGenerationConfig.Size, qrGenerationConfig.Size)

		if err != nil {
			logger.Error("could not resize final logo image ", zap.Error(err))
			return nil
		}
	}

	return payload

}
