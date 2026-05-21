package paymentintent

import "time"

type PaymentIntent struct {
	Id               string
	PaymentReference string
	OrderId          string
	CustomerId       string
	Amount           int
	Currency         string
	Status           string
	CreatedAt        *time.Time
	UpdatedAt        time.Time
}
