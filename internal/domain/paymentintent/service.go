package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/domain/bank"
	"ledger/internal/domain/idempotency"
	"ledger/internal/domain/paymentevent"
	"ledger/internal/jobs/queue"
	"ledger/internal/platform/notifier"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo      Repository
	eventRepo paymentevent.Repository
	bankRepo  bank.Repository

	uow UnitOfWork

	idempotencyService *idempotency.Service

	queue    queue.QueueInterface
	notifier notifier.Notifier
}

func NewService(
	repo Repository,
	paymentEventRepo paymentevent.Repository,
	bankRepo bank.Repository,
	uow UnitOfWork,
	idempotencyService *idempotency.Service,
	queueClient queue.QueueInterface,
	notifier notifier.Notifier,
) *Service {
	return &Service{
		repo,
		paymentEventRepo,
		bankRepo,
		uow,
		idempotencyService,
		queueClient,
		notifier,
	}
}

func (s *Service) GetByRef(ctx context.Context, ref string) (*PaymentIntent, error) {
	pi, err := s.repo.GetByRef(ctx, ref)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Println("failed to get payment intent for payment ref: ", ref, " error: ", err)
		return nil, ErrInternal
	}

	return &pi, nil
}

func (s *Service) getReservedIntent(ctx context.Context, idempotencyKey, requestHash string) (*PaymentIntent, error) {
	entry, err := s.idempotencyService.GetKeyReserve(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Println("Redis internal error:", err)
		return nil, ErrInternal
	}

	if entry == nil {
		return nil, nil
	}

	if entry.RequestHash != requestHash {
		log.Printf("Idempotency key reuse with different request hash. Key: %s", idempotencyKey)
		return nil, ErrIdempotencyKeyReuse
	}

	pi, err := s.repo.GetByRef(ctx, entry.PaymentRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("missing payment intent for known payment_ref: %s", entry.PaymentRef)
			return nil, ErrInconsistentState
		}

		log.Println("failed to get payment intent for payment ref: ", entry.PaymentRef, " error: ", err)
		return nil, ErrInternal
	}

	return &pi, nil
}

func (s *Service) startOperation(ctx context.Context, paymentRef, operation string) error {
	return s.repo.UpdateOperation(ctx, paymentRef, operation)
}
