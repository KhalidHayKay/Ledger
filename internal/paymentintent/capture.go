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

	paymentIntent, err := s.repo.GetByPaymentRef(ctx, paymentRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Payment intent not found for payment reference: %s", paymentRef)
			return PaymentIntent{}, false, ErrNotFound
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}

	paymentProcess, err := s.processRepo.GetByIntentAndStatus(ctx, paymentIntent.Id, paymentIntent.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("No payment process found with payment reference id %s and status %s",
				paymentIntent.Id, paymentIntent.Status)
			return PaymentIntent{}, false, ErrNotFound
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return PaymentIntent{}, false, ErrInternal
	}

	payment, err := s.bankRepo.Capture(ctx, idempotencyKey, paymentProcess.ExternalId, amount)
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

		process, err := r.PaymentProcess.Create(ctx, paymentIntent.Id, PaymentStatusCaptured, payment.CaptureId)
		if err != nil {
			log.Printf("Error creating payment process: %s", err)
			return ErrInternal
		}

		paymentIntent.Amount = payment.Amount
		paymentIntent.CurrentPaymentProcess = &process

		return nil
	})

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, paymentRef)
	if err != nil {
		log.Printf("Error reserving idempotency key: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	return paymentIntent, false, nil
}
