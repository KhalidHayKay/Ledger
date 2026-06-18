package paymentintent

import (
	"context"
	"ledger/internal/bank"
	"ledger/pkg/utils"
	"log"
	"strconv"
)

func (s *Service) Create(
	ctx context.Context, idempotencyKey, requestHash string, input CreateInput,
) (PaymentIntent, bool, error) {
	reservedPI, err := s.getReservedIntent(ctx, idempotencyKey, requestHash)
	if err != nil {
		return PaymentIntent{}, false, err
	}

	if reservedPI != nil {
		return *reservedPI, true, nil
	}

	var pi PaymentIntent
	err = s.uow.RunInTx(ctx, func(r Repos) error {
		pi, err = r.PaymentIntent.Create(
			ctx, input.Amount.Figure, input.Amount.Currency, input.OrderId, input.CustomerId,
		)
		if err != nil {
			log.Printf("Error creating payment intent: %s", err)
			return err
		}

		id, err := strconv.ParseInt(pi.Id, 10, 64)
		if err != nil {
			log.Printf("Error parsing payment intent ID: %s", err)
			return err
		}

		pi.PaymentRef = utils.GeneratePaymentRef(id)

		err = r.PaymentIntent.CreateReference(
			ctx, pi.Id, pi.PaymentRef,
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

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, pi.PaymentRef)
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
		log.Printf("Bank authorization failed for payment reference %s: %s", pi.PaymentRef, err)
		return PaymentIntent{}, false, ErrBankDeclined
	}

	err = s.uow.RunInTx(ctx, func(r Repos) error {
		event, err := r.PaymentEvent.Create(ctx, pi.Id, PaymentStatusAuthorized, payment.AuthorizationId)
		if err != nil {
			log.Printf("Error creating payment event state: %s", err)
			return err
		}

		err = r.PaymentIntent.UpdateState(ctx, pi.PaymentRef, event.State)
		if err != nil {
			log.Printf("Error updating payment intent state: %s", err)
			return err
		}

		pi.Status = event.State
		pi.CurrentEvent = nil

		return nil
	})

	return pi, false, nil
}
