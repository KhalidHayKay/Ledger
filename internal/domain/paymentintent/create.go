package paymentintent

import (
	"context"
	"ledger/internal/domain/bank"
	"ledger/internal/jobs/tasks"
	"ledger/pkg/uuid"
	"log"
	"strconv"
	"time"
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

	pi, err := s.createPendingIntent(ctx, input)
	if err != nil {
		return PaymentIntent{}, false, err
	}

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, pi.PaymentRef)
	if err != nil {
		return PaymentIntent{}, false, ErrInternal
	}

	// subscribe before enqueuing - avoid missing the signal
	resultCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sub := s.notifier.Subscribe(pi.PaymentRef)
	defer sub.Close()

	// Error not handled yet
	err = s.queue.EnqueueCreate(ctx, tasks.CreatePayload{
		IdempotencyKey: idempotencyKey,
		IntentId:       pi.Id,
		PaymentRef:     pi.PaymentRef,
		BankInput: &bank.AuthorizeInput{
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
	})
	if err != nil {
		return PaymentIntent{}, false, ErrInternal
	}

	result, err := sub.Wait(resultCtx)
	if err != nil {
		return pi, false, nil
	}

	return result, false, nil
}

func (s *Service) createPendingIntent(ctx context.Context, input CreateInput) (PaymentIntent, error) {
	var pi PaymentIntent
	err := s.uow.RunInTx(ctx, func(r TxRepos) error {
		pi, err := r.PaymentIntent.Create(
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

		pi.PaymentRef = uuid.GeneratePaymentRef(id)

		err = r.PaymentIntent.CreateReference(
			ctx, pi.Id, pi.PaymentRef,
		)
		if err != nil {
			log.Printf("Error creating payment intent: %s", err)
			return err
		}

		return nil
	})

	return pi, err
}
