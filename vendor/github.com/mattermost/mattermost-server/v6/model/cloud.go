// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import "strings"

const (
	EventTypeFailedPayment         = "failed-payment"
	EventTypeFailedPaymentNoCard   = "failed-payment-no-card"
	EventTypeSendAdminWelcomeEmail = "send-admin-welcome-email"
	EventTypeTrialWillEnd          = "trial-will-end"
	EventTypeTrialEnded            = "trial-ended"
	JoinLimitation                 = "join"
	InviteLimitation               = "invite"
)

var MockCWS string

type BillingScheme string

const (
	BillingSchemePerSeat    = BillingScheme("per_seat")
	BillingSchemeFlatFee    = BillingScheme("flat_fee")
	BillingSchemeSalesServe = BillingScheme("sales_serve")
)

type RecurringInterval string

const (
	RecurringIntervalYearly  = RecurringInterval("year")
	RecurringIntervalMonthly = RecurringInterval("month")
)

type SubscriptionFamily string

const (
	SubscriptionFamilyCloud  = SubscriptionFamily("cloud")
	SubscriptionFamilyOnPrem = SubscriptionFamily("on-prem")
)

// Product model represents a product on the cloud system.
type Product struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	PricePerSeat      float64            `json:"price_per_seat"`
	AddOns            []*AddOn           `json:"add_ons"`
	SKU               string             `json:"sku"`
	PriceID           string             `json:"price_id"`
	Family            SubscriptionFamily `json:"product_family"`
	RecurringInterval RecurringInterval  `json:"recurring_interval"`
	BillingScheme     BillingScheme      `json:"billing_scheme"`
}

// AddOn represents an addon to a product.
type AddOn struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	PricePerSeat float64 `json:"price_per_seat"`
}

// StripeSetupIntent represents the SetupIntent model from Stripe for updating payment methods.
type StripeSetupIntent struct {
	ID           string `json:"id"`
	ClientSecret string `json:"client_secret"`
}

// ConfirmPaymentMethodRequest contains the fields for the customer payment update API.
type ConfirmPaymentMethodRequest struct {
	StripeSetupIntentID string `json:"stripe_setup_intent_id"`
	SubscriptionID      string `json:"subscription_id"`
}

// Customer model represents a customer on the system.
type CloudCustomer struct {
	CloudCustomerInfo
	ID             string         `json:"id"`
	CreatorID      string         `json:"creator_id"`
	CreateAt       int64          `json:"create_at"`
	BillingAddress *Address       `json:"billing_address"`
	CompanyAddress *Address       `json:"company_address"`
	PaymentMethod  *PaymentMethod `json:"payment_method"`
}

// CloudCustomerInfo represents editable info of a customer.
type CloudCustomerInfo struct {
	Name             string `json:"name"`
	Email            string `json:"email,omitempty"`
	ContactFirstName string `json:"contact_first_name,omitempty"`
	ContactLastName  string `json:"contact_last_name,omitempty"`
	NumEmployees     int    `json:"num_employees"`
}

// Address model represents a customer's address.
type Address struct {
	City       string `json:"city"`
	Country    string `json:"country"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	PostalCode string `json:"postal_code"`
	State      string `json:"state"`
}

// PaymentMethod represents methods of payment for a customer.
type PaymentMethod struct {
	Type      string `json:"type"`
	LastFour  string `json:"last_four"`
	ExpMonth  int    `json:"exp_month"`
	ExpYear   int    `json:"exp_year"`
	CardBrand string `json:"card_brand"`
	Name      string `json:"name"`
}

// Subscription model represents a subscription on the system.
type Subscription struct {
	ID          string   `json:"id"`
	CustomerID  string   `json:"customer_id"`
	ProductID   string   `json:"product_id"`
	AddOns      []string `json:"add_ons"`
	StartAt     int64    `json:"start_at"`
	EndAt       int64    `json:"end_at"`
	CreateAt    int64    `json:"create_at"`
	Seats       int      `json:"seats"`
	Status      string   `json:"status"`
	DNS         string   `json:"dns"`
	IsPaidTier  string   `json:"is_paid_tier"`
	LastInvoice *Invoice `json:"last_invoice"`
	IsFreeTrial string   `json:"is_free_trial"`
	TrialEndAt  int64    `json:"trial_end_at"`
}

// GetWorkSpaceNameFromDNS returns the work space name. For example from test.mattermost.cloud.com, it returns test
func (s *Subscription) GetWorkSpaceNameFromDNS() string {
	return strings.Split(s.DNS, ".")[0]
}

// Invoice model represents a cloud invoice
type Invoice struct {
	ID                 string             `json:"id"`
	Number             string             `json:"number"`
	CreateAt           int64              `json:"create_at"`
	Total              int64              `json:"total"`
	Tax                int64              `json:"tax"`
	Status             string             `json:"status"`
	Description        string             `json:"description"`
	PeriodStart        int64              `json:"period_start"`
	PeriodEnd          int64              `json:"period_end"`
	SubscriptionID     string             `json:"subscription_id"`
	Items              []*InvoiceLineItem `json:"line_items"`
	CurrentProductName string             `json:"current_product_name"`
}

// InvoiceLineItem model represents a cloud invoice lineitem tied to an invoice.
type InvoiceLineItem struct {
	PriceID      string                 `json:"price_id"`
	Total        int64                  `json:"total"`
	Quantity     float64                `json:"quantity"`
	PricePerUnit int64                  `json:"price_per_unit"`
	Description  string                 `json:"description"`
	Type         string                 `json:"type"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type CWSWebhookPayload struct {
	Event                             string               `json:"event"`
	FailedPayment                     *FailedPayment       `json:"failed_payment"`
	CloudWorkspaceOwner               *CloudWorkspaceOwner `json:"cloud_workspace_owner"`
	SubscriptionTrialEndUnixTimeStamp int64                `json:"trial_end_time_stamp"`
}

type FailedPayment struct {
	CardBrand      string `json:"card_brand"`
	LastFour       string `json:"last_four"`
	FailureMessage string `json:"failure_message"`
}

// CloudWorkspaceOwner is part of the CWS Webhook payload that contains information about the user that created the workspace from the CWS
type CloudWorkspaceOwner struct {
	UserName string `json:"username"`
}
type SubscriptionStats struct {
	RemainingSeats int    `json:"remaining_seats"`
	IsPaidTier     string `json:"is_paid_tier"`
	IsFreeTrial    string `json:"is_free_trial"`
}

type SubscriptionChange struct {
	ProductID string `json:"product_id"`
}
