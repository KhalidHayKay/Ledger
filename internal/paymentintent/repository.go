package paymentintent

import "context"

type Repository interface {
	Create(ctx context.Context, amount int, currency, orderId, customerId string) PaymentIntent
}
