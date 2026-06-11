package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/bank"
	"ledger/pkg/utils"
	"log"
	"strconv"
)

func (s *Service) Create(
	ctx context.Context, idempotencyKey, requestHash string, input CreateInput,
) (PaymentIntent, bool, error) {
	reservedPaymentIntent, err := s.getReserved(ctx, idempotencyKey, requestHash)
	if err != nil && !errors.Is(err, ErrReserveNotFound) {
		return PaymentIntent{}, false, err
	}

	if reservedPaymentIntent != nil {
		return *reservedPaymentIntent, true, nil
	}

	var paymentIntent PaymentIntent
	err = s.uow.RunInTx(ctx, func(r Repos) error {
		paymentIntent, err = s.repo.Create(
			ctx, input.Amount.Figure, input.Amount.Currency, input.OrderId, input.CustomerId,
		)
		if err != nil {
			log.Printf("Error creating payment intent: %s", err)
			return err
		}

		id, err := strconv.ParseInt(paymentIntent.Id, 10, 64)
		if err != nil {
			log.Printf("Error parsing payment intent ID: %s", err)
			return err
		}

		paymentIntent.PaymentRef = utils.GeneratePaymentRef(id)

		err = s.repo.CreateReference(
			ctx, paymentIntent.Id, paymentIntent.PaymentRef,
		)
		if err != nil {
			log.Printf("Error creating payment intent: %s", err)
			return err
		}

		return nil
	})
	if err != nil {
		return PaymentIntent{}, false, err
	}

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentIntent.PaymentRef)
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
		log.Printf("Bank authorization failed for payment reference %s: %s", paymentIntent.PaymentRef, err)
		return PaymentIntent{}, false, ErrBankDeclined
	}

	err = s.uow.RunInTx(ctx, func(r Repos) error {
		err := r.PaymentIntent.UpdateState(ctx, paymentIntent.PaymentRef, PaymentStatusAuthorized)
		if err != nil {
			log.Printf("Error updating payment intent state: %s", err)
			return err
		}

		process, err := r.PaymentProcess.Create(ctx, paymentIntent.Id, PaymentStatusAuthorized, payment.AuthorizationId)
		if err != nil {
			log.Printf("Error creating payment process state: %s", err)
			return err
		}

		paymentIntent.CurrentPaymentProcess = &process

		return nil
	})

	return paymentIntent, false, nil
}
