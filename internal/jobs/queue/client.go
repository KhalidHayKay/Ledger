package queue

import (
	"context"
	"ledger/internal/jobs/tasks"

	"github.com/hibiken/asynq"
)

type QueueInterface interface {
	EnqueueCreate(ctx context.Context, p tasks.CreatePayload) error
	EnqueueCapture(ctx context.Context, p tasks.CapturePayload) error
}

type Client struct {
	client *asynq.Client
}

func NewClient(client *asynq.Client) *Client {
	return &Client{client}
}

func (c *Client) Close() {
	err := c.client.Close()
	if err != nil {
		panic(err)
	}
}

func (c *Client) enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) error {
	_, err := c.client.EnqueueContext(ctx, task, opts...)
	return err
}
