package object

// LeadFormsForm struct.
type LeadFormsForm struct {
	FormID        int         `json:"form_id"`
	GroupID       int         `json:"group_id"`
	Photo         interface{} `json:"photo"`
	Name          string      `json:"name"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Confirmation  string      `json:"confirmation"`
	SiteLinkURL   string      `json:"site_link_url"`
	PolicyLinkURL string      `json:"policy_link_url"`
	Questions     []struct {
		Type    string `json:"type"`
		Key     string `json:"key"`
		Label   string `json:"label,omitempty"`
		Options []struct {
			Label string `json:"label"`
			Key   string `json:"key"`
		} `json:"options,omitempty"`
	} `json:"questions"`
	Active       int    `json:"active"`
	LeadsCount   int    `json:"leads_count"`
	PixelCode    string `json:"pixel_code"`
	OncePerUser  int    `json:"once_per_user"`
	NotifyAdmins string `json:"notify_admins"`
	NotifyEmails string `json:"notify_emails"`
	URL          string `json:"url"`
}

// LeadFormsLead struct.
type LeadFormsLead struct {
	LeadID  string `json:"lead_id"`
	UserID  string `json:"user_id"`
	Date    string `json:"date"`
	Answers []struct {
		Key    string `json:"key"`
		Answer struct {
			Value string `json:"value"`
		} `json:"answer"`
	} `json:"answers"`
	AdID string `json:"ad_id"`
}
