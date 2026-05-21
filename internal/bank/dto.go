package bank

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
	Card           Card
	Amount         Amount
	IdempotencyKey string
}
