package object // import "github.com/SevereCloud/vksdk/v2/object"

// LeadsChecked struct.
type LeadsChecked struct {
	Reason    string `json:"reason"` // Reason why user can't start the lead
	Result    string `json:"result"`
	Sid       string `json:"sid"`        // Session ID
	StartLink string `json:"start_link"` // URL user should open to start the lead
}

// LeadsComplete struct.
type LeadsComplete struct {
	Cost     int         `json:"cost"`  // Offer cost
	Limit    int         `json:"limit"` // Offer limit
	Spent    int         `json:"spent"` // Amount of spent votes
	Success  BaseBoolInt `json:"success"`
	TestMode BaseBoolInt `json:"test_mode"` // Information whether test mode is enabled
}

// LeadsEntry struct.
type LeadsEntry struct {
	Aid       int         `json:"aid"`        // Application ID
	Comment   string      `json:"comment"`    // Comment text
	Date      int         `json:"date"`       // Date when the action has been started in Unixtime
	Sid       string      `json:"sid"`        // Session string ID
	StartDate int         `json:"start_date"` // Start date in Unixtime (for status=2)
	Status    int         `json:"status"`     // Action type
	TestMode  BaseBoolInt `json:"test_mode"`  // Information whether test mode is enabled
	UID       int         `json:"uid"`        // User ID
}

// LeadsLead struct.
type LeadsLead struct {
	Completed   int           `json:"completed"` // Completed offers number
	Cost        int           `json:"cost"`      // Offer cost
	Days        LeadsLeadDays `json:"days"`
	Impressions int           `json:"impressions"` // Impressions number
	Limit       int           `json:"limit"`       // Lead limit
	Spent       int           `json:"spent"`       // Amount of spent votes
	Started     int           `json:"started"`     // Started offers number
}

// LeadsLeadDays struct.
type LeadsLeadDays struct {
	Completed   int `json:"completed"`   // Completed offers number
	Impressions int `json:"impressions"` // Impressions number
	Spent       int `json:"spent"`       // Amount of spent votes
	Started     int `json:"started"`     // Started offers number
}

// LeadsStart struct.
type LeadsStart struct {
	TestMode BaseBoolInt `json:"test_mode"` // Information whether test mode is enabled
	VkSid    string      `json:"vk_sid"`    // Session data
}
