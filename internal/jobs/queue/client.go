package queue

import (
	"context"

	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
}

func NewClient(redisAddr string) (*Client, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &Client{client: client}, nil
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) error {
	_, err := c.client.EnqueueContext(ctx, task, opts...)
	return err
}
