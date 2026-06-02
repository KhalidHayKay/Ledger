package paymentintent

import (
	"context"
	"ledger/utils"
	"strconv"

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
	defer tx.Rollback(ctx)

	var intentId int64
	err = tx.QueryRow(ctx, `
		INSERT INTO payment_intents (
			amount,
			currency,
			order_id,
			customer_id,
			status
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

	paymentReference := utils.GeneratePaymentRef(
		strconv.FormatInt(int64(intentId), 10),
	)

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
	var intent PaymentIntent

	err := r.pgsql.QueryRow(ctx, `
		UPDATE payment_intents
		SET
			bank_authorization_id = $1,
			status = $2
		WHERE payment_reference = $3
		RETURNING
			payment_reference,
			amount,
			currency,
			order_id,
			customer_id,
			status,
			created_at
	`, bankAuthId, PaymentStatusAuthorized, paymentRef).Scan(
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

	return intent, nil
}

func (r *PostgresRepo) GetByPaymentRef(ctx context.Context, paymentReference string) (PaymentIntent, error) {
	var intent PaymentIntent

	err := r.pgsql.QueryRow(ctx, `
		SELECT
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

	return intent, nil
}
