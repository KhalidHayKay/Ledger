package paymentintent

import "context"

type Repository interface {
	Create(ctx context.Context, paymentId string, amount int, currency, orderId, customerId string) (PaymentIntent, error)
}
