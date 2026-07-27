package bank

import "time"

type CardExpiry struct {
	Month int
	Year  int
}

type Card struct {
	Number string
	CVV    string
	Expiry CardExpiry
}

type Amount struct {
	Figure   int
	Currency string
}

type AuthorizeInput struct {
	Card   Card
	Amount Amount
}

type Payment struct {
	AuthorizationId string     `json:"authorization_id,omitempty"`
	CaptureId       string     `json:"capture_id,omitempty"`
	VoidId          string     `json:"void_id,omitempty"`
	RefundId        string     `json:"refund_id,omitempty"`
	Amount          int        `json:"amount,omitempty"`
	Currency        string     `json:"currency,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CapturedAt      *time.Time `json:"captured_at,omitempty"`
	VoidedAt        *time.Time `json:"voided_at,omitempty"`
	RefundedAt      *time.Time `json:"refunded_at,omitempty"`
}
