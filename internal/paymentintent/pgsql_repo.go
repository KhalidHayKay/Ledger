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

func (r *PostgresRepo) Create(
	ctx context.Context,
	paymentReference string,
	amount int, currency, orderId, customerId string,
) (PaymentIntent, error) {
	var intent PaymentIntent

	err := r.pgsql.QueryRow(ctx, `
		INSERT INTO payment_intents (
			payment_reference, amount, currency, order_id, customer_id, status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, payment_reference, amount, currency, order_id, customer_id, status, created_at, updated_at
	`,
		paymentReference,
		amount,
		currency,
		orderId,
		customerId,
		PaymentStatusPending,
	).Scan(
		&intent.Id,
		&intent.PaymentReference,
		&intent.Amount,
		&intent.Currency,
		&intent.OrderId,
		&intent.CustomerId,
		&intent.Status,
		&intent.CreatedAt,
		&intent.UpdatedAt,
	)
	if err != nil {
		return PaymentIntent{}, err
	}

	return intent, nil
}
