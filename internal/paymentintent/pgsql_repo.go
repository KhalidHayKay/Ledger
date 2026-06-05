package paymentintent

import (
	"context"
	"ledger/utils"

	"github.com/jackc/pgx/v5"
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
	amount int,
	currency, orderId, customerId string,
) (string, error) {
	tx, err := r.pgsql.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var intentId int64
	err = tx.QueryRow(ctx, `
		INSERT INTO payment_intents (
			amount, currency, order_id, customer_id, status
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`,
		amount,
		currency,
		orderId,
		customerId,
		PaymentStatusPending,
	).Scan(&intentId)
	if err != nil {
		return "", err
	}

	paymentReference := utils.GeneratePaymentRef(intentId)

	_, err = tx.Exec(ctx, `
		UPDATE payment_intents
		SET payment_reference = $1
		WHERE id = $2
	`,
		paymentReference,
		intentId,
	)
	if err != nil {
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", err
	}

	return paymentReference, nil
}

func (r *PostgresRepo) UpdateBankAuth(ctx context.Context, paymentRef, bankAuthId string) (PaymentIntent, error) {
	tx, err := r.pgsql.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PaymentIntent{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var intent PaymentIntent
	err = tx.QueryRow(ctx, `
		UPDATE payment_intents
		SET status = $1
		WHERE payment_reference = $2
		RETURNING
			id,
			payment_reference,
			amount,
			currency,
			order_id,
			customer_id,
			status,
			created_at
	`, PaymentStatusAuthorized, paymentRef).Scan(
		&intent.Id,
		&intent.PaymentReference,
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

	var process PaymentProcess
	err = tx.QueryRow(ctx, `
		INSERT INTO payment_processes (
			payment_intent_id, type, external_id
		)
		VALUES ($1, $2, $3)
		RETURNING id, payment_intent_id, type, external_id, created_at
	`,
		intent.Id,
		PaymentStatusAuthorized,
		bankAuthId,
	).Scan(
		&process.Id,
		&process.PaymentIntentId,
		&process.Type,
		&process.ExternalId,
		&process.CreatedAt,
	)
	if err != nil {
		return PaymentIntent{}, err
	}
	intent.CurrentPaymentProcess = &process

	err = tx.Commit(ctx)
	if err != nil {
		return PaymentIntent{}, err
	}

	return intent, nil
}

func (r *PostgresRepo) GetByPaymentRef(ctx context.Context, paymentReference string) (PaymentIntent, error) {
	var intent PaymentIntent

	err := r.pgsql.QueryRow(ctx, `
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
	`, paymentReference).Scan(
		&intent.Id,
		&intent.PaymentReference,
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

	process, err := r.GetCurrentProcess(ctx, intent.Id, intent.Status)
	if err != nil {
		return PaymentIntent{}, err
	}

	intent.CurrentPaymentProcess = &process

	return intent, nil
}

func (r *PostgresRepo) GetCurrentProcess(ctx context.Context, paymentIntentId, paymentIntentStatus string) (PaymentProcess, error) {
	var process PaymentProcess
	err := r.pgsql.QueryRow(ctx, `
		SELECT
			id,
			payment_intent_id,
			type,
			external_id,
			metadata,
			created_at
		FROM payment_processes
		WHERE payment_intent_id = $1
			AND type = $2
	`, paymentIntentId, paymentIntentStatus).Scan(
		&process.Id,
		&process.PaymentIntentId,
		&process.Type,
		&process.ExternalId,
		&process.Metadata,
		&process.CreatedAt,
	)
	if err != nil {
		return PaymentProcess{}, err
	}

	return process, nil
}
