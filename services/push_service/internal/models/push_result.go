package models

// PushResult captures the delivery outcome per device token.
type PushResult struct {
	Token     string `json:"token"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

const (
	// ResultDelivered indicates the push was acknowledged by the provider.
	ResultDelivered = "delivered"
	// ResultFailed indicates an unrecoverable failure.
	ResultFailed = "failed"
)
