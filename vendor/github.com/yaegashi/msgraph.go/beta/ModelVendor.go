// Code generated by msgraph.go/gen DO NOT EDIT.

package msgraph

import "time"

// Vendor undocumented
type Vendor struct {
	// Entity is the base model of Vendor
	Entity
	// Number undocumented
	Number *string `json:"number,omitempty"`
	// DisplayName undocumented
	DisplayName *string `json:"displayName,omitempty"`
	// Address undocumented
	Address *PostalAddressType `json:"address,omitempty"`
	// PhoneNumber undocumented
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	// Email undocumented
	Email *string `json:"email,omitempty"`
	// Website undocumented
	Website *string `json:"website,omitempty"`
	// TaxRegistrationNumber undocumented
	TaxRegistrationNumber *string `json:"taxRegistrationNumber,omitempty"`
	// CurrencyID undocumented
	CurrencyID *UUID `json:"currencyId,omitempty"`
	// CurrencyCode undocumented
	CurrencyCode *string `json:"currencyCode,omitempty"`
	// PaymentTermsID undocumented
	PaymentTermsID *UUID `json:"paymentTermsId,omitempty"`
	// PaymentMethodID undocumented
	PaymentMethodID *UUID `json:"paymentMethodId,omitempty"`
	// TaxLiable undocumented
	TaxLiable *bool `json:"taxLiable,omitempty"`
	// Blocked undocumented
	Blocked *string `json:"blocked,omitempty"`
	// Balance undocumented
	Balance *int `json:"balance,omitempty"`
	// LastModifiedDateTime undocumented
	LastModifiedDateTime *time.Time `json:"lastModifiedDateTime,omitempty"`
	// Picture undocumented
	Picture []Picture `json:"picture,omitempty"`
	// Currency undocumented
	Currency *Currency `json:"currency,omitempty"`
	// PaymentTerm undocumented
	PaymentTerm *PaymentTerm `json:"paymentTerm,omitempty"`
	// PaymentMethod undocumented
	PaymentMethod *PaymentMethod `json:"paymentMethod,omitempty"`
}
