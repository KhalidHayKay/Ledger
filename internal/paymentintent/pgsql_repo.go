package paymentintent

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	pgsql *pgxpool.Pool
}

func NewPostgresRepo(pgsql *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pgsql}
}

func (r *PostgresRepo) Create(ctx context.Context, amount int, currency, orderId, customerId string) PaymentIntent {
	return PaymentIntent{}
}
