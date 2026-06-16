package paymentevent

import "context"

type Repository interface {
	Create(ctx context.Context, intentId, status, bankAuthId string) (PaymentEvent, error)
	// GetByIntent(ctx context.Context, intentId string) ([]PaymentEvent, error)
	GetByIntentAndStatus(ctx context.Context, intentId, status string) (PaymentEvent, error)
}
