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
)

type PaymentIntent struct {
	Id               string    `json:"id,omitempty"`
	ClientId         string    `json:"client_id"`
	PaymentReference string    `json:"payment_reference"`
	Amount           int       `json:"amount"`
	Currency         string    `json:"currency"`
	OrderId          string    `json:"order_id"`
	CustomerId       string    `json:"customer_id"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}
