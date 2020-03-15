// Code generated by msgraph-generate.go DO NOT EDIT.

package msgraph

import "time"

// ContentClassification undocumented
type ContentClassification struct {
	// Object is the base model of ContentClassification
	Object
	// SensitiveTypeID undocumented
	SensitiveTypeID *string `json:"sensitiveTypeId,omitempty"`
	// UniqueCount undocumented
	UniqueCount *int `json:"uniqueCount,omitempty"`
	// Confidence undocumented
	Confidence *int `json:"confidence,omitempty"`
	// Matches undocumented
	Matches []MatchLocation `json:"matches,omitempty"`
}

// ContentInfo undocumented
type ContentInfo struct {
	// Object is the base model of ContentInfo
	Object
	// Format undocumented
	Format *ContentFormat `json:"format,omitempty"`
	// State undocumented
	State *ContentState `json:"state,omitempty"`
	// Identifier undocumented
	Identifier *string `json:"identifier,omitempty"`
	// Metadata undocumented
	Metadata []KeyValuePair `json:"metadata,omitempty"`
}

// ContentMetadata undocumented
type ContentMetadata struct {
	// Object is the base model of ContentMetadata
	Object
}

// ContentProperties undocumented
type ContentProperties struct {
	// Object is the base model of ContentProperties
	Object
	// Extensions undocumented
	Extensions []string `json:"extensions,omitempty"`
	// Metadata undocumented
	Metadata *ContentMetadata `json:"metadata,omitempty"`
	// LastModifiedDateTime undocumented
	LastModifiedDateTime *time.Time `json:"lastModifiedDateTime,omitempty"`
	// LastModifiedBy undocumented
	LastModifiedBy *string `json:"lastModifiedBy,omitempty"`
}

// ContentType undocumented
type ContentType struct {
	// Entity is the base model of ContentType
	Entity
	// Description undocumented
	Description *string `json:"description,omitempty"`
	// Group undocumented
	Group *string `json:"group,omitempty"`
	// Hidden undocumented
	Hidden *bool `json:"hidden,omitempty"`
	// InheritedFrom undocumented
	InheritedFrom *ItemReference `json:"inheritedFrom,omitempty"`
	// Name undocumented
	Name *string `json:"name,omitempty"`
	// Order undocumented
	Order *ContentTypeOrder `json:"order,omitempty"`
	// ParentID undocumented
	ParentID *string `json:"parentId,omitempty"`
	// ReadOnly undocumented
	ReadOnly *bool `json:"readOnly,omitempty"`
	// Sealed undocumented
	Sealed *bool `json:"sealed,omitempty"`
	// ColumnLinks undocumented
	ColumnLinks []ColumnLink `json:"columnLinks,omitempty"`
}

// ContentTypeInfo undocumented
type ContentTypeInfo struct {
	// Object is the base model of ContentTypeInfo
	Object
	// ID undocumented
	ID *string `json:"id,omitempty"`
	// Name undocumented
	Name *string `json:"name,omitempty"`
}

// ContentTypeOrder undocumented
type ContentTypeOrder struct {
	// Object is the base model of ContentTypeOrder
	Object
	// Default undocumented
	Default *bool `json:"default,omitempty"`
	// Position undocumented
	Position *int `json:"position,omitempty"`
}
