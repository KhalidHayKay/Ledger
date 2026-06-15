package paymentintent

import "errors"

var (
	ErrNotFound            = errors.New("payment intent not found")
	ErrReserveNotFound     = errors.New("no payment intent reserve")
	ErrIdempotencyKeyReuse = errors.New("idempotency key reused")
	ErrBankDeclined        = errors.New("bank declined payment")
	ErrInconsistentState   = errors.New("payment intent state inconsistent")
	ErrInternal            = errors.New("internal error")
	ErrInvalidStatus       = errors.New("invalid payment intent status for this operation")
)
