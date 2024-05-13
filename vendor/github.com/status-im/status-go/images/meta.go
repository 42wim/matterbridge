package images

const (
	UNKNOWN ImageType = 1 + iota

	// Raster image types
	JPEG
	PNG
	GIF
	WEBP
	ICO
)

const (
	MaxJpegQuality = 80
	MinJpegQuality = 50

	SmallDim = ResizeDimension(80)
	LargeDim = ResizeDimension(240)

	BannerDim = ResizeDimension(800)

	SmallDimName = "thumbnail"
	LargeDimName = "large"

	BannerIdentityName = "banner"
)

var (
	// ResizeDimensions list of all available image resize sizes
	ResizeDimensions = []ResizeDimension{SmallDim, LargeDim}

	// DimensionSizeLimit the size limits imposed on each resize dimension
	// Figures are based on the following sample data https://github.com/status-im/status-mobile/issues/11047#issuecomment-694970473
	DimensionSizeLimit = map[ResizeDimension]FileSizeLimits{
		SmallDim: {
			Ideal: 2560, // Base on the largest sample image at quality 60% (2,554 bytes ∴ 1024 * 2.5)
			Max:   5632, // Base on the largest sample image at quality 80% + 50% margin (3,683 bytes * 1.5 ≈ 5500 ∴ 1024 * 5.5)
		},
		LargeDim: {
			Ideal: 16384, // Base on the largest sample image at quality 60% (16,143 bytes ∴ 1024 * 16)
			Max:   38400, // Base on the largest sample image at quality 80% + 50% margin (24,290 bytes * 1.5 ≈ 37500 ∴ 1024 * 37.5)
		},
	}

	// ResizeDimensionToName maps a ResizeDimension to its assigned string name
	ResizeDimensionToName = map[ResizeDimension]string{
		SmallDim: SmallDimName,
		LargeDim: LargeDimName,
	}

	// NameToResizeDimension maps a string name to its assigned ResizeDimension
	NameToResizeDimension = map[string]ResizeDimension{
		SmallDimName: SmallDim,
		LargeDimName: LargeDim,
	}
)

type FileSizeLimits struct {
	Ideal int
	Max   int
}

type ImageType uint
type ResizeDimension uint

func GetBannerDimensionLimits() FileSizeLimits {
	return FileSizeLimits{
		Ideal: 307200, // We want to save space and traffic but keep to maximum compression
		Max:   460800, // Can't go bigger than 450 KB
	}
}
