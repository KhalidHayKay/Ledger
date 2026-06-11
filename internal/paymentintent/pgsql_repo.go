package paymentintent

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
}

type PostgresRepo struct {
	db DBTX
}

func NewPostgresRepo(db DBTX) *PostgresRepo {
	return &PostgresRepo{db}
}

func (r *PostgresRepo) Tx(ctx context.Context, fn func(Repository) error) error {
	db, ok := r.db.(*pgxpool.Pool)
	if !ok {
		return errors.New(
			"cannot start tx: already inside a transaction")
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(NewPostgresRepo(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepo) Create(
	ctx context.Context,
	amount int,
	currency, orderId, customerId string,
) (PaymentIntent, error) {
	var intent PaymentIntent
	err := r.db.QueryRow(ctx, `
		INSERT INTO payment_intents (
			amount, currency, order_id, customer_id, status
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, amount, currency, order_id, customer_id, status, created_at
	`,
		amount,
		currency,
		orderId,
		customerId,
		PaymentStatusPending,
	).Scan(
		&intent.Id,
		&intent.Amount,
		&intent.Currency,
		&intent.OrderId,
		&intent.CustomerId,
		&intent.Status,
		&intent.CreatedAt,
	)
	if err != nil {
		return PaymentIntent{}, err
	}

	return intent, nil
}

func (r *PostgresRepo) CreateReference(ctx context.Context, id, generatedRef string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE payment_intents
		SET payment_reference = $1
		WHERE id = $2
	`,
		generatedRef,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepo) UpdateState(ctx context.Context, paymentRef, state string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE payment_intents
		SET status = $1
		WHERE payment_reference = $2
	`, state, paymentRef)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepo) GetByPaymentRef(ctx context.Context, paymentRef string) (PaymentIntent, error) {
	var intent PaymentIntent

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			payment_reference,
			amount,
			currency,
			order_id,
			customer_id,
			status,
			created_at
		 FROM payment_intents
		 WHERE payment_reference = $1
	`, paymentRef).Scan(
		&intent.Id,
		&intent.PaymentRef,
		&intent.Amount,
		&intent.Currency,
		&intent.OrderId,
		&intent.CustomerId,
		&intent.Status,
		&intent.CreatedAt,
	)
	if err != nil {
		return PaymentIntent{}, err
	}

	return intent, nil
}
