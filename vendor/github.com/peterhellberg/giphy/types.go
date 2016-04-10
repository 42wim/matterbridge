package giphy

import "encoding/json"

// Search represents a search response from the Giphy API
type Search struct {
	Data       []Data     `json:"data"`
	Meta       Meta       `json:"meta"`
	Pagination Pagination `json:"pagination"`
}

// GIF represents a ID response from the Giphy API
type GIF struct {
	Data    Data
	RawData json.RawMessage `json:"data"`
	Meta    Meta            `json:"meta"`
}

// Random represents a random response from the Giphy API
type Random struct {
	Data    RandomData
	RawData json.RawMessage `json:"data"`
	Meta    Meta            `json:"meta"`
}

// Translate represents a translate response from the Giphy API
type Translate struct {
	Data
	RawData json.RawMessage `json:"data"`
	Meta    Meta            `json:"meta"`
}

// Trending represents a trending response from the Giphy API
type Trending struct {
	Data       []Data     `json:"data"`
	Meta       Meta       `json:"meta"`
	Pagination Pagination `json:"pagination"`
}

// Data contains all the fields in a data response from the Giphy API
type Data struct {
	Type             string `json:"type"`
	ID               string `json:"id"`
	URL              string `json:"url"`
	BitlyGifURL      string `json:"bitly_gif_url"`
	BitlyURL         string `json:"bitly_url"`
	EmbedURL         string `json:"embed_url"`
	Username         string `json:"username"`
	Source           string `json:"source"`
	Rating           string `json:"rating"`
	Caption          string `json:"caption"`
	ContentURL       string `json:"content_url"`
	ImportDatetime   string `json:"import_datetime"`
	TrendingDatetime string `json:"trending_datetime"`
	Images           Images `json:"images"`
}

// RandomData represents data section in random response from the Giphy API
type RandomData struct {
	Type                         string   `json:"type"`
	ID                           string   `json:"id"`
	URL                          string   `json:"url"`
	ImageOriginalURL             string   `json:"image_original_url"`
	ImageURL                     string   `json:"image_url"`
	ImageMp4URL                  string   `json:"image_mp4_url"`
	ImageFrames                  string   `json:"image_frames"`
	ImageWidth                   string   `json:"image_width"`
	ImageHeight                  string   `json:"image_height"`
	FixedHeightDownsampledURL    string   `json:"fixed_height_downsampled_url"`
	FixedHeightDownsampledWidth  string   `json:"fixed_height_downsampled_width"`
	FixedHeightDownsampledHeight string   `json:"fixed_height_downsampled_height"`
	FixedWidthDownsampledURL     string   `json:"fixed_width_downsampled_url"`
	FixedWidthDownsampledWidth   string   `json:"fixed_width_downsampled_width"`
	FixedWidthDownsampledHeight  string   `json:"fixed_width_downsampled_height"`
	Rating                       string   `json:"rating"`
	Username                     string   `json:"username"`
	Caption                      string   `json:"caption"`
	Tags                         []string `json:"tags"`
}

// Images represents all the different types of images
type Images struct {
	FixedHeight            Image `json:"fixed_height"`
	FixedHeightStill       Image `json:"fixed_height_still"`
	FixedHeightDownsampled Image `json:"fixed_height_downsampled"`
	FixedWidth             Image `json:"fixed_width"`
	FixedWidthStill        Image `json:"fixed_width_still"`
	FixedWidthDownsampled  Image `json:"fixed_width_downsampled"`
	Downsized              Image `json:"downsized"`
	DownsizedStill         Image `json:"downsized_still"`
	Original               Image `json:"original"`
	OriginalStill          Image `json:"original_still"`
}

// Image represents an image
type Image struct {
	URL    string `json:"url"`
	Width  string `json:"width"`
	Height string `json:"height"`
	Size   string `json:"size,omitempty"`
	Frames string `json:"frames,omitempty"`
	Mp4    string `json:"mp4,omitempty"`
}

// Pagination represents the pagination section in a Giphy API response
type Pagination struct {
	TotalCount int `json:"total_count"`
	Count      int `json:"count"`
	Offset     int `json:"offset"`
}

// Meta represents the meta section in a Giphy API response
type Meta struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}
