package paymentintent

import (
	"errors"
	"testing"
)

func TestBankDeclinedErrorPreservesReason(t *testing.T) {
	reason := "bank declined: code=invalid_card message=insufficient funds"

	err := NewBankDeclinedError(reason)
	if !errors.Is(err, ErrBankDeclined) {
		t.Fatalf("expected error to match ErrBankDeclined")
	}

	if err.Error() != reason {
		t.Fatalf("expected error message %q, got %q", reason, err.Error())
	}
}
