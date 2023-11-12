package images

type CroppedImage struct {
	ImagePath string `json:"imagePath"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}
