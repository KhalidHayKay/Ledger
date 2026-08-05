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

	// subscribing before enqueuing to avoid missing the signal
	resultCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sub := s.notifier.Subscribe(pi.PaymentRef)
	defer sub.Close()

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
		log.Printf("Error enqueuing create payment task: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	result, err := sub.Wait(resultCtx)
	if err != nil || result == "" {
		return pi, false, nil
	}

	if result != PaymentStatusAuthorized {
		log.Printf("Payment failed for %v: %s", pi.PaymentRef, result)
		return PaymentIntent{}, false, NewBankDeclinedError(result)
	}

	pi.Status = result
	return pi, false, nil
}

func (s *Service) createPendingIntent(ctx context.Context, input CreateInput) (PaymentIntent, error) {
	var pi PaymentIntent
	err := s.uow.RunInTx(ctx, func(r TxRepos) error {
		var err error
		pi, err = r.PaymentIntent.Create(
			ctx, input.Amount.Figure, input.Amount.Currency, input.OrderId, input.CustomerId,
		)
		if err != nil {
			return err
		}

		id, err := strconv.ParseInt(pi.Id, 10, 64)
		if err != nil {
			return err
		}

		ref := uuid.GeneratePaymentRef(id)

		err = r.PaymentIntent.CreateReference(
			ctx, pi.Id, ref,
		)
		if err != nil {
			return err
		}

		err = r.PaymentIntent.UpdateOperation(ctx, ref, PaymentAuthorizationOp)

		pi.PaymentRef = ref
		pi.CurrentOp = PaymentAuthorizationOp

		return nil
	})
	if err != nil {
		log.Printf("Error creating payment intent: %s", err)
		return PaymentIntent{}, err
	}

	return pi, err
}
