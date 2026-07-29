package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"ledger/internal/domain/bank"
	"ledger/internal/domain/paymentintent"
	"ledger/internal/jobs/tasks"
	"ledger/internal/platform/notifier"
	"log"

	"github.com/hibiken/asynq"
)

type PaymentWorker struct {
	bankRepo bank.Repository
	uow      paymentintent.UnitOfWork
	notifier notifier.Notifier
}

func NewPaymentWorker(
	bankRepo bank.Repository,
	uow paymentintent.UnitOfWork,
	notifier notifier.Notifier,
) *PaymentWorker {
	return &PaymentWorker{bankRepo, uow, notifier}
}

func (w *PaymentWorker) HandleCreate(ctx context.Context, t *asynq.Task) error {
	var p tasks.CreatePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("%v: %w", err, asynq.SkipRetry)
	}

	log.Println("payload in job worker: ", p)

	payment, err := w.bankRepo.Authorize(ctx, p.IdempotencyKey, *p.BankInput)
	if err != nil {
		return err
	}

	err = w.uow.RunInTx(ctx, func(r paymentintent.TxRepos) error {
		event, err := r.PaymentEvent.Create(
			ctx, p.IntentId,
			paymentintent.PaymentStatusAuthorized,
			payment.AuthorizationId,
		)
		if err != nil {
			return err
		}

		return r.PaymentIntent.UpdateState(ctx, p.PaymentRef, event.State)
	})
	if err != nil {
		log.Printf("Error creating payment event for state to %v: %s", paymentintent.PaymentStatusAuthorized, err)
		return err
	}

	log.Printf("Publishing payment status for %v: %v", p.PaymentRef, paymentintent.PaymentStatusAuthorized)
	w.notifier.Publish(ctx, p.PaymentRef, paymentintent.PaymentStatusAuthorized)
	return nil
}
