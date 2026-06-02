package paymentintent

import "context"

type Repository interface {
	Create(ctx context.Context, amount int, currency, orderId, customerId string) (string, error)
	UpdateBankAuth(ctx context.Context, paymentRef, bankAuthId string) (PaymentIntent, error)
	GetByPaymentRef(ctx context.Context, paymentReference string) (PaymentIntent, error)
}
