package images

import (
	"bytes"
	"image"
)

func GenerateImageVariants(cImg image.Image) ([]IdentityImage, error) {
	var iis []IdentityImage
	var err error

	for _, s := range ResizeDimensions {
		rImg := Resize(s, cImg)

		bb := bytes.NewBuffer([]byte{})
		err = EncodeToBestSize(bb, rImg, s)
		if err != nil {
			return nil, err
		}

		ii := IdentityImage{
			Name:         ResizeDimensionToName[s],
			Payload:      bb.Bytes(),
			Width:        rImg.Bounds().Dx(),
			Height:       rImg.Bounds().Dy(),
			FileSize:     bb.Len(),
			ResizeTarget: int(s),
		}

		iis = append(iis, ii)
	}

	return iis, nil
}

func GenerateIdentityImages(filepath string, aX, aY, bX, bY int) ([]IdentityImage, error) {
	img, err := Decode(filepath)
	if err != nil {
		return nil, err
	}

	cropRect := image.Rectangle{
		Min: image.Point{X: aX, Y: aY},
		Max: image.Point{X: bX, Y: bY},
	}
	cImg, err := Crop(img, cropRect)
	if err != nil {
		return nil, err
	}

	return GenerateImageVariants(cImg)
}

func GenerateIdentityImagesFromURL(url string) ([]IdentityImage, error) {
	img, err := DecodeFromURL(url)
	if err != nil {
		return nil, err
	}

	cImg, err := CropCenter(img)
	if err != nil {
		return nil, err
	}

	return GenerateImageVariants(cImg)
}

func GenerateBannerImage(filepath string, aX, aY, bX, bY int) (*IdentityImage, error) {
	img, err := Decode(filepath)
	if err != nil {
		return nil, err
	}

	cropRect := image.Rectangle{
		Min: image.Point{X: aX, Y: aY},
		Max: image.Point{X: bX, Y: bY},
	}
	croppedImg, err := Crop(img, cropRect)
	if err != nil {
		return nil, err
	}

	resizedImg := ShrinkOnly(BannerDim, croppedImg)

	sizeLimits := GetBannerDimensionLimits()

	bb := bytes.NewBuffer([]byte{})
	err = EncodeToLimits(bb, resizedImg, sizeLimits)
	if err != nil {
		return nil, err
	}

	ii := &IdentityImage{
		Name:         BannerIdentityName,
		Payload:      bb.Bytes(),
		Width:        resizedImg.Bounds().Dx(),
		Height:       resizedImg.Bounds().Dy(),
		FileSize:     bb.Len(),
		ResizeTarget: int(BannerDim),
	}

	return ii, nil
}
