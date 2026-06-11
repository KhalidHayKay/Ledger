package paymentintent

import (
	"context"
	"encoding/json"
	"fmt"
	"ledger/app/config"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"ledger/internal/paymentprocess"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestCreateReplaysExistingPaymentIntent(t *testing.T) {
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
		GetByPaymentRefFn: func(ctx context.Context, paymentRef string) (PaymentIntent, error) {
			return PaymentIntent{}, nil
		},
		CreateFn: func(ctx context.Context, amount int, currency, orderId, customerId string) (PaymentIntent, error) {
			t.Fatalf("Create should not be called for reserved payment intent")
			return PaymentIntent{}, nil
		},
	}

	uow := &uowMock{
		RunInTxFn: func(ctx context.Context, fn func(Repos) error) error {
			t.Fatal("RunInTx should not be called for reserved payment intent")
			return nil
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(repo, &bankRepoMock{}, uow, idempotencyService)

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
	config.Env = &config.EnvType{App: config.AppConfig{Key: "test-key"}}

	tests := []struct {
		id             int64
		idempotencyKey string
		requestHash    string
		paymentRef     string
		bankAuthId     string
	}{
		{1, "idmkey-1", "req-hash-1", "payment-ref-1", "bank-authorization-id-1"},
		{2, "idmkey-2", "req-hash-2", "payment-ref-2", "bank-authorization-id-2"},
		{3, "idmkey-3", "req-hash-3", "payment-ref-3", "bank-authorization-id-3"},
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
					return bank.Payment{AuthorizationId: tt.bankAuthId}, nil
				},
			}

			repo := &repoMock{
				CreateFn: func(ctx context.Context, amount int, currency, orderId, customerId string) (PaymentIntent, error) {
					return PaymentIntent{Id: fmt.Sprintf("%v", tt.id), PaymentRef: tt.paymentRef}, nil
				},
				CreateReferenceFn: func(ctx context.Context, id, generatedRef string) error {
					return nil
				},
				UpdateStateFn: func(ctx context.Context, paymentRef, state string) error {
					return nil
				},
			}

			ppRepo := &ppRepoMock{
				CreateFn: func(ctx context.Context, intentId, status, bankAuthId string) (paymentprocess.PaymentProcess, error) {
					return paymentprocess.PaymentProcess{}, nil
				},
			}

			idempotencyService := idempotency.NewService(idempotencyRepo)

			uow := &uowMock{
				RunInTxFn: func(ctx context.Context, fn func(Repos) error) error {
					return fn(Repos{
						PaymentIntent:  repo,
						PaymentProcess: ppRepo,
					})
				},
			}

			service := NewService(repo, bankRepo, uow, idempotencyService)

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

			if paymentIntent.Id != fmt.Sprintf("%v", tt.id) {
				t.Errorf("Expected payment ID to be %q, got %q", fmt.Sprintf("%v", tt.id), paymentIntent.Id)
			}

			// if paymentIntent.BankAuthorizationId != tt.bankAuthId {
			// 	t.Errorf("Expected bank authorization ID to be %q, got %q", tt.bankAuthId, paymentIntent.BankAuthorizationId)
			// }
		})
	}
}
