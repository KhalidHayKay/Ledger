package paymentintent

import (
	"ledger/internal/paymentprocess"
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
	Id         string     `json:"id,omitempty"`
	PaymentRef string     `json:"payment_reference"`
	Amount     int        `json:"amount"`
	Currency   string     `json:"currency"`
	OrderId    string     `json:"order_id"`
	CustomerId string     `json:"customer_id"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`

	CurrentPaymentProcess *paymentprocess.PaymentProcess `json:"payment_process,omitempty"`
}
