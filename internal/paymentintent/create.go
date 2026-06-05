package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/bank"
	"log"
)

func (s *Service) Create(
	ctx context.Context, idempotencyKey, requestHash string, input CreateInput,
) (PaymentIntent, bool, error) {
	paymentIntent, err := s.getReserved(ctx, idempotencyKey, requestHash)
	if err != nil && !errors.Is(err, ErrReserveNotFound) {
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

	updatedPaymentIntent, err := s.repo.UpdateBankAuth(ctx, paymentRef, payment.AuthorizationId)
	if err != nil {
		log.Printf("Error updating bank authorization for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}

	return updatedPaymentIntent, false, nil
}
