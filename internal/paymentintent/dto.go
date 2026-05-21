package paymentintent

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CardR struct {
	Number      string `json:"number"`
	CVV         string `json:"cvv"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

func (c CardR) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Number, validation.Required),
		validation.Field(&c.CVV, validation.Required),
		validation.Field(&c.ExpiryMonth,
			validation.Required,
			validation.Min(1),
			validation.Max(12),
		),
		validation.Field(&c.ExpiryYear,
			validation.Required,
			validation.Min(time.Now().Year()),
		),
	)
}

type CreatePaymentIntentRequest struct {
	Card       *Card  `json:"card"`
	Amount     int    `json:"amount"`
	Currency   string `json:"currency"`
	OrderId    string `json:"order_id"`
	CustomerId string `json:"customer_id"`
}

func (r *CreatePaymentIntentRequest) Validate() error {
	return validation.ValidateStruct(
		r,
		validation.Field(&r.Card, validation.Required),
		validation.Field(&r.Amount, validation.Required),
		validation.Field(&r.Currency, validation.Required),
		validation.Field(&r.OrderId, validation.Required),
		validation.Field(&r.CustomerId, validation.Required),
	)
}

// Bussiness DTO

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

type CreateInput struct {
	Card           Card
	Amount         Amount
	OrderId        string
	CustomerId     string
	IdempotencyKey string
}
