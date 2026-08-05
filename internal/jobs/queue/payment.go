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

	err = c.Enqueue(ctx, task, asynq.Queue("critical"), asynq.MaxRetry(5))
	if err != nil {

		return err
	}

	log.Printf("Enqueued create payment task for intent %s", p.IntentId)
	return nil
}
