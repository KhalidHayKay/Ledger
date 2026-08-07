package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

const (
	saveKeyCall  = "SaveKey"
	getByKeyCall = "GetByKey"
)

type repoMock struct {
	SaveKeyFn   func(ctx context.Context, idempotencyKey string, entry string) error
	GetByKeyFn  func(ctx context.Context, idempotencyKey string) (string, error)
	RemoveKeyFn func(ctx context.Context, idempotencyKey string) error

	calls []string
}

func (r *repoMock) SaveKey(ctx context.Context, idempotencyKey string, entry string) error {
	r.calls = append(r.calls, saveKeyCall)
	return r.SaveKeyFn(ctx, idempotencyKey, entry)
}

func (r *repoMock) GetByKey(ctx context.Context, idempotencyKey string) (string, error) {
	r.calls = append(r.calls, getByKeyCall)
	return r.GetByKeyFn(ctx, idempotencyKey)
}

func (r *repoMock) RemoveKey(ctx context.Context, idempotencyKey string) error {
	r.calls = append(r.calls, "RemoveKey")
	return r.RemoveKeyFn(ctx, idempotencyKey)
}

func TestReserveKeyHappyPath(t *testing.T) {
	key := "idmkey"
	formattedKey := fmt.Sprintf("idempotency-key:%v", key)

	repo := &repoMock{
		SaveKeyFn: func(ctx context.Context, idempotencyKey, entry string) error {
			if formattedKey != idempotencyKey {
				t.Fatalf("expected key %q, got %q", formattedKey, idempotencyKey)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.ReserveKey(context.Background(), key, "request-hash", "payment-ref")
	if err != nil {
		t.Errorf("Expected no error on happy path, but got: %v", err)
	}

	if len(repo.calls) != 1 {
		t.Errorf("Expected SaveKey to be called once, but was called %d times", len(repo.calls))
	}
}

func TestGetByKeyHappyPath(t *testing.T) {
	key := "idmkey"
	formattedKey := fmt.Sprintf("idempotency-key:%v", key)
	mockEntry := Entry{
		RequestHash: "request-hash",
		PaymentRef:  "payment-ref",
	}

	repo := &repoMock{
		GetByKeyFn: func(ctx context.Context, idempotencyKey string) (string, error) {
			if formattedKey != idempotencyKey {
				t.Fatalf("expected key %q, got %q", formattedKey, idempotencyKey)
			}

			entryEncode, _ := json.Marshal(mockEntry)
			return string(entryEncode), nil
		},
	}

	service := NewService(repo)

	entry, err := service.GetKeyReserve(context.Background(), key)
	if err != nil {
		t.Errorf("Expected no error on happy path, but got: %v", err)
	}

	if len(repo.calls) != 1 {
		t.Errorf("Expected GetByKey to be called once, but was called %d times", len(repo.calls))
	}

	if *entry != mockEntry {
		t.Errorf("Expected result of GetKeyReserve to be %v. Got %v", mockEntry, &entry)
	}
}
