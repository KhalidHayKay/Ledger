package paymentevent

import "time"

type PaymentEvent struct {
	Id              string    `json:"id,omitempty"`
	PaymentIntentId string    `json:"payment_intent_id"`
	State           string    `json:"state"`
	ExternalId      string    `json:"external_id,omitempty"`
	Metadata        []byte    `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
