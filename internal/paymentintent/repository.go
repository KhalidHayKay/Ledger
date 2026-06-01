package paymentintent

import "context"

type Repository interface {
	Create(ctx context.Context, paymentReference string, amount int, currency, orderId, customerId string) (PaymentIntent, error)
	GetByPaymentRef(ctx context.Context, paymentReference string) (PaymentIntent, error)
}
