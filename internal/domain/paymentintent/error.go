package paymentintent

import "errors"

var (
	ErrNotFound = errors.New("payment intent not found")

	ErrIdempotencyKeyReuse = errors.New("idempotency key reused")

	ErrBankDeclined = errors.New("bank declined payment")

	ErrInternal = errors.New("internal error")

	ErrOperationNotAllowed = errors.New("operation not allowed on payment intent: intent not found or not in a valid state")

	ErrInconsistentState = errors.New("payment intent inconsistent with idempotency state")
)
