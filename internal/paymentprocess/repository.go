package paymentprocess

import "context"

type Repository interface {
	Create(ctx context.Context, intentId, status, bankAuthId string) (PaymentProcess, error)
}
