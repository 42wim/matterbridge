package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// LeadFormsCreateResponse struct.
type LeadFormsCreateResponse struct {
	FormID int    `json:"form_id"`
	URL    string `json:"url"`
}

// LeadFormsCreate leadForms.create.
//
// https://dev.vk.com/method/leadForms.create
func (vk *VK) LeadFormsCreate(params Params) (response LeadFormsCreateResponse, err error) {
	err = vk.RequestUnmarshal("leadForms.create", &response, params)
	return
}

// LeadFormsDeleteResponse struct.
type LeadFormsDeleteResponse struct {
	FormID int `json:"form_id"`
}

// LeadFormsDelete leadForms.delete.
//
// https://dev.vk.com/method/leadForms.delete
func (vk *VK) LeadFormsDelete(params Params) (response LeadFormsDeleteResponse, err error) {
	err = vk.RequestUnmarshal("leadForms.delete", &response, params)
	return
}

// LeadFormsGetResponse struct.
type LeadFormsGetResponse object.LeadFormsForm

// LeadFormsGet leadForms.get.
//
// https://dev.vk.com/method/leadForms.get
func (vk *VK) LeadFormsGet(params Params) (response LeadFormsGetResponse, err error) {
	err = vk.RequestUnmarshal("leadForms.get", &response, params)
	return
}

// LeadFormsGetLeadsResponse struct.
type LeadFormsGetLeadsResponse struct {
	Leads []object.LeadFormsLead `json:"leads"`
}

// LeadFormsGetLeads leadForms.getLeads.
//
// https://dev.vk.com/method/leadForms.getLeads
func (vk *VK) LeadFormsGetLeads(params Params) (response LeadFormsGetLeadsResponse, err error) {
	err = vk.RequestUnmarshal("leadForms.getLeads", &response, params)
	return
}

// LeadFormsGetUploadURL leadForms.getUploadURL.
//
// https://dev.vk.com/method/leadForms.getUploadURL
func (vk *VK) LeadFormsGetUploadURL(params Params) (response string, err error) {
	err = vk.RequestUnmarshal("leadForms.getUploadURL", &response, params)
	return
}

// LeadFormsListResponse struct.
type LeadFormsListResponse []object.LeadFormsForm

// LeadFormsList leadForms.list.
//
// https://dev.vk.com/method/leadForms.list
func (vk *VK) LeadFormsList(params Params) (response LeadFormsListResponse, err error) {
	err = vk.RequestUnmarshal("leadForms.list", &response, params)
	return
}

// LeadFormsUpdateResponse struct.
type LeadFormsUpdateResponse struct {
	FormID int    `json:"form_id"`
	URL    string `json:"url"`
}

// LeadFormsUpdate leadForms.update.
//
// https://dev.vk.com/method/leadForms.update
func (vk *VK) LeadFormsUpdate(params Params) (response LeadFormsUpdateResponse, err error) {
	err = vk.RequestUnmarshal("leadForms.update", &response, params)
	return
}
