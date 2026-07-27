package bank

import "context"

type Repository interface {
	Authorize(ctx context.Context, idempotencyKey string, input AuthorizeInput) (Payment, error)
	Capture(ctx context.Context, idempotencyKey, authorizationId string, amount int) (Payment, error)
	Void(ctx context.Context, idempotencyKey, authorizationId string) (Payment, error)
	Refund(ctx context.Context, idempotencyKey, captureId string, amount int) (Payment, error)
}
