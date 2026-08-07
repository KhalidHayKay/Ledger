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

	payment, err := w.bankRepo.Authorize(ctx, p.IdempotencyKey, *p.BankInput)

	if err := w.handleTransientError(ctx, p.PaymentRef, p.IntentId, err); err != nil {
		return err
	}

	if err := w.handleTerminalError(ctx, payment, p.PaymentRef, p.IntentId, err); err != nil {
		return err
	}

	err = w.uow.RunInTx(ctx, func(r paymentintent.TxRepos) error {
		event, err := r.PaymentEvent.Create(
			ctx, p.IntentId,
			paymentintent.PaymentStatusAuthorized,
			payment.AuthorizationId,
			"", // metadata
		)
		if err != nil {
			return err
		}

		err = r.PaymentIntent.UpdateOperation(ctx, p.PaymentRef, "")
		if err != nil {
			return err
		}

		return r.PaymentIntent.UpdateState(ctx, p.PaymentRef, event.State)
	})
	if err != nil {
		log.Printf("Error creating payment event for state to %v: %s", paymentintent.PaymentStatusAuthorized, err)
		return err
	}

	if pubErr := w.notifier.Publish(ctx, p.PaymentRef, paymentintent.PaymentStatusAuthorized); pubErr != nil {
		log.Printf("Error publishing authorized status for %v: %s", p.PaymentRef, pubErr)
	}

	log.Printf("Payment authorized for %v: %s", p.PaymentRef, payment.AuthorizationId)
	return nil
}

func (w *PaymentWorker) HandleCapture(ctx context.Context, t *asynq.Task) error {
	var p tasks.CapturePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("%v: %w", err, asynq.SkipRetry)
	}

	payment, err := w.bankRepo.Capture(ctx, p.IdempotencyKey, p.StateId, p.Amount)

	if err := w.handleTransientError(ctx, p.PaymentRef, p.IntentId, err); err != nil {
		return err
	}

	if err := w.handleTerminalError(ctx, payment, p.PaymentRef, p.IntentId, err); err != nil {
		return err
	}

	err = w.uow.RunInTx(ctx, func(r paymentintent.TxRepos) error {
		event, err := r.PaymentEvent.Create(
			ctx, p.IntentId,
			paymentintent.PaymentStatusCaptured,
			payment.CaptureId, "",
		)
		if err != nil {
			return err
		}

		err = r.PaymentIntent.UpdateOperation(ctx, p.PaymentRef, "")
		if err != nil {
			return err
		}

		return r.PaymentIntent.UpdateState(ctx, p.PaymentRef, event.State)
	})
	if err != nil {
		log.Printf("Error creating payment event for state to %v: %s", paymentintent.PaymentStatusCaptured, err)
		return err
	}

	if pubErr := w.notifier.Publish(ctx, p.PaymentRef, paymentintent.PaymentStatusCaptured); pubErr != nil {
		log.Printf("Error publishing captured status for %v: %s", p.PaymentRef, pubErr)
	}

	log.Printf("Payment captured for %v: %s", p.PaymentRef, payment.CaptureId)
	return nil
}
