package paymentintent

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
)

func (s *Service) Capture(ctx context.Context,
	idempotencyKey, requestHash, paymentRef, amount string,
) (PaymentIntent, bool, error) {
	reservedPaymentIntent, err := s.getReserved(ctx, idempotencyKey, requestHash)
	if err != nil && !errors.Is(err, ErrReserveNotFound) {
		return PaymentIntent{}, false, err
	}

	if reservedPaymentIntent != nil {
		return *reservedPaymentIntent, true, nil
	}

	paymentIntent, err := s.repo.GetWithEvent(ctx, paymentRef, PaymentStatusAuthorized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("operation not allowed: intent=%s, required_status=%s", paymentRef, PaymentStatusAuthorized)
			return PaymentIntent{}, false, ErrOperationNotAllowed
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}

	payment, err := s.bankRepo.Capture(ctx, idempotencyKey, paymentIntent.CurrentEvent.ExternalId, amount)
	if err != nil {
		log.Printf("Bank capture failed for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrBankDeclined
	}

	err = s.uow.RunInTx(ctx, func(r Repos) error {
		err = r.PaymentIntent.UpdateState(ctx, paymentRef, PaymentStatusCaptured)
		if err != nil {
			log.Printf("Error updating bank capture for payment reference %s: %s", paymentRef, err)
			return ErrInternal
		}

		process, err := r.PaymentEvent.Create(ctx, paymentIntent.Id, PaymentStatusCaptured, payment.CaptureId)
		if err != nil {
			log.Printf("Error creating payment process: %s", err)
			return ErrInternal
		}

		paymentIntent.Amount = payment.Amount
		paymentIntent.CurrentEvent = &process

		return nil
	})

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentRef)
	if err != nil {
		log.Printf("Error reserving idempotency key: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	return paymentIntent, false, nil
}
