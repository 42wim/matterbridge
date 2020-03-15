// Code generated by msgraph-generate.go DO NOT EDIT.

package msgraph

import "time"

// ProvisioningObjectSummary undocumented
type ProvisioningObjectSummary struct {
	// Entity is the base model of ProvisioningObjectSummary
	Entity
	// ActivityDateTime undocumented
	ActivityDateTime *time.Time `json:"activityDateTime,omitempty"`
	// TenantID undocumented
	TenantID *string `json:"tenantId,omitempty"`
	// JobID undocumented
	JobID *string `json:"jobId,omitempty"`
	// CycleID undocumented
	CycleID *string `json:"cycleId,omitempty"`
	// ChangeID undocumented
	ChangeID *string `json:"changeId,omitempty"`
	// Action undocumented
	Action *string `json:"action,omitempty"`
	// DurationInMilliseconds undocumented
	DurationInMilliseconds *int `json:"durationInMilliseconds,omitempty"`
	// InitiatedBy undocumented
	InitiatedBy *Initiator `json:"initiatedBy,omitempty"`
	// SourceSystem undocumented
	SourceSystem *ProvisioningSystemDetails `json:"sourceSystem,omitempty"`
	// TargetSystem undocumented
	TargetSystem *ProvisioningSystemDetails `json:"targetSystem,omitempty"`
	// SourceIdentity undocumented
	SourceIdentity *ProvisionedIdentity `json:"sourceIdentity,omitempty"`
	// TargetIdentity undocumented
	TargetIdentity *ProvisionedIdentity `json:"targetIdentity,omitempty"`
	// StatusInfo undocumented
	StatusInfo *StatusBase `json:"statusInfo,omitempty"`
	// ProvisioningSteps undocumented
	ProvisioningSteps []ProvisioningStep `json:"provisioningSteps,omitempty"`
	// ModifiedProperties undocumented
	ModifiedProperties []ModifiedProperty `json:"modifiedProperties,omitempty"`
}

// ProvisioningStep undocumented
type ProvisioningStep struct {
	// Object is the base model of ProvisioningStep
	Object
	// Name undocumented
	Name *string `json:"name,omitempty"`
	// Status undocumented
	Status *ProvisioningResult `json:"status,omitempty"`
	// Description undocumented
	Description *string `json:"description,omitempty"`
	// Details undocumented
	Details *DetailsInfo `json:"details,omitempty"`
	// ProvisioningStepType undocumented
	ProvisioningStepType *ProvisioningStepType `json:"provisioningStepType,omitempty"`
}

// ProvisioningSystemDetails undocumented
type ProvisioningSystemDetails struct {
	// Object is the base model of ProvisioningSystemDetails
	Object
	// ID undocumented
	ID *string `json:"id,omitempty"`
	// DisplayName undocumented
	DisplayName *string `json:"displayName,omitempty"`
	// Details undocumented
	Details *DetailsInfo `json:"details,omitempty"`
}
