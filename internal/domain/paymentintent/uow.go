package paymentintent

import (
	"context"
	"ledger/internal/domain/paymentevent"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TxRepos struct {
	PaymentIntent Repository
	PaymentEvent  paymentevent.Repository
}

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(TxRepos) error) error
}

type PgsqlUoW struct{ db *pgxpool.Pool }

func NewPgsqlUoW(db *pgxpool.Pool) *PgsqlUoW { return &PgsqlUoW{db} }

func (u *PgsqlUoW) RunInTx(ctx context.Context, fn func(TxRepos) error) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return err
	}
	repos := TxRepos{
		PaymentIntent: NewPostgresRepo(tx),
		PaymentEvent:  paymentevent.NewPostgresRepo(tx),
	}
	if err := fn(repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
