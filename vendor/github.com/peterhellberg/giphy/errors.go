package giphy

import "errors"

var (
	// ErrNoImageFound is the error returned when no image was found
	ErrNoImageFound = errors.New("no image found")

	// ErrUnknown is used for unknown errors from the Giphy API
	ErrUnknown = errors.New("unknown error")

	// ErrNoTrendingImagesFound is returned when no trending images were found
	ErrNoTrendingImagesFound = errors.New("no trending images found")

	// ErrNoRawData is returned if there was no data property in response
	ErrNoRawData = errors.New("no raw data")
)
