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
	Id        string    `json:"id"`
	Amount    int       `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
