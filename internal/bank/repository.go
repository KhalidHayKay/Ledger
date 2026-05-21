package bank

import "context"

type Repository interface {
	Authorize(ctx context.Context, input AuthorizeInput) (Payment, error)
}
