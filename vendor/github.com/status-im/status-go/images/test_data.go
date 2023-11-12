package images

// Test data that would typically only exist in a test file, used for exporting sample data outside the package.
var (
	testJpegBytes    = []byte{0xff, 0xd8, 0xff, 0xdb, 0x00, 0x84, 0x00, 0x50, 0x37, 0x3c, 0x46, 0x3c, 0x32, 0x50}
	testPngBytes     = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48}
	testGifBytes     = []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x00, 0x01, 0x00, 0x01, 0x84, 0x1f, 0x00, 0xff}
	testWebpBytes    = []byte{0x52, 0x49, 0x46, 0x46, 0x90, 0x49, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56, 0x50}
	testAacBytes     = []byte{0xff, 0xf1, 0x50, 0x80, 0x1c, 0x3f, 0xfc, 0xda, 0x00, 0x4c, 0x61, 0x76, 0x63, 0x35}
	testLogoBytes, _ = Asset("_assets/tests/qr/status.png")
)

func SampleIdentityImages() []IdentityImage {
	return []IdentityImage{
		{
			Name:         SmallDimName,
			Payload:      testJpegBytes,
			Width:        80,
			Height:       80,
			FileSize:     256,
			ResizeTarget: 80,
			Clock:        0,
		},
		{
			Name:         LargeDimName,
			Payload:      testPngBytes,
			Width:        240,
			Height:       300,
			FileSize:     1024,
			ResizeTarget: 240,
			Clock:        0,
		},
	}
}

func SampleIdentityImageForQRCode() []IdentityImage {
	return []IdentityImage{
		{
			Name:         LargeDimName,
			Payload:      testLogoBytes,
			Width:        240,
			Height:       300,
			FileSize:     1024,
			ResizeTarget: 240,
			Clock:        0,
		},
	}
}
