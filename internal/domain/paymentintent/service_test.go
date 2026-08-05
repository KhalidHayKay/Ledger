package paymentintent

import (
	"context"
	"encoding/json"
	"ledger/internal/domain/bank"
	"ledger/internal/domain/idempotency"
	"ledger/internal/domain/paymentevent"
	"ledger/internal/jobs/queue"
	"ledger/internal/platform/notifier"
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
	GetByRefFn        func(ctx context.Context, paymentRef string) (PaymentIntent, error)
	GetWithEventFn    func(ctx context.Context, paymentRef, state string) (PaymentIntent, error)

	calls []string
}

// type notifierMock struct {
// 	publishFn func(ctx context.Context, channel, result string) error
// 	waitFn    func(ctx context.Context) (string, error)
// }

// func (n *notifierMock) Subscribe(channel string) notifier.Subscription {
// 	return &testSubscription{waitFn: n.waitFn}
// }

// func (n *notifierMock) Publish(ctx context.Context, channel, result string) error {
// 	if n.publishFn != nil {
// 		return n.publishFn(ctx, channel, result)
// 	}
// 	return nil
// }

// type testSubscription struct {
// 	waitFn func(ctx context.Context) (string, error)
// }

// func (s *testSubscription) Wait(ctx context.Context) (string, error) {
// 	if s.waitFn != nil {
// 		return s.waitFn(ctx)
// 	}
// 	return "", nil
// }

// func (s *testSubscription) Close() {}

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

func (r *repoMock) GetByRef(ctx context.Context, paymentRef string) (PaymentIntent, error) {
	r.calls = append(r.calls, getByPaymentRefCall)
	return r.GetByRefFn(ctx, paymentRef)
}

func (r *repoMock) GetWithEvent(ctx context.Context, paymentRef, state string) (PaymentIntent, error) {
	r.calls = append(r.calls, getByPaymentRefCall)
	return r.GetWithEventFn(ctx, paymentRef, state)
}

// Payment event mocks
type paymentEventRepoMock struct {
	CreateFn               func(ctx context.Context, intentId, status, providerStateId, metadata string) (paymentevent.PaymentEvent, error)
	GetByIntentAndStatusFn func(ctx context.Context, paymentIntentId, paymentIntentStatus string) (paymentevent.PaymentEvent, error)

	calls []string
}

func (r *paymentEventRepoMock) Create(ctx context.Context, intentId, status, providerStateId, metadata string) (paymentevent.PaymentEvent, error) {
	r.calls = append(r.calls, createCall)
	return r.CreateFn(ctx, intentId, status, providerStateId, metadata)
}

func (r *paymentEventRepoMock) GetByIntentAndStatus(ctx context.Context, paymentIntentId, paymentIntentStatus string) (paymentevent.PaymentEvent, error) {
	r.calls = append(r.calls, getByPaymentRefCall)
	return r.GetByIntentAndStatusFn(ctx, paymentIntentId, paymentIntentStatus)
}

// Bank Repo Mocks
type bankRepoMock struct {
	AuthorizeFn func(ctx context.Context, idempotencyKey string, input bank.AuthorizeInput) (bank.Payment, error)
	CaptureFn   func(ctx context.Context, idempotencyKey, authorizationId string, amount int) (bank.Payment, error)
	VoidFn      func(ctx context.Context, idempotencyKey, authorizationId string) (bank.Payment, error)
	RefundFn    func(ctx context.Context, idempotencyKey, captureId string, amount int) (bank.Payment, error)

	calls []string
}

func (r *bankRepoMock) Authorize(ctx context.Context, idempotencyKey string, input bank.AuthorizeInput) (bank.Payment, error) {
	r.calls = append(r.calls, authorizeCall)
	return r.AuthorizeFn(ctx, idempotencyKey, input)
}

func (r *bankRepoMock) Capture(ctx context.Context, idempotencyKey, authorizationId string, amount int) (bank.Payment, error) {
	r.calls = append(r.calls, captureCall)
	return r.CaptureFn(ctx, idempotencyKey, authorizationId, amount)
}

func (r *bankRepoMock) Void(ctx context.Context, idempotencyKey, authorizationId string) (bank.Payment, error) {
	r.calls = append(r.calls, voidCall)
	return r.VoidFn(ctx, idempotencyKey, authorizationId)
}

func (r *bankRepoMock) Refund(ctx context.Context, idempotencyKey, captureId string, amount int) (bank.Payment, error) {
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
	RunInTxFn func(ctx context.Context, fn func(TxRepos) error) error
}

func (u *uowMock) RunInTx(ctx context.Context, fn func(TxRepos) error) error {
	return u.RunInTxFn(ctx, fn)
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
		GetByRefFn: func(ctx context.Context, paymentRef string) (PaymentIntent, error) {
			return PaymentIntent{
				PaymentRef: paymentRef,
			}, nil
		},
	}

	uow := &uowMock{
		RunInTxFn: func(ctx context.Context, fn func(TxRepos) error) error {
			return fn(TxRepos{
				PaymentIntent: repo,
			})
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)
	service := NewService(
		repo,
		&paymentEventRepoMock{},
		&bankRepoMock{},
		uow,
		idempotencyService,
		&queue.Client{},
		&notifier.RedisNotifier{},
	)

	paymentIntent, err := service.getReservedIntent(context.Background(), "idm-key", requestHash)
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
		GetByRefFn: func(ctx context.Context, paymentRef string) (PaymentIntent, error) {
			t.Errorf("Expected an error boundary to block GetByPaymentRef from getting called for idempotency key reuse")
			return PaymentIntent{}, nil
		},
	}

	uow := &uowMock{
		RunInTxFn: func(ctx context.Context, fn func(TxRepos) error) error {
			return fn(TxRepos{
				PaymentIntent: repo,
			})
		},
	}

	idempotencyService := idempotency.NewService(idempotencyRepo)

	service := NewService(
		repo,
		&paymentEventRepoMock{},
		&bankRepoMock{},
		uow,
		idempotencyService,
		&queue.Client{},
		&notifier.RedisNotifier{},
	)

	_, err := service.getReservedIntent(context.Background(), "idm-key", requestHash)
	if err == nil {
		t.Fatalf("Expected error due to idempotency key reuse with different request hash, got nil")
	}
}
