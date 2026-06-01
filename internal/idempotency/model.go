package idempotency

type Entry struct {
	RequestHash string
	PaymentRef  string
}
