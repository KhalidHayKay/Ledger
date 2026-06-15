package paymentprocess

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *PostgresRepo) Create(ctx context.Context, intentId, status, bankAuthId string) (PaymentProcess, error) {
	var process PaymentProcess
	err := r.db.QueryRow(ctx, `
		INSERT INTO payment_processes (
			payment_intent_id, type, external_id
		)
		VALUES ($1, $2, $3)
		RETURNING id::text, payment_intent_id, type, external_id, metadata, created_at
	`,
		intentId,
		status,
		bankAuthId,
	).Scan(
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

func (r *PostgresRepo) GetCurrentProcess(ctx context.Context, paymentIntentId, paymentIntentStatus string) (PaymentProcess, error) {
	var process PaymentProcess
	err := r.db.QueryRow(ctx, `
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

func (r *PostgresRepo) GetByIntentAndStatus(ctx context.Context, intentId, status string) (PaymentProcess, error) {
	var p PaymentProcess
	err := r.db.QueryRow(ctx, `
        SELECT id, payment_intent_id, type, external_id, created_at
        FROM payment_processes
        WHERE payment_intent_id = $1
          AND type = $2
    `, intentId, status).Scan(
		&p.Id,
		&p.PaymentIntentId,
		&p.Type,
		&p.ExternalId,
		&p.CreatedAt,
	)

	if err != nil {
		return PaymentProcess{}, err
	}

	return p, nil
}
