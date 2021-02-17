package object

// SearchHint struct.
type SearchHint struct {
	Description string      `json:"description"`      // Object description
	Global      int         `json:"global,omitempty"` // Information whether the object has been found globally
	Group       GroupsGroup `json:"group,omitempty"`
	Profile     UsersUser   `json:"profile,omitempty"`
	Section     string      `json:"section"`
	Type        string      `json:"type"`
}
