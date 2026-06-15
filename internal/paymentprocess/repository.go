package paymentprocess

import "context"

type Repository interface {
	Create(ctx context.Context, intentId, status, bankAuthId string) (PaymentProcess, error)
	// GetByIntent(ctx context.Context, intentId string) ([]PaymentProcess, error)
	GetByIntentAndStatus(ctx context.Context, intentId, status string) (PaymentProcess, error)
}
