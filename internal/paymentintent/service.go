package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"ledger/internal/paymentprocess"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type Repos struct {
	PaymentIntent  Repository
	PaymentProcess paymentprocess.Repository
}

type UoW interface {
	RunInTx(ctx context.Context, fn func(Repos) error) error
}

type Service struct {
	repo     Repository
	bankRepo bank.Repository

	uow UoW

	idempotencyService *idempotency.Service
}

func NewService(
	repo Repository,
	bankRepo bank.Repository,
	uow UoW,
	idempotencyService *idempotency.Service,
) *Service {
	return &Service{repo, bankRepo, uow, idempotencyService}
}

func (s *Service) Cancel() {
	//
}

func (s *Service) Refund() {
	//
}

func (s *Service) getReserved(ctx context.Context, idempotencyKey, requestHash string) (*PaymentIntent, error) {
	entry, err := s.idempotencyService.GetKeyReserve(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Println("Redis internal error:", err)
		return nil, ErrInternal
	}

	if entry != nil {
		if entry.RequestHash != requestHash {
			log.Printf("Idempotency key reuse with different request hash. Key: %s", idempotencyKey)
			return nil, ErrIdempotencyKeyReuse
		}

		paymentIntent, err := s.repo.GetByPaymentRef(ctx, entry.PaymentRef)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Printf("missing payment intent for known payment_ref: %s", entry.PaymentRef)
				return nil, ErrInconsistentState
			}

			log.Println("failed to get payment intent for payment ref: ", entry.PaymentRef, " error: ", err)
			return nil, ErrInternal
		}

		return &paymentIntent, nil
	}

	return nil, ErrReserveNotFound
}
