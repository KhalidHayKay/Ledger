package paymentevent

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

func (r *PostgresRepo) Create(ctx context.Context, intentId, status, providerStateId string) (PaymentEvent, error) {
	var event PaymentEvent
	err := r.db.QueryRow(ctx, `
		INSERT INTO payment_events (
			payment_intent_id, state, external_state_id
		)
		VALUES ($1, $2, $3)
		RETURNING id::text, payment_intent_id, state, external_state_id, metadata, created_at
	`,
		intentId,
		status,
		providerStateId,
	).Scan(
		&event.Id,
		&event.PaymentIntentId,
		&event.State,
		&event.ExternalStateId,
		&event.Metadata,
		&event.CreatedAt,
	)
	if err != nil {
		return PaymentEvent{}, err
	}

	return event, nil
}

func (r *PostgresRepo) GetByIntentAndStatus(ctx context.Context, intentId, status string) (PaymentEvent, error) {
	var p PaymentEvent
	err := r.db.QueryRow(ctx, `
        SELECT id, payment_intent_id, state, external_state_id, created_at
        FROM payment_events
        WHERE payment_intent_id = $1
          AND state = $2
    `, intentId, status).Scan(
		&p.Id,
		&p.PaymentIntentId,
		&p.State,
		&p.ExternalStateId,
		&p.CreatedAt,
	)

	if err != nil {
		return PaymentEvent{}, err
	}

	return p, nil
}
