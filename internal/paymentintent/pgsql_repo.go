package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/paymentevent"

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

func (r *PostgresRepo) GetByRef(ctx context.Context, paymentRef string) (PaymentIntent, error) {
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

func (r *PostgresRepo) GetWithEvent(ctx context.Context, paymentRef, state string) (PaymentIntent, error) {
	query := `SELECT
			pi.id,
			pi.payment_reference,
			pi.amount,
			pi.currency,
			pi.order_id,
			pi.customer_id,
			pi.status,
			pi.created_at,
			pi.updated_at,

			pe.id,
			pe.payment_intent_id,
			pe.state,
			pe.external_state_id,
			pe.metadata,
			pe.created_at
		FROM payment_intents pi
		JOIN payment_events pe
			ON pe.payment_intent_id = pi.id
		WHERE pi.payment_reference = $1
			AND pi.status = $2
			AND pe.state = $2
	`

	var pi PaymentIntent
	var pe paymentevent.PaymentEvent

	err := r.db.QueryRow(ctx, query, paymentRef, state).Scan(
		&pi.Id,
		&pi.PaymentRef,
		&pi.Amount,
		&pi.Currency,
		&pi.OrderId,
		&pi.CustomerId,
		&pi.Status,
		&pi.CreatedAt,
		&pi.UpdatedAt,

		&pe.Id,
		&pe.PaymentIntentId,
		&pe.State,
		&pe.ExternalStateId,
		&pe.Metadata,
		&pe.CreatedAt,
	)
	if err != nil {
		return PaymentIntent{}, err
	}

	pi.CurrentEvent = &pe
	return pi, nil
}
