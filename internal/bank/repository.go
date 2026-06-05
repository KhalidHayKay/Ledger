package bank

import "context"

type Repository interface {
	Authorize(ctx context.Context, idempotencyKey string, input AuthorizeInput) (Payment, error)
	Capture(ctx context.Context, idempotencyKey, authorizationId, amount string) (Payment, error)
	Void(ctx context.Context, idempotencyKey, authorizationId string) (Payment, error)
	Refund(ctx context.Context, idempotencyKey, captureId, amount string) (Payment, error)
}
