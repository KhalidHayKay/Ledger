package paymentintent

import "errors"

var (
	ErrPaymentIntentNotFound        = errors.New("payment intent not found")
	ErrPaymentIntentReserveNotFound = errors.New("no payment intent reserve")
	ErrIdempotencyKeyReuse          = errors.New("idempotency key reused")
	ErrBankDeclined                 = errors.New("bank declined payment")
	ErrInconsistentState            = errors.New("payment intent state inconsistent")
	ErrInternal                     = errors.New("internal error")
)
