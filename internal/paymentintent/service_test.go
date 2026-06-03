package paymentintent

import (
	"context"
	"encoding/json"
	"fmt"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"testing"

	"github.com/redis/go-redis/v9"
)

const (
	createCall          = "Create"
	updateBankAuthCall  = "UpdateBankAuth"
	getByPaymentRefCall = "GetByPaymentRef"

	authorizeCall = "Authorize"

	saveKeyCall  = "SaveKey"
	getByKeyCall = "GetByKey"
)

type repoMock struct {
	CreateFn          func(ctx context.Context, amount int, currency, orderId, customerId string) (string, error)
	UpdateBankAuthFn  func(ctx context.Context, paymentRef, bankAuthId string) (PaymentIntent, error)
	GetByPaymentRefFn func(ctx context.Context, paymentReference string) (PaymentIntent, error)

	calls []string
}

func (r *repoMock) Create(ctx context.Context, amount int, currency, orderId, customerId string) (string, error) {
	r.calls = append(r.calls, createCall)
	return r.CreateFn(ctx, amount, currency, orderId, customerId)
}

func (r *repoMock) UpdateBankAuth(ctx context.Context, paymentRef, bankAuthId string) (PaymentIntent, error) {
	r.calls = append(r.calls, updateBankAuthCall)
	return r.UpdateBankAuthFn(ctx, paymentRef, bankAuthId)
}

func (r *repoMock) GetByPaymentRef(ctx context.Context, paymentReference string) (PaymentIntent, error) {
	r.calls = append(r.calls, getByPaymentRefCall)
	return r.GetByPaymentRefFn(ctx, paymentReference)
}

// Bank Repo Mocks
type bankRepoMock struct {
	AuthorizeFn func(ctx context.Context, idempotencyKey string, input bank.AuthorizeInput) (bank.Payment, error)

	calls []string
}

func (r *bankRepoMock) Authorize(ctx context.Context, idempotencyKey string, input bank.AuthorizeInput) (bank.Payment, error) {
	r.calls = append(r.calls, authorizeCall)
	return r.AuthorizeFn(ctx, idempotencyKey, input)
}

// Idempotency Repo Mocks
type idempotencyRepoMock struct {
	SaveKeyFn  func(ctx context.Context, idempotencyKey string, entry string) error
	GetByKeyFn func(ctx context.Context, idempotencyKey string) (string, error)

	calls []string
}

func (r *idempotencyRepoMock) SaveKey(ctx context.Context, idempotencyKey string, entry string) error {
	r.calls = append(r.calls, saveKeyCall)
	return r.SaveKeyFn(ctx, idempotencyKey, entry)
}

func (r *idempotencyRepoMock) GetByKey(ctx context.Context, idempotencyKey string) (string, error) {
	r.calls = append(r.calls, getByKeyCall)
	return r.GetByKeyFn(ctx, idempotencyKey)
}

