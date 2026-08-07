package paymentintent

import (
	"context"
	"errors"
	"ledger/internal/jobs/tasks"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Service) Capture(ctx context.Context,
	idempotencyKey, requestHash, paymentRef string, amount int,
) (PaymentIntent, bool, error) {
	reservedPI, err := s.getReservedIntent(ctx, idempotencyKey, requestHash)
	if err != nil {
		return PaymentIntent{}, false, err
	}

	if reservedPI != nil {
		return *reservedPI, true, nil
	}

	pi, err := s.repo.GetWithEvent(ctx, paymentRef, PaymentStatusAuthorized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("operation not allowed: intent=%s, required_status=%s", paymentRef, PaymentStatusAuthorized)
			return PaymentIntent{}, false, ErrOperationNotAllowed
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}

	err = s.startOperation(ctx, pi.PaymentRef, PaymentCaptureOp)
	if err != nil {
		log.Printf("Error starting capture process for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}
	pi.CurrentOp = PaymentCaptureOp

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentRef)
	if err != nil {
		log.Printf("Error reserving idempotency key: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	// subscribing before enqueuing to avoid missing the signal
	resultCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sub := s.notifier.Subscribe(pi.PaymentRef)
	defer sub.Close()

	err = s.queue.EnqueueCapture(ctx, tasks.CapturePayload{
		IdempotencyKey: idempotencyKey,
		IntentId:       pi.Id,
		PaymentRef:     pi.PaymentRef,
		StateId:        pi.CurrentEvent.ExternalStateId,
		Amount:         pi.Amount,
	})
	if err != nil {
		log.Printf("Error enqueuing capture payment task: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	result, err := sub.Wait(resultCtx)
	if err != nil || result == "" {
		return pi, false, nil
	}

	if result != PaymentStatusCaptured {
		log.Printf("Payment failed for %v: %s", pi.PaymentRef, result)
		return PaymentIntent{}, false, NewBankDeclinedError(result)
	}

	pi.Status = result
	pi.CurrentOp = ""

	return pi, false, nil
}
