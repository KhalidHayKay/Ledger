package paymentintent

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
)

func (s *Service) Cancel(ctx context.Context,
	idempotencyKey, requestHash, paymentRef string,
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

	payment, err := s.bankRepo.Void(ctx, idempotencyKey, pi.CurrentEvent.ExternalStateId)
	if err != nil {
		log.Printf("Bank failed to void payment: %s", err)
		return PaymentIntent{}, false, ErrBankDeclined
	}

	err = s.uow.RunInTx(ctx, func(r TxRepos) error {
		event, err := r.PaymentEvent.Create(ctx, pi.Id, PaymentStatusCanceled, payment.VoidId, "")
		if err != nil {
			return err
		}

		err = r.PaymentIntent.UpdateState(ctx, pi.PaymentRef, event.State)
		if err != nil {
			return err
		}

		pi.Status = event.State
		pi.CurrentEvent = nil

		return nil
	})
	if err != nil {
		log.Printf("Error creating payment event for state to %v: %s", PaymentStatusCanceled, err)
		return PaymentIntent{}, false, ErrInternal
	}

	err = s.idempotencyService.ReserveKey(ctx, idempotencyKey, requestHash, pi.PaymentRef)
	if err != nil {
		log.Printf("Error reserving refunded payment intent: %s", err)
		return PaymentIntent{}, false, ErrInternal
	}

	return pi, false, nil
}
