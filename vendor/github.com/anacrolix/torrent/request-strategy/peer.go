package request_strategy

import (
	"github.com/RoaringBitmap/roaring"
)

type PeerRequestState struct {
	Interested bool
	// Expecting. TODO: This should be ordered so webseed requesters initiate in the same order they
	// were assigned.
	Requests roaring.Bitmap
	// Cancelled and waiting response
	Cancelled roaring.Bitmap
}
