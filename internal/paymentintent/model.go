package paymentintent

import (
	"time"
)

const (
	PaymentStatusPending    string = "pending"
	PaymentStatusAuthorized string = "authorized"
	PaymentStatusCaptured   string = "captured"
	PaymentStatusRefunded   string = "refunded"
	PaymentStatusVoided     string = "voided"
	PaymentStatusFailed     string = "failed"
)

type PaymentIntent struct {
	Id               string     `json:"id,omitempty"`
	PaymentReference string     `json:"payment_reference"`
	Amount           int        `json:"amount"`
	Currency         string     `json:"currency"`
	OrderId          string     `json:"order_id"`
	CustomerId       string     `json:"customer_id"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`

	CurrentPaymentProcess *PaymentProcess `json:"payment_process,omitempty"`
}

type PaymentProcess struct {
	Id              string    `json:"id,omitempty"`
	PaymentIntentId string    `json:"payment_intent_id"`
	Type            string    `json:"type"`
	ExternalId      string    `json:"external_id,omitempty"`
	Metadata        []byte    `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
