package socketmode

type Response struct {
	EnvelopeID string      `json:"envelope_id"`
	Payload    interface{} `json:"payload,omitempty"`
}
