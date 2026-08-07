package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/jobs/tasks"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Service) Refund(ctx context.Context,
	idempotencyKey, requestHash, paymentRef string,
	amount int,
) (PaymentIntent, bool, error) {
	reservedPI, err := s.getReservedIntent(ctx, idempotencyKey, requestHash)
	if err != nil {
		return PaymentIntent{}, false, err
	}

	if reservedPI != nil {
		return *reservedPI, true, nil
	}

	pi, err := s.repo.GetWithEvent(ctx, paymentRef, PaymentStatusCaptured)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("operation not allowed: intent=%s, required_status=%s", paymentRef, PaymentStatusCaptured)
			return PaymentIntent{}, false, ErrOperationNotAllowed
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}

	err = s.startOperation(ctx, pi.PaymentRef, PaymentRefundOp)
	if err != nil {
		log.Printf("Error starting refund process for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}
	pi.CurrentOp = PaymentRefundOp

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentRef)
	if err != nil {
		log.Printf("Error reserving idempotency key: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	resultCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sub := s.notifier.Subscribe(pi.PaymentRef)
	defer sub.Close()

	err = s.queue.EnqueueRefund(ctx, tasks.RefundPayload{
		IdempotencyKey: idempotencyKey,
		IntentId:       pi.Id,
		PaymentRef:     pi.PaymentRef,
		StateId:        pi.CurrentEvent.ExternalStateId,
		Amount:         amount,
	})
	if err != nil {
		log.Printf("Error enqueuing refund payment task: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	result, err := sub.Wait(resultCtx)
	if err != nil || result == "" {
		return pi, false, nil
	}

	if result != PaymentStatusRefunded {
		log.Printf("Payment failed for %v: %s", pi.PaymentRef, result)
		return PaymentIntent{}, false, NewBankDeclinedError(result)
	}

	pi.Status = result
	pi.CurrentOp = ""

	return pi, false, nil
}
