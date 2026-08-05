package paymentintent

import "context"

type Repository interface {
	Create(ctx context.Context, amount int, currency, orderId, customerId string) (PaymentIntent, error)
	CreateReference(ctx context.Context, id, generatedRef string) error
	UpdateState(ctx context.Context, paymentRef, state string) error
	GetByRef(ctx context.Context, paymentRef string) (PaymentIntent, error)
	GetWithEvent(ctx context.Context, paymentRef, state string) (PaymentIntent, error)
	UpdateOperation(ctx context.Context, paymentRef, operation string) error
}
