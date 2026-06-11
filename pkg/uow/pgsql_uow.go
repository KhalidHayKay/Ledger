package uow

import (
	"context"
	"ledger/internal/paymentintent"
	"ledger/internal/paymentprocess"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgsqlUoW struct{ db *pgxpool.Pool }

func NewPgsqlUoW(db *pgxpool.Pool) *PgsqlUoW { return &PgsqlUoW{db} }

func (u *PgsqlUoW) RunInTx(ctx context.Context, fn func(Repos) error) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return err
	}

	repos := Repos{
		PaymentIntent:  paymentintent.NewPostgresRepo(tx),
		PaymentProcess: paymentprocess.NewPostgresRepo(tx),
	}

	if err := fn(repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
