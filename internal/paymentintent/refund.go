package paymentintent

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
)

func (s *Service) Refund(ctx context.Context, idempotencyKey, paymentRef string, amount int) (string, error) {
	paymentIntent, err := s.repo.GetWithEvent(ctx, paymentRef, PaymentStatusCaptured)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("operation not allowed: intent=%s, required_status=%s", paymentRef, PaymentStatusAuthorized)
			return "", ErrOperationNotAllowed
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return "", ErrInternal
	}

	payment, err := s.bankRepo.Refund(ctx, idempotencyKey, paymentIntent.CurrentEvent.ExternalStateId, amount)
	if err != nil {
		log.Printf("Bank refund failed: %s", err)
	}

	return payment.RefundId, nil
}
