package worker

import (
	"context"
	"encoding/json"
	"errors"
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
	if err != nil {
		if !bank.IsTerminal(err) {
			log.Printf("Transient bank error for %v, will retry: %s", p.PaymentRef, err)
			return err
		}

		reason := err.Error()
		var apiErr *bank.APIError
		if errors.As(err, &apiErr) {
			reason = apiErr.Message
		}

		txErr := w.uow.RunInTx(ctx, func(r paymentintent.TxRepos) error {
			ev, err := r.PaymentEvent.Create(
				ctx, p.IntentId,
				paymentintent.PaymentStatusFailed,
				payment.AuthorizationId,
				reason,
			)
			if err != nil {
				return err
			}
			return r.PaymentIntent.UpdateState(ctx, p.PaymentRef, ev.State)
		})
		if txErr != nil {
			log.Printf("Error recording terminal failure for %v: %s", p.PaymentRef, txErr)
			return txErr
		}

		if pubErr := w.notifier.Publish(ctx, p.PaymentRef, reason); pubErr != nil {
			log.Printf("Error publishing failed status for %v: %s", p.PaymentRef, pubErr)
		}

		log.Printf("Payment declined for %v: %s", p.PaymentRef, err)
		return fmt.Errorf("%v: %w", err, asynq.SkipRetry)
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
		return r.PaymentIntent.UpdateState(ctx, p.PaymentRef, event.State)
	})
	if err != nil {
		log.Printf("Error creating payment event for state to %v: %s", paymentintent.PaymentStatusAuthorized, err)
		return err
	}

	if pubErr := w.notifier.Publish(ctx, p.PaymentRef, paymentintent.PaymentStatusAuthorized); pubErr != nil {
		log.Printf("Error publishing authorized status for %v: %s", p.PaymentRef, pubErr)
	}

	return nil
}
