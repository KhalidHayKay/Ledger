package uow

import (
	"context"
	"ledger/internal/paymentevent"
	"ledger/internal/paymentintent"
)

type Repos struct {
	PaymentIntent paymentintent.Repository
	PaymentEvent  paymentevent.Repository
}

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(repos Repos) error) error
}
