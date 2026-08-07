package queue

import (
	"context"
	"ledger/internal/jobs/tasks"
	"log"

	"github.com/hibiken/asynq"
)

func (c *Client) EnqueueCreate(ctx context.Context, p tasks.CreatePayload) error {
	task, err := tasks.NewCreatePaymentTask(p)
	if err != nil {
		return err
	}

	err = c.enqueue(ctx, task, asynq.Queue("critical"), asynq.MaxRetry(5))
	if err != nil {

		return err
	}

	log.Printf("Enqueued create payment task for intent %s", p.IntentId)
	return nil
}

func (c *Client) EnqueueCapture(ctx context.Context, p tasks.CapturePayload) error {
	task, err := tasks.NewCapturePaymentTask(p)
	if err != nil {
		return err
	}

	err = c.enqueue(ctx, task, asynq.Queue("critical"), asynq.MaxRetry(5))
	if err != nil {

		return err
	}

	log.Printf("Enqueued capture payment task for intent %s", p.IntentId)
	return nil
}

func (c *Client) EnqueueRefund(ctx context.Context, p tasks.RefundPayload) error {
	task, err := tasks.NewRefundPaymentTask(p)
	if err != nil {
		return err
	}

	err = c.enqueue(ctx, task, asynq.Queue("critical"), asynq.MaxRetry(5))
	if err != nil {
		return err
	}

	log.Printf("Enqueued refund payment task for intent %s", p.IntentId)
	return nil
}

func (c *Client) EnqueueCancel(ctx context.Context, p tasks.CancelPayload) error {
	task, err := tasks.NewCancelPaymentTask(p)
	if err != nil {
		return err
	}

	err = c.enqueue(ctx, task, asynq.Queue("critical"), asynq.MaxRetry(5))
	if err != nil {
		return err
	}

	log.Printf("Enqueued cancel payment task for intent %s", p.IntentId)
	return nil
}
