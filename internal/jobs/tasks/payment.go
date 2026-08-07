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
	IdempotencyKey string
	IntentId       string
	PaymentRef     string
	StateId        string
	Amount         int
}

type RefundPayload struct {
	IdempotencyKey string
	IntentId       string
	PaymentRef     string
	StateId        string
	Amount         int
}

type CancelPayload struct {
	IdempotencyKey string
	IntentId       string
	PaymentRef     string
	StateId        string
}

func NewCreatePaymentTask(p CreatePayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCreatePayment, payload), nil
}

func NewCapturePaymentTask(p CapturePayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCapturePayment, payload), nil
}

func NewRefundPaymentTask(p RefundPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeRefundPayment, payload), nil
}

func NewCancelPaymentTask(p CancelPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCancelPayment, payload), nil
}
