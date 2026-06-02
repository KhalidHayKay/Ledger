package paymentintent

import (
	"context"
	"encoding/json"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"testing"
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

func TestCreateReturnsReserveByDefautl(t *testing.T) {
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

	bankRepo := &bankRepoMock{
		AuthorizeFn: func(
			ctx context.Context, idempotencyKey string, input bank.AuthorizeInput,
		) (bank.Payment, error) {
			return bank.Payment{Id: bankAuthId}, nil
		},
	}

	repo := &repoMock{
		GetByPaymentRefFn: func(ctx context.Context, paymentReference string) (PaymentIntent, error) {
			return PaymentIntent{
				BankAuthorizationId: bankAuthId,
				PaymentReference:    paymentRef,
			}, nil
		},
		CreateFn: func(ctx context.Context, amount int, currency, orderId, customerId string) (string, error) {
			t.Fatalf("Create should not be called for reserved payment intent")
			return "", nil
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(repo, bankRepo, idempotencyService)

	paymentIntent, err := service.getReserved(context.Background(), "idm-key", requestHash)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if paymentIntent.BankAuthorizationId != bankAuthId {
		t.Errorf("Expected bank auth id to be %q, got %q", bankAuthId, paymentIntent.BankAuthorizationId)
	}

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
