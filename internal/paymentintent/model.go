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
	Id                  string     `json:"id,omitempty"`
	PaymentReference    string     `json:"payment_reference"`
	BankAuthorizationId string     `json:"bank_authorization_id,omitempty"`
	Amount              int        `json:"amount"`
	Currency            string     `json:"currency"`
	OrderId             string     `json:"order_id"`
	CustomerId          string     `json:"customer_id"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}
