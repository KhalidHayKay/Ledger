package paymentintent

// func (s *Service) Capture(ctx context.Context, idempotencyKey, paymentRef string) (PaymentIntent, error) {
// 	paymentIntent, err := s.repo.GetByPaymentRef(ctx, paymentRef)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			log.Printf("Payment intent not found for payment reference: %s", paymentRef)
// 			return PaymentIntent{}, ErrNotFound
// 		}

// 		log.Printf("Error retrieving payment intent for payment reference %s: %s", paymentRef, err)
// 		return PaymentIntent{}, ErrInternal
// 	}

// 	capture, err := s.bankRepo.Capture(ctx, idempotencyKey, paymentIntent.BankAuthorizationId)
// 	if err != nil {
// 		log.Printf("Bank capture failed for payment reference %s: %s", paymentRef, err)
// 		return PaymentIntent{}, ErrBankDeclined
// 	}

// 	updatedPaymentIntent, err := s.repo.UpdateBankCapture(ctx, paymentRef, capture.CaptureId)
// 	if err != nil {
// 		log.Printf("Error updating bank capture for payment reference %s: %s", paymentRef, err)
// 		return PaymentIntent{}, ErrInternal
// 	}

// 	return updatedPaymentIntent, nil
// }
