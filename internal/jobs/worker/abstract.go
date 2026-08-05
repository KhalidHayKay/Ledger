package worker

import (
	"context"
	"errors"
	"fmt"
	"ledger/internal/domain/bank"
	"ledger/internal/domain/paymentintent"
	"log"

	"github.com/hibiken/asynq"
)

func (w *PaymentWorker) terminateRetry(ctx context.Context, paymentRef, intentId string) error {
	err := w.uow.RunInTx(ctx, func(r paymentintent.TxRepos) error {
		ev, err := r.PaymentEvent.Create(
			ctx, intentId,
			paymentintent.PaymentStatusFailed,
			"", // no authorization ID
			"bank unreachable after max retries",
		)
		if err != nil {
			return err
		}
		return r.PaymentIntent.UpdateState(ctx, paymentRef, ev.State)
	})
	if err != nil {
		log.Printf("Error recording exhausted-retry failure for %v: %s", paymentRef, err)
		return err
	}

	log.Printf("Retries exhausted for %v: %s", paymentRef, err)
	return fmt.Errorf("%v: %w", err, asynq.SkipRetry)
}

func (w *PaymentWorker) handleTerminalError(
	ctx context.Context,
	payment bank.Payment,
	paymentRef, intentId string,
	err error,
) error {
	if err == nil {
		return nil
	}

	if !bank.IsTerminal(err) {
		return nil
	}

	reason := err.Error()
	var apiErr *bank.APIError
	if errors.As(err, &apiErr) {
		reason = apiErr.Message
	}

	txErr := w.uow.RunInTx(ctx, func(r paymentintent.TxRepos) error {
		ev, err := r.PaymentEvent.Create(
			ctx, intentId,
			paymentintent.PaymentStatusFailed,
			payment.AuthorizationId,
			reason,
		)
		if err != nil {
			return err
		}
		return r.PaymentIntent.UpdateState(ctx, paymentRef, ev.State)
	})
	if txErr != nil {
		log.Printf("Error recording terminal failure for %v: %s", paymentRef, txErr)
		return txErr
	}

	if pubErr := w.notifier.Publish(ctx, paymentRef, reason); pubErr != nil {
		log.Printf("Error publishing failed status for %v: %s", paymentRef, pubErr)
	}

	log.Printf("Payment declined for %v: %s", paymentRef, err)
	return fmt.Errorf("%v: %w", err, asynq.SkipRetry)
}

func (w *PaymentWorker) handleTransientError(
	ctx context.Context,
	paymentRef, intentId string,
	err error,
) error {
	if err == nil {
		return nil
	}

	if bank.IsTerminal(err) {
		return nil
	}

	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)

	if retried >= maxRetry {
		if pubErr := w.notifier.Publish(ctx, paymentRef, paymentintent.PaymentStatusFailed); pubErr != nil {
			log.Printf("Error publishing exhausted-retry status for %v: %s", paymentRef, pubErr)
		}
		return w.terminateRetry(ctx, paymentRef, intentId)
	}

	log.Printf("Transient bank error for %v, will retry: %s", paymentRef, err)
	return err
}
