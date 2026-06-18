package paymentintent

import (
	"context"
	"errors"
	"log"

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

	payment, err := s.bankRepo.Capture(ctx, idempotencyKey, pi.CurrentEvent.ExternalStateId, amount)
	if err != nil {
		log.Printf("Bank capture failed for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrBankDeclined
	}

	err = s.uow.RunInTx(ctx, func(r Repos) error {
		event, err := r.PaymentEvent.Create(ctx, pi.Id, PaymentStatusCaptured, payment.CaptureId)
		if err != nil {
			log.Printf("Error creating payment event: %s", err)
			return ErrInternal
		}

		err = r.PaymentIntent.UpdateState(ctx, paymentRef, event.State)
		if err != nil {
			log.Printf("Error updating bank capture for payment reference %s: %s", paymentRef, err)
			return ErrInternal
		}

		pi.Status = event.State
		pi.CurrentEvent = nil

		return nil
	})

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentRef)
	if err != nil {
		log.Printf("Error reserving idempotency key: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	return pi, false, nil
}
