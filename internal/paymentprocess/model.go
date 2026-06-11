package paymentprocess

import "time"

type PaymentProcess struct {
	Id              string    `json:"id,omitempty"`
	PaymentIntentId string    `json:"payment_intent_id"`
	Type            string    `json:"type"`
	ExternalId      string    `json:"external_id,omitempty"`
	Metadata        []byte    `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
