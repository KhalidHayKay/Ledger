package paymentintent

import (
	"context"
	"encoding/json"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"ledger/internal/paymentprocess"
	"testing"
)

const (
	createCall          = "Create"
	updateStateCall     = "UpdateState"
	getByPaymentRefCall = "GetByPaymentRef"

	authorizeCall = "Authorize"

	captureCall = "Capture"
	voidCall    = "Void"
	refundCall  = "Refund"

	saveKeyCall  = "SaveKey"
	getByKeyCall = "GetByKey"
)

type repoMock struct {
	CreateFn          func(ctx context.Context, amount int, currency, orderId, customerId string) (PaymentIntent, error)
	CreateReferenceFn func(ctx context.Context, id, generatedRef string) error
	UpdateStateFn     func(ctx context.Context, paymentRef, state string) error
	GetByPaymentRefFn func(ctx context.Context, paymentRef string) (PaymentIntent, error)

	calls []string
}

func (r *repoMock) Create(ctx context.Context, amount int, currency, orderId, customerId string) (PaymentIntent, error) {
	r.calls = append(r.calls, createCall)
	return r.CreateFn(ctx, amount, currency, orderId, customerId)
}

func (r *repoMock) CreateReference(ctx context.Context, id, generatedRef string) error {
	r.calls = append(r.calls, createCall)
	return r.CreateReferenceFn(ctx, id, generatedRef)
}

func (r *repoMock) UpdateState(ctx context.Context, paymentRef, state string) error {
	r.calls = append(r.calls, updateStateCall)
	return r.UpdateStateFn(ctx, paymentRef, state)
}

func (r *repoMock) GetByPaymentRef(ctx context.Context, paymentRef string) (PaymentIntent, error) {
	r.calls = append(r.calls, getByPaymentRefCall)
	return r.GetByPaymentRefFn(ctx, paymentRef)
}

// Payment process mocks
type ppRepoMock struct {
	CreateFn               func(ctx context.Context, intentId, status, bankAuthId string) (paymentprocess.PaymentProcess, error)
	GetByIntentAndStatusFn func(ctx context.Context, paymentIntentId, paymentIntentStatus string) (paymentprocess.PaymentProcess, error)

	calls []string
}

func (m *ppRepoMock) Create(ctx context.Context, intentId, status, bankAuthId string) (paymentprocess.PaymentProcess, error) {
	m.calls = append(m.calls, createCall)
	return m.CreateFn(ctx, intentId, status, bankAuthId)
}

func (m *ppRepoMock) GetByIntentAndStatus(ctx context.Context, paymentIntentId, paymentIntentStatus string) (paymentprocess.PaymentProcess, error) {
	m.calls = append(m.calls, getByPaymentRefCall)
	return m.GetByIntentAndStatusFn(ctx, paymentIntentId, paymentIntentStatus)
}

// Bank Repo Mocks
type bankRepoMock struct {
	AuthorizeFn func(ctx context.Context, idempotencyKey string, input bank.AuthorizeInput) (bank.Payment, error)
	CaptureFn   func(ctx context.Context, idempotencyKey, authorizationId, amount string) (bank.Payment, error)
	VoidFn      func(ctx context.Context, idempotencyKey, authorizationId string) (bank.Payment, error)
	RefundFn    func(ctx context.Context, idempotencyKey, captureId, amount string) (bank.Payment, error)

	calls []string
}

func (r *bankRepoMock) Authorize(ctx context.Context, idempotencyKey string, input bank.AuthorizeInput) (bank.Payment, error) {
	r.calls = append(r.calls, authorizeCall)
	return r.AuthorizeFn(ctx, idempotencyKey, input)
}

func (r *bankRepoMock) Capture(ctx context.Context, idempotencyKey, authorizationId, amount string) (bank.Payment, error) {
	r.calls = append(r.calls, captureCall)
	return r.CaptureFn(ctx, idempotencyKey, authorizationId, amount)
}

func (r *bankRepoMock) Void(ctx context.Context, idempotencyKey, authorizationId string) (bank.Payment, error) {
	r.calls = append(r.calls, voidCall)
	return r.VoidFn(ctx, idempotencyKey, authorizationId)
}

func (r *bankRepoMock) Refund(ctx context.Context, idempotencyKey, captureId, amount string) (bank.Payment, error) {
	r.calls = append(r.calls, refundCall)
	return r.RefundFn(ctx, idempotencyKey, captureId, amount)
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

// UoW Mock
type uowMock struct {
	RunInTxFn func(ctx context.Context, fn func(Repos) error) error
}

func (m *uowMock) RunInTx(ctx context.Context, fn func(Repos) error) error {
	return m.RunInTxFn(ctx, fn)
}

func TestGetReservedHappyPath(t *testing.T) {
	requestHash := "req-hash"
	paymentRef := "payment-ref"
	mockEntry := idempotency.Entry{
		RequestHash: requestHash,
		PaymentRef:  paymentRef,
	}

	idempotencyRepo := &idempotencyRepoMock{
		GetByKeyFn: func(ctx context.Context, idempotencyKey string) (string, error) {
			entryEncode, _ := json.Marshal(mockEntry)
			return string(entryEncode), nil
		},
	}

	repo := &repoMock{
		GetByPaymentRefFn: func(ctx context.Context, paymentRef string) (PaymentIntent, error) {
			return PaymentIntent{
				PaymentRef: paymentRef,
			}, nil
		},
	}

	uow := &uowMock{
		RunInTxFn: func(ctx context.Context, fn func(Repos) error) error {
			return fn(Repos{
				PaymentIntent: repo,
			})
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(repo, &ppRepoMock{}, &bankRepoMock{}, uow, idempotencyService)

	paymentIntent, err := service.getReserved(context.Background(), "idm-key", requestHash)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if paymentIntent.PaymentRef != paymentRef {
		t.Fatalf("Expected payment reference to be %q, got %q", paymentRef, paymentIntent.PaymentRef)
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
		GetByPaymentRefFn: func(ctx context.Context, paymentRef string) (PaymentIntent, error) {
			t.Errorf("Expected an error boundary to block GetByPaymentRef from getting called for idempotency key reuse")
			return PaymentIntent{}, nil
		},
	}

	uow := &uowMock{
		RunInTxFn: func(ctx context.Context, fn func(Repos) error) error {
			return fn(Repos{
				PaymentIntent: repo,
			})
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(repo, &ppRepoMock{}, &bankRepoMock{}, uow, idempotencyService)

	_, err := service.getReserved(context.Background(), "idm-key", requestHash)
	if err == nil {
		t.Fatalf("Expected error due to idempotency key reuse with different request hash, got nil")
	}
}
