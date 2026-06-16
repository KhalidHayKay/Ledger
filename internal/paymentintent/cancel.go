package paymentintent

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
)

func (s *Service) Cancel(ctx context.Context, idempotencyKey, paymentRef string) (string, error) {
	paymentIntent, err := s.repo.GetWithEvent(ctx, paymentRef, PaymentStatusAuthorized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("operation not allowed: intent=%s, required_status=%s", paymentRef, PaymentStatusAuthorized)
			return "", ErrOperationNotAllowed
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return "", ErrInternal
	}

	payment, err := s.bankRepo.Void(ctx, idempotencyKey, paymentIntent.CurrentEvent.ExternalId)
	if err != nil {
		log.Printf("Bank failed to void payment: %s", err)
		return "", ErrBankDeclined
	}

	return payment.VoidId, nil
}
