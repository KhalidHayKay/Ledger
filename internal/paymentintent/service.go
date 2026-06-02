package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo     Repository
	bankRepo bank.Repository

	idempotencyService *idempotency.Service
}

func NewService(
	repo Repository,
	bankRepo bank.Repository,
	idempotencyService *idempotency.Service,
) *Service {
	return &Service{repo, bankRepo, idempotencyService}
}

func (s *Service) Create(
	ctx context.Context, idempotencyKey, requestHash string, input CreateInput,
) (PaymentIntent, bool, error) {
	paymentIntent, err := s.getReserved(ctx, idempotencyKey, requestHash)
	if err != nil && !errors.Is(err, ErrPaymentIntentReserveNotFound) {
		return PaymentIntent{}, false, err
	}

	if paymentIntent != nil {
		return *paymentIntent, true, nil
	}

	paymentRef, err := s.repo.Create(
		ctx, input.Amount.Figure, input.Amount.Currency, input.OrderId, input.CustomerId,
	)
	if err != nil {
		log.Printf("Error creating payment intent: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentRef)
	if err != nil {
		return PaymentIntent{}, false, ErrInternal
	}

	payment, err := s.bankRepo.Authorize(
		ctx,
		idempotencyKey,
		bank.AuthorizeInput{
			Card: bank.Card{
				Number: input.Card.Number,
				CVV:    input.Card.CVV,
				Expiry: bank.CardExpiry{
					Month: input.Card.Expiry.Month,
					Year:  input.Card.Expiry.Year,
				},
			},
			Amount: bank.Amount{
				Figure:   input.Amount.Figure,
				Currency: input.Amount.Currency,
			},
		},
	)
	if err != nil {
		log.Printf("Bank authorization failed for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrBankDeclined
	}

	updatedPaymentIntent, err := s.repo.UpdateBankAuth(ctx, paymentRef, payment.Id)
	if err != nil {
		log.Printf("Error updating bank authorization for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}

	return updatedPaymentIntent, false, nil
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

	return nil, ErrPaymentIntentReserveNotFound
}
