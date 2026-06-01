package bank

import "context"

type Repository interface {
	Authorize(ctx context.Context, idempotencyKey string, input AuthorizeInput) (Payment, error)
}
