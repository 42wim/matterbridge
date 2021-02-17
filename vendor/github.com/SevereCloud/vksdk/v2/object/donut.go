package object

// DonutDonatorSubscriptionInfo struct.
type DonutDonatorSubscriptionInfo struct {
	OwnerID         int    `json:"owner_id"`
	NextPaymentDate int    `json:"next_payment_date"`
	Amount          int    `json:"amount"`
	Status          string `json:"status"`
}
