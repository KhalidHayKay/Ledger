package uow

import (
	"context"
	"ledger/internal/paymentintent"
	"ledger/internal/paymentprocess"
)

type Repos struct {
	PaymentIntent  paymentintent.Repository
	PaymentProcess paymentprocess.Repository
}

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(repos Repos) error) error
}
