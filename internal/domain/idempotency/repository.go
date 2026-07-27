package idempotency

import "context"

type Repository interface {
	SaveKey(ctx context.Context, idempotencyKey string, entry string) error
	GetByKey(ctx context.Context, idempotencyKey string) (string, error)
}
