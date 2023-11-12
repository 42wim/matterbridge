package signal

const (
	walletEvent = "wallet"
)

type UnsignedTransactions struct {
	Type         string   `json:"type"`
	Transactions []string `json:"transactions"`
}

// SendWalletEvent sends event from services/wallet/events.
func SendWalletEvent(event interface{}) {
	send(walletEvent, event)
}

func SendTransactionsForSigningEvent(transactions []string) {
	send(walletEvent, UnsignedTransactions{Type: "sing-transactions", Transactions: transactions})
}
