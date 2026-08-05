package bank

import (
	"errors"
	"fmt"
)

// Sentinels used for asynq retry classification.
var (
	ErrInternal      = errors.New("internal error")
	ErrBankTransient = errors.New("bank transient error")
	ErrBankTerminal  = errors.New("bank terminal error")
)

// APIError carries FicMart's structured 4xx error body.
type APIError struct {
	StatusCode int
	Code       string // e.g. "invalid_card"
	Message    string // e.g. "Available balance is less than requested amount"
}

func (e *APIError) Error() string {
	return fmt.Sprintf("bank declined: code=%s message=%s", e.Code, e.Message)
}

// Unwrap lets errors.Is(err, ErrBankTerminal) succeed for any *APIError,
// so callers can classify without knowing the concrete type.
func (e *APIError) Unwrap() error {
	return ErrBankTerminal
}

func IsTerminal(err error) bool {
	return errors.Is(err, ErrInternal) || errors.Is(err, ErrBankTerminal)
}
