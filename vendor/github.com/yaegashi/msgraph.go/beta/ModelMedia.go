// Code generated by msgraph-generate.go DO NOT EDIT.

package msgraph

// MediaConfig undocumented
type MediaConfig struct {
	// Object is the base model of MediaConfig
	Object
	// RemoveFromDefaultAudioGroup undocumented
	RemoveFromDefaultAudioGroup *bool `json:"removeFromDefaultAudioGroup,omitempty"`
}

// MediaContentRatingAustralia undocumented
type MediaContentRatingAustralia struct {
	// Object is the base model of MediaContentRatingAustralia
	Object
	// MovieRating Movies rating selected for Australia
	MovieRating *RatingAustraliaMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for Australia
	TvRating *RatingAustraliaTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingCanada undocumented
type MediaContentRatingCanada struct {
	// Object is the base model of MediaContentRatingCanada
	Object
	// MovieRating Movies rating selected for Canada
	MovieRating *RatingCanadaMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for Canada
	TvRating *RatingCanadaTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingFrance undocumented
type MediaContentRatingFrance struct {
	// Object is the base model of MediaContentRatingFrance
	Object
	// MovieRating Movies rating selected for France
	MovieRating *RatingFranceMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for France
	TvRating *RatingFranceTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingGermany undocumented
type MediaContentRatingGermany struct {
	// Object is the base model of MediaContentRatingGermany
	Object
	// MovieRating Movies rating selected for Germany
	MovieRating *RatingGermanyMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for Germany
	TvRating *RatingGermanyTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingIreland undocumented
type MediaContentRatingIreland struct {
	// Object is the base model of MediaContentRatingIreland
	Object
	// MovieRating Movies rating selected for Ireland
	MovieRating *RatingIrelandMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for Ireland
	TvRating *RatingIrelandTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingJapan undocumented
type MediaContentRatingJapan struct {
	// Object is the base model of MediaContentRatingJapan
	Object
	// MovieRating Movies rating selected for Japan
	MovieRating *RatingJapanMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for Japan
	TvRating *RatingJapanTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingNewZealand undocumented
type MediaContentRatingNewZealand struct {
	// Object is the base model of MediaContentRatingNewZealand
	Object
	// MovieRating Movies rating selected for New Zealand
	MovieRating *RatingNewZealandMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for New Zealand
	TvRating *RatingNewZealandTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingUnitedKingdom undocumented
type MediaContentRatingUnitedKingdom struct {
	// Object is the base model of MediaContentRatingUnitedKingdom
	Object
	// MovieRating Movies rating selected for United Kingdom
	MovieRating *RatingUnitedKingdomMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for United Kingdom
	TvRating *RatingUnitedKingdomTelevisionType `json:"tvRating,omitempty"`
}

// MediaContentRatingUnitedStates undocumented
type MediaContentRatingUnitedStates struct {
	// Object is the base model of MediaContentRatingUnitedStates
	Object
	// MovieRating Movies rating selected for United States
	MovieRating *RatingUnitedStatesMoviesType `json:"movieRating,omitempty"`
	// TvRating TV rating selected for United States
	TvRating *RatingUnitedStatesTelevisionType `json:"tvRating,omitempty"`
}

// MediaInfo undocumented
type MediaInfo struct {
	// Object is the base model of MediaInfo
	Object
	// URI undocumented
	URI *string `json:"uri,omitempty"`
	// ResourceID undocumented
	ResourceID *string `json:"resourceId,omitempty"`
}

// MediaPrompt undocumented
type MediaPrompt struct {
	// Prompt is the base model of MediaPrompt
	Prompt
	// MediaInfo undocumented
	MediaInfo *MediaInfo `json:"mediaInfo,omitempty"`
	// Loop undocumented
	Loop *int `json:"loop,omitempty"`
}

// MediaStream undocumented
type MediaStream struct {
	// Object is the base model of MediaStream
	Object
	// MediaType undocumented
	MediaType *Modality `json:"mediaType,omitempty"`
	// Label undocumented
	Label *string `json:"label,omitempty"`
	// SourceID undocumented
	SourceID *string `json:"sourceId,omitempty"`
	// Direction undocumented
	Direction *MediaDirection `json:"direction,omitempty"`
	// ServerMuted undocumented
	ServerMuted *bool `json:"serverMuted,omitempty"`
}
