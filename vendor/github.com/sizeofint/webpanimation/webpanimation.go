package webpanimation

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io"
)

type WebpAnimation struct {
	WebPAnimEncoderOptions *WebPAnimEncoderOptions
	Width                  int
	Height                 int
	loopCount              int
	AnimationEncoder       *WebPAnimEncoder
	WebPData               *WebPData
	WebPMux                *WebPMux
	WebPPictures           []*WebPPicture
}

// NewWebpAnimation Initialize animation
func NewWebpAnimation(width, height, loopCount int) *WebpAnimation {
	webpAnimation := &WebpAnimation{loopCount: loopCount, Width: width, Height: height}
	webpAnimation.WebPAnimEncoderOptions = &WebPAnimEncoderOptions{}

	WebPAnimEncoderOptionsInitInternal(webpAnimation.WebPAnimEncoderOptions)

	webpAnimation.AnimationEncoder = WebPAnimEncoderNewInternal(width, height, webpAnimation.WebPAnimEncoderOptions)
	return webpAnimation
}

// ReleaseMemory release memory
func (wpa *WebpAnimation) ReleaseMemory() {
	WebPDataClear(wpa.WebPData)
	WebPMuxDelete(wpa.WebPMux)
	for _, webpPicture := range wpa.WebPPictures {
		WebPPictureFree(webpPicture)
	}
	WebPAnimEncoderDelete(wpa.AnimationEncoder)
}

// AddFrame add frame to animation
func (wpa *WebpAnimation) AddFrame(img image.Image, timestamp int, webpcfg WebPConfig) error {
	var webPPicture *WebPPicture = nil
	var m *image.RGBA
	if img != nil {
		if v, ok := img.(*image.RGBA); ok {
			m = v
		} else {
			b := img.Bounds()
			m = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
			draw.Draw(m, m.Bounds(), img, b.Min, draw.Src)
		}

		webPPicture = &WebPPicture{}
		wpa.WebPPictures = append(wpa.WebPPictures, webPPicture)
		webPPicture.SetUseArgb(1)
		webPPicture.SetHeight(wpa.Height)
		webPPicture.SetWidth(wpa.Width)
		err := WebPPictureImportRGBA(m.Pix, m.Stride, webPPicture)
		if err != nil {
			return err
		}
	}

	res := WebPAnimEncoderAdd(wpa.AnimationEncoder, webPPicture, timestamp, webpcfg)
	if res == 0 {
		return errors.New("Failed to add frame in animation ecoder")
	}
	return nil
}

// Encode encode animation
func (wpa *WebpAnimation) Encode(w io.Writer) error {
	wpa.WebPData = &WebPData{}

	WebPDataInit(wpa.WebPData)

	WebPAnimEncoderAssemble(wpa.AnimationEncoder, wpa.WebPData)

	if wpa.loopCount > 0 {
		wpa.WebPMux = WebPMuxCreateInternal(wpa.WebPData, 1)
		if wpa.WebPMux == nil {
			return errors.New("ERROR: Could not re-mux to add loop count/metadata.")
		}
		WebPDataClear(wpa.WebPData)

		webPMuxAnimNewParams := WebPMuxAnimParams{}
		muxErr := WebPMuxGetAnimationParams(wpa.WebPMux, &webPMuxAnimNewParams)
		if muxErr != WebpMuxOk {
			return errors.New("Could not fetch loop count")
		}
		webPMuxAnimNewParams.SetLoopCount(wpa.loopCount)

		muxErr = WebPMuxSetAnimationParams(wpa.WebPMux, &webPMuxAnimNewParams)
		if muxErr != WebpMuxOk {
			return errors.New(fmt.Sprint("Could not update loop count, code:", muxErr))
		}

		muxErr = WebPMuxAssemble(wpa.WebPMux, wpa.WebPData)
		if muxErr != WebpMuxOk {
			return errors.New("Could not assemble when re-muxing to add")
		}

	}
	_, err := w.Write(wpa.WebPData.GetBytes())
	return err
}
