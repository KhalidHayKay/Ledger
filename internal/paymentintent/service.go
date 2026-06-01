package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"log"
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

func (s *Service) Create(ctx context.Context, idempotencyKey, requestHash string, input CreateInput) (PaymentIntent, bool, error) {
	entry, err := s.idempotencyService.GetKeyReserve(ctx, idempotencyKey)
	if err != nil && err.Error() != "redis: nil" {
		log.Println("Redis error:", err)
		return PaymentIntent{}, false, err
	}

	if entry != nil {
		if entry.RequestHash != requestHash {
			return PaymentIntent{}, false, errors.New("KEY_REUSED_ERROR")
		}

		paymentIntent, err := s.repo.GetByPaymentRef(ctx, entry.PaymentRef)
		if err != nil {
			log.Println("failed to get payment intent by payment ref: ", entry.PaymentRef, " error: ", err)
			return PaymentIntent{}, false, err
		}

		return paymentIntent, true, nil
	}

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, "")
	if err != nil {
		return PaymentIntent{}, false, err
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
		return PaymentIntent{}, false, err
	}

	paymentIntent, err := s.repo.Create(
		ctx, payment.Id, input.Amount.Figure, input.Amount.Currency, input.OrderId, input.CustomerId,
	)
	if err != nil {
		return PaymentIntent{}, false, err
	}

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentIntent.PaymentReference)
	if err != nil {
		return PaymentIntent{}, false, err
	}

	return paymentIntent, false, nil
}
