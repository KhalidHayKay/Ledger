package tasks

import (
	"encoding/json"
	"ledger/internal/domain/bank"

	"github.com/hibiken/asynq"
)

const (
	TypeCreatePayment  = "payment:create"
	TypeCapturePayment = "payment:capture"
	TypeRefundPayment  = "payment:refund"
	TypeCancelPayment  = "payment:cancel"
)

type CreatePayload struct {
	IdempotencyKey string
	IntentId       string
	PaymentRef     string
	BankInput      *bank.AuthorizeInput
}

type CapturePayload struct {
	PaymentRef      string
	IdempotencyKey  string
	AuthorizationId string
	Amount          int
}

func NewCreatePaymentTask(p CreatePayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCreatePayment, payload), nil
}

// I think handler shoild not be here
// type BankHandler struct {
// 	bankRepo bank.Repository
// }

// func NewBankHandler(bankRepo bank.Repository) *BankHandler {
// 	return &BankHandler{bankRepo}
// }

// func (h *BankHandler) BankRequest(ctx context.Context, t *asynq.Task) error {
// 	var p BankRequestPayload
// 	if err := json.Unmarshal(t.Payload(), &p); err != nil {
// 		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
// 	}

// 	/////

// 	return nil
// }