func TestGetReservedHappyPath(t *testing.T) {
	requestHash := "req-hash"
	paymentRef := "payment-ref"
	mockEntry := idempotency.Entry{
		RequestHash: requestHash,
		PaymentRef:  paymentRef,
	}
	bankAuthId := "payment-id-from-bank"

	idempotencyRepo := &idempotencyRepoMock{
		GetByKeyFn: func(ctx context.Context, idempotencyKey string) (string, error) {
			entryEncode, _ := json.Marshal(mockEntry)
			return string(entryEncode), nil
		},
	}

	repo := &repoMock{
		GetByPaymentRefFn: func(ctx context.Context, paymentReference string) (PaymentIntent, error) {
			return PaymentIntent{
				BankAuthorizationId: bankAuthId,
				PaymentReference:    paymentRef,
			}, nil
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(repo, &bankRepoMock{}, idempotencyService)

	paymentIntent, err := service.getReserved(context.Background(), "idm-key", requestHash)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if paymentIntent.PaymentReference != paymentRef {
		t.Fatalf("Expected payment reference to be %q, got %q", paymentRef, paymentIntent.PaymentReference)
	}
}

func TestGetReservedThrowsErrorOnIdempotencyKeyReuse(t *testing.T) {
	requestHash := "req-hash"
	paymentRef := "payment-ref"
	mockEntry := idempotency.Entry{
		RequestHash: requestHash,
		PaymentRef:  paymentRef,
	}

	idempotencyRepo := &idempotencyRepoMock{
		GetByKeyFn: func(ctx context.Context, idempotencyKey string) (string, error) {
			mockEntry.RequestHash = "changed-request-hash"
			entryEncode, err := json.Marshal(mockEntry)
			if err != nil {
				t.Fatalf("Failed to marshal mock entry: %v", err)
			}
			return string(entryEncode), nil
		},
	}

	repo := &repoMock{
		GetByPaymentRefFn: func(ctx context.Context, paymentReference string) (PaymentIntent, error) {
			t.Errorf("Expected an error boundary to block GetByPaymentRef from getting called for idempotency key reuse")
			return PaymentIntent{}, nil
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(repo, &bankRepoMock{}, idempotencyService)

	_, err := service.getReserved(context.Background(), "idm-key", requestHash)
	if err == nil {
		t.Fatalf("Expected error due to idempotency key reuse with different request hash, got nil")
	}
}

func TestCreateRelaysExistingPaymentIntent(t *testing.T) {
	requestHash := "req-hash"
	mockEntry := idempotency.Entry{
		RequestHash: requestHash,
	}

	idempotencyRepo := &idempotencyRepoMock{
		GetByKeyFn: func(ctx context.Context, idempotencyKey string) (string, error) {
			entryEncode, err := json.Marshal(mockEntry)
			if err != nil {
				t.Fatalf("Failed to marshal mock entry: %v", err)
			}
			return string(entryEncode), nil
		},
	}

	repo := &repoMock{
		GetByPaymentRefFn: func(ctx context.Context, paymentReference string) (PaymentIntent, error) {
			return PaymentIntent{}, nil
		},
		CreateFn: func(ctx context.Context, amount int, currency, orderId, customerId string) (string, error) {
			t.Fatalf("Create should not be called for reserved payment intent")
			return "", nil
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(repo, &bankRepoMock{}, idempotencyService)

	_, relayed, err := service.Create(
		context.Background(),
		"idmkey",
		"req-hash",
		CreateInput{},
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !relayed {
		t.Errorf("Expected Create to return relayed=true for reserved payment intent, got false")
	}
}

func TestNewPaymentIntentIsCreatedForDifferentIdempotencyKeys(t *testing.T) {
	tests := []struct {
		idempotencyKey string
		requestHash    string
		paymentRef     string
		bankAuthId     string
	}{
		{"idmkey-1", "req-hash-1", "payment-ref-1", "bank-authorization-id-1"},
		{"idmkey-2", "req-hash-2", "payment-ref-2", "bank-authorization-id-2"},
		{"idmkey-3", "req-hash-3", "payment-ref-3", "bank-authorization-id-3"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("With Idempotency Key \"%v\"", tt.idempotencyKey), func(t *testing.T) {
			idempotencyRepo := &idempotencyRepoMock{
				GetByKeyFn: func(ctx context.Context, idempotencyKey string) (string, error) {
					return "", redis.Nil
				},
				SaveKeyFn: func(ctx context.Context, idempotencyKey, entry string) error {
					return nil
				},
			}

			bankRepo := &bankRepoMock{
				AuthorizeFn: func(
					ctx context.Context, idempotencyKey string, input bank.AuthorizeInput,
				) (bank.Payment, error) {
					return bank.Payment{Id: tt.bankAuthId}, nil
				},
			}

			repo := &repoMock{
				CreateFn: func(ctx context.Context, amount int, currency, orderId, customerId string) (string, error) {
					return tt.paymentRef, nil
				},
				UpdateBankAuthFn: func(ctx context.Context, paymentRef, bankAuthId string) (PaymentIntent, error) {
					return PaymentIntent{
						PaymentReference:    paymentRef,
						BankAuthorizationId: bankAuthId,
					}, nil
				},
			}

			idempotencyService := idempotency.NewService(idempotencyRepo)

			service := NewService(repo, bankRepo, idempotencyService)

			paymentIntent, relayed, err := service.Create(
				context.Background(),
				tt.idempotencyKey,
				tt.requestHash,
				CreateInput{},
			)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if relayed {
				t.Errorf("Expected Create to return relayed=false for new payment intent, got true")
			}

			if paymentIntent.PaymentReference != tt.paymentRef {
				t.Errorf("Expected payment reference to be %q, got %q", tt.paymentRef, paymentIntent.PaymentReference)
			}

			if paymentIntent.BankAuthorizationId != tt.bankAuthId {
				t.Errorf("Expected bank authorization ID to be %q, got %q", tt.bankAuthId, paymentIntent.BankAuthorizationId)
			}
		})
	}
}
