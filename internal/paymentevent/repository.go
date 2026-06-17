package paymentevent

import "context"

type Repository interface {
	Create(ctx context.Context, intentId, status, providerStateId string) (PaymentEvent, error)
	GetByIntentAndStatus(ctx context.Context, intentId, status string) (PaymentEvent, error)
}
