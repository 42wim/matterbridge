// Code generated by msgraph.go/gen DO NOT EDIT.

package msgraph

// ChoiceColumn undocumented
type ChoiceColumn struct {
	// Object is the base model of ChoiceColumn
	Object
	// AllowTextEntry undocumented
	AllowTextEntry *bool `json:"allowTextEntry,omitempty"`
	// Choices undocumented
	Choices []string `json:"choices,omitempty"`
	// DisplayAs undocumented
	DisplayAs *string `json:"displayAs,omitempty"`
}
