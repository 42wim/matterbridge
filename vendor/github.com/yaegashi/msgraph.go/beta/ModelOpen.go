// Code generated by msgraph.go/gen DO NOT EDIT.

package msgraph

// OpenShift undocumented
type OpenShift struct {
	// ChangeTrackedEntity is the base model of OpenShift
	ChangeTrackedEntity
	// SharedOpenShift undocumented
	SharedOpenShift *OpenShiftItem `json:"sharedOpenShift,omitempty"`
	// DraftOpenShift undocumented
	DraftOpenShift *OpenShiftItem `json:"draftOpenShift,omitempty"`
	// SchedulingGroupID undocumented
	SchedulingGroupID *string `json:"schedulingGroupId,omitempty"`
}

// OpenShiftChangeRequestObject undocumented
type OpenShiftChangeRequestObject struct {
	// ScheduleChangeRequestObject is the base model of OpenShiftChangeRequestObject
	ScheduleChangeRequestObject
	// OpenShiftID undocumented
	OpenShiftID *string `json:"openShiftId,omitempty"`
}

// OpenShiftItem undocumented
type OpenShiftItem struct {
	// ShiftItem is the base model of OpenShiftItem
	ShiftItem
	// OpenSlotCount undocumented
	OpenSlotCount *int `json:"openSlotCount,omitempty"`
}

// OpenTypeExtension undocumented
type OpenTypeExtension struct {
	// Extension is the base model of OpenTypeExtension
	Extension
	// ExtensionName undocumented
	ExtensionName *string `json:"extensionName,omitempty"`
}
