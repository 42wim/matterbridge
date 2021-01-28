package object // import "github.com/SevereCloud/vksdk/v2/object"

// OrdersAmount struct.
type OrdersAmount struct {
	Amounts  []OrdersAmountItem `json:"amounts"`
	Currency string             `json:"currency"` // Currency name
}

// OrdersAmountItem struct.
type OrdersAmountItem struct {
	Amount      int    `json:"amount"`      // Votes amount in user's currency
	Description string `json:"description"` // Amount description
	Votes       string `json:"votes"`       // Votes number
}

// OrdersOrder struct.
type OrdersOrder struct {
	Amount              int    `json:"amount"`                // Amount
	AppOrderID          int    `json:"app_order_id"`          // App order ID
	CancelTransactionID int    `json:"cancel_transaction_id"` // Cancel transaction ID
	Date                int    `json:"date"`                  // Date of creation in Unixtime
	ID                  int    `json:"id"`                    // Order ID
	Item                string `json:"item"`                  // Order item
	ReceiverID          int    `json:"receiver_id"`           // Receiver ID
	Status              string `json:"status"`                // Order status
	TransactionID       int    `json:"transaction_id"`        // Transaction ID
	UserID              int    `json:"user_id"`               // User ID
}

// OrdersSubscription struct.
type OrdersSubscription struct {
	CancelReason    string      `json:"cancel_reason"`     // Cancel reason
	CreateTime      int         `json:"create_time"`       // Date of creation in Unixtime
	ID              int         `json:"id"`                // Subscription ID
	ItemID          string      `json:"item_id"`           // Subscription order item
	NextBillTime    int         `json:"next_bill_time"`    // Date of next bill in Unixtime
	Period          int         `json:"period"`            // Subscription period
	PeriodStartTime int         `json:"period_start_time"` // Date of last period start in Unixtime
	Price           int         `json:"price"`             // Subscription price
	Status          string      `json:"status"`            // Subscription status
	PendingCancel   BaseBoolInt `json:"pending_cancel"`    // Pending cancel state
	TestMode        BaseBoolInt `json:"test_mode"`         // Is test subscription
	TrialExpireTime int         `json:"trial_expire_time"` // Date of trial expire in Unixtime
	UpdateTime      int         `json:"update_time"`       // Date of last change in Unixtime
}
