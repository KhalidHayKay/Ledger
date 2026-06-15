package paymentintent

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
)

func (s *Service) Refund(ctx context.Context, idempotencyKey, amount, paymentRef string) (string, error) {
	paymentIntent, err := s.repo.GetByPaymentRef(ctx, paymentRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Payment intent not found for payment reference: %s", paymentRef)
			return "", ErrNotFound
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return "", ErrInternal
	}

	switch paymentIntent.Status {
	case PaymentStatusAuthorized:
		log.Printf("Cannot refund. Payment intent with payment reference: %s is has not been captured", paymentRef)
		return "", ErrInvalidStatus
	case PaymentStatusVoided:
		log.Printf("Payment intent already voided for payment reference: %s", paymentRef)
		return "", ErrInvalidStatus
	case PaymentStatusRefunded:
		log.Printf("Payment intent already refunded for payment reference: %s", paymentRef)
		return "", ErrInvalidStatus
	case PaymentStatusFailed:
		log.Printf("Payment intent already failed for payment reference: %s", paymentRef)
		return "", ErrInvalidStatus
	}

	process, err := s.processRepo.GetByIntentAndStatus(ctx, paymentIntent.Id, PaymentStatusCaptured)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("No payment process found with payment reference id %s and status %s",
				paymentIntent.Id, paymentIntent.Status)
			return "", ErrNotFound
		}

		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
		return "", ErrInternal
	}

	payment, err := s.bankRepo.Refund(ctx, idempotencyKey, process.ExternalId, amount)
	if err != nil {
		log.Printf("Bank refund failed: %s", err)
	}

	return payment.RefundId, nil
}
