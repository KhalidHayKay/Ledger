package queue

import (
	"context"
	"ledger/internal/jobs/tasks"

	"github.com/hibiken/asynq"
)

func (c *Client) EnqueueCreate(ctx context.Context, p tasks.CreatePayload) error {
	task, err := tasks.NewCreatePaymentTask(p)
	if err != nil {
		return err
	}
	return c.Enqueue(ctx, task, asynq.Queue("critical"))
}
