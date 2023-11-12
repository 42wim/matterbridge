package types

// SubscriptionOptions represents the parameters passed to Subscribe()
// to customize the subscription behavior.
type SubscriptionOptions struct {
	PrivateKeyID string
	SymKeyID     string
	PoW          float64
	PubsubTopic  string
	Topics       [][]byte
}
